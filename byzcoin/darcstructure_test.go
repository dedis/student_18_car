package byzcoin

import (
	"github.com/dedis/cothority/skipchain"
	"github.com/dedis/kyber/suites"
	"testing"
	"time"

	"github.com/dedis/cothority/darc"
	"github.com/dedis/cothority/darc/expression"
	"github.com/dedis/onet"
	"github.com/stretchr/testify/require"
	"github.com/dedis/cothority/byzcoin"
)

var testInterval = 500 * time.Millisecond
var tSuite = suites.MustFind("Ed25519")
var dummyContract = "dummy"


func TestService_DarcStructure(t *testing.T) {
	s := newSer(t, 1, testInterval)
	defer s.local.CloseAll()
	t.Log("Genesis Darc")
	t.Log(s.darc.String())

	admin, darcAdmin := createAdminDarc(t, s)
	t.Log("Admin Darc")
	t.Log(darcAdmin.String())

	_, darcUser := createUserDarc(t, s , darcAdmin, admin)
	t.Log("User Darc")
	t.Log(darcUser.String())

	_, darcReader := createReaderDarc(t, s, darcAdmin, admin, darcUser)
	t.Log("Reader Darc")
	t.Log(darcReader.String())

	_, darcGarage := createGarageDarc(t, s, darcAdmin, admin, darcUser)
	t.Log("Garage Darc")
	t.Log(darcGarage.String())

	darcCar := createCarDarc(t, s, darcAdmin,
		admin, darcReader, darcGarage)
	t.Log("Car Darc")
	t.Log(darcCar.String())
}

// Spawn an Admin Darc from the Genesis Darc, giving the service as input
//and returning the new darc together with the admin signer
func createAdminDarc(t *testing.T, s *ser) (darc.Signer, *darc.Darc){
	// Spawn Admin darc with a new owner/signer, but delegate its spawn
	// rule to the first darc or the new owner/signer
	admin := darc.NewSignerEd25519(nil, nil)
	idAdmin := []darc.Identity{admin.Identity()}
	darcAdmin := darc.NewDarc(darc.InitRules(idAdmin, idAdmin),
		[]byte("Admin darc"))
	darcAdmin.Rules.AddRule("spawn:darc", expression.InitOrExpr(s.darc.GetIdentityString(), admin.Identity().String()))
	darcAdmin.Rules.AddRule("invoke:evolve", expression.InitOrExpr(s.darc.GetIdentityString(), admin.Identity().String()))
	darcAdminBuf, err := darcAdmin.ToProto()
	require.Nil(t, err)
	darcAdminCopy, err := darc.NewFromProtobuf(darcAdminBuf)
	require.Nil(t, err)
	require.True(t, darcAdmin.Equal(darcAdminCopy))
	ctx := byzcoin.ClientTransaction{
		Instructions: []byzcoin.Instruction{{
			InstanceID: byzcoin.NewInstanceID(s.darc.GetBaseID()),
			Nonce:      byzcoin.GenNonce(),
			Index:      0,
			Length:     1,
			Spawn: &byzcoin.Spawn{
				ContractID: byzcoin.ContractDarcID,
				Args: []byzcoin.Argument{{
					Name:  "darc",
					Value: darcAdminBuf,
				}},
			},
		}},
	}
	require.Nil(t, ctx.Instructions[0].SignBy(s.darc.GetBaseID(), s.signer))
	s.sendTx(t, ctx)
	pr := s.waitProof(t, byzcoin.NewInstanceID(darcAdmin.GetBaseID()))
	require.True(t, pr.InclusionProof.Match())
	return admin,darcAdmin
}

// Spawn an User Darc from the Admin Darc(input), giving the service and Admin Signer as input as well
//and returning the new darc together with the user signer
func createUserDarc(t *testing.T, s *ser, darcAdmin *darc.Darc, admin darc.Signer) (darc.Signer, *darc.Darc){

	// Spawn User darc from the Admin one, but sign the request with
	// the signer of the first darc to test delegation
	user := darc.NewSignerEd25519(nil, nil)
	idUser := []darc.Identity{user.Identity()}
	darcUser := darc.NewDarc(darc.InitRules(idUser, idUser),
		[]byte("User darc"))
	darcUser.Rules.AddRule("invoke:evolve", expression.InitOrExpr(user.Identity().String()))
	darcUserBuf, err := darcUser.ToProto()
	require.Nil(t, err)
	darcUserCopy, err := darc.NewFromProtobuf(darcUserBuf)
	require.Nil(t, err)
	require.True(t, darcUser.Equal(darcUserCopy))
	ctx := byzcoin.ClientTransaction{
		Instructions: []byzcoin.Instruction{{
			InstanceID: byzcoin.NewInstanceID(darcAdmin.GetBaseID()),
			Nonce:      byzcoin.GenNonce(),
			Index:      0,
			Length:     1,
			Spawn: &byzcoin.Spawn{
				ContractID: byzcoin.ContractDarcID,
				Args: []byzcoin.Argument{{
					Name:  "darc",
					Value: darcUserBuf,
				}},
			},
		}},
	}
	require.Nil(t, ctx.Instructions[0].SignBy(darcAdmin.GetBaseID(), admin))
	s.sendTx(t, ctx)
	pr := s.waitProof(t, byzcoin.NewInstanceID(darcUser.GetBaseID()))
	require.True(t, pr.InclusionProof.Match())
	return user, darcUser
}

