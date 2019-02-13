package car

import (
	"bytes"
	"errors"
	"github.com/dedis/kyber/suites"
	"strings"
	"time"

	"github.com/dedis/cothority/byzcoin"
	"github.com/dedis/cothority/darc"
	"github.com/dedis/cothority/darc/expression"
)

var testInterval = 500 * time.Millisecond
var tSuite = suites.MustFind("Ed25519")

// Spawn a new Admin Darc with a rule to spawn other darcs
func spawnAdminDarc(controlDarc *darc.Darc, user darc.Signer) (byzcoin.ClientTransaction, *darc.Darc, error){

	var ctx byzcoin.ClientTransaction
	idUser := []darc.Identity{user.Identity()}
	newDarc := darc.NewDarc(darc.InitRules(idUser, idUser),
		[]byte("Admin darc"))
	newDarc.Rules.AddRule("spawn:darc", expression.InitOrExpr(controlDarc.GetIdentityString(), user.Identity().String()))
	darcUserBuf, err := newDarc.ToProto()
	if err != nil {
		return ctx, nil, err
	}
	ctx = newSpawnDarcTransaction(controlDarc, darcUserBuf)

	return ctx, newDarc, err
}

// Spawn a new Darc from an existing control Darc(input)
func spawnDarc(controlDarc *darc.Darc, idString string, role string) (byzcoin.ClientTransaction, *darc.Darc, error){

	var ctx byzcoin.ClientTransaction
	//rules for the new Reader Darc
	rs := darc.NewRules()
	if err := rs.AddRule("invoke:evolve", expression.InitAndExpr(idString)); err != nil {
		panic("add rule should never fail on an empty rule list: " + err.Error())
	}
	if err := rs.AddRule("_sign", expression.InitAndExpr(idString)); err != nil {
		panic("add rule should never fail on an empty rule list: " + err.Error())
	}
	newDarc := darc.NewDarc(rs,
		[]byte(role + " darc"))
	newDarcBuf, err := newDarc.ToProto()
	if err != nil {
		return ctx, nil, err
	}
	ctx = newSpawnDarcTransaction(controlDarc, newDarcBuf)

	return ctx, newDarc, err
}


func spawnCarDarc( darcAdmin *darc.Darc, darcReader *darc.Darc,
	darcGarage *darc.Darc) (byzcoin.ClientTransaction, *darc.Darc, error) {

	var ctx byzcoin.ClientTransaction
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
		return ctx, nil, err
	}
	ctx = newSpawnDarcTransaction(darcAdmin, darcCarBuf)

	return ctx, darcCar, err
}

/*
Adding and removing Members for garage and reader darcs
assume the following rule:

(Pub_m1 | Pub_m2 | ...) & DarcUser

where m1 is member1, ...
 */

func (s *ser) removeSigner(d *darc.Darc,
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
	exp, err := removeSignerFromExpr(oldExp, signerToBeRemoved)
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


func removeSignerFromExpr(oldExp expression.Expr, removedMember darc.Signer) (expression.Expr, error){

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