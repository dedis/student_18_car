package byzcoin

import(
	"github.com/dedis/cothority/byzcoin"
	"time"
)


// PROTOSTART
// package keyvalue;
//
// option java_package = "ch.epfl.dedis.template.proto";
// option java_outer_classname = "KeyValueProto";

// KeyValue is created as a structure here, as go's map returns the
// elements in a random order and as such is not suitable for use in a
// system that needs to return always the same state.

//For the Car contract Key will be the VIN, and value the car struct
type KeyValue struct {
	Key   string
	Value []byte
}

// KeyValueData is the structure that will hold all key/value pairs.
type KeyValueData struct {
	Storage []KeyValue
}

type Report struct {
	Date time.Time
	GarageId string
	WriteInstanceID byzcoin.InstanceID
}

type Car struct {
	Vin string
	Reports []Report
}

type WriteData struct {
	ECOScore string
	Mileage string
	Warranty bool
}