// Spawn a Reader Darc from the Admin Darc(input), giving the service and Admin Signer as input as well
//and returning the new darc together with the user signer
func createReaderDarc(t *testing.T, s *ser, darcAdmin *darc.Darc, admin darc.Signer, userDarc *darc.Darc) (darc.Signer, *darc.Darc){

	// Spawn Reader darc from the Admin one, but sign the request with
	// the signer of the first darc to test delegation
	reader := darc.NewSignerEd25519(nil, nil)

	//rules for the new Reader Darc
	rs := darc.NewRules()
	if err := rs.AddRule("invoke:evolve", expression.InitAndExpr(userDarc.GetIdentityString())); err != nil {
		panic("add rule should never fail on an empty rule list: " + err.Error())
	}
	if err := rs.AddRule("_sign", expression.InitAndExpr(reader.Identity().String(),userDarc.GetIdentityString())); err != nil {
		panic("add rule should never fail on an empty rule list: " + err.Error())
	}

	darcReader := darc.NewDarc(rs,
		[]byte("Reader darc"))
	darcReaderBuf, err := darcReader.ToProto()
	require.Nil(t, err)
	darcReaderCopy, err := darc.NewFromProtobuf(darcReaderBuf)
	require.Nil(t, err)
	require.True(t, darcReader.Equal(darcReaderCopy))
	ctx := byzcoin.ClientTransaction{
		Instructions: []byzcoin.Instruction{{
			InstanceID: byzcoin.NewInstanceID(darcAdmin.GetBaseID()),
			Nonce:      byzcoin.GenNonce(),
			Index:      0,
			Length:     1,
			Spawn: &byzcoin.Spawn{
				ContractID: byzcoin.ContractDarcID,
				Args: []byzcoin.Argument{{
					Name:  "darc",
					Value: darcReaderBuf,
				}},
			},
		}},
	}
	require.Nil(t, ctx.Instructions[0].SignBy(darcAdmin.GetBaseID(), admin))
	s.sendTx(t, ctx)
	pr := s.waitProof(t, byzcoin.NewInstanceID(darcReader.GetBaseID()))
	require.True(t, pr.InclusionProof.Match())
	return reader, darcReader
}

// Spawn a Reader Darc from the Admin Darc(input), giving the service and Admin Signer as input as well
//and returning the new darc together with the user signer
func createGarageDarc(t *testing.T, s *ser, darcAdmin *darc.Darc, admin darc.Signer, userDarc *darc.Darc) (darc.Signer, *darc.Darc) {

	// Spawn Reader darc from the Admin one, but sign the request with
	// the signer of the first darc to test delegation
	garage := darc.NewSignerEd25519(nil, nil)

	//rules for the new Garage Darc
	rs := darc.NewRules()
	if err := rs.AddRule("invoke:evolve", expression.InitAndExpr(userDarc.GetIdentityString())); err != nil {
		panic("add rule should never fail on an empty rule list: " + err.Error())
	}
	if err := rs.AddRule("_sign", expression.InitAndExpr(garage.Identity().String(), userDarc.GetIdentityString())); err != nil {
		panic("add rule should never fail on an empty rule list: " + err.Error())
	}

	darcGarage := darc.NewDarc(rs,
		[]byte("Garage darc"))
	darcGarageBuf, err := darcGarage.ToProto()
	require.Nil(t, err)
	darcGarageCopy, err := darc.NewFromProtobuf(darcGarageBuf)
	require.Nil(t, err)
	require.True(t, darcGarage.Equal(darcGarageCopy))
	ctx := byzcoin.ClientTransaction{
		Instructions: []byzcoin.Instruction{{
			InstanceID: byzcoin.NewInstanceID(darcAdmin.GetBaseID()),
			Nonce:      byzcoin.GenNonce(),
			Index:      0,
			Length:     1,
			Spawn: &byzcoin.Spawn{
				ContractID: byzcoin.ContractDarcID,
				Args: []byzcoin.Argument{{
					Name:  "darc",
					Value: darcGarageBuf,
				}},
			},
		}},
	}
	require.Nil(t, ctx.Instructions[0].SignBy(darcAdmin.GetBaseID(), admin))
	s.sendTx(t, ctx)
	pr := s.waitProof(t, byzcoin.NewInstanceID(darcGarage.GetBaseID()))
	require.True(t, pr.InclusionProof.Match())
	return garage, darcGarage
}

