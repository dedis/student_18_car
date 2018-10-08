package contracts

import (
	"github.com/dedis/cothority/byzcoin"
	"github.com/dedis/onet"
	"github.com/dedis/onet/log"
)

// This service is only used because we need to register our contracts to
// the ByzCoin service. So we create this stub and add contracts to it
// from the `contracts` directory.

func init() {
	_, err := onet.RegisterNewService("contracts", newService)
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
	byzcoin.RegisterContract(c, ContractValueID, ContractValue)
	byzcoin.RegisterContract(c, ContractCoinID, ContractCoin)
	return s, nil
}
