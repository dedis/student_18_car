package byzcoin

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"github.com/dedis/cothority"
	"github.com/dedis/cothority/byzcoin"
	"github.com/dedis/cothority/calypso"
	"github.com/dedis/cothority/darc"
	"github.com/dedis/kyber/util/random"
	"github.com/dedis/protobuf"
	"io"
	"time"
)


func NewCar(VIN string) (Car) {
	var c Car
	c.Vin = VIN
	c.Reports = []Report{}
	return c
}

//creating transaction with spawn:car instruction
func (s *ser) createCarInstance(car Car,
	controlDarc *darc.Darc, signer darc.Signer) (byzcoin.InstanceID, error) {

	var instID byzcoin.InstanceID
	carBuf, err := protobuf.Encode(&car)
	if err != nil {
		return instID, err
	}
	ctx := byzcoin.ClientTransaction{
		Instructions: []byzcoin.Instruction{{
			InstanceID: byzcoin.NewInstanceID(controlDarc.GetBaseID()),
			Nonce:      byzcoin.Nonce{},
			Index:      0,
			Length:     1,
			Spawn: &byzcoin.Spawn{
				ContractID: ContractCarID,
				Args:       byzcoin.Arguments{{Name: "car", Value: carBuf}},
			},
		}},
	}

	// Sending this transaction to ByzCoin
	_, err = s.signAndSendTransaction(ctx, signer, controlDarc, ctx.Instructions[0].DeriveID("").Slice())
	if err != nil {
		return instID, err
	}

	return ctx.Instructions[0].DeriveID(""), err
}


func (s *ser) addReport(instID byzcoin.InstanceID,
	controlDarc *darc.Darc, wData SecretData, signerG darc.Signer, signerO darc.Signer) error{

	//creating a Calypso Write Instance
	//key := []byte("secret key")
	key := random.Bits(128, true, random.New())
	_, wInstance, err := s.addWrite(key, wData, controlDarc, signerG, signerO)
	if err != nil {
		return err
	}

	//creating new Report to be added in the list of the reports in the instance
	var newReport Report
	newReport.Date = time.Now().String()
	newReport.WriteInstanceID = wInstance.Slice()
	newReport.GarageId = signerG.Identity().String()

	reportBuf, err := protobuf.Encode(&newReport)
	if err != nil {
		return err
	}

	ctx := byzcoin.ClientTransaction{
		Instructions: []byzcoin.Instruction{{
			InstanceID: instID,
			Nonce:      byzcoin.Nonce{},
			Index:      0,
			Length:     1,
			Invoke: &byzcoin.Invoke{
				Command: "addReport",
				Args:    byzcoin.Arguments{{Name: "report", Value: reportBuf}},
			},
		}},
	}
	// And we need to sign the instruction with the signer that has his
	// public key stored in the darc.
	err = ctx.Instructions[0].SignBy(controlDarc.GetBaseID(), signerG, signerO)
	if err != nil {
		return err
	}

	_, err = s.cl.AddTransactionAndWait(ctx,5)
	if err != nil {
		return err
	}

	_, err = s.cl.GetProof(ctx.Instructions[0].InstanceID.Slice())
	if err != nil {
		return err
	}

	return err
}

//instID is the ID of the car instance that contains reports with calypso secrets that should be read
func (s *ser) readReports(instID byzcoin.InstanceID,
	controlDarc *darc.Darc, signerR darc.Signer, signerO darc.Signer) ([]SecretData, error) {

	var secretsList []SecretData
	//getting the car data from the car instance and storing it in the carData variable
	resp, err := s.cl.GetProof(instID.Slice())
	if err != nil {
		return secretsList, err
	}
	_,values, err := resp.Proof.KeyValue()
	if err != nil {
		return secretsList, err
	}
	var carData Car
	err = protobuf.Decode(values[0], &carData)
	if err != nil {
		return secretsList, err
	}

	//if there is any report, create a calypso read instance in order to read the encrypted secret
	if len(carData.Reports)>0 {

		for i := 0; i < len(carData.Reports); i++ {

			//get the proof for the write instance
			resp, err = s.cl.GetProof(carData.Reports[0].WriteInstanceID)
			if err != nil {
				return secretsList, err
			}

			prWr:= resp.Proof
			//add read instance
			prRe, err := s.addRead(&prWr, controlDarc, signerR, signerO)

			dk, err := s.servicesCal[0].DecryptKey(&calypso.DecryptKey{Read: *prRe, Write: prWr})
			if err != nil {
				return secretsList, err
			}
			if dk.X.Equal(s.ltsReply.X) != true {
				return secretsList, errors.New("the points are not derived from the same group")
			}
			key, err := calypso.DecodeKey(cothority.Suite, s.ltsReply.X, dk.Cs, dk.XhatEnc, s.signer.Ed25519.Secret)
			if err != nil {
				return secretsList, err
			}
			//now that we have the symetric key, we can decrypt the secret

			//getting the write structure from the proof
			var write calypso.Write
			err = prWr.ContractValue(cothority.Suite, calypso.ContractWriteID, &write)
			if err != nil {
				return secretsList, err
			}


			//decrypting the secret and placing it in a SecretData structure
			plainText, err := decrypt(write.Data, key)
			if err != nil {
				return secretsList, err
			}
			var secret SecretData
			err = protobuf.Decode(plainText, &secret)
			if err != nil {
				return secretsList, err
			}
			secretsList = append(secretsList, secret)
		}

		return secretsList, err
	}
	return secretsList, err
}

