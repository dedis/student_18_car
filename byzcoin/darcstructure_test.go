package byzcoin

import (
	"github.com/dedis/cothority/darc"
	"github.com/dedis/protobuf"
	"github.com/stretchr/testify/require"
	"testing"
)


func TestService_DarcStructure(t *testing.T) {
	s := newSer(t, testInterval)
	defer s.local.CloseAll()
	t.Log("Genesis Darc")
	t.Log(s.gDarc.String())

	admin := darc.NewSignerEd25519(nil, nil)
	darcAdmin, err := s.createAdminDarc(admin)
	require.Nil(t,err)
	t.Log("Admin Darc")
	t.Log(darcAdmin.String())

	user, darcUser := s.createUserDarc(t, darcAdmin, admin)
	t.Log("User Darc")
	t.Log(darcUser.String())

	darcReader := s.createReaderDarc(t, darcAdmin, admin, darcUser)
	t.Log("Reader Darc")
	t.Log(darcReader.String())

	//newGarage := darc.NewSignerEd25519(nil, nil)
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
	evolvedCar_darc := s.addSigner(t, darcGarage, newGarage, user)
	t.Log("Evolved Reader Darc")
	t.Log(evolvedCar_darc.String())

	car := NewCar("123A2314")

	//carTemp := Car{}

	cInstance, err := s.createCarInstance(t, car,
		darcCar, admin)
	require.Nil(t, err)
	t.Log("Car Instance")
	t.Log(cInstance.String())

	resp, err := s.cl.GetProof(cInstance.Slice())
	require.Nil(t, err)
	//key,_,err := resp.Proof.KeyValue()

	_,values, err := resp.Proof.KeyValue()
	require.Nil(t, err)

	var carData Car
	err = protobuf.Decode(values[0], &carData)
	require.Nil(t, err)


	//err = resp.Proof.ContractValue(cothority.Suite, ContractCarID, &carTemp)
	//require.Nil(t, err)
	t.Log("Car Instance VIN")
	t.Log(carData.Vin)



	var wData WriteData
	wData.ECOScore = "2310"
	wData.Mileage = "100 000"
	wData.Warranty = true

	s.addReport(t, cInstance,
		darcCar, wData, newGarage, user)

	resp, err = s.cl.GetProof(cInstance.Slice())
	require.Nil(t, err)
	//key,_,err := resp.Proof.KeyValue()

	_,values, err = resp.Proof.KeyValue()
	require.Nil(t, err)

	var carData2 Car
	err = protobuf.Decode(values[0], &carData2)
	require.Nil(t, err)

	t.Log("Car Instance Report")
	t.Log(carData2.Reports)



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
