package main

import (
	"errors"
	"github.com/BurntSushi/toml"
	"github.com/dedis/cothority/byzcoin"
	"github.com/dedis/cothority/darc"
	"github.com/dedis/cothority/darc/expression"
	"github.com/dedis/onet"
	"github.com/dedis/onet/log"
	"github.com/dedis/onet/simul/monitor"
	"time"
)

/*
 * Defines the simulation for the service-template
 */

func init() {
	onet.SimulationRegister("SimulationCarService", NewSimulationService)
}

// SimulationService only holds the BFTree simulation
type SimulationService struct {
	onet.SimulationBFTree
	Transactions  int
	BlockInterval string
	BatchSize     int
	Keep          bool
	Delay         int
}

// NewSimulationService returns the new simulation, where all fields are
// initialised using the config-file
func NewSimulationService(config string) (onet.Simulation, error) {
	es := &SimulationService{}
	_, err := toml.Decode(config, es)
	if err != nil {
		return nil, err
	}
	return es, nil
}

// Setup creates the tree used for that simulation
func (s *SimulationService) Setup(dir string, hosts []string) (
	*onet.SimulationConfig, error) {
	sc := &onet.SimulationConfig{}
	s.CreateRoster(sc, hosts, 2000)
	err := s.CreateTree(sc)
	if err != nil {
		return nil, err
	}
	return sc, nil
}

// Node can be used to initialize each node before it will be run
// by the server. Here we call the 'Node'-method of the
// SimulationBFTree structure which will load the roster- and the
// tree-structure to speed up the first round.
func (s *SimulationService) Node(config *onet.SimulationConfig) error {
	index, _ := config.Roster.Search(config.Server.ServerIdentity.ID)
	if index < 0 {
		log.Fatal("Didn't find this node in roster")
	}
	log.Lvl3("Initializing node-index", index)
	return s.SimulationBFTree.Node(config)
}


