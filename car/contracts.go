package car

import (
	"errors"
	"github.com/dedis/cothority/byzcoin"
	"github.com/dedis/cothority/darc"
	"github.com/dedis/protobuf"
)


var ContractCarID = "car"

// ContractCar can only spawn new Car instances and will store the arguments in
// the data field.
func ContractCar(cdb byzcoin.ReadOnlyStateTrie, inst byzcoin.Instruction,
	cIn []byzcoin.Coin) (scs []byzcoin.StateChange, cOut []byzcoin.Coin, err error) {

	cOut = cIn
	err = inst.VerifyDarcSignature(cdb)
	if err != nil {
		return
	}

	var darcID darc.ID
	_, _, darcID, err = cdb.GetValues(inst.InstanceID.Slice())
	if err != nil {
		return
	}

	switch inst.GetType() {
	//spawns new car instance
	case byzcoin.SpawnType:
		if inst.Spawn.ContractID != ContractCarID {
			return nil, nil, errors.New("can only spawn car instances")
		}
		//cBuf is the value of the argument with name car
		cBuf := inst.Spawn.Args.Search("car")
		if cBuf == nil || len(cBuf) == 0 {
			return nil, nil, errors.New("need a car argument")
		}
		//verify that is car
		car := Car{}
		err = protobuf.Decode(cBuf, &car)
		if err != nil {
			return nil, nil, errors.New("not a car")
		}

		instID := inst.DeriveID("")
		//creating the Car Instance in the global state
		scs = []byzcoin.StateChange{
			byzcoin.NewStateChange(byzcoin.Create, instID,
				inst.Spawn.ContractID, cBuf, darcID),
		}
		return
	//updates the car instance by adding a new report
	case byzcoin.InvokeType:
		if inst.Invoke.Command != "addReport" {
			return nil, nil, errors.New("Value contract can only add Reports")
		}
		//getting the Car Data from the car instance
		var carBuf []byte
		carBuf, _, _, err = cdb.GetValues(inst.InstanceID.Slice())
		car := Car{}
		err = protobuf.Decode(carBuf, &car)
		if err != nil {
			return
		}
		//adding reports to the car data
		err = car.Add(inst.Invoke.Args)
		if err != nil {
			return
		}
		carBuf, err = protobuf.Encode(&car)
		if err != nil {
			return
		}
		//updating the car instance so that it contains the new reports
		scs = []byzcoin.StateChange{
			byzcoin.NewStateChange(byzcoin.Update, inst.InstanceID,
				ContractCarID, carBuf, darcID),
		}
		return
	default:
		panic("should not get here")
	}
}

//the arguments should have "report" as name and the Report marshaled in bytes as value
func (car *Car) Add(args byzcoin.Arguments) error{
	var report Report
	var err error
	for _, rep := range args {
		if rep.Name == "report" {
			err = protobuf.Decode(rep.Value, &report)
			car.Reports = append(car.Reports, report)
		}
	}
	return err
}