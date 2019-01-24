package car

import (
	"github.com/dedis/cothority/byzcoin"
	"github.com/dedis/cothority/calypso"
	"github.com/dedis/cothority/darc"
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
	servers    []*onet.Server
	services []*byzcoin.Service
	roster   *onet.Roster
	signer   darc.Signer
	gDarc    *darc.Darc
	interval time.Duration

	cl         *byzcoin.Client
	gbReply    *byzcoin.CreateGenesisBlockResponse
	genesisMsg *byzcoin.CreateGenesisBlock
	servicesCal   []*calypso.Service
	ltsReply   *calypso.CreateLTSReply
}

func (s *ser) service() *byzcoin.Service {
	return s.services[0]
}

func registerCarContract(servers []*onet.Server) {
	// For testing - there must be a better way to do that. But putting
	// services []skipchain.GetService in the method signature doesn't work :(
	for _, s := range servers {
		byzcoin.RegisterContract(s, ContractCarID, ContractCar)
	}
}

func newSer(t *testing.T, interval time.Duration) *ser {
	return newSerN(t, interval, 4, false)
}


func newSerN(t *testing.T, interval time.Duration, nodes int, viewchange bool) *ser {
	s := &ser{
		local:  onet.NewLocalTestT(tSuite, t),
		signer: darc.NewSignerEd25519(nil, nil),
	}
	s.servers, s.roster, _ = s.local.GenTree(nodes, true)

	for _, sv := range s.local.GetServices(s.servers, byzcoin.ByzCoinID) {
		service := sv.(*byzcoin.Service)
		s.services = append(s.services, service)
	}

	services := s.local.GetServices(s.servers, onet.ServiceFactory.ServiceID(calypso.ServiceName))
	for _, ser := range services {
		s.servicesCal = append(s.servicesCal, ser.(*calypso.Service))
	}

	registerCarContract(s.servers)
	s.createGenesis(t, interval)

	//resp, err := s.services[0].CreateGenesisBlock(s.genesisMsg)
	//require.Nil(t, err)
	//s.sb = resp.Skipblock

	var err error
	// Start DKG
	s.ltsReply, err = s.servicesCal[0].CreateLTS(&calypso.CreateLTS{Roster: *s.roster, BCID: s.gbReply.Skipblock.Hash})
	require.Nil(t, err)

	return s
}

func (s *ser) createGenesis(t *testing.T, interval time.Duration) {
	var err error
	s.genesisMsg, err = byzcoin.DefaultGenesisMsg(byzcoin.CurrentVersion, s.roster,
		[]string{"spawn:darc"}, s.signer.Identity())
	require.Nil(t, err)
	s.gDarc = &s.genesisMsg.GenesisDarc
	s.genesisMsg.BlockInterval = interval
	s.interval = s.genesisMsg.BlockInterval

	s.cl, s.gbReply, err = byzcoin.NewLedger(s.genesisMsg, false)
	require.Nil(t, err)
}