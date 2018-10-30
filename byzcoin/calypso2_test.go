package byzcoin

import (
	"github.com/dedis/cothority"
	"github.com/dedis/cothority/byzcoin"
	"github.com/dedis/cothority/calypso"
	"github.com/dedis/cothority/darc"
	"github.com/dedis/onet"
	"github.com/dedis/protobuf"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

// TestService_DecryptKey is an end-to-end test that logs two write and read
// requests and make sure that we can decrypt the secret afterwards.
func TestService_DecryptKey(t *testing.T) {
	s := newTS(t, 5)
	defer s.closeAll(t)

	key1 := []byte("secret key 1")
	prWr1 := s.addWrite(t, key1)
	prRe1 := s.addRead(t, prWr1)
	key2 := []byte("secret key 2")
	prWr2 := s.addWrite(t, key2)
	prRe2 := s.addRead(t, prWr2)

	_, err := s.services[0].DecryptKey(&calypso.DecryptKey{Read: *prRe1, Write: *prWr2})
	require.NotNil(t, err)
	_, err = s.services[0].DecryptKey(&calypso.DecryptKey{Read: *prRe2, Write: *prWr1})
	require.NotNil(t, err)

	dk1, err := s.services[0].DecryptKey(&calypso.DecryptKey{Read: *prRe1, Write: *prWr1})
	require.Nil(t, err)
	require.True(t, dk1.X.Equal(s.ltsReply.X))
	keyCopy1, err := calypso.DecodeKey(cothority.Suite, s.ltsReply.X, dk1.Cs, dk1.XhatEnc, s.signer.Ed25519.Secret)
	require.Nil(t, err)
	require.Equal(t, key1, keyCopy1)

	dk2, err := s.services[0].DecryptKey(&calypso.DecryptKey{Read: *prRe2, Write: *prWr2})
	require.Nil(t, err)
	require.True(t, dk2.X.Equal(s.ltsReply.X))
	keyCopy2, err := calypso.DecodeKey(cothority.Suite, s.ltsReply.X, dk2.Cs, dk2.XhatEnc, s.signer.Ed25519.Secret)
	require.Nil(t, err)
	require.Equal(t, key2, keyCopy2)
}

type ts struct {
	local      *onet.LocalTest
	servers    []*onet.Server
	services   []*calypso.Service
	roster     *onet.Roster
	ltsReply   *calypso.CreateLTSReply
	signer     darc.Signer
	cl         *byzcoin.Client
	gbReply    *byzcoin.CreateGenesisBlockResponse
	genesisMsg *byzcoin.CreateGenesisBlock
	gDarc      *darc.Darc
}

func (s *ts) addRead(t *testing.T, write *byzcoin.Proof) *byzcoin.Proof {
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


func newTS(t *testing.T, nodes int) ts {
	s := ts{}
	s.local = onet.NewLocalTestT(cothority.Suite, t)

	// Create the service
	s.servers, s.roster, _ = s.local.GenTree(nodes, true)

	services := s.local.GetServices(s.servers, onet.ServiceFactory.ServiceID(calypso.ServiceName))
	for _, ser := range services {
		s.services = append(s.services, ser.(*calypso.Service))
	}

	// Create the skipchain
	s.signer = darc.NewSignerEd25519(nil, nil)
	s.createGenesis(t)

	// Start DKG
	var err error
	s.ltsReply, err = s.services[0].CreateLTS(&calypso.CreateLTS{Roster: *s.roster, BCID: s.gbReply.Skipblock.Hash})
	require.Nil(t, err)

	return s
}

func (s *ts) createGenesis(t *testing.T) {
	var err error
	s.genesisMsg, err = byzcoin.DefaultGenesisMsg(byzcoin.CurrentVersion, s.roster,
		[]string{"spawn:" + calypso.ContractWriteID, "spawn:" + calypso.ContractReadID}, s.signer.Identity())
	require.Nil(t, err)
	s.gDarc = &s.genesisMsg.GenesisDarc
	s.genesisMsg.BlockInterval = time.Second

	s.cl, s.gbReply, err = byzcoin.NewLedger(s.genesisMsg, false)
	require.Nil(t, err)
}


func (s *ts) addWrite(t *testing.T, key []byte) *byzcoin.Proof {
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
}

func (s *ts) closeAll(t *testing.T) {
	require.Nil(t, s.cl.Close())
	s.local.CloseAll()
}