// Run is used on the destination machines and runs a number of
// rounds
func (s *SimulationService) Run(config *onet.SimulationConfig) error {
	size := config.Tree.Size()
	log.Lvl2("Size is:", size, "rounds:", s.Rounds, "transactions:", s.Transactions)
	signer := darc.NewSignerEd25519(nil, nil)

	// Create the ledger
	gm, err := byzcoin.DefaultGenesisMsg(byzcoin.CurrentVersion, config.Roster,
		[]string{"spawn:darc"}, signer.Identity())
	if err != nil {
		return errors.New("couldn't setup genesis message: " + err.Error())
	}

	// Set block interval from the simulation config.
	blockInterval, err := time.ParseDuration(s.BlockInterval)
	if err != nil {
		return errors.New("parse duration of BlockInterval failed: " + err.Error())
	}
	gm.BlockInterval = blockInterval

	c, _, err := byzcoin.NewLedger(gm, s.Keep)
	if err != nil {
		return errors.New("couldn't create genesis block: " + err.Error())
	}

	//create a darc for the Admin(not the genesis one) to be able to spawn new darcs (reader/garage/car...)
	admin := darc.NewSignerEd25519(nil, nil)
	ctx, adminDarc, err := spawnDarcTxn(gm.GenesisDarc, admin)
	if err != nil {
		return errors.New("couldn't create transaction: " + err.Error())
	}
	// Now sign all the instructions
	for i := range ctx.Instructions {
		if err = byzcoin.SignInstruction(&ctx.Instructions[i], gm.GenesisDarc.GetBaseID(), signer); err != nil {
			return errors.New("signing of instruction failed: " + err.Error())
		}
	}
	// Send the instructions.
	_, err = c.AddTransactionAndWait(ctx, 2)
	if err != nil {
		return errors.New("couldn't create admin darc: " + err.Error())
	}

	//create user darc, which will be used as reader, owner and garage for simplicity
	user := darc.NewSignerEd25519(nil, nil)
	ctx, userDarc, err := spawnDarcTxn(adminDarc, user)
	if err != nil {
		return errors.New("couldn't create transaction: " + err.Error())
	}
	// Now sign all the instructions
	for i := range ctx.Instructions {
		if err = byzcoin.SignInstruction(&ctx.Instructions[i], adminDarc.GetBaseID(), admin); err != nil {
			return errors.New("signing of instruction failed: " + err.Error())
		}
	}
	// Send the instructions.
	_, err = c.AddTransactionAndWait(ctx, 2)
	if err != nil {
		return errors.New("couldn't create user darc: " + err.Error())
	}


	//create cars
	for round := 0; round < s.Rounds; round++ {
		log.Lvl1("Starting round", round)
		roundM := monitor.NewTimeMeasure("round")

		if s.Transactions < 3 {
			log.Warn("The 'send_sum' measurement will be very skewed, as the last transaction")
			log.Warn("is not measured.")
		}

		txs := s.Transactions / s.BatchSize
		insts := s.BatchSize
		log.Lvlf1("Sending %d transactions with %d instructions each", txs, insts)
		tx := byzcoin.ClientTransaction{}
		// Inverse the prepare/send loop, so that the last transaction is not sent,
		// but can be sent in the 'confirm' phase using 'AddTransactionAndWait'.
		counterID := 0
		for t := 0; t < txs; t++ {
			if len(tx.Instructions) > 0 {
				log.Lvlf1("Sending transaction %d", t)
				send := monitor.NewTimeMeasure("send")
				_, err = c.AddTransaction(tx)
				if err != nil {
					return errors.New("couldn't add transfer transaction: " + err.Error())
				}
				send.Record()
				tx.Instructions = byzcoin.Instructions{}
			}

			prepare := monitor.NewTimeMeasure("prepare")
			for i := 0; i < insts; i++ {

				inst,_,err := spawnCarDarc( &adminDarc,
					&userDarc, counterID)
				if err != nil {
					return errors.New("instruction error: " + err.Error())
				}
				tx.Instructions = append(tx.Instructions, inst)
				err = byzcoin.SignInstruction(&tx.Instructions[i], adminDarc.GetBaseID(), admin)
				if err != nil {
					return errors.New("signature error: " + err.Error())
				}
			}
			prepare.Record()
		}
		// Confirm the transaction by sending the last transaction using
		// AddTransactionAndWait. There is a small error in measurement,
		// as we're missing one of the AddTransaction call in the measurements.
		confirm := monitor.NewTimeMeasure("confirm")
		log.Lvl1("Sending last transaction and waiting")
		_, err = c.AddTransactionAndWait(tx, 20)
		if err != nil {
			return errors.New("while adding transaction and waiting: " + err.Error())
		}
		//todo should i check some proof for any car?
		//proof, err := c.GetProof(coinAddr2.Slice())
		//if err != nil {
		//	return errors.New("couldn't get proof for transaction: " + err.Error())
		//}
		//_, v0, _, _, err := proof.Proof.KeyValue()
		//if err != nil {
		//	return errors.New("proof doesn't hold transaction: " + err.Error())
		//}
		//var account byzcoin.Coin
		//err = protobuf.Decode(v0, &account)
		//if err != nil {
		//	return errors.New("couldn't decode account: " + err.Error())
		//}
		//log.Lvlf1("Account has %d", account.Value)
		//if account.Value != uint64(s.Transactions*(round+1)) {
		//	return errors.New("account has wrong amount")
		//}
		confirm.Record()
		roundM.Record()

		// This sleep is needed to wait for the propagation to finish
		// on all the nodes. Otherwise the simulation manager
		// (runsimul.go in onet) might close some nodes and cause
		// skipblock propagation to fail.
		time.Sleep(blockInterval)
	}
	// We wait a bit before closing because c.GetProof is sent to the
	// leader, but at this point some of the children might still be doing
	// updateCollection. If we stop the simulation immediately, then the
	// database gets closed and updateCollection on the children fails to
	// complete.
	time.Sleep(time.Second)
	return nil
}



func spawnDarcTxn(controlDarc darc.Darc, newDracSigner darc.Signer)  (byzcoin.ClientTransaction, darc.Darc, error){
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

func newSpawnDarcTransaction(controlDarc *darc.Darc, newDarcBuf []byte) byzcoin.ClientTransaction{

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


func spawnCarDarc( controlDarc *darc.Darc,
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
		[]byte("Car darc" + string(id)))
	darcCarBuf, err := darcCar.ToProto()
	//ctx := newSpawnDarcTransaction(controlDarc, darcCarBuf)
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

