package byzcoin

import (
	"errors"
	"github.com/dedis/cothority/byzcoin"
	"github.com/dedis/cothority/darc"
	"github.com/dedis/protobuf"
)


var ContractCarID = "car"

// ContractCar can only spawn new Car instances and will store the arguments in
// the data field.
func ContractCar(cdb byzcoin.CollectionView, inst byzcoin.Instruction,
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
	case byzcoin.SpawnType:
		if inst.Spawn.ContractID != ContractCarID {
			return nil, nil, errors.New("can only spawn car instances")
		}

		//c is the value of the argument with name car
		c := inst.Spawn.Args.Search("car")
		if c == nil || len(c) == 0 {
			return nil, nil, errors.New("need a car argument")
		}

		instID := inst.DeriveID("")

		scs = []byzcoin.StateChange{
			byzcoin.NewStateChange(byzcoin.Create, instID,
				inst.Spawn.ContractID, c, darcID),
		}
		return

	case byzcoin.InvokeType:
		if inst.Invoke.Command != "addReport" {
			return nil, nil, errors.New("Value contract can only add Reports")
		}
		var carBuf []byte
		carBuf, _, _, err = cdb.GetValues(inst.InstanceID.Slice())
		car := Car{}
		err = protobuf.Decode(carBuf, &car)
		if err != nil {
			return
		}
		err = car.Update(inst.Invoke.Args)
		if err != nil {
			return
		}
		carBuf, err = protobuf.Encode(&car)
		if err != nil {
			return
		}
		scs = []byzcoin.StateChange{
			byzcoin.NewStateChange(byzcoin.Update, inst.InstanceID,
				ContractCarID, carBuf, darcID),
		}

		return
	default:
		panic("should not get here")
	}
}

func (car *Car) Update(args byzcoin.Arguments) error{
	var report Report
	var err error
	for _, kv := range args {
		if kv.Name == "report" {
			err = protobuf.Decode(kv.Value, &report)
			car.Reports = append(car.Reports, report)
		}
	}
	return err
}