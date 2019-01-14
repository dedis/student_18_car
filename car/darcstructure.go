package car

import (
	"bytes"
	"errors"
	"github.com/dedis/kyber/suites"
	"strings"
	"testing"
	"time"

	"github.com/dedis/cothority/byzcoin"
	"github.com/dedis/cothority/darc"
	"github.com/dedis/cothority/darc/expression"
)

var testInterval = 500 * time.Millisecond
var tSuite = suites.MustFind("Ed25519")

// Spawn an Admin Darc from the Genesis Darc, giving the admin signer as input
//and returning the new darc as output
func (s *ser) spawnAdminDarc(admin darc.Signer)  (*darc.Darc, error){
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

	//creating a transaction with spawn:darc instruction
	ctx := newSpawnDarcTransaction(s.gDarc, darcAdminBuf)

	//Signing and sending a transaction to ByzCoin and waiting for it to be included in the ledger
	_, err = s.signAndSendTransaction(ctx, s.signer, s.gDarc, byzcoin.NewInstanceID(darcAdmin.GetBaseID()).Slice())
	if err != nil {
		return nil, err
	}

	return darcAdmin, err
}

// Spawn an User(owner) Darc from the Admin Darc(input), giving the Admin Signer and User(owner) Signer as input as well
//and returning the new darc
func (s *ser) spawnUserDarc(darcAdmin *darc.Darc, admin darc.Signer, user darc.Signer) (*darc.Darc, error){

	// Spawn User darc from the Admin one, but sign the request with
	// the signer of the first darc to test delegation
	idUser := []darc.Identity{user.Identity()}
	darcUser := darc.NewDarc(darc.InitRules(idUser, idUser),
		[]byte("User darc"))
	darcUser.Rules.AddRule("invoke:evolve", expression.InitOrExpr(user.Identity().String()))
	darcUserBuf, err := darcUser.ToProto()
	if err != nil {
		return nil, err
	}

	ctx := newSpawnDarcTransaction(darcAdmin, darcUserBuf)

	//Signing and sending a transaction to ByzCoin and waiting for it to be included in the ledger
	_, err = s.signAndSendTransaction(ctx, admin, darcAdmin, byzcoin.NewInstanceID(darcUser.GetBaseID()).Slice())
	if err != nil {
		return nil, err
	}

	return darcUser, err
}

// Spawn a Reader Darc from the Admin Darc(input), giving the service and Admin Signer as input as well
//and returning the new darc together with the user signer
func (s *ser) spawnReaderDarc( darcAdmin *darc.Darc,
	admin darc.Signer, userDarc *darc.Darc) (*darc.Darc, error){

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
	if err != nil {
		return nil, err
	}

	ctx := newSpawnDarcTransaction(darcAdmin, darcReaderBuf)

	_, err = s.signAndSendTransaction(ctx, admin, darcAdmin, byzcoin.NewInstanceID(darcReader.GetBaseID()).Slice())
	if err != nil {
		return nil, err
	}

	return darcReader, err
}

// Spawn a Reader Darc from the Admin Darc(input), giving the service and Admin Signer as input as well
//and returning the new darc together with the user signer
func (s *ser) spawnGarageDarc( darcAdmin *darc.Darc,
	admin darc.Signer, userDarc *darc.Darc) (*darc.Darc, error) {

	// Spawn Reader darc from the Admin one, but sign the request with
	// the signer of the first darc to test delegation

	//rules for the new Garage Darc
	rs := darc.NewRules()
	if err := rs.AddRule("invoke:evolve", expression.InitAndExpr(userDarc.GetIdentityString())); err != nil {
		panic("add rule should never fail on an empty rule list: " + err.Error())
	}
	if err := rs.AddRule("_sign", expression.InitAndExpr(userDarc.GetIdentityString())); err != nil {
		panic("add rule should never fail on an empty rule list: " + err.Error())
	}

	darcGarage := darc.NewDarc(rs,
		[]byte("Garage darc"))
	darcGarageBuf, err := darcGarage.ToProto()
	if err != nil {
		return nil, err
	}

	ctx := newSpawnDarcTransaction(darcAdmin, darcGarageBuf)

	_, err = s.signAndSendTransaction(ctx, admin, darcAdmin, byzcoin.NewInstanceID(darcGarage.GetBaseID()).Slice())
	if err != nil {
		return nil, err
	}

	return darcGarage, err
}

