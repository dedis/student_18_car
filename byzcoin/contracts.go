package byzcoin

import (
	"errors"
	"github.com/dedis/cothority/byzcoin"
	"github.com/dedis/cothority/darc"
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

		scs = []byzcoin.StateChange{
			byzcoin.NewStateChange(byzcoin.Create, byzcoin.NewInstanceID(inst.Hash()),
				inst.Spawn.ContractID, c, darcID),
		}
		return

	case byzcoin.InvokeType:
		if inst.Invoke.Command != "addReport" {
			return nil, nil, errors.New("Value contract can only add Reports")
		}

		c := inst.Spawn.Args.Search("car")
		if c == nil || len(c) == 0 {
			return nil, nil, errors.New("need a car argument")
		}

		scs = []byzcoin.StateChange{
			byzcoin.NewStateChange(byzcoin.Update, inst.InstanceID,
				ContractCarID, c, darcID),
		}
		return
	default:
		panic("should not get here")
	}
}

// Update goes through all the arguments and:
//  - updates the value if the key already exists
//  - deletes the keyvalue if the value is empty
//  - adds a new keyValue if the key does not exist yet
/*func (cs *KeyValueData) Update(args byzcoin.Arguments) {
	for _, kv := range args {
		var updated bool
		for i, stored := range cs.Storage {
			if stored.Key == kv.Name {
				updated = true
				if kv.Value == nil || len(kv.Value) == 0 {
					cs.Storage = append(cs.Storage[0:i], cs.Storage[i+1:]...)
					break
				}
				cs.Storage[i].Value = kv.Value
			}

		}
		if !updated {
			cs.Storage = append(cs.Storage, KeyValue{kv.Name, kv.Value})
		}
	}
}*/