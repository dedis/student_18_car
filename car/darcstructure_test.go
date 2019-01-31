package car

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
	darcAdmin, err := s.spawnAdminDarc(admin)
	require.Nil(t,err)
	t.Log("Admin Darc")
	t.Log(darcAdmin.String())

	user := darc.NewSignerEd25519(nil, nil)
	darcUser,err := s.spawnUserDarc(darcAdmin, admin, user)
	require.Nil(t,err)
	t.Log("User Darc")
	t.Log(darcUser.String())

	darcReader,err := s.spawnReaderDarc(darcAdmin, admin, darcUser)
	require.Nil(t,err)
	t.Log("Reader Darc")
	t.Log(darcReader.String())

	//newGarage := darc.NewSignerEd25519(nil, nil)
	darcGarage,err := s.spawnGarageDarc(darcAdmin, admin, darcUser)
	require.Nil(t,err)
	t.Log("Garage Darc")
	t.Log(darcGarage.String())

	darcCar,err := s.spawnCarDarc(darcAdmin,
		admin, darcReader, darcGarage)
	require.Nil(t,err)
	t.Log("Car Darc")
	t.Log(darcCar.String())

	newReader := darc.NewSignerEd25519(nil, nil)
	evolved_darc, err := s.addSigner(darcReader, newReader, user)
	require.Nil(t,err)
	t.Log("Evolved Reader Darc")
	t.Log(evolved_darc.String())

	newReader2 := darc.NewSignerEd25519(nil, nil)
	evolved_darc2, err := s.addSigner(evolved_darc, newReader2, user)
	require.Nil(t,err)
	t.Log("Evolved Reader Darc 2")
	t.Log(evolved_darc2.String())

	newReader3 := darc.NewSignerEd25519(nil, nil)
	evolved_darc3, err := s.addSigner(evolved_darc2, newReader3, user)
	require.Nil(t,err)
	t.Log("Evolved Reader Darc 2")
	t.Log(evolved_darc3.String())

	evolved_darc4, err := s.removeSigner(evolved_darc3, newReader3, user)
	require.Nil(t,err)
	t.Log("Evolved Reader Darc 3")
	t.Log(evolved_darc4.String())

	newGarage := darc.NewSignerEd25519(nil, nil)
	evolvedGarage_darc, err := s.addSigner(darcGarage, newGarage, user)
	require.Nil(t,err)
	t.Log("Evolved Garage Darc")
	t.Log(evolvedGarage_darc.String())

	car := NewCar("123A2314")

	//carTemp := Car{}

	cInstance, err := s.createCarInstance(car,
		darcCar, admin)
	require.Nil(t, err)
	t.Log("Car Instance")
	t.Log(cInstance.String())

	resp, err := s.cl.GetProof(cInstance.Slice())
	require.Nil(t, err)
	//key,_,err := resp.Proof.KeyValue()

	_,value, _, _, err := resp.Proof.KeyValue()
	require.Nil(t, err)

	var carData Car
	err = protobuf.Decode(value, &carData)
	require.Nil(t, err)


	//err = resp.Proof.ContractValue(cothority.Suite, ContractCarID, &carTemp)
	//require.Nil(t, err)
	t.Log("Car Instance VIN")
	t.Log(carData.Vin)


	var wData SecretData
	wData.ECOScore = "2310"
	wData.Mileage = "100 000"
	wData.Warranty = true

	s.addReport(cInstance,
		darcCar, wData, newGarage, user)

	resp, err = s.cl.GetProof(cInstance.Slice())
	require.Nil(t, err)

	_,value, _, _, err = resp.Proof.KeyValue()
	require.Nil(t, err)

	var carData2 Car
	err = protobuf.Decode(value, &carData2)
	require.Nil(t, err)

	t.Log("Car Instance Report")
	t.Log(carData2.Reports)


	secrets, err := s.readReports(cInstance, darcCar, newReader, user)
	require.Nil(t, err)

	t.Log("Mileage")
	t.Log(secrets)
}
