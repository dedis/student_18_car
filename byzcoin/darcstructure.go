package byzcoin

import (
	"github.com/dedis/kyber/suites"
	"testing"
	"time"

	"github.com/dedis/cothority/byzcoin"
	"github.com/dedis/cothority/darc"
	"github.com/dedis/cothority/darc/expression"
	"github.com/stretchr/testify/require"
)

var testInterval = 500 * time.Millisecond
var tSuite = suites.MustFind("Ed25519")

// Spawn an Admin Darc from the Genesis Darc, giving the service as input
//and returning the new darc together with the admin signer
func createAdminDarc(t *testing.T, s *ser) (darc.Signer, *darc.Darc){
	// Spawn Admin darc with a new owner/signer, but delegate its spawn
	// rule to the first darc or the new owner/signer
	admin := darc.NewSignerEd25519(nil, nil)
	idAdmin := []darc.Identity{admin.Identity()}
	darcAdmin := darc.NewDarc(darc.InitRules(idAdmin, idAdmin),
		[]byte("Admin darc"))
	darcAdmin.Rules.AddRule("spawn:darc", expression.InitOrExpr(s.gDarc.GetIdentityString(), admin.Identity().String()))
	darcAdmin.Rules.AddRule("invoke:evolve", expression.InitOrExpr(s.gDarc.GetIdentityString(), admin.Identity().String()))
	darcAdminBuf, err := darcAdmin.ToProto()
	require.Nil(t, err)
	darcAdminCopy, err := darc.NewFromProtobuf(darcAdminBuf)
	require.Nil(t, err)
	require.True(t, darcAdmin.Equal(darcAdminCopy))
	ctx := byzcoin.ClientTransaction{
		Instructions: []byzcoin.Instruction{{
			InstanceID: byzcoin.NewInstanceID(s.gDarc.GetBaseID()),
			Nonce:      byzcoin.GenNonce(),
			Index:      0,
			Length:     1,
			Spawn: &byzcoin.Spawn{
				ContractID: byzcoin.ContractDarcID,
				Args: []byzcoin.Argument{{
					Name:  "darc",
					Value: darcAdminBuf,
				}},
			},
		}},
	}
	require.Nil(t, ctx.Instructions[0].SignBy(s.gDarc.GetBaseID(), s.signer))
	s.sendTx(t, ctx)
	pr := s.waitProof(t, byzcoin.NewInstanceID(darcAdmin.GetBaseID()))
	require.True(t, pr.InclusionProof.Match())
	return admin,darcAdmin
}

// Spawn an User Darc from the Admin Darc(input), giving the service and Admin Signer as input as well
//and returning the new darc together with the user signer
func createUserDarc(t *testing.T, s *ser, darcAdmin *darc.Darc, admin darc.Signer) (darc.Signer, *darc.Darc){

	// Spawn User darc from the Admin one, but sign the request with
	// the signer of the first darc to test delegation
	user := darc.NewSignerEd25519(nil, nil)
	idUser := []darc.Identity{user.Identity()}
	darcUser := darc.NewDarc(darc.InitRules(idUser, idUser),
		[]byte("User darc"))
	darcUser.Rules.AddRule("invoke:evolve", expression.InitOrExpr(user.Identity().String()))
	darcUserBuf, err := darcUser.ToProto()
	require.Nil(t, err)
	darcUserCopy, err := darc.NewFromProtobuf(darcUserBuf)
	require.Nil(t, err)
	require.True(t, darcUser.Equal(darcUserCopy))
	ctx := byzcoin.ClientTransaction{
		Instructions: []byzcoin.Instruction{{
			InstanceID: byzcoin.NewInstanceID(darcAdmin.GetBaseID()),
			Nonce:      byzcoin.GenNonce(),
			Index:      0,
			Length:     1,
			Spawn: &byzcoin.Spawn{
				ContractID: byzcoin.ContractDarcID,
				Args: []byzcoin.Argument{{
					Name:  "darc",
					Value: darcUserBuf,
				}},
			},
		}},
	}
	require.Nil(t, ctx.Instructions[0].SignBy(darcAdmin.GetBaseID(), admin))
	s.sendTx(t, ctx)
	pr := s.waitProof(t, byzcoin.NewInstanceID(darcUser.GetBaseID()))
	require.True(t, pr.InclusionProof.Match())
	return user, darcUser
}

