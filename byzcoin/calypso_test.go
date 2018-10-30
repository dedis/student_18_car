package byzcoin



/*type ts struct {
	local      *onet.LocalTest
	servers    []*onet.Server
	services   []*calypso.Service
	roster     *onet.Roster
	ltsReply   *calypso.CreateLTSReply
	signer     darc.Signer
	cl         *byzcoin.Client
	gbReply    *byzcoin.CreateGenesisBlockResponse
	genesisMsg *byzcoin.CreateGenesisBlock
	gDarc      *darc.Darc
}

func TestService_DecryptKey(t *testing.T) {
	s := newTS(t, 5)
	defer s.closeAll(t)

	key1 := []byte("secret key 1")
	prWr1 := s.addWriteAndWait(t, key1)
	prRe1 := s.addReadAndWait(t, prWr1)
	key2 := []byte("secret key 2")
	prWr2 := s.addWriteAndWait(t, key2)
	prRe2 := s.addReadAndWait(t, prWr2)

	_, err := s.services[0].DecryptKey(&calypso.DecryptKey{Read: *prRe1, Write: *prWr2})
	require.NotNil(t, err)
	_, err = s.services[0].DecryptKey(&calypso.DecryptKey{Read: *prRe2, Write: *prWr1})
	require.NotNil(t, err)

	dk1, err := s.services[0].DecryptKey(&calypso.DecryptKey{Read: *prRe1, Write: *prWr1})
	require.Nil(t, err)
	require.True(t, dk1.X.Equal(s.ltsReply.X))
	keyCopy1, err := calypso.DecodeKey(cothority.Suite, s.ltsReply.X, dk1.Cs, dk1.XhatEnc, s.signer.Ed25519.Secret)
	require.Nil(t, err)
	require.Equal(t, key1, keyCopy1)

	dk2, err := s.services[0].DecryptKey(&calypso.DecryptKey{Read: *prRe2, Write: *prWr2})
	require.Nil(t, err)
	require.True(t, dk2.X.Equal(s.ltsReply.X))
	keyCopy2, err := calypso.DecodeKey(cothority.Suite, s.ltsReply.X, dk2.Cs, dk2.XhatEnc, s.signer.Ed25519.Secret)
	require.Nil(t, err)
	require.Equal(t, key2, keyCopy2)
}

func newTS(t *testing.T, nodes int) ts {
	s := ts{}
	s.local = onet.NewLocalTestT(cothority.Suite, t)
	// Create the service
	s.servers, s.roster, _ = s.local.GenTree(nodes, true)
	services := s.local.GetServices(s.servers, onet.ServiceFactory.ServiceID(calypso.ServiceName))

	for _, ser := range services {
		s.services = append(s.services, ser.(*calypso.Service))
	}

	// Create the skipchain
	s.signer = darc.NewSignerEd25519(nil, nil)
	s.createGenesis(t)

	// Start DKG
	var err error
	s.ltsReply, err = s.services[0].CreateLTS(&calypso.CreateLTS{Roster: *s.roster, BCID: s.gbReply.Skipblock.Hash})
	require.Nil(t, err)

	return s
}

func (s *ts) createGenesis(t *testing.T) {
	var err error
	s.genesisMsg, err = byzcoin.DefaultGenesisMsg(byzcoin.CurrentVersion, s.roster,
		[]string{"spawn:" + calypso.ContractWriteID, "spawn:" + calypso.ContractReadID}, s.signer.Identity())
	require.Nil(t, err)
	s.gDarc = &s.genesisMsg.GenesisDarc
	s.genesisMsg.BlockInterval = time.Second

	s.cl, s.gbReply, err = byzcoin.NewLedger(s.genesisMsg, false)
	require.Nil(t, err)
}

func (s *ts) addWriteAndWait(t *testing.T, key []byte) *byzcoin.Proof {
	instID := s.addWrite(t, key)
	return s.waitInstID(t, instID)
}

func (s *ts) addWrite(t *testing.T, key []byte) byzcoin.InstanceID {
	write := calypso.NewWrite(cothority.Suite, s.ltsReply.LTSID, s.gDarc.GetBaseID(), s.ltsReply.X, key)
	writeBuf, err := protobuf.Encode(write)
	require.Nil(t, err)

	ctx := byzcoin.ClientTransaction{
		Instructions: byzcoin.Instructions{{
			InstanceID: byzcoin.NewInstanceID(s.gDarc.GetBaseID()),
			Nonce:      byzcoin.Nonce{},
			Index:      0,
			Length:     1,
			Spawn: &byzcoin.Spawn{
				ContractID: calypso.ContractWriteID,
				Args:       byzcoin.Arguments{{Name: "write", Value: writeBuf}},
			},
		}},
	}
	require.Nil(t, ctx.Instructions[0].SignBy(s.gDarc.GetID(), s.signer))
	_, err = s.cl.AddTransaction(ctx)
	require.Nil(t, err)
	return ctx.Instructions[0].DeriveID("")
}

func (s *ts) addRead(t *testing.T, write *byzcoin.Proof) byzcoin.InstanceID {
	var readBuf []byte
	read := &calypso.Read{
		Write: byzcoin.NewInstanceID(write.InclusionProof.Key),
		Xc:    s.signer.Ed25519.Point,
	}
	var err error
	readBuf, err = protobuf.Encode(read)
	require.Nil(t, err)
	ctx := byzcoin.ClientTransaction{
		Instructions: byzcoin.Instructions{{
			InstanceID: byzcoin.NewInstanceID(write.InclusionProof.Key),
			Nonce:      byzcoin.Nonce{},
			Index:      0,
			Length:     1,
			Spawn: &byzcoin.Spawn{
				ContractID: calypso.ContractReadID,
				Args:       byzcoin.Arguments{{Name: "read", Value: readBuf}},
			},
		}},
	}
	//require.Nil(t, ctx.Instructions[0].SignBy(s.gDarc.GetID(), s.signer))
	//_, err = s.cl.AddTransactionAndWait(ctx, 5)
	//require.Nil(t, err)
	//return ctx.Instructions[0].DeriveID("")

	require.Nil(t, ctx.Instructions[0].SignBy(s.gDarc.GetID(), s.signer))
	_, err = s.cl.AddTransaction(ctx)
	require.Nil(t, err)
	return ctx.Instructions[0].DeriveID("")


}

func (s *ts) addReadAndWait(t *testing.T, write *byzcoin.Proof) *byzcoin.Proof {
	instID := s.addRead(t, write)
	p, err := s.cl.GetProof(instID.Slice())
	require.Nil(t, err)
	return &p.Proof
}

func (s *ts) waitInstID(t *testing.T, instID byzcoin.InstanceID) *byzcoin.Proof {
	var err error
	var pr *byzcoin.Proof
	for i := 0; i < 10; i++ {
		pr, err = s.cl.WaitProof(instID, s.genesisMsg.BlockInterval, nil)
		if err == nil {
			require.Nil(t, pr.Verify(s.gbReply.Skipblock.Hash))
			break
		}
	}
	if err != nil {
		require.Fail(t, "didn't find proof")
	}
	return pr
}

func (s *ts) closeAll(t *testing.T) {
	require.Nil(t, s.cl.Close())
	s.local.CloseAll()
}


// Spawn an Admin Darc from the Genesis Darc, giving the service as input
//and returning the new darc together with the admin signer
/*func createAdminDarcTS(t *testing.T, s ts) (darc.Signer, *darc.Darc){
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
	require.Nil(t, ctx.Instructions[0].SignBy(s.gDarc.GetID(), s.signer))
	_, err = s.cl.AddTransaction(ctx)
	require.Nil(t, err)
	return admin,darcAdmin
}

// Spawn an User Darc from the Admin Darc(input), giving the service and Admin Signer as input as well
//and returning the new darc together with the user signer
func createUserDarcTS(t *testing.T, s ts, darcAdmin *darc.Darc, admin darc.Signer) (darc.Signer, *darc.Darc){

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
	_, err = s.cl.AddTransaction(ctx)
	require.Nil(t, err)
	return user, darcUser
}

// Spawn a Reader Darc from the Admin Darc(input), giving the service and Admin Signer as input as well
//and returning the new darc together with the user signer
func createReaderDarcTS(t *testing.T, s ts, darcAdmin *darc.Darc, admin darc.Signer, userDarc *darc.Darc) (*darc.Darc){

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
	_, err = s.cl.AddTransaction(ctx)
	require.Nil(t, err)
	return darcReader
}

// Spawn a Reader Darc from the Admin Darc(input), giving the service and Admin Signer as input as well
//and returning the new darc together with the user signer
func createGarageDarcTS(t *testing.T, s ts, darcAdmin *darc.Darc, admin darc.Signer, userDarc *darc.Darc) (*darc.Darc) {

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
	_, err = s.cl.AddTransaction(ctx)
	require.Nil(t, err)
	return darcGarage
}

func createCarDarcTS(t *testing.T, s ts, darcAdmin *darc.Darc,
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
	_, err = s.cl.AddTransaction(ctx)
	require.Nil(t, err)
	return darcCar
}

func (s *ts) createCarInstanceTS(t *testing.T, args byzcoin.Arguments,
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
	_, err := s.cl.AddTransaction(ctx)
	require.Nil(t, err)

	return ctx.Instructions[0].DeriveID("")
}
*/


