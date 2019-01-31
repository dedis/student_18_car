package car

import (
	"github.com/dedis/cothority/byzcoin"
	"github.com/dedis/onet"
	"github.com/dedis/onet/log"
)


var ServiceName = "Calypso_Car"
// This service is only used because we need to register our contracts to
// the ByzCoin service. So we create this stub and add contracts to it
// from the `contracts` directory.

func init() {
	_, err := onet.RegisterNewService(ServiceName, newService)
	log.ErrFatal(err)
}

// Service is only used to being able to store our contracts
type Service struct {
	// We need to embed the ServiceProcessor, so that incoming messages
	// are correctly handled.
	*onet.ServiceProcessor
}


func newService(c *onet.Context) (onet.Service, error) {
	s := &Service{
		ServiceProcessor: onet.NewServiceProcessor(c),
	}
	byzcoin.RegisterContract(c, ContractCarID, ContractCar)
	return s, nil
}

