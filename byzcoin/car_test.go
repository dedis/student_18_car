package byzcoin

import (
	"github.com/dedis/cothority"
	"github.com/dedis/cothority/byzcoin"
	"github.com/dedis/cothority/calypso"
	"github.com/dedis/cothority/darc"
	"github.com/dedis/protobuf"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)


func TestService_Car(t *testing.T) {

}



func NewCar(VIN string) (Car) {
	var c Car
	c.Vin = VIN
	c.Reports = []Report{}
	return c
}

func (s *ser) createCarInstance(t *testing.T, car Car,
	d *darc.Darc, signer darc.Signer) (byzcoin.InstanceID, error) {

	carBuf, err := protobuf.Encode(&car)
	require.Nil(t, err)
	//if err != nil {
	//	return nil, err
	//}
	ctx := byzcoin.ClientTransaction{
		Instructions: []byzcoin.Instruction{{
			InstanceID: byzcoin.NewInstanceID(d.GetBaseID()),
			Nonce:      byzcoin.Nonce{},
			Index:      0,
			Length:     1,
			Spawn: &byzcoin.Spawn{
				ContractID: ContractCarID,
				Args:       byzcoin.Arguments{{Name: "car", Value: carBuf}},
			},
		}},
	}
	// And we need to sign the instruction with the signer that has his
	// public key stored in the darc.
	require.Nil(t, ctx.Instructions[0].SignBy(d.GetBaseID(), signer))


	// Sending this transaction to ByzCoin does not directly include it in the
	// global state - first we must wait for the new block to be created.
	_, err = s.cl.AddTransactionAndWait(ctx, 5)
	require.Nil(t, err)


	_, err = s.cl.GetProof(ctx.Instructions[0].DeriveID("").Slice())
	//_, err = s.cl.GetProof(ctx.Instructions[0].InstanceID.Slice())
	require.Nil(t, err)

	return ctx.Instructions[0].DeriveID(""), err
	//return ctx.Instructions[0].InstanceID, err
}


func (s *ser) addReport(t *testing.T, instID byzcoin.InstanceID,
	controlDarc *darc.Darc, wData WriteData, signerG darc.Signer, signerO darc.Signer) {

	//creating a Calypso Write Instance
	key1 := []byte("secret key 1")
	_, wInstance := s.addWrite(t, key1, wData)

	//creating new Report to be added in the list of the reports in the instance
	var newReport Report
	newReport.Date = time.Now()
	newReport.WriteInstanceID = wInstance
	newReport.GarageId = signerG.Identity().String()


	reportBuf, err := protobuf.Encode(&newReport)
	require.Nil(t, err)

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
	require.Nil(t, ctx.Instructions[0].SignBy(controlDarc.GetBaseID(), signerG, signerO))
	
	_, err = s.cl.AddTransactionAndWait(ctx,5)
	require.Nil(t, err)

	_, err = s.cl.GetProof(ctx.Instructions[0].InstanceID.Slice())
	require.Nil(t, err)

	//_, err = s.cl.AddTransaction(ctx)
	//require.Nil(t, err)
	//_, err = s.cl.GetProof(instID.Slice())
	require.Nil(t, err)
}




// TestService_DecryptKey is an end-to-end test that logs two write and read
// requests and make sure that we can decrypt the secret afterwards.
/*func TestService_CarDecryptKey(t *testing.T) {
	s := newSer(t, testInterval)
	defer s.local.CloseAll()

	key1 := []byte("secret key 1")
	prWr1 := s.addWrite(t, key1)
	prRe1 := s.addRead(t, prWr1)
	key2 := []byte("secret key 2")
	prWr2 := s.addWrite(t, key2)
	prRe2 := s.addRead(t, prWr2)

	_, err := s.servicesCal[0].DecryptKey(&calypso.DecryptKey{Read: *prRe1, Write: *prWr2})
	require.NotNil(t, err)
	_, err = s.servicesCal[0].DecryptKey(&calypso.DecryptKey{Read: *prRe2, Write: *prWr1})
	require.NotNil(t, err)

	dk1, err := s.servicesCal[0].DecryptKey(&calypso.DecryptKey{Read: *prRe1, Write: *prWr1})
	require.Nil(t, err)
	require.True(t, dk1.X.Equal(s.ltsReply.X))
	keyCopy1, err := calypso.DecodeKey(cothority.Suite, s.ltsReply.X, dk1.Cs, dk1.XhatEnc, s.signer.Ed25519.Secret)
	require.Nil(t, err)
	require.Equal(t, key1, keyCopy1)

	dk2, err := s.servicesCal[0].DecryptKey(&calypso.DecryptKey{Read: *prRe2, Write: *prWr2})
	require.Nil(t, err)
	require.True(t, dk2.X.Equal(s.ltsReply.X))
	keyCopy2, err := calypso.DecodeKey(cothority.Suite, s.ltsReply.X, dk2.Cs, dk2.XhatEnc, s.signer.Ed25519.Secret)
	require.Nil(t, err)
	require.Equal(t, key2, keyCopy2)
}*/

