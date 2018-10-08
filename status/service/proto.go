package status

import (
	"github.com/dedis/onet"
	"github.com/dedis/onet/network"
)

// PROTOSTART
// type :map\[string\]onet.Status:map<string, onet.Status>
// package status;
//
// option java_package = "ch.epfl.dedis.proto";
// option java_outer_classname = "StatusProto";
//
// import "onet.proto";
// import "network.proto";

// Request is what the Status service is expected to receive from clients.
type Request struct {
}

// Response is what the Status service will reply to clients.
type Response struct {
	Status         map[string]*onet.Status
	ServerIdentity *network.ServerIdentity
}