// Spawn a Reader Darc from the Admin Darc(input), giving the service and Admin Signer as input as well
//and returning the new darc together with the user signer
func createReaderDarc(t *testing.T, s *ser, darcAdmin *darc.Darc, admin darc.Signer, userDarc *darc.Darc) (darc.Signer, *darc.Darc){

	// Spawn Reader darc from the Admin one, but sign the request with
	// the signer of the first darc to test delegation
	reader := darc.NewSignerEd25519(nil, nil)

	//rules for the new Reader Darc
	rs := darc.NewRules()
	if err := rs.AddRule("invoke:evolve", expression.InitAndExpr(userDarc.GetIdentityString())); err != nil {
		panic("add rule should never fail on an empty rule list: " + err.Error())
	}
	if err := rs.AddRule("_sign", expression.InitAndExpr(reader.Identity().String(),userDarc.GetIdentityString())); err != nil {
		panic("add rule should never fail on an empty rule list: " + err.Error())
	}

	darcReader := darc.NewDarc(rs,
		[]byte("Reader darc"))
	darcReaderBuf, err := darcReader.ToProto()
	require.Nil(t, err)
	darcReaderCopy, err := darc.NewFromProtobuf(darcReaderBuf)
	require.Nil(t, err)
	require.True(t, darcReader.Equal(darcReaderCopy))
	ctx := byzcoin.ClientTransaction{
		Instructions: []byzcoin.Instruction{{
			InstanceID: byzcoin.NewInstanceID(darcAdmin.GetBaseID()),
			Nonce:      byzcoin.GenNonce(),
			Index:      0,
			Length:     1,
			Spawn: &byzcoin.Spawn{
				ContractID: byzcoin.ContractDarcID,
				Args: []byzcoin.Argument{{
					Name:  "darc",
					Value: darcReaderBuf,
				}},
			},
		}},
	}
	require.Nil(t, ctx.Instructions[0].SignBy(darcAdmin.GetBaseID(), admin))
	s.sendTx(t, ctx)
	pr := s.waitProof(t, byzcoin.NewInstanceID(darcReader.GetBaseID()))
	require.True(t, pr.InclusionProof.Match())
	return reader, darcReader
}

// Spawn a Reader Darc from the Admin Darc(input), giving the service and Admin Signer as input as well
//and returning the new darc together with the user signer
func createGarageDarc(t *testing.T, s *ser, darcAdmin *darc.Darc, admin darc.Signer, userDarc *darc.Darc) (darc.Signer, *darc.Darc) {

	// Spawn Reader darc from the Admin one, but sign the request with
	// the signer of the first darc to test delegation
	garage := darc.NewSignerEd25519(nil, nil)

	//rules for the new Garage Darc
	rs := darc.NewRules()
	if err := rs.AddRule("invoke:evolve", expression.InitAndExpr(userDarc.GetIdentityString())); err != nil {
		panic("add rule should never fail on an empty rule list: " + err.Error())
	}
	if err := rs.AddRule("_sign", expression.InitAndExpr(garage.Identity().String(), userDarc.GetIdentityString())); err != nil {
		panic("add rule should never fail on an empty rule list: " + err.Error())
	}

	darcGarage := darc.NewDarc(rs,
		[]byte("Garage darc"))
	darcGarageBuf, err := darcGarage.ToProto()
	require.Nil(t, err)
	darcGarageCopy, err := darc.NewFromProtobuf(darcGarageBuf)
	require.Nil(t, err)
	require.True(t, darcGarage.Equal(darcGarageCopy))
	ctx := byzcoin.ClientTransaction{
		Instructions: []byzcoin.Instruction{{
			InstanceID: byzcoin.NewInstanceID(darcAdmin.GetBaseID()),
			Nonce:      byzcoin.GenNonce(),
			Index:      0,
			Length:     1,
			Spawn: &byzcoin.Spawn{
				ContractID: byzcoin.ContractDarcID,
				Args: []byzcoin.Argument{{
					Name:  "darc",
					Value: darcGarageBuf,
				}},
			},
		}},
	}
	require.Nil(t, ctx.Instructions[0].SignBy(darcAdmin.GetBaseID(), admin))
	s.sendTx(t, ctx)
	pr := s.waitProof(t, byzcoin.NewInstanceID(darcGarage.GetBaseID()))
	require.True(t, pr.InclusionProof.Match())
	return garage, darcGarage
}

