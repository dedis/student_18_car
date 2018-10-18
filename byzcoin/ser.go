package byzcoin

import (
	"github.com/dedis/cothority/byzcoin"
	"github.com/dedis/cothority/darc"
	"github.com/dedis/cothority/skipchain"
	"github.com/dedis/onet"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

// This service is only used because we need to register our contracts to
// the ByzCoin service. So we create this stub and add contracts to it
// from the `contracts` directory.

type ser struct {
	local    *onet.LocalTest
	hosts    []*onet.Server
	roster   *onet.Roster
	services []*byzcoin.Service
	sb       *skipchain.SkipBlock
	value    []byte
	gDarc     *darc.Darc
	signer   darc.Signer
	interval time.Duration
	cl      *byzcoin.Client
}

func (s *ser) service() *byzcoin.Service {
	return s.services[0]
}

func (s *ser) sendTx(t *testing.T, ctx byzcoin.ClientTransaction) {
	s.sendTxTo(t, ctx, 0)
}

func (s *ser) sendTxTo(t *testing.T, ctx byzcoin.ClientTransaction, idx int) {
	_, err := s.services[idx].AddTransaction(&byzcoin.AddTxRequest{
		Version:     byzcoin.CurrentVersion,
		SkipchainID: s.sb.SkipChainID(),
		Transaction: ctx,
	})
	require.Nil(t, err)
}
func (s *ser) waitProof(t *testing.T, id byzcoin.InstanceID) byzcoin.Proof {
	return s.waitProofWithIdx(t, id.Slice(), 0)
}

func (s *ser) waitProofWithIdx(t *testing.T, key []byte, idx int) byzcoin.Proof {
	var pr byzcoin.Proof
	var ok bool
	for i := 0; i < 10; i++ {
		// wait for the block to be processed
		time.Sleep(2 * s.interval)

		resp, err := s.services[idx].GetProof(&byzcoin.GetProof{
			Version: byzcoin.CurrentVersion,
			Key:     key,
			ID:      s.sb.SkipChainID(),
		})
		require.Nil(t, err)
		pr = resp.Proof
		if pr.InclusionProof.Match() {
			ok = true
			break
		}
	}

	require.True(t, ok, "got not match")
	return pr
}

func registerCarContract(servers []*onet.Server) {
	// For testing - there must be a better way to do that. But putting
	// services []skipchain.GetService in the method signature doesn't work :(
	for _, s := range servers {
		byzcoin.RegisterContract(s, ContractCarID, ContractCar)
	}
}

func createOneClientTx(dID darc.ID, kind string, value []byte, signer darc.Signer) (byzcoin.ClientTransaction, error) {
	instr, err := createInstr(dID, kind, value, signer)
	if err != nil {
		return byzcoin.ClientTransaction{}, err
	}
	t := byzcoin.ClientTransaction{
		Instructions: []byzcoin.Instruction{instr},
	}
	return t, err
}

func createInstr(dID darc.ID, contractID string, value []byte, signer darc.Signer) (byzcoin.Instruction, error) {
	instr := byzcoin.Instruction{
		InstanceID: byzcoin.NewInstanceID(dID),
		Spawn: &byzcoin.Spawn{
			ContractID: contractID,
			Args:       byzcoin.Arguments{{Name: "data", Value: value}},
		},
		Nonce:  byzcoin.GenNonce(),
		Index:  0,
		Length: 1,
	}
	err := instr.SignBy(dID, signer)
	return instr, err
}

func newSer(t *testing.T, step int, interval time.Duration) *ser {
	return newSerN(t, step, interval, 4, false)
}

func newSerN(t *testing.T, step int, interval time.Duration, n int, viewchange bool) *ser {
	s := &ser{
		local:  onet.NewLocalTestT(tSuite, t),
		value:  []byte("anyvalue"),
		signer: darc.NewSignerEd25519(nil, nil),
	}
	s.hosts, s.roster, _ = s.local.GenTree(n, true)
	for _, sv := range s.local.GetServices(s.hosts, byzcoin.ByzCoinID) {
		service := sv.(*byzcoin.Service)
		s.services = append(s.services, service)
	}
	registerCarContract(s.hosts)

	genesisMsg, err := byzcoin.DefaultGenesisMsg(byzcoin.CurrentVersion, s.roster,
		[]string{"spawn:darc"}, s.signer.Identity())
	require.Nil(t, err)
	s.gDarc = &genesisMsg.GenesisDarc

	genesisMsg.BlockInterval = interval
	s.interval = genesisMsg.BlockInterval

	for i := 0; i < step; i++ {
		switch i {
		case 0:
			resp, err := s.service().CreateGenesisBlock(genesisMsg)
			require.Nil(t, err)
			s.sb = resp.Skipblock
			/*case 1:
				tx, err := createOneClientTx(s.gDarc.GetBaseID(), dummyContract, s.value, s.signer)
				require.Nil(t, err)
				s.tx = tx
				_, err = s.service().AddTransaction(&byzcoin.AddTxRequest{
					Version:       byzcoin.CurrentVersion,
					SkipchainID:   s.sb.SkipChainID(),
					Transaction:   tx,
					InclusionWait: 10,
				})
				require.Nil(t, err)*/
		default:
			require.Fail(t, "no such step")
		}
	}
	return s
}
