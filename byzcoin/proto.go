package byzcoin

// PROTOSTART
// package keyvalue;
//
// option java_package = "ch.epfl.dedis.template.proto";
// option java_outer_classname = "KeyValueProto";

// KeyValue is created as a structure here, as go's map returns the
// elements in a random order and as such is not suitable for use in a
// system that needs to return always the same state.
type KeyValue struct {
	Key   string
	Value []byte
}

// KeyValueData is the structure that will hold all key/value pairs.
type KeyValueData struct {
	Storage []KeyValue
}