func createCarDarc(t *testing.T, s *ser, darcAdmin *darc.Darc,
	admin darc.Signer, darcReader *darc.Darc, darcGarage *darc.Darc) (*darc.Darc) {

	// Spawn Car darc from the Admin one

	//rules for the new Car Darc
	rs := darc.NewRules()
	if err := rs.AddRule("spawn:car", expression.InitAndExpr(darcAdmin.GetIdentityString())); err != nil {
		panic("add rule should never fail on an empty rule list: " + err.Error())
	}
	if err := rs.AddRule("spawn:calypsoRead", expression.InitAndExpr(darcReader.GetIdentityString())); err != nil {
		panic("add rule should never fail on an empty rule list: " + err.Error())
	}
	if err := rs.AddRule("invoke:addreport", expression.InitAndExpr(darcGarage.GetIdentityString())); err != nil {
		panic("add rule should never fail on an empty rule list: " + err.Error())
	}

	darcCar := darc.NewDarc(rs,
		[]byte("Car darc"))
	darcCarBuf, err := darcCar.ToProto()
	require.Nil(t, err)
	darcCarCopy, err := darc.NewFromProtobuf(darcCarBuf)
	require.Nil(t, err)
	require.True(t, darcCar.Equal(darcCarCopy))
	ctx := byzcoin.ClientTransaction{
		Instructions: []byzcoin.Instruction{{
			InstanceID: byzcoin.NewInstanceID(darcAdmin.GetBaseID()),
			Nonce:      byzcoin.GenNonce(),
			Index:      0,
			Length:     1,
			Spawn: &byzcoin.Spawn{
				ContractID: byzcoin.ContractDarcID,
				Args: []byzcoin.Argument{{
					Name:  "darc",
					Value: darcCarBuf,
				}},
			},
		}},
	}
	require.Nil(t, ctx.Instructions[0].SignBy(darcAdmin.GetBaseID(), admin))
	s.sendTx(t, ctx)
	pr := s.waitProof(t, byzcoin.NewInstanceID(darcCar.GetBaseID()))
	require.True(t, pr.InclusionProof.Match())
	return darcCar
}



func (s *ser) createCarInstance(t *testing.T, args byzcoin.Arguments,
	d *darc.Darc, signer darc.Signer) byzcoin.InstanceID {
	ctx := byzcoin.ClientTransaction{
		Instructions: []byzcoin.Instruction{{
			InstanceID: byzcoin.NewInstanceID(d.GetBaseID()),
			Nonce:      byzcoin.Nonce{},
			Index:      0,
			Length:     1,
			Spawn: &byzcoin.Spawn{
				ContractID: ContractCarID,
				Args:       args,
			},
		}},
	}

	// And we need to sign the instruction with the signer that has his
	// public key stored in the darc.
	require.Nil(t, ctx.Instructions[0].SignBy(d.GetBaseID(), signer))

	// Sending this transaction to ByzCoin does not directly include it in the
	// global state - first we must wait for the new block to be created.
	s.sendTx(t, ctx)

	//wait for the proof for instance ID
	pr := s.waitProof(t, ctx.Instructions[0].InstanceID)
	require.True(t, pr.InclusionProof.Match())

	return ctx.Instructions[0].DeriveID("")
}