func (s *ser) addRead(t *testing.T, write *byzcoin.Proof) *byzcoin.Proof {
	var readBuf []byte
	read := &calypso.Read{
		Write: byzcoin.NewInstanceID(write.InclusionProof.Key),
		Xc:    s.signer.Ed25519.Point,
	}
	var err error
	readBuf, err = protobuf.Encode(read)
	require.Nil(t, err)
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
	require.Nil(t, ctx.Instructions[0].SignBy(s.gDarc.GetID(), s.signer))

	_, err = s.cl.AddTransactionAndWait(ctx, 5)
	require.Nil(t, err)
	instID := ctx.Instructions[0].DeriveID("")

	resp, err := s.cl.GetProof(instID.Slice())
	require.Nil(t, err)

	return &resp.Proof
}

func (s *ser) addWrite(t *testing.T, key []byte, wData WriteData) (*byzcoin.Proof, byzcoin.InstanceID) {
	write := calypso.NewWrite(cothority.Suite, s.ltsReply.LTSID, s.gDarc.GetBaseID(), s.ltsReply.X, key)
	var er error
	write.Data, er = protobuf.Encode(&wData)
	require.Nil(t, er)
	writeBuf, err := protobuf.Encode(write)
	require.Nil(t, err)

	ctx := byzcoin.ClientTransaction{
		Instructions: byzcoin.Instructions{{
			InstanceID: byzcoin.NewInstanceID(s.gDarc.GetBaseID()),
			Nonce:      byzcoin.Nonce{},
			Index:      0,
			Length:     1,
			Spawn: &byzcoin.Spawn{
				ContractID: calypso.ContractWriteID,
				Args:       byzcoin.Arguments{{Name: "write", Value: writeBuf}},
			},
		}},
	}
	require.Nil(t, ctx.Instructions[0].SignBy(s.gDarc.GetID(), s.signer))

	_, err = s.cl.AddTransactionAndWait(ctx, 5)
	require.Nil(t, err)
	instID := ctx.Instructions[0].DeriveID("")

	resp, err := s.cl.GetProof(instID.Slice())
	require.Nil(t, err)

	return &resp.Proof, instID
}

/*func (s *ser) addWrite(t *testing.T, key []byte) (*byzcoin.Proof) {
	write := calypso.NewWrite(cothority.Suite, s.ltsReply.LTSID, s.gDarc.GetBaseID(), s.ltsReply.X, key)
	writeBuf, err := protobuf.Encode(write)
	require.Nil(t, err)

	ctx := byzcoin.ClientTransaction{
		Instructions: byzcoin.Instructions{{
			InstanceID: byzcoin.NewInstanceID(s.gDarc.GetBaseID()),
			Nonce:      byzcoin.Nonce{},
			Index:      0,
			Length:     1,
			Spawn: &byzcoin.Spawn{
				ContractID: calypso.ContractWriteID,
				Args:       byzcoin.Arguments{{Name: "write", Value: writeBuf}},
			},
		}},
	}
	require.Nil(t, ctx.Instructions[0].SignBy(s.gDarc.GetID(), s.signer))

	_, err = s.cl.AddTransactionAndWait(ctx, 5)
	require.Nil(t, err)
	instID := ctx.Instructions[0].DeriveID("")

	resp, err := s.cl.GetProof(instID.Slice())
	require.Nil(t, err)

	return &resp.Proof
}*/