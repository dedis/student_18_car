package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"github.com/dedis/cothority"
	"github.com/dedis/cothority/byzcoin"
	"github.com/dedis/cothority/calypso"
	"github.com/dedis/cothority/darc"
	"github.com/dedis/cothority/darc/expression"
	"github.com/dedis/protobuf"
	"github.com/dedis/student_18_car/car"
	"io"
	"strconv"
	"time"
)

//create the new darc and
func spawnDarcTxn(controlDarc darc.Darc, newDracSigner darc.Signer) (byzcoin.ClientTransaction, darc.Darc, error) {
	var err error
	idAdmin := []darc.Identity{newDracSigner.Identity()}
	darcAdmin := darc.NewDarc(darc.InitRules(idAdmin, idAdmin),
		[]byte("Admin darc"))
	darcAdmin.Rules.AddRule("spawn:darc",
		expression.InitOrExpr(controlDarc.GetIdentityString(), newDracSigner.Identity().String()))
	darcAdmin.Rules.AddRule("invoke:evolve",
		expression.InitOrExpr(controlDarc.GetIdentityString(), newDracSigner.Identity().String()))
	darcAdminBuf, err := darcAdmin.ToProto()

	//creating a transaction with spawn:darc instruction
	ctx := newSpawnDarcTransaction(&controlDarc, darcAdminBuf)

	return ctx, *darcAdmin, err
}

//create new client transaction with instruction to spawn a darc
func newSpawnDarcTransaction(controlDarc *darc.Darc, newDarcBuf []byte) byzcoin.ClientTransaction {

	ctx := byzcoin.ClientTransaction{
		Instructions: []byzcoin.Instruction{{
			InstanceID: byzcoin.NewInstanceID(controlDarc.GetBaseID()),
			Nonce:      byzcoin.GenNonce(),
			Index:      0,
			Length:     1,
			Spawn: &byzcoin.Spawn{
				ContractID: byzcoin.ContractDarcID,
				Args: []byzcoin.Argument{{
					Name:  "darc",
					Value: newDarcBuf,
				}},
			},
		}},
	}
	return ctx
}

func spawnCarDarc(controlDarc *darc.Darc,
	darcOwner *darc.Darc, id int) (byzcoin.Instruction, *darc.Darc, error) {

	//rules for the new Car Darc
	rs := darc.NewRules()
	if err := rs.AddRule("spawn:car", expression.InitAndExpr(controlDarc.GetIdentityString())); err != nil {
		panic("add rule should never fail on an empty rule list: " + err.Error())
	}
	if err := rs.AddRule("spawn:calypsoRead", expression.InitAndExpr(darcOwner.GetIdentityString())); err != nil {
		panic("add rule should never fail on an empty rule list: " + err.Error())
	}
	if err := rs.AddRule("invoke:addReport", expression.InitAndExpr(darcOwner.GetIdentityString())); err != nil {
		panic("add rule should never fail on an empty rule list: " + err.Error())
	}
	if err := rs.AddRule("spawn:calypsoWrite", expression.InitAndExpr(darcOwner.GetIdentityString())); err != nil {
		panic("add rule should never fail on an empty rule list: " + err.Error())
	}

	//todo will it cause problems to create many car darcs with same rules and description? this is why i added ID
	darcCar := darc.NewDarc(rs,
		[]byte("Car darc"+strconv.Itoa(id)))
	darcCarBuf, err := darcCar.ToProto()
	inst := byzcoin.Instruction{
		InstanceID: byzcoin.NewInstanceID(controlDarc.GetBaseID()),
		Nonce:      byzcoin.GenNonce(),
		Index:      0,
		Length:     1,
		Spawn: &byzcoin.Spawn{
			ContractID: byzcoin.ContractDarcID,
			Args: []byzcoin.Argument{{
				Name:  "darc",
				Value: darcCarBuf,
			}},
		},
	}
	return inst, darcCar, err
}

//creating transaction with spawn:car instruction
func createCarInstanceInstr(car car.Car,
	controlDarc *darc.Darc) (byzcoin.Instruction, error) {

	var ContractCarID = "car"
	carBuf, err := protobuf.Encode(&car)

	instr := byzcoin.Instruction{
		InstanceID: byzcoin.NewInstanceID(controlDarc.GetBaseID()),
		Nonce:      byzcoin.Nonce{},
		Index:      0,
		Length:     1,
		Spawn: &byzcoin.Spawn{
			ContractID: ContractCarID,
			Args:       byzcoin.Arguments{{Name: "car", Value: carBuf}},
		},
	}
	return instr, err
}


func addWrite(key []byte, wData car.SecretData,
	controlDarc *darc.Darc, lts calypso.CreateLTSReply) (byzcoin.Instruction, error) {

	var instr byzcoin.Instruction
	write := calypso.NewWrite(cothority.Suite, lts.LTSID, controlDarc.GetBaseID(), lts.X, key)
	var err error

	writeDataBuf, err := protobuf.Encode(&wData)
	if err != nil {
		return instr, err
	}
	write.Data, err = encrypt(writeDataBuf, key)
	if err != nil {
		return instr, err
	}

	writeBuf, err := protobuf.Encode(write)
	if err != nil {
		return instr, err
	}
	instr = byzcoin.Instruction{
		InstanceID: byzcoin.NewInstanceID(controlDarc.GetBaseID()),
		Nonce:      byzcoin.Nonce{},
		Index:      0,
		Length:     1,
		Spawn: &byzcoin.Spawn{
			ContractID: calypso.ContractWriteID,
			Args:       byzcoin.Arguments{{Name: "write", Value: writeBuf}},
		},
	}
	return instr, err
}





func addReport(carInstID byzcoin.InstanceID, writeInstanceID byzcoin.InstanceID,
	signer darc.Signer) (byzcoin.Instruction, error){
	var instr byzcoin.Instruction
	//creating new Report to be added in the list of the reports in the instance
	var newReport car.Report
	newReport.Date = time.Now().String()
	newReport.WriteInstanceID = writeInstanceID.Slice()
	newReport.GarageId = signer.Identity().String()

	reportBuf, err := protobuf.Encode(&newReport)
	if err != nil {
		return instr, err
	}
	instr = byzcoin.Instruction{
		InstanceID: carInstID,
		Nonce:      byzcoin.Nonce{},
		Index:      0,
		Length:     1,
		Invoke: &byzcoin.Invoke{
			Command: "addReport",
			Args:    byzcoin.Arguments{{Name: "report", Value: reportBuf}},
		},
	}
	return instr, err
}

func addRead( write *byzcoin.Proof,
	signer darc.Signer) (byzcoin.Instruction, error) {
	var readBuf []byte
	var instr byzcoin.Instruction
	read := &calypso.Read{
		Write: byzcoin.NewInstanceID(write.InclusionProof.Key()),
		Xc:    signer.Ed25519.Point,
	}
	var err error
	readBuf, err = protobuf.Encode(read)
	if err != nil {
		return instr, err
	}
	instr = byzcoin.Instruction{
			InstanceID: byzcoin.NewInstanceID(write.InclusionProof.Key()),
			Nonce:      byzcoin.Nonce{},
			Index:      0,
			Length:     1,
			Spawn: &byzcoin.Spawn{
				ContractID: calypso.ContractReadID,
				Args:       byzcoin.Arguments{{Name: "read", Value: readBuf}},
			},
		}

	return instr, err
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