func createCarDarc(t *testing.T, s *ser, darcAdmin *darc.Darc,
	admin darc.Signer, darcReader *darc.Darc, darcGarage *darc.Darc) (*darc.Darc) {

	// Spawn Car darc from the Admin one

	//rules for the new Car Darc
	rs := darc.NewRules()
	if err := rs.AddRule("spawn:car", expression.InitAndExpr(darcAdmin.GetIdentityString())); err != nil {
		panic("add rule should never fail on an empty rule list: " + err.Error())
	}
	if err := rs.AddRule("spawn:calypsoReade", expression.InitAndExpr(darcReader.GetIdentityString())); err != nil {
		panic("add rule should never fail on an empty rule list: " + err.Error())
	}
	if err := rs.AddRule("invoke:addreport", expression.InitAndExpr(darcGarage.GetIdentityString())); err != nil {
		panic("add rule should never fail on an empty rule list: " + err.Error())
	}

	darcCar := darc.NewDarc(rs,
		[]byte("Car darc"))
	darcCarBuf, err := darcCar.ToProto()
	require.Nil(t, err)
	darcCarCopy, err := darc.NewFromProtobuf(darcCarBuf)
	require.Nil(t, err)
	require.True(t, darcCar.Equal(darcCarCopy))
	ctx := byzcoin.ClientTransaction{
		Instructions: []byzcoin.Instruction{{
			InstanceID: byzcoin.NewInstanceID(darcAdmin.GetBaseID()),
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
		}},
	}
	require.Nil(t, ctx.Instructions[0].SignBy(darcAdmin.GetBaseID(), admin))
	s.sendTx(t, ctx)
	pr := s.waitProof(t, byzcoin.NewInstanceID(darcCar.GetBaseID()))
	require.True(t, pr.InclusionProof.Match())
	return darcCar
}

type ser struct {
	local    *onet.LocalTest
	hosts    []*onet.Server
	roster   *onet.Roster
	services []*byzcoin.Service
	sb       *skipchain.SkipBlock
	value    []byte
	darc     *darc.Darc
	signer   darc.Signer
	tx       byzcoin.ClientTransaction
	interval time.Duration
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
func dummyContractFunc(cdb byzcoin.CollectionView, inst byzcoin.Instruction,
	c []byzcoin.Coin) ([]byzcoin.StateChange, []byzcoin.Coin, error) {
	err := inst.VerifyDarcSignature(cdb)
	if err != nil {
		return nil, nil, err
	}

	_, _, darcID, err := cdb.GetValues(inst.InstanceID.Slice())
	if err != nil {
		return nil, nil, err
	}

	switch inst.GetType() {
	case byzcoin.SpawnType:
		return []byzcoin.StateChange{
			byzcoin.NewStateChange(byzcoin.Create, byzcoin.NewInstanceID(inst.Hash()),
				inst.Spawn.ContractID, inst.Spawn.Args[0].Value, darcID),
		}, nil, nil
	case byzcoin.DeleteType:
		return []byzcoin.StateChange{
			byzcoin.NewStateChange(byzcoin.Remove, inst.InstanceID, "", nil, darcID),
		}, nil, nil
	default:
		panic("should not get here")
	}
}
func registerDummy(servers []*onet.Server) {
	// For testing - there must be a better way to do that. But putting
	// services []skipchain.GetService in the method signature doesn't work :(
	for _, s := range servers {
		byzcoin.RegisterContract(s, dummyContract, dummyContractFunc)
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
	registerDummy(s.hosts)

	genesisMsg, err := byzcoin.DefaultGenesisMsg(byzcoin.CurrentVersion, s.roster,
		[]string{"spawn:darc"}, s.signer.Identity())
	require.Nil(t, err)
	s.darc = &genesisMsg.GenesisDarc

	genesisMsg.BlockInterval = interval
	s.interval = genesisMsg.BlockInterval

	for i := 0; i < step; i++ {
		switch i {
		case 0:
			resp, err := s.service().CreateGenesisBlock(genesisMsg)
			require.Nil(t, err)
			s.sb = resp.Skipblock
		case 1:
			tx, err := createOneClientTx(s.darc.GetBaseID(), dummyContract, s.value, s.signer)
			require.Nil(t, err)
			s.tx = tx
			_, err = s.service().AddTransaction(&byzcoin.AddTxRequest{
				Version:       byzcoin.CurrentVersion,
				SkipchainID:   s.sb.SkipChainID(),
				Transaction:   tx,
				InclusionWait: 10,
			})
			require.Nil(t, err)
		default:
			require.Fail(t, "no such step")
		}
	}
	return s
}


