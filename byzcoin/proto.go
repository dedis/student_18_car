package byzcoin

// PROTOSTART
// package car;
// type :time.Time:sfixed64
//
// option java_package = "ch.epfl.dedis.template.proto";
// option java_outer_classname = "CarProto";


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