package byzcoin

import (
	"github.com/dedis/cothority/darc"
	"testing"
)


func TestService_DarcStructure(t *testing.T) {
	s := newSer(t, testInterval)
	defer s.local.CloseAll()
	t.Log("Genesis Darc")
	t.Log(s.gDarc.String())

	admin, darcAdmin := s.createAdminDarc(t)
	t.Log("Admin Darc")
	t.Log(darcAdmin.String())

	user, darcUser := s.createUserDarc(t, darcAdmin, admin)
	t.Log("User Darc")
	t.Log(darcUser.String())

	darcReader := s.createReaderDarc(t, darcAdmin, admin, darcUser)
	t.Log("Reader Darc")
	t.Log(darcReader.String())

	darcGarage := s.createGarageDarc(t, darcAdmin, admin, darcUser)
	t.Log("Garage Darc")
	t.Log(darcGarage.String())

	darcCar := s.createCarDarc(t, darcAdmin,
		admin, darcReader, darcGarage)
	t.Log("Car Darc")
	t.Log(darcCar.String())

	newReader := darc.NewSignerEd25519(nil, nil)
	evolved_darc := s.addSigner(t, darcReader, newReader, user)
	t.Log("Evolved Reader Darc")
	t.Log(evolved_darc.String())

	newGarage := darc.NewSignerEd25519(nil, nil)
	evolvedGarage_darc := s.addSigner(t, darcGarage, newGarage, user)
	t.Log("Evolved Garage Darc")
	t.Log(evolvedGarage_darc.String())

	t.Log("Evolved Reader Darc")
	t.Log(darcReader.String())

	/*car := NewCar("123454321324")

	//carTemp := Car{}

	cInstance, err := s.createCarInstance(t, car,
		darcCar, admin)
	require.Nil(t, err)
	t.Log("Car Instance")
	t.Log(cInstance.String())

	//resp, err := s.cl.GetProof(cInstance.Slice())
	//require.Nil(t, err)
	//_,vals,err := resp.Proof.KeyValue()


	/*err = protobuf.Decode(vals[0], &carTemp)
	require.Nil(t, err)
	t.Log("Key")
	t.Log(carTemp.VIN)


	//err = resp.Proof.ContractValue(cothority.Suite, ContractCarID, &carTemp)
	//require.Nil(t, err)
	t.Log("Car Instance VIN")
	t.Log(carTemp.VIN)*/



	/*var wData WriteData
	wData.ECOScore = "2310"
	wData.Mileage = "100 000"
	wData.Warranty = true

	s.addReport(t, cInstance,
		darcCar, newGarage, wData)

	t.Log("Car Instance")
	t.Log(cInstance.String())*/
	/*//VIN number
	vin := "123454321324"
	args := byzcoin.Arguments{
		{
			Name:  "car",
			Value: []byte(vin),
		},
	}
	cInstance := s.createCarInstance(t, args, darcCar, admin)
	t.Log("Car Instance")
	t.Log(cInstance.String())*/
}
