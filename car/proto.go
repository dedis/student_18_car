package car

// PROTOSTART
// package car;
// type :time.Time:sfixed64
//
// option java_package = "ch.epfl.dedis.template.proto";
// option java_outer_classname = "CarProto";


type Report struct {
	Date string
	GarageId string
	WriteInstanceID []byte
}

type Car struct {
	Vin string
	Reports []Report
}

type SecretData struct {
	ECOScore string
	Mileage string
	Warranty bool
	CheckNote string
}



//todo send prop
//todo list prop
//todo repl