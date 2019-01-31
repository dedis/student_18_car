package main

import (
	"errors"
	"github.com/BurntSushi/toml"
	"github.com/dedis/cothority/byzcoin"
	"github.com/dedis/cothority/calypso"
	"github.com/dedis/cothority/darc"
	"github.com/dedis/kyber/util/random"
	"github.com/dedis/onet"
	"github.com/dedis/onet/log"
	"github.com/dedis/onet/simul/monitor"
	"github.com/dedis/student_18_car/car"
	"time"
)

/*
 * Defines the simulation for the service-template
 */

func init() {
	onet.SimulationRegister("SimulationCarService", NewSimulationService)
}

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

	var carDarcs []darc.Darc
	var carInstances []byzcoin.InstanceID
	var writeInstances []byzcoin.InstanceID
	var wData car.SecretData
	wData.ECOScore = "2310"
	wData.Mileage = "100 000"
	wData.Warranty = true

	// Measuring the time it takes to setup the blockchain
	bChainSetup := monitor.NewTimeMeasure("bChainSetup")
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
	//todo is this the way i should use calypso???
	calypsoClient := calypso.NewClient(c)
	lts, err := calypsoClient.CreateLTS()
	if err != nil {
		return errors.New("couldn't create genesis block: " + err.Error())
	}
	bChainSetup.Record()

	// Measuring the time it takes to prepare the system
	prepare := monitor.NewTimeMeasure("prepare")
	// Spawn an Admin Darc from the genesis darc
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

	// Create user darc, which will be used as reader, owner and garage for simplicity
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

	//create #(s.Transactions) car darcs and store them in []carDarcs
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
			_, err = c.AddTransaction(tx)
			if err != nil {
				return errors.New("couldn't add transaction: " + err.Error())
			}
			tx.Instructions = byzcoin.Instructions{}
		}
		for i := 0; i < insts; i++ {
			counterID++
			inst, carDarc, err := spawnCarDarc(&adminDarc,
				&userDarc, counterID)
			if err != nil {
				return errors.New("instruction error: " + err.Error())
			}
			carDarcs = append(carDarcs, *carDarc)
			tx.Instructions = append(tx.Instructions, inst)
			err = byzcoin.SignInstruction(&tx.Instructions[i], adminDarc.GetBaseID(), admin)
			if err != nil {
				return errors.New("signature error: " + err.Error())
			}
		}
	}
	log.Lvl1("Sending last transaction and waiting")
	_, err = c.AddTransactionAndWait(tx, 20)
	if err != nil {
		return errors.New("while adding transaction and waiting: " + err.Error())
	}



	//create #(s.Transactions) car Instances and store them in []carInstances
	log.Lvlf1("Sending %d transactions with %d instructions each", txs, insts)
	tx = byzcoin.ClientTransaction{}
	for t := 0; t < txs; t++ {
		if len(tx.Instructions) > 0 {
			log.Lvlf1("Sending transaction %d", t)
			_, err = c.AddTransaction(tx)
			if err != nil {
				return errors.New("couldn't add transaction: " + err.Error())
			}
			tx.Instructions = byzcoin.Instructions{}
		}
		for i := 0; i < insts; i++ {
			counterID++
			c := car.NewCar(string(counterID))
			instr, err := createCarInstanceInstr(c, &carDarcs[t*insts+i])
			if err != nil {
				return errors.New("instruction error: " + err.Error())
			}
			carInstances = append(carInstances, instr.DeriveID(""))
			tx.Instructions = append(tx.Instructions, instr)
			err = byzcoin.SignInstruction(&tx.Instructions[i], carDarcs[t*insts+i].GetBaseID(), admin)
			if err != nil {
				return errors.New("signature error: " + err.Error())
			}
		}
	}
	// Confirm the transaction by sending the last transaction using
	// AddTransactionAndWait. There is a small error in measurement,
	// as we're missing one of the AddTransaction call in the measurements.
	log.Lvl1("Sending last transaction and waiting")
	_, err = c.AddTransactionAndWait(tx, 20)
	if err != nil {
		return errors.New("while adding transaction and waiting: " + err.Error())
	}

	prepare.Record()


	addReports := monitor.NewTimeMeasure("addReports")
	//create #(s.Transactions) car Instances and store them in []writeInstances
	log.Lvlf1("Sending %d transactions with %d instructions each", txs, insts)
	tx = byzcoin.ClientTransaction{}
	for t := 0; t < txs; t++ {
		if len(tx.Instructions) > 0 {
			log.Lvlf1("Sending transaction %d", t)
			_, err = c.AddTransaction(tx)
			if err != nil {
				return errors.New("couldn't add transaction: " + err.Error())
			}
			tx.Instructions = byzcoin.Instructions{}
		}
		for i := 0; i < insts; i++ {
			counterID++
			key := random.Bits(128, true, random.New())
			instr, err := addWrite(key, wData,
				&carDarcs[t*insts+i], *lts)
			if err != nil {
				return errors.New("instruction error: " + err.Error())
			}
			writeInstances = append(writeInstances, instr.DeriveID(""))
			tx.Instructions = append(tx.Instructions, instr)
			err = byzcoin.SignInstruction(&tx.Instructions[i], carDarcs[t*insts+i].GetBaseID(), user)
			if err != nil {
				return errors.New("signature error: " + err.Error())
			}
		}
	}
	// Confirm the transaction by sending the last transaction using
	// AddTransactionAndWait. There is a small error in measurement,
	// as we're missing one of the AddTransaction call in the measurements.
	log.Lvl1("Sending last transaction and waiting")
	_, err = c.AddTransactionAndWait(tx, 20)
	if err != nil {
		return errors.New("while adding transaction and waiting: " + err.Error())
	}


	//add a report for each car instance
	log.Lvlf1("Sending %d transactions with %d instructions each", txs, insts)
	tx = byzcoin.ClientTransaction{}
	for t := 0; t < txs; t++ {
		if len(tx.Instructions) > 0 {
			log.Lvlf1("Sending transaction %d", t)
			_, err = c.AddTransaction(tx)
			if err != nil {
				return errors.New("couldn't add transaction: " + err.Error())
			}
			tx.Instructions = byzcoin.Instructions{}
		}
		for i := 0; i < insts; i++ {
			instruction, err := addReport(carInstances[t*insts+i], writeInstances[t*insts+i], user)
			if err != nil {
				return errors.New("instruction error: " + err.Error())
			}
			tx.Instructions = append(tx.Instructions, instruction)
			err = byzcoin.SignInstruction(&tx.Instructions[i], carDarcs[t*insts+i].GetBaseID(), user)
			if err != nil {
				return errors.New("signature error: " + err.Error())
			}
		}
	}
	// Confirm the transaction by sending the last transaction using
	// AddTransactionAndWait. There is a small error in measurement,
	// as we're missing one of the AddTransaction call in the measurements.
	log.Lvl1("Sending last transaction and waiting")
	_, err = c.AddTransactionAndWait(tx, 20)
	if err != nil {
		return errors.New("while adding transaction and waiting: " + err.Error())
	}
	addReports.Record()



	readReports := monitor.NewTimeMeasure("readReports")

	readReports.Record()


	// This sleep is needed to wait for the propagation to finish
	// on all the nodes. Otherwise the simulation manager
	// (runsimul.go in onet) might close some nodes and cause
	// skipblock propagation to fail.
	time.Sleep(blockInterval)


	// We wait a bit before closing because c.GetProof is sent to the
	// leader, but at this point some of the children might still be doing
	// updateCollection. If we stop the simulation immediately, then the
	// database gets closed and updateCollection on the children fails to
	// complete.
	time.Sleep(time.Second)
	return nil
}






