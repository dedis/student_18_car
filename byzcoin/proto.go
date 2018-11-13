package byzcoin

// PROTOSTART
// package car;
//
// option java_package = "ch.epfl.dedis.template.proto";
// option java_outer_classname = "CarProto";

type KeyValue struct {
	Key   string
	Value []byte
}

// KeyValueData is the structure that will hold all key/value pairs.
type KeyValueData struct {
	Storage []KeyValue
}

//todo Report.java
type Report struct {
	Date string
	GarageId string
	//todo there is an error when i run make proto
	//WriteInstanceID byzcoin.InstanceID
	WriteInstanceID []byte
}
//todo Car.java and CarInstance.java
type Car struct {
	Vin string
	Reports []Report
}
//todo SecretData.java
type SecretData struct {
	ECOScore string
	Mileage string
	Warranty bool
	CheckNote string
}