func (s *ser) addRead( write *byzcoin.Proof,
	controlDarc *darc.Darc, signerR darc.Signer, signerO darc.Signer) (*byzcoin.Proof, error) {
	var readBuf []byte
	read := &calypso.Read{
		Write: byzcoin.NewInstanceID(write.InclusionProof.Key),
		Xc:    s.signer.Ed25519.Point,
	}
	var err error
	readBuf, err = protobuf.Encode(read)
	if err != nil {
		return nil, err
	}
	ctx := byzcoin.ClientTransaction{
		Instructions: byzcoin.Instructions{{
			InstanceID: byzcoin.NewInstanceID(write.InclusionProof.Key),
			Nonce:      byzcoin.Nonce{},
			Index:      0,
			Length:     1,
			Spawn: &byzcoin.Spawn{
				ContractID: calypso.ContractReadID,
				Args:       byzcoin.Arguments{{Name: "read", Value: readBuf}},
			},
		}},
	}
	err = ctx.Instructions[0].SignBy(controlDarc.GetID(), signerR, signerO)
	if err != nil {
		return nil, err
	}

	_, err = s.cl.AddTransactionAndWait(ctx, 5)
	if err != nil {
		return nil, err
	}

	resp, err := s.cl.GetProof(ctx.Instructions[0].DeriveID("").Slice())
	if err != nil {
		return nil, err
	}

	return &resp.Proof, err
}

func (s *ser) addWrite(key []byte, wData SecretData,
	controlDarc *darc.Darc, signerG darc.Signer, signerO darc.Signer) (*byzcoin.Proof, byzcoin.InstanceID, error) {

	var instID byzcoin.InstanceID
	write := calypso.NewWrite(cothority.Suite, s.ltsReply.LTSID, controlDarc.GetBaseID(), s.ltsReply.X, key)
	var err error

	writeDataBuf, err := protobuf.Encode(&wData)
	if err != nil {
		return nil, instID, err
	}
	write.Data, err = encrypt(writeDataBuf, key)
	if err != nil {
		return nil, instID, err
	}

	writeBuf, err := protobuf.Encode(write)
	if err != nil {
		return nil, instID, err
	}

	ctx := byzcoin.ClientTransaction{
		Instructions: byzcoin.Instructions{{
			InstanceID: byzcoin.NewInstanceID(controlDarc.GetBaseID()),
			Nonce:      byzcoin.Nonce{},
			Index:      0,
			Length:     1,
			Spawn: &byzcoin.Spawn{
				ContractID: calypso.ContractWriteID,
				Args:       byzcoin.Arguments{{Name: "write", Value: writeBuf}},
			},
		}},
	}

	err = ctx.Instructions[0].SignBy(controlDarc.GetID(), signerG, signerO)
	if err != nil {
		return nil, instID, err
	}

	_, err = s.cl.AddTransactionAndWait(ctx, 5)
	if err != nil {
		return nil, instID, err
	}
	instID = ctx.Instructions[0].DeriveID("")

	resp, err := s.cl.GetProof(instID.Slice())
	if err != nil {
		return nil, instID, err
	}

	return &resp.Proof, instID, err
}

//Symmetric encryption AES
func encrypt(plaintext []byte, key []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

//Symmetric decryption AES
func decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}