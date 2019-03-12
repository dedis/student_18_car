package car

import (
	"github.com/dedis/cothority/byzcoin"
	"github.com/dedis/cothority/darc"
	"github.com/stretchr/testify/require"
	"testing"
)


func TestService_DarcStructure(t *testing.T) {
	s := newSer(t, testInterval)
	defer s.local.CloseAll()

	//creating admin and admin darc
	admin := darc.NewSignerEd25519(nil, nil)
	ctx, darcAdmin, err := spawnAdminDarc(s.gDarc, admin)
	require.Nil(t,err)
	_, err = s.signAndSendTransaction(ctx, s.signer, s.gDarc, byzcoin.NewInstanceID(darcAdmin.GetBaseID()).Slice())
	require.Nil(t,err)

	//creating user and user darc
	user := darc.NewSignerEd25519(nil, nil)
	ctx, darcUser,err := spawnDarc(darcAdmin, user.Identity().String(), "User")
	require.Nil(t,err)
	_, err = s.signAndSendTransaction(ctx, admin, darcAdmin, byzcoin.NewInstanceID(darcUser.GetBaseID()).Slice())
	require.Nil(t,err)

	//creating reader darc with rules initialized with the user darc
	ctx, darcReader,err := spawnDarc(darcAdmin, darcUser.GetIdentityString(), "Reader")
	require.Nil(t,err)
	_, err = s.signAndSendTransaction(ctx, admin, darcAdmin, byzcoin.NewInstanceID(darcReader.GetBaseID()).Slice())
	require.Nil(t,err)

	//creating garage darc with rules initialized with the user darc
	ctx, darcGarage,err := spawnDarc(darcAdmin, darcUser.GetIdentityString(), "Garage")
	require.Nil(t,err)
	_, err = s.signAndSendTransaction(ctx, admin, darcAdmin, byzcoin.NewInstanceID(darcGarage.GetBaseID()).Slice())
	require.Nil(t,err)

	//create car darc
	ctx, darcCar,err := spawnCarDarc(darcAdmin, darcReader, darcGarage)
	require.Nil(t,err)
	_, err = s.signAndSendTransaction(ctx, admin, darcAdmin, byzcoin.NewInstanceID(darcCar.GetBaseID()).Slice())
	require.Nil(t,err)

	//evolve the reader darc by adding a new reader
	newReader := darc.NewSignerEd25519(nil, nil)
	evolved_darc, err := s.addSigner(darcReader, newReader, user)
	require.Nil(t,err)

	//add a new reader
	newReader2 := darc.NewSignerEd25519(nil, nil)
	evolved_darc2, err := s.addSigner(evolved_darc, newReader2, user)
	require.Nil(t,err)

	//remove the second reader
	_, err = s.removeSigner(evolved_darc2, newReader2, user)
	require.Nil(t,err)

	//evolve the garage darc by adding a new garage person
	newGarage := darc.NewSignerEd25519(nil, nil)
	_, err = s.addSigner(darcGarage, newGarage, user)
	require.Nil(t,err)

	//create a car object and then spawn a car instance
	car := NewCar("123A2314")
	cInstance, err := s.createCarInstance(car,
		darcCar, admin)
	require.Nil(t, err)

	//check if the car instance is on the blockchain
	_, err = s.cl.GetProof(cInstance.Slice())
	require.Nil(t, err)

	//adding report for the car instance
	var wData SecretData
	wData.ECOScore = "2310"
	wData.Mileage = "100 000"
	wData.Warranty = true

	s.addReport(cInstance,
		darcCar, wData, newGarage, user)

	//reading the reports for the car instance
	secrets, err := s.readReports(cInstance, darcCar, newReader, user)
	require.Nil(t, err)

	require.Equal(t, secrets[0].Mileage, wData.Mileage)
}
