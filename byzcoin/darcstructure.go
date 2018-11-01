package byzcoin

import (
	"github.com/dedis/kyber/suites"
	"strings"
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
func (s *ser) createAdminDarc(admin darc.Signer)  (*darc.Darc, error){
	// Spawn Admin darc with a new owner/signer, but delegate its spawn
	// rule to the first darc or the new owner/signer
	var err error
	idAdmin := []darc.Identity{admin.Identity()}
	darcAdmin := darc.NewDarc(darc.InitRules(idAdmin, idAdmin),
		[]byte("Admin darc"))
	darcAdmin.Rules.AddRule("spawn:darc", expression.InitOrExpr(s.gDarc.GetIdentityString(), admin.Identity().String()))
	darcAdmin.Rules.AddRule("invoke:evolve", expression.InitOrExpr(s.gDarc.GetIdentityString(), admin.Identity().String()))
	darcAdminBuf, err := darcAdmin.ToProto()
	if err != nil {
		return nil, err
	}
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
	err = ctx.Instructions[0].SignBy(s.gDarc.GetBaseID(), s.signer)
	if err != nil {
		return nil, err
	}
	_, err = s.cl.AddTransactionAndWait(ctx, 5)
	if err != nil {
		return nil, err
	}

	_, err = s.cl.GetProof(byzcoin.NewInstanceID(darcAdmin.GetBaseID()).Slice())
	if err != nil {
		return nil, err
	}

	return darcAdmin, err
}

// Spawn an User Darc from the Admin Darc(input), giving the service and Admin Signer as input as well
//and returning the new darc together with the user signer
func (s *ser) createUserDarc(t *testing.T, darcAdmin *darc.Darc, admin darc.Signer) (darc.Signer, *darc.Darc){

	// Spawn User darc from the Admin one, but sign the request with
	// the signer of the first darc to test delegation
	user := darc.NewSignerEd25519(nil, nil)
	idUser := []darc.Identity{user.Identity()}
	darcUser := darc.NewDarc(darc.InitRules(idUser, idUser),
		[]byte("User darc"))
	darcUser.Rules.AddRule("invoke:evolve", expression.InitOrExpr(user.Identity().String()))
	darcUserBuf, err := darcUser.ToProto()
	require.Nil(t, err)
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

	_, err = s.cl.AddTransactionAndWait(ctx, 5)
	require.Nil(t, err)

	_, err = s.cl.GetProof(byzcoin.NewInstanceID(darcUser.GetBaseID()).Slice())
	require.Nil(t, err)

	//s.sendTx(t, ctx)
	//pr := s.waitProof(t, byzcoin.NewInstanceID(darcUser.GetBaseID()))
	//require.True(t, pr.InclusionProof.Match())
	return user, darcUser
}

// Spawn a Reader Darc from the Admin Darc(input), giving the service and Admin Signer as input as well
//and returning the new darc together with the user signer
func (s *ser) createReaderDarc(t *testing.T, darcAdmin *darc.Darc, admin darc.Signer, userDarc *darc.Darc) (*darc.Darc){

	// Spawn Reader darc from the Admin one, but sign the request with
	// the signer of the first darc to test delegation

	//rules for the new Reader Darc
	rs := darc.NewRules()
	if err := rs.AddRule("invoke:evolve", expression.InitAndExpr(userDarc.GetIdentityString())); err != nil {
		panic("add rule should never fail on an empty rule list: " + err.Error())
	}
	if err := rs.AddRule("_sign", expression.InitAndExpr(userDarc.GetIdentityString())); err != nil {
		panic("add rule should never fail on an empty rule list: " + err.Error())
	}

	darcReader := darc.NewDarc(rs,
		[]byte("Reader darc"))
	darcReaderBuf, err := darcReader.ToProto()
	require.Nil(t, err)
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

	_, err = s.cl.AddTransactionAndWait(ctx, 5)
	require.Nil(t, err)

	_, err = s.cl.GetProof(byzcoin.NewInstanceID(darcReader.GetBaseID()).Slice())
	require.Nil(t, err)


	//s.sendTx(t, ctx)
	//pr := s.waitProof(t, byzcoin.NewInstanceID(darcReader.GetBaseID()))
	//require.True(t, pr.InclusionProof.Match())
	return darcReader
}

// Spawn a Reader Darc from the Admin Darc(input), giving the service and Admin Signer as input as well
//and returning the new darc together with the user signer
func (s *ser) createGarageDarc(t *testing.T, darcAdmin *darc.Darc,
	admin darc.Signer, userDarc *darc.Darc) (*darc.Darc) {

	// Spawn Reader darc from the Admin one, but sign the request with
	// the signer of the first darc to test delegation

	//rules for the new Garage Darc
	rs := darc.NewRules()
	if err := rs.AddRule("invoke:evolve", expression.InitAndExpr(userDarc.GetIdentityString())); err != nil {
		panic("add rule should never fail on an empty rule list: " + err.Error())
	}
	//if err := rs.AddRule("_sign", expression.InitAndExpr(garage.Identity().String())); err != nil {
	//	panic("add rule should never fail on an empty rule list: " + err.Error())

	if err := rs.AddRule("_sign", expression.InitAndExpr(userDarc.GetIdentityString())); err != nil {
		panic("add rule should never fail on an empty rule list: " + err.Error())
	}

	darcGarage := darc.NewDarc(rs,
		[]byte("Garage darc"))
	darcGarageBuf, err := darcGarage.ToProto()
	require.Nil(t, err)

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

	_, err = s.cl.AddTransactionAndWait(ctx, 5)
	require.Nil(t, err)

	_, err = s.cl.GetProof(byzcoin.NewInstanceID(darcGarage.GetBaseID()).Slice())
	require.Nil(t, err)




	//s.sendTx(t, ctx)
	//pr := s.waitProof(t, byzcoin.NewInstanceID(darcGarage.GetBaseID()))
	//require.True(t, pr.InclusionProof.Match())
	return darcGarage
}

func (s *ser) createCarDarc(t *testing.T, darcAdmin *darc.Darc,
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
	if err := rs.AddRule("invoke:addReport", expression.InitAndExpr(darcGarage.GetIdentityString())); err != nil {
		panic("add rule should never fail on an empty rule list: " + err.Error())
	}

	darcCar := darc.NewDarc(rs,
		[]byte("Car darc"))
	darcCarBuf, err := darcCar.ToProto()
	require.Nil(t, err)

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

	_, err = s.cl.AddTransactionAndWait(ctx, 5)
	require.Nil(t, err)

	_, err = s.cl.GetProof(byzcoin.NewInstanceID(darcCar.GetBaseID()).Slice())
	require.Nil(t, err)


	//s.sendTx(t, ctx)
	//pr := s.waitProof(t, byzcoin.NewInstanceID(darcCar.GetBaseID()))
	//require.True(t, pr.InclusionProof.Match())
	return darcCar
}

func (s *ser) addSigner(t *testing.T, d *darc.Darc,
	newReader darc.Signer, signer darc.Signer) (*darc.Darc){

	d2 := d.Copy()
	require.Nil(t, d2.EvolveFrom(d))
	oldExp := d.Rules.GetSignExpr()

	exp:= addSignerToExpr(oldExp, newReader)

	//TODO check if the signer is allowed to update

	d2.Rules.UpdateSign([]byte(exp))
	pr := s.testDarcEvolution(t, *d2, false, signer)

	require.True(t, pr.InclusionProof.Match())
	_, vs, err := pr.KeyValue()
	require.Nil(t, err)
	d22, err := darc.NewFromProtobuf(vs[0])
	require.Nil(t, err)
	require.True(t, d22.Equal(d2))
	return d2
}

// Creating the Expression When a new Member(Reader/Garage) is added
// to have the format
// (Pub_m1 | Pub_m2 | ...) & DarcUser

func addSignerToExpr(oldExp expression.Expr, newMember darc.Signer) (expression.Expr){

	exp := ""

	//if there is only one Identity in the expression
	if !(strings.Contains(string(oldExp), "&")){
		exp = string(oldExp) + " & " + newMember.Identity().String()
		return []byte(exp)
	}

	//TODO check if there are more than one & in the expression
	//otherwise split the expression into members connected with OR (Pub_m1 | Pub_m2 | ...)
	//and the Darc of the user that owns the Car and then add a new  Pub_m_n
	dPart := ""
	otherPart := ""
	oldExpStr := strings.Split(string(oldExp), "&")

	//checking which part of the expression contains the darc Identity
	if strings.Contains(oldExpStr[0], "darc"){
		dPart = removeSpace(oldExpStr[0])
		otherPart = removeSpace(oldExpStr[1])
	}
	if strings.Contains(oldExpStr[1], "darc"){
		dPart = removeSpace(oldExpStr[1])
		otherPart = removeSpace(oldExpStr[0])
	}

	if strings.Contains(otherPart, "("){
		temp:= strings.Split(otherPart, ")")
		otherPart = removeSpace(temp[0]) + " | " + newMember.Identity().String() + ")"
	}
	if !(strings.Contains(otherPart, "(")){
		otherPart = "(" + otherPart + " | " + newMember.Identity().String() + ")"

	}
	//return the oldExpr if the new expression is not valid
	//TODO return an error instead
	if dPart == "" || otherPart == ""{
		return oldExp
	}

	exp = removeSpace(otherPart) + " & " + removeSpace(dPart)
	return []byte(exp)

}

func removeSpace(str string) (string) {
	if strings.HasSuffix(str, " ") {
		str = strings.TrimSuffix(str, " ")
	}
	if strings.HasPrefix(str, " ") {
		str = strings.TrimPrefix(str, " ")
	}
	return str
}

func darcToTx(t *testing.T, d2 darc.Darc, signer darc.Signer) byzcoin.ClientTransaction {
	d2Buf, err := d2.ToProto()
	require.Nil(t, err)
	invoke := byzcoin.Invoke{
		Command: "evolve",
		Args: []byzcoin.Argument{
			byzcoin.Argument{
				Name:  "darc",
				Value: d2Buf,
			},
		},
	}
	instr := byzcoin.Instruction{
		InstanceID: byzcoin.NewInstanceID(d2.GetBaseID()),
		Nonce:      byzcoin.GenNonce(),
		Index:      0,
		Length:     1,
		Invoke:     &invoke,
	}
	require.Nil(t, instr.SignBy(d2.GetBaseID(), signer))
	return byzcoin.ClientTransaction{
		Instructions: []byzcoin.Instruction{instr},
	}
}

func (s *ser) testDarcEvolution(t *testing.T, d2 darc.Darc, fail bool, signer darc.Signer) (pr *byzcoin.Proof) {
	ctx := darcToTx(t, d2, signer)

	_, err := s.cl.AddTransactionAndWait(ctx, 5)
	require.Nil(t, err)

	resp, err := s.cl.GetProof(d2.GetBaseID())
	require.Nil(t, err)

	pr = &resp.Proof

	return
}