func (s *ser) spawnCarDarc( darcAdmin *darc.Darc, admin darc.Signer,
	darcReader *darc.Darc, darcGarage *darc.Darc) (*darc.Darc, error) {

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
	if err := rs.AddRule("spawn:calypsoWrite", expression.InitAndExpr(darcGarage.GetIdentityString())); err != nil {
		panic("add rule should never fail on an empty rule list: " + err.Error())
	}

	darcCar := darc.NewDarc(rs,
		[]byte("Car darc"))
	darcCarBuf, err := darcCar.ToProto()
	if err != nil {
		return nil, err
	}

	ctx := newSpawnDarcTransaction(darcAdmin, darcCarBuf)

	_, err = s.signAndSendTransaction(ctx, admin, darcAdmin, byzcoin.NewInstanceID(darcCar.GetBaseID()).Slice())

	if err != nil {
		return nil, err
	}

	return darcCar, err
}

func (s *ser) removeSigner(t *testing.T, d *darc.Darc,
	signerToBeRemoved darc.Signer, signer darc.Signer) (*darc.Darc, error){

	if !(bytes.Equal(d.Description, []byte("Reader darc")) || bytes.Equal(d.Description, []byte("Garage darc"))) {
		return nil, errors.New("Signer can be removed only from a Reader or Garage Darc")
	}

	d2 := d.Copy()
	err := d2.EvolveFrom(d)
	if err != nil {
		return nil, err
	}
	//updating the sign expression
	oldExp := d.Rules.GetSignExpr()
	exp, err := removeSignerFromExpr(t, oldExp, signerToBeRemoved)
	if err != nil {
		return nil, err
	}
	d2.Rules.UpdateSign([]byte(exp))

	pr,err := s.evolveDarc(d2, signer)
	if err != nil {
		return nil, err
	}

	if pr.InclusionProof.Match(d2.GetBaseID()) != true {
		return nil, errors.New("absence of the key in the collection")
	}
	_, vs, _, _, err := pr.KeyValue()
	if err != nil {
		return nil, err
	}
	d22, err := darc.NewFromProtobuf(vs)
	if err != nil {
		return nil, err
	}
	if d22.Equal(d2) != true {
		return nil, errors.New("evolved and original darc don't point to the same data")
	}

	return d2, err
}


func removeSignerFromExpr(t *testing.T, oldExp expression.Expr, removedMember darc.Signer) (expression.Expr, error){

	exp := ""
	var err error

	//if the member to be removed doesn't exist in the expression, nothing needs to be done (old expr returned)
	if !(strings.Contains(string(oldExp), removedMember.Identity().String())) {
		return oldExp, err
	}

	//otherwise split the expression into members connected with OR (Pub_m1 | Pub_m2 | ...)
	//and the Darc of the user that owns the Car
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

	if dPart == "" || otherPart == ""{
		return nil, errors.New("invalid expression")
	}

	//if there aren't brackets it means there should be only the darc identity after the removal
	if !(strings.Contains(otherPart, "(")){
		exp = removeSpace(dPart)
		return []byte(exp), err
	} else {
		temp_before := " | " + removedMember.Identity().String()
		temp_after := removedMember.Identity().String() + " | "


		if (strings.Contains(otherPart, temp_before)) {
			otherPart := strings.Replace(otherPart, temp_before, "",-1)
			if !(strings.Contains(otherPart, "|")) {
				otherPart = strings.Trim(otherPart,"(")
				otherPart = strings.Trim(otherPart,")")
			}
			exp = removeSpace(otherPart) + " & " + removeSpace(dPart)
			return []byte(exp), err
		}

		if (strings.Contains(otherPart, temp_after)) {
			otherPart := strings.Replace(otherPart, temp_after, "",-1)
			t.Log("here")
			t.Log(otherPart)
			if !(strings.Contains(otherPart, "|")) {
				otherPart = strings.Trim(otherPart,"(")
				otherPart = strings.Trim(otherPart,")")
			}
			exp = removeSpace(otherPart) + " & " + removeSpace(dPart)
			return []byte(exp), err
		}

	}
	exp = removeSpace(otherPart) + " & " + removeSpace(dPart)
	return []byte(exp), err
}


