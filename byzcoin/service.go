package byzcoin

import (
	"github.com/dedis/cothority/byzcoin"
	"github.com/dedis/onet"
	"github.com/dedis/onet/log"
)


var ServiceName = "Calypso_Car"
// This service is only used because we need to register our contracts to
// the ByzCoin service. So we create this stub and add contracts to it
// from the `contracts` directory.

/*type vData struct {
	Proof     byzcoin.Proof
	Ephemeral kyber.Point
	Signature *darc.Signature
}*/

func init() {
	_, err := onet.RegisterNewService(ServiceName, newService)
	log.ErrFatal(err)

	/*var err2 error
	calypsoID, err2 = onet.RegisterNewService(ServiceName, newServiceCalypso)
	log.ErrFatal(err2)*/
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
	byzcoin.RegisterContract(c, ContractKeyValueID, ContractKeyValue)
	byzcoin.RegisterContract(c, ContractCarID, ContractCar)
	return s, nil
}
