package byzcoin

import (
	"github.com/dedis/cothority/byzcoin"
	"testing"
)


func TestService_DarcStructure(t *testing.T) {
	s := newSer(t, 1, testInterval)
	defer s.local.CloseAll()
	t.Log("Genesis Darc")
	t.Log(s.gDarc.String())

	admin, darcAdmin := createAdminDarc(t, s)
	t.Log("Admin Darc")
	t.Log(darcAdmin.String())

	_, darcUser := createUserDarc(t, s , darcAdmin, admin)
	t.Log("User Darc")
	t.Log(darcUser.String())

	_, darcReader := createReaderDarc(t, s, darcAdmin, admin, darcUser)
	t.Log("Reader Darc")
	t.Log(darcReader.String())

	_, darcGarage := createGarageDarc(t, s, darcAdmin, admin, darcUser)
	t.Log("Garage Darc")
	t.Log(darcGarage.String())

	darcCar := createCarDarc(t, s, darcAdmin,
		admin, darcReader, darcGarage)
	t.Log("Car Darc")
	t.Log(darcCar.String())

	//VIN number
	vin := "123454321324"
	args := byzcoin.Arguments{
		{
			Name:  "car",
			Value: []byte(vin),
		},
	}
	cInstance := s.createCarInstance(t, args, darcCar, admin)
	t.Log("Car Instance")
	t.Log(cInstance.String())
}