func (s *ser) addSigner(d *darc.Darc,
	newSigner darc.Signer, signer darc.Signer) (*darc.Darc, error){

	if !(bytes.Equal(d.Description, []byte("Reader darc")) || bytes.Equal(d.Description, []byte("Garage darc"))) {
		return nil, errors.New("Signer can be added only for a Reader or Garage Darc")
	}

	d2 := d.Copy()
	err := d2.EvolveFrom(d)
	if err != nil {
		return nil, err
	}
	//updating the sign expression
	oldExp := d.Rules.GetSignExpr()
	exp, err := addSignerToExpr(oldExp, newSigner)
	if err != nil {
		return nil, err
	}
	d2.Rules.UpdateSign([]byte(exp))

	pr,err := s.evolveDarc(d2, signer)
	if err != nil {
		return nil, err
	}

	if pr.InclusionProof.Match(d2.GetBaseID()) != true {
		return nil, errors.New("absence of the key in the collection")
	}
	_, vs, _, _, err := pr.KeyValue()
	if err != nil {
		return nil, err
	}
	d22, err := darc.NewFromProtobuf(vs)
	if err != nil {
		return nil, err
	}
	if d22.Equal(d2) != true {
		return nil, errors.New("evolved and original darc don't point to the same data")
	}

	return d2, err
}

// Creating the Expression When a new Member(Reader/Garage) is added
// to have the format
// (Pub_m1 | Pub_m2 | ...) & DarcUser

func addSignerToExpr(oldExp expression.Expr, newMember darc.Signer) (expression.Expr, error){

	exp := ""
	var err error
	//if there is only one Identity in the expression
	if !(strings.Contains(string(oldExp), "&")){
		exp = string(oldExp) + " & " + newMember.Identity().String()
		return []byte(exp), err
	}

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

	if dPart == "" || otherPart == ""{
		return nil, errors.New("invalid expression")
	}

	exp = removeSpace(otherPart) + " & " + removeSpace(dPart)
	return []byte(exp), err

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

func (s *ser) evolveDarc(d2 *darc.Darc, signer darc.Signer) (*byzcoin.Proof, error) {
	d2Buf, err := d2.ToProto()
	if err != nil {
		return nil, err
	}
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
	ctx := byzcoin.ClientTransaction{
		Instructions: []byzcoin.Instruction{instr}}

	pr,err := s.signAndSendTransaction(ctx, signer, d2, d2.GetBaseID())
	if err != nil {
		return nil, err
	}
	return pr,err
}

//returns a transaction with spawn:darc instruction
func newSpawnDarcTransaction(controlDarc *darc.Darc, newDarcBuf []byte) byzcoin.ClientTransaction{

	ctx := byzcoin.ClientTransaction{
		Instructions: []byzcoin.Instruction{{
			InstanceID: byzcoin.NewInstanceID(controlDarc.GetBaseID()),
			Nonce:      byzcoin.GenNonce(),
			Index:      0,
			Length:     1,
			Spawn: &byzcoin.Spawn{
				ContractID: byzcoin.ContractDarcID,
				Args: []byzcoin.Argument{{
					Name:  "darc",
					Value: newDarcBuf,
				}},
			},
		}},
	}
	return ctx
}

//Signing and sending a transaction to ByzCoin and waiting for it to be included in the ledger
func (s *ser)signAndSendTransaction(ctx byzcoin.ClientTransaction, txnSigner darc.Signer,
	controlDarc *darc.Darc, instanceKey []byte) (*byzcoin.Proof, error) {

	// Sign the instruction with the signer that has his
	// public key stored in the darc.
	err := ctx.Instructions[0].SignBy(controlDarc.GetBaseID(), txnSigner)
	if err != nil {
		return nil, err
	}

	// Sending this transaction to ByzCoin and waiting for it to be included
	// in the ledger, up to a maximum of 5 block intervals
	_, err = s.cl.AddTransactionAndWait(ctx, 5)
	if err != nil {
		return nil, err
	}

	//GetProof returns a proof for the key stored in the skipchain
	resp, err := s.cl.GetProof(instanceKey)
	if err != nil {
		return nil, err
	}

	return &resp.Proof, err
}