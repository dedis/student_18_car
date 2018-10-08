/*
Cisc is the Cisc Identity SkipChain to store information in a skipchain and
being able to retrieve it.

This is only one part of the system - the other part being the cothority that
holds the skipchain and answers to requests from the cisc-binary.
*/
package main

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/dedis/cothority"
	"github.com/dedis/cothority/identity"
	"github.com/dedis/cothority/pop/service"
	status "github.com/dedis/cothority/status/service"
	"github.com/dedis/kyber"
	"github.com/dedis/kyber/sign/schnorr"
	"github.com/dedis/kyber/util/encoding"
	"github.com/dedis/kyber/util/key"
	"github.com/dedis/onet"
	"github.com/dedis/onet/app"
	"github.com/dedis/onet/log"
	"github.com/dedis/onet/network"
	"github.com/qantik/qrgo"
	"gopkg.in/urfave/cli.v1"
)

func main() {
	app := cli.NewApp()
	app.Name = "SSH keystore client"
	app.Usage = "Connects to a ssh-keystore-server and updates/changes information"
	app.Version = "0.3"
	app.Commands = getCommands()
	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:  "debug, d",
			Value: 0,
			Usage: "debug-level: 1 for terse, 5 for maximal",
		},
		cli.StringFlag{
			Name:  "config, c",
			Value: "~/.cisc",
			Usage: "The configuration-directory of cisc",
		},
		cli.StringFlag{
			Name:  "config-ssh, cs",
			Value: "~/.ssh",
			Usage: "The configuration-directory of the ssh-directory",
		},
	}
	app.Before = func(c *cli.Context) error {
		log.SetDebugVisible(c.Int("debug"))
		return nil
	}
	log.ErrFatal(app.Run(os.Args))
}

/*
 * Admins commands
 */
func linkPin(c *cli.Context) error {
	if c.NArg() < 1 || c.NArg() > 2 {
		return errors.New("please give the following arguments: ip:port [PIN]")
	}

	var pin string
	if c.NArg() == 2 {
		pin = c.Args().Get(1)
	}
	addrStr := c.Args().First()
	addr := network.NewAddress(network.PlainTCP, addrStr)
	si := &network.ServerIdentity{Address: addr}

	cfg := loadConfigAdminOrFail(c)

	kp := key.NewKeyPair(cothority.Suite)
	client := onet.NewClient(cothority.Suite, identity.ServiceName)
	if err := client.SendProtobuf(si, &identity.PinRequest{PIN: pin, Public: kp.Public}, nil); err != nil {
		// Compare by string because we are on the client, and we will
		// be receiving a new error made locally by onet, not the original error.
		if strings.Contains(err.Error(), identity.ErrorReadPIN.Error()) {
			log.Info("Please read PIN in server-log")
			return nil
		}
		return err
	}
	log.Info("Successfully linked with", addr)
	cfg.KeyPairs[addrStr] = kp
	cfg.saveConfig(c)
	return nil
}

func getClient(c *cli.Context, arg string) (*ciscConfig, *network.ServerIdentity, *key.Pair, error) {
	if c.NArg() != 2 {
		return nil, nil, nil, errors.New("please give the following arguments: " + arg + " ip:port")
	}
	addrStr := c.Args().Get(1)
	addr := network.NewAddress(network.PlainTCP, addrStr)
	si := &network.ServerIdentity{Address: addr}

	cfg := loadConfigAdminOrFail(c)
	kp, ok := cfg.KeyPairs[addrStr]
	if !ok {
		return cfg, si, nil, errors.New("not linked")
	}

	return cfg, si, kp, nil
}
func linkAddFinal(c *cli.Context) error {
	cfg, si, kp, err := getClient(c, "final_statement.toml")
	if err != nil || cfg == nil {
		return err
	}
	client := onet.NewClient(cothority.Suite, identity.ServiceName)
	finalName := c.Args().First()
	buf, err := ioutil.ReadFile(finalName)
	log.ErrFatal(err)
	final, err := service.NewFinalStatementFromToml(buf)
	log.ErrFatal(err)
	if err := final.Verify(); err != nil {
		log.Error("Signature s invalid")
		return err
	}
	hash, err := final.Hash()
	if err != nil {
		log.Error("error while Hashing")
		return err
	}
	sig, err := schnorr.Sign(cothority.Suite, kp.Private, hash)
	if err != nil {
		return err
	}

	cerr := client.SendProtobuf(si,
		&identity.StoreKeys{Type: identity.PoPAuth, Final: final,
			Publics: nil, Sig: sig}, nil)
	return cerr
}

func linkAddPublic(c *cli.Context) error {
	cfg, si, kp, err := getClient(c, "public_key")
	if err != nil || cfg == nil {
		return err
	}
	client := onet.NewClient(cothority.Suite, identity.ServiceName)

	pub, err := encoding.StringHexToPoint(cothority.Suite, c.Args().Get(0))
	if err != nil {
		return err
	}

	h := cothority.Suite.Hash()
	_, err = pub.MarshalTo(h)
	if err != nil {
		return err
	}
	sig, err := schnorr.Sign(cothority.Suite, kp.Private, h.Sum(nil))
	if err != nil {
		return err
	}

	cerr := client.SendProtobuf(si,
		&identity.StoreKeys{
			Type:    identity.PublicAuth,
			Final:   nil,
			Publics: []kyber.Point{pub},
			Sig:     sig,
		}, nil)
	return cerr
}

func linkPair(c *cli.Context) error {
	kp := key.NewKeyPair(cothority.Suite)

	secStr, err := encoding.ScalarToStringHex(nil, kp.Private)
	if err != nil {
		return err
	}
	pubStr, err := encoding.PointToStringHex(nil, kp.Public)
	if err != nil {
		return err
	}
	log.Infof("Private: %s\nPublic: %s", secStr, pubStr)
	return nil
}

func linkList(c *cli.Context) error {
	cfg := loadConfigOrFail(c)
	for host := range cfg.KeyPairs {
		log.Info("Host:", host)
	}
	return nil
}

/*
 * Identity-related commands
 */
func scCreate(c *cli.Context) error {
	if c.NArg() < 1 {
		return errors.New("Please give a group-definition")
	}

	cfg := loadConfigAdminOrFail(c)
	group := getGroup(c)
	var atts []kyber.Point

	addrStr := group.Roster.List[0].Address.NetworkAddress()
	var typ identity.AuthType
	kp := cfg.KeyPairs[addrStr]
	if kp != nil {
		log.Info("Found full link to conode:", addrStr, kp.Public)
		typ = identity.PublicAuth
	} else if c.String("private") != "" {
		log.Info("Signing with given private key")
		typ = identity.PublicAuth
		kp = &key.Pair{}
		var err error
		kp.Private, err = encoding.StringHexToScalar(cothority.Suite, c.String("private"))
		if err != nil {
			return err
		}
		kp.Public = cothority.Suite.Point().Mul(kp.Private, nil)
	} else if c.String("token") != "" {
		b, err := ioutil.ReadFile(c.String("token"))
		if err != nil {
			return err
		}
		popToken, err := service.NewPopTokenFromToml(b)
		if err != nil {
			return err
		}
		if popToken == nil {
			return errors.New("couldn't read pop-token from " + c.String("token"))
		}
		typ = identity.PoPAuth
		kp = &key.Pair{
			Public:  popToken.Public,
			Private: popToken.Private,
		}
		atts = popToken.Final.Attendees
		log.Info("Found PoP-link to conode:", addrStr)
	} else {
		return errors.New("didn't find any authentication method")
	}

	name := c.String("name")
	if name == "" {
		var err error
		name, err = os.Hostname()
		if err != nil {
			return err
		}
	}
	log.Info("Creating new blockchain-identity for", name, "in roster", group.Roster.List)

	thr := c.Int("threshold")
	id := identity.NewIdentity(group.Roster, thr, name, nil)
	cfg.Identities = append(cfg.Identities, id)
	log.ErrFatal(id.CreateIdentity(typ, atts, kp.Private))
	log.Infof("New cisc-id is: %x", id.ID)
	return cfg.saveConfig(c)
}

func scJoin(c *cli.Context) error {
	log.Info("Connecting")
	name, err := os.Hostname()
	log.ErrFatal(err)
	switch c.NArg() {
	case 2:
		// We'll get all arguments after
	case 3:
		name = c.Args().Get(2)
	default:
		return errors.New("Please give the following arguments: group.toml id [hostname]")
	}
	group := getGroup(c)
	idBytes, err := hex.DecodeString(c.Args().Get(1))
	log.ErrFatal(err)
	sbid := identity.ID(idBytes)
	id := identity.NewIdentity(group.Roster, 0, name, nil)
	cfg := newCiscConfig(id)
	log.ErrFatal(id.AttachToIdentity(sbid))
	log.Infof("Public key: %s",
		id.Proposed.Device[id.DeviceName].Point.String())
	return cfg.saveConfig(c)
}

func scLeave(c *cli.Context) error {
	if c.NArg() == 0 {
		return errors.New("Please give device that you want to remove from identity")
	}
	cfg := loadConfigOrFail(c)
	var id *identity.Identity
	if len(cfg.Identities) == 1 {
		id = cfg.Identities[0]
	} else if c.NArg() == 1 {
		scList(c)
		return errors.New("Have more than one identity, please chose")
	} else {
		var err error
		id, err = cfg.findSC(c.Args().Get(1))
		if err != nil {
			return err
		}
		if id == nil {
			scList(c)
			return errors.New("Didn't find skipchain with id " + c.Args().Get(1))
		}
	}
	dev := c.Args().First()
	if _, ok := id.Data.Device[dev]; !ok {
		log.Error("Didn't find", dev, "in config. Available devices:")
		dataList(c)
		return errors.New("device not found in config")
	}
	prop := id.GetProposed()
	delete(prop.Device, dev)
	for _, s := range id.Data.GetSuffixColumn("ssh", dev) {
		delete(prop.Storage, "ssh:"+dev+":"+s)
	}
	cfg.proposeSendVoteUpdate(id, prop)
	return nil
}

func scList(c *cli.Context) error {
	cfg := loadConfigOrFail(c)
	if len(cfg.Identities) > 0 {
		log.Info("Full identities we're part of:")
		for _, id := range cfg.Identities {
			log.Infof("Name: %s - ID: %x", id.DeviceName, id.ID)
		}
	}
	if len(cfg.Follow) > 0 {
		log.Info("Identities we're following:")
		for _, i := range cfg.Follow {
			log.Infof("Devices: %s - ID: %x", i.Data.Device, i.ID)
		}
	}
	return nil
}

func scQrcode(c *cli.Context) error {
	cfg := loadConfigOrFail(c)
	id, err := cfg.findSC(c.Args().First())
	if err != nil {
		return err
	}
	if id == nil {
		scList(c)
		return errors.New("Please chose one of the existing skipchain-ids")
	}
	scid := []byte(id.ID)
	address := strings.Split(id.Data.Roster.RandomServerIdentity().Address.NetworkAddress(), ":")

	// Get our local IP address - this can be different from the public IP
	// address returned by a service like `whatsmyip`, because we're behind
	// a router.
	if address[0] == "localhost" && c.Bool("e") {
		conn, err := net.Dial("udp", "8.8.8.8:80")
		if err != nil {
			return err
		}

		localAddr := conn.LocalAddr().(*net.UDPAddr)
		conn.Close()
		address[0] = localAddr.IP.String()
	}

	str := fmt.Sprintf("cisc://%s/%x", address[0]+":"+address[1],
		scid)
	log.Info("QrCode for", str)
	qr, err := qrgo.NewQR(str)
	log.ErrFatal(err)
	qr.OutputTerminal()
	return nil
}

func scRosterShow(c *cli.Context) error {
	cfg := loadConfigOrFail(c)
	id, err := cfg.findSC(c.Args().Get(1))
	if err != nil {
		return err
	}
	if id == nil {
		scList(c)
		return errors.New("Please chose one of the existing skipchain-ids")
	}
	log.Infof("Roster for %x:", id.ID)
	var buf bytes.Buffer
	err = toml.NewEncoder(&buf).Encode(id.Roster().Toml(cothority.Suite))
	if err != nil {
		log.Error(err)
	} else {
		log.Info(buf.String())
	}
	return nil
}

func scRosterSet(c *cli.Context) error {
	cfg := loadConfigOrFail(c)
	id, err := cfg.findSC(c.Args().Get(1))
	if err != nil {
		return err
	}
	if id == nil {
		scList(c)
		return errors.New("Please chose one of the existing skipchain-ids")
	}
	group := getGroup(c)
	prop := id.GetProposed()
	prop.Roster = group.Roster
	cfg.proposeSendVoteUpdate(id, prop)
	log.Info("Proposed new roster for skipchain")
	if id.Proposed == nil {
		log.Info("New roster has been accepted")
	}
	return nil
}

func scRosterAdd(c *cli.Context) error {
	si := getServerIdentity(c)
	if si == nil {
		return errors.New("Please give either --toml or --addr as argument")
	}

	cfg := loadConfigOrFail(c)
	id, err := cfg.findSC(c.Args().Get(1))
	if err != nil {
		return err
	}
	if id == nil {
		scList(c)
		return errors.New("Please chose one of the existing skipchain-ids")
	}

	prop := id.GetProposed()
	prop.Roster = onet.NewRoster(append(id.Roster().List, si))
	cfg.proposeSendVoteUpdate(id, prop)
	log.Info("Proposed new roster for skipchain")
	if id.Proposed == nil {
		log.Info("New roster has been accepted")
	}
	return nil
}

func scRosterRemove(c *cli.Context) error {
	si := getServerIdentity(c)
	if si == nil {
		return errors.New("Please give either --toml or --addr as argument")
	}
	cfg := loadConfigOrFail(c)
	id, err := cfg.findSC(c.Args().Get(1))
	if err != nil {
		return err
	}
	if id == nil {
		scList(c)
		return errors.New("Please chose one of the existing skipchain-ids")
	}

	prop := id.GetProposed()
	roster := onet.NewRoster(id.Roster().List)
	index := -1
	for i, s := range roster.List {
		if s.Equal(si) {
			index = i
		}
	}
	if index == -1 {
		return errors.New("Couldn't find this node in the roster")
	}
	roster.List = append(roster.List[0:index], roster.List[index+1:]...)
	prop.Roster = roster
	cfg.proposeSendVoteUpdate(id, prop)
	log.Info("Proposed new roster for skipchain")
	if id.Proposed == nil {
		log.Info("New roster has been accepted")
	}
	return nil
}

func getServerIdentity(c *cli.Context) *network.ServerIdentity {
	if file := c.String("toml"); file != "" {
		// Suppose it's a roster file
		g, err := getGroupString(file)
		log.ErrFatal(err)
		return g.Roster.List[0]
	}
	if addr := c.String("addr"); addr != "" {
		// Go and get this conode's public key
		if !strings.Contains(addr, "://") {
			addr = "tls://" + addr
		}
		si := network.NewServerIdentity(nil, network.Address(addr))
		resp, err := status.NewClient().Request(si)
		log.ErrFatal(err)
		return resp.ServerIdentity
	}
	return nil
}

/*
 * Commands related to the data
 */
func dataUpdate(c *cli.Context) error {
	cfg := loadConfigOrFail(c)
	id, err := cfg.findSC(c.Args().First())
	if err != nil {
		return err
	}
	if id == nil {
		scList(c)
		return errors.New("Please indicate skipchain-id")
	}
	log.Info("Successfully updated")
	if err := cfg.saveConfig(c); err != nil {
		return err
	}
	if id.Proposed != nil {
		cfg.showDifference(id)
	} else {
		cfg.showKeys(id)
	}
	return nil
}
func dataList(c *cli.Context) error {
	log.Info("Listing data on the identity-skipchain")
	cfg := loadConfigOrFail(c)
	id, err := cfg.findSC(c.Args().First())
	if err != nil {
		return err
	}
	if id == nil {
		scList(c)
		return errors.New("Please indicate skipchain-id")
	}
	log.Info("Account name:", id.DeviceName)
	log.Infof("Identity-ID: %x", id.ID)
	if c.Bool("d") {
		log.Info(id.Data.Storage)
	} else {
		cfg.showKeys(id)
	}
	log.Info("Roster is:", id.Data.Roster.List)
	if c.Bool("p") {
		if id.Proposed != nil {
			log.Infof("Proposed data: %s", id.Proposed)
		} else {
			log.Info("No proposed data")
		}
	}
	return nil
}
func dataClear(c *cli.Context) error {
	cfg := loadConfigOrFail(c)
	id, err := cfg.findSC(c.Args().First())
	if err != nil {
		return err
	}
	if id == nil {
		scList(c)
		return errors.New("Please indicate skipchain-id")
	}
	if err := id.DataUpdate(); err != nil {
		return err
	}
	id.Proposed = nil
	if err := id.ProposeSend(id.Proposed); err != nil {
		return err
	}
	log.Infof("Cleared proposed-data of skipchain %x", id.ID)
	return cfg.saveConfig(c)
}
func dataVote(c *cli.Context) error {
	cfg := loadConfigOrFail(c)
	id, err := cfg.findSC(c.Args().First())
	if err != nil {
		return err
	}
	if id == nil {
		scList(c)
		return errors.New("Please indicate skipchain-id")
	}
	if id.Proposed == nil {
		log.Info("No proposed data")
		return nil
	}

	if c.Bool("no") {
		return nil
	} else if !c.Bool("yes") {
		cfg.showDifference(id)
		if !app.InputYN(true, "Do you want to accept the changes") {
			return nil
		}
	}
	if err := id.ProposeVote(true); err != nil {
		return err
	}
	dataList(c)
	return cfg.saveConfig(c)
}

/*
 * Commands related to the key/value storage and retrieval
 */
func kvList(c *cli.Context) error {
	cfg := loadConfigOrFail(c)
	id, err := cfg.findSC(c.Args().First())
	if id == nil {
		scList(c)
		return errors.New("Please give skipchain-id")
	}
	if err != nil {
		return err
	}
	// only print value for the given key
	if c.String("key") != "" {
		val, ok := id.Data.Storage[c.String("key")]
		if !ok {
			return errors.New("key does not exists")
		}
		fmt.Println(val)
		return nil
	}

	// print everything
	for k, v := range id.Data.Storage {
		log.Infof("%s: %s", k, v)
	}
	return nil
}
func kvValue(c *cli.Context) error {
	if c.NArg() < 1 {
		return errors.New("please give key to search")
	}
	cfg := loadConfigOrFail(c)
	id, err := cfg.findSC(c.Args().Get(1))
	if err != nil {
		return err
	}
	if id == nil {
		scList(c)
		return errors.New("Please give skipchain-id")
	}
	key := c.Args().First()
	value, ok := id.Data.Storage[key]
	if ok {
		log.Infof("Data[%s] = %s", key, value)
	} else {
		log.Infof("Key '%s' does not exist", key)
	}
	return nil
}
func kvAdd(c *cli.Context) error {
	cfg := loadConfigOrFail(c)
	if c.NArg() < 2 {
		return errors.New("Please give a key value pair")
	}
	id, err := cfg.findSC(c.Args().Get(2))
	if err != nil {
		return err
	}
	if id == nil {
		scList(c)
		return errors.New("Please give skipchain-id")
	}
	key := c.Args().Get(0)
	value := c.Args().Get(1)
	prop := id.GetProposed()
	prop.Storage[key] = value
	return addKv(c, cfg, id, prop)
}

// kvAddFile reads the input file, and stores it in the data. The key is the
// name of the file by default or overridden by the key flag. Do not use with
// big files as it reads all at once.
func kvAddFile(c *cli.Context) error {
	cfg := loadConfigOrFail(c)
	if c.NArg() < 1 {
		return errors.New("Missing argument: file name")
	}
	id, err := cfg.findSC(c.Args().Get(1))
	if err != nil {
		scList(c)
		return errors.New("Please give skipchain-id")
	}
	// read file name
	file := c.Args().First()
	// determine the key
	key := path.Base(file)
	if c.String("key") != "" {
		key = c.String("key")
	}
	// read file
	fd, err := os.Open(file)
	log.ErrFatal(err)
	buff, err := ioutil.ReadAll(fd)
	log.ErrFatal(err)

	// store it
	log.Info("File will be stored under key: " + key)
	prop := id.GetProposed()
	prop.Storage[key] = string(buff)
	return addKv(c, cfg, id, prop)
}

func addKv(c *cli.Context, cfg *ciscConfig, id *identity.Identity, prop *identity.Data) error {
	cfg.proposeSendVoteUpdate(id, prop)
	if id.Proposed == nil {
		log.Info("Stored key-value pair")
	} else {
		log.Info("Voted for key-value pair - need confirmation")
	}
	return cfg.saveConfig(c)

}
func kvDel(c *cli.Context) error {
	cfg := loadConfigOrFail(c)
	if c.NArg() < 1 {
		return errors.New("Please give a key to delete")
	}
	id, err := cfg.findSC(c.Args().Get(1))
	if err != nil {
		return err
	}
	if id == nil {
		scList(c)
		return errors.New("Please give skipchain-id")
	}
	key := c.Args().First()
	prop := id.GetProposed()
	if _, ok := prop.Storage[key]; !ok {
		return errors.New("Didn't find key " + key + " in the config")
	}
	delete(prop.Storage, key)
	cfg.proposeSendVoteUpdate(id, prop)
	return cfg.saveConfig(c)
}
func kvAddWeb(c *cli.Context) error {
	if c.NArg() < 1 {
		return errors.New("Please give an html file to add")
	}
	if c.Bool("inline") {
		return errors.New("inline Not implemented yet")
		// https://github.com/remy/inliner
	}
	cfg := loadConfigOrFail(c)
	id, err := cfg.findSC(c.Args().Get(1))
	if err != nil {
		return err
	}
	if id == nil {
		scList(c)
		return errors.New("Please give skipchain-id")
	}
	name := c.Args().First()
	log.Info("Reading file", name)
	data, err := ioutil.ReadFile(name)
	if err != nil {
		return nil
	}
	prop := id.GetProposed()
	prop.Storage["html:"+path.Dir(name)+":"+path.Base(name)] = string(data)
	cfg.proposeSendVoteUpdate(id, prop)
	return cfg.saveConfig(c)
}

/*
 * Commands related to the ssh-handling. All ssh-keys are stored in the
 * identity-sc as
 *
 *   ssh:device:server = ssh_public_key
 *
 * where 'ssh' is a fixed string, 'device' is the device where the private
 * key is stored and 'server' the server that should add the public key to
 * its authorized_keys.
 *
 * For safety reasons, this function saves to authorized_keys.cisc instead
 * of overwriting authorized_keys. If authorized_keys doesn't exist,
 * a symbolic link to authorized_keys.cisc is created.
 *
 * If you want to use your own authorized_keys but also allow keys in
 * authorized_keys.cisc to log in to your system, you can add the following
 * line to /etc/ssh/sshd_config
 *
 *   AuthorizedKeysFile ~/.ssh/authorized_keys ~/.ssh/authorized_keys.cisc
 */
func sshAdd(c *cli.Context) error {
	cfg := loadConfigOrFail(c)
	id, err := cfg.findSC(c.Args().Get(1))
	if err != nil {
		return err
	}
	if id == nil {
		scList(c)
		return errors.New("Please give skipchain-id")
	}
	sshDir, sshConfig := sshDirConfig(c)
	if c.NArg() != 1 {
		return errors.New("Please give the hostname as argument")
	}

	// Get the current configuration
	sc, err := NewSSHConfigFromFile(sshConfig)
	if err != nil {
		return err
	}

	// Add a new host-entry
	hostname := c.Args().First()
	alias := c.String("a")
	if alias == "" {
		alias = hostname
	}
	filePub := path.Join(sshDir, "key_"+alias+".pub")
	idPriv := "key_" + alias
	filePriv := path.Join(sshDir, idPriv)
	if err := makeSSHKeyPair(c.Int("sec"), filePub, filePriv); err != nil {
		return err
	}
	host := NewSSHHost(alias, "HostName "+hostname,
		"IdentityFile "+filePriv)
	if port := c.String("p"); port != "" {
		host.AddConfig("Port " + port)
	}
	if user := c.String("u"); user != "" {
		host.AddConfig("User " + user)
	}
	sc.AddHost(host)
	if err := ioutil.WriteFile(sshConfig, []byte(sc.String()), 0600); err != nil {
		return err
	}

	// Propose the new configuration
	prop := id.GetProposed()
	key := strings.Join([]string{"ssh", id.DeviceName, hostname}, ":")
	pub, err := ioutil.ReadFile(filePub)
	if err != nil {
		return err
	}
	prop.Storage[key] = strings.TrimSpace(string(pub))
	cfg.proposeSendVoteUpdate(id, prop)
	return cfg.saveConfig(c)
}
func sshLs(c *cli.Context) error {
	cfg := loadConfigOrFail(c)
	id, err := cfg.findSC(c.Args().First())
	if err != nil {
		return err
	}
	if id == nil {
		scList(c)
		return errors.New("Please give skipchain-id")
	}
	var devs []string
	if c.Bool("a") {
		devs = id.Data.GetSuffixColumn("ssh")
	} else {
		devs = []string{id.DeviceName}
	}
	for _, dev := range devs {
		for _, pub := range id.Data.GetSuffixColumn("ssh", dev) {
			log.Infof("SSH-key for device %s: %s", dev, pub)
		}
	}
	return nil
}
func sshDel(c *cli.Context) error {
	cfg := loadConfigOrFail(c)
	id, err := cfg.findSC(c.Args().Get(1))
	if err != nil {
		return err
	}
	if id == nil {
		scList(c)
		return errors.New("Please give skipchain-id")
	}
	_, sshConfig := sshDirConfig(c)
	if c.NArg() == 0 {
		return errors.New("Please give alias or host to delete from ssh")
	}
	sc, err := NewSSHConfigFromFile(sshConfig)
	if err != nil {
		return err
	}
	// Converting ah to a hostname if found in ssh-config
	host := sc.ConvertAliasToHostname(c.Args().First())
	if len(id.Data.GetValue("ssh", id.DeviceName, host)) == 0 {
		log.Error("Didn't find alias or host", host, "here is what I know:")
		sshLs(c)
		return errors.New("unknown alias or host")
	}

	sc.DelHost(host)
	if err := ioutil.WriteFile(sshConfig, []byte(sc.String()), 0600); err != nil {
		return err
	}
	prop := id.GetProposed()
	delete(prop.Storage, "ssh:"+id.DeviceName+":"+host)
	cfg.proposeSendVoteUpdate(id, prop)
	return cfg.saveConfig(c)
}
func sshRotate(c *cli.Context) error {
	return errors.New("Not yet implemented")
}
func sshSync(c *cli.Context) error {
	return errors.New("Not yet implemented")
}

func followAdd(c *cli.Context) error {
	if c.NArg() < 2 {
		return errors.New("Please give a group-definition, an ID, and optionally a service-name of the skipchain to follow")
	}
	cfg, _ := loadConfig(c)
	group := getGroup(c)
	idBytes, err := hex.DecodeString(c.Args().Get(1))
	if err != nil {
		return err
	}
	id := identity.ID(idBytes)
	newID, err := identity.NewIdentityFromRoster(group.Roster, id)
	if err != nil {
		return err
	}
	if c.NArg() == 3 {
		newID.DeviceName = c.Args().Get(2)
	} else {
		var err error
		newID.DeviceName, err = os.Hostname()
		if err != nil {
			return err
		}
		log.Info("Using", newID.DeviceName, "as the device-name.")
	}
	cfg.Follow = append(cfg.Follow, newID)
	cfg.writeAuthorizedKeys(c)
	// Identity needs to exist, else saving/loading will fail. For
	// followers it doesn't matter if the identity will be overwritten,
	// as it is not used.
	cfg.Identities = append(cfg.Identities, newID)
	return cfg.saveConfig(c)
}
func followDel(c *cli.Context) error {
	if c.NArg() != 1 {
		return errors.New("Please give id of skipchain to unfollow")
	}
	cfg := loadConfigOrFail(c)
	idBytes, err := hex.DecodeString(c.Args().First())
	if err != nil {
		return err
	}
	idDel := identity.ID(idBytes)
	newSlice := cfg.Follow[:0]
	for _, id := range cfg.Follow {
		if !bytes.Equal(id.ID, idDel) {
			newSlice = append(newSlice, id)
		}
	}
	cfg.Follow = newSlice
	cfg.writeAuthorizedKeys(c)
	return cfg.saveConfig(c)
}
func followList(c *cli.Context) error {
	cfg := loadConfigOrFail(c)
	for _, id := range cfg.Follow {
		if c.Bool("id-only") {
			fmt.Printf("%x\n", id.ID)
			continue
		}
		log.Infof("SCID: %x", id.ID)
		server := id.DeviceName
		log.Infof("Server %s is asked to accept ssh-keys from %s:",
			server,
			id.Data.GetIntermediateColumn("ssh", server))
	}
	return nil
}
func followUpdate(c *cli.Context) error {
	cfg := loadConfigOrFail(c)
	for _, f := range cfg.Follow {
		if err := f.DataUpdate(); err != nil {
			return err
		}
	}
	cfg.writeAuthorizedKeys(c)
	return cfg.saveConfig(c)
}

/*
 * Commands related to certificate management
 * Request, add, retrieve, revoke, renew, list the certificates
 */

// Request a Certificate to Letsencrypt Ca and store it in the skipchain
// It receives as argument the domain name the certificate path where
// the keys and the fullchain will be stored the and the www path to complete
// the challenge
func certRequest(c *cli.Context) error {
	return errors.New("no longer implemented")
	// cfg := loadConfigOrFail(c)
	// if c.NArg() < 3 {
	// 	return errors.New("Please give a domain name, the path to the certificate repository and the path to the www folder")
	// }

	// domain := c.Args().Get(0)
	// certDir := c.Args().Get(1)
	// wwwDir := c.Args().Get(2)
	// certPath := path.Join(certDir, domain)

	// // Request Certificate (see certificate.go)
	// cert, err := getCert(wwwDir, certDir, domain)
	// if err != nil {
	// 	// Delete generated files if an error happens
	// 	os.Remove(path.Join(certPath, "registerkey.pem"))
	// 	os.Remove(path.Join(certPath, "privkey.pem"))
	// 	return errors.New("Error in requesting certificate: " + err.Error())
	// }

	// // Check the validity of the certificate(see certificate.go)
	// log.Info("Verify the validity of the cert:")
	// if !isValid(cert) {
	// 	return errors.New("Certificate is not valid, can't add it to proposal storage")
	// }

	// id, err := cfg.findSCOrList(c, c.Args().Get(3))
	// if err != nil {
	// 	return err
	// }

	// prop := id.GetProposed()
	// log.Info("Valid Certificate, added to proposal storage")
	// prop.Storage[domain] = string(cert)
	// cfg.CertPath[domain] = certPath

	// // Send the certificate to proposal
	// cfg.proposeSendVoteUpdate(id, prop)
	// return cfg.saveConfig(c)
}

// List all the certificates stored in the skipchain, by giving -v it displays
// the fullchain.pem by giving -p only the domain certificate and by giving -c
// it only displays the chain certificate
func certList(c *cli.Context) error {
	return errors.New("no longer implemented")
	// cfg := loadConfigOrFail(c)

	// id, err := cfg.findSCOrList(c, c.Args().First())
	// if err != nil {
	// 	return err
	// }

	// log.Infof("config for id %x", id.ID)

	// for k, v := range id.Data.Storage {
	// 	if isCert([]byte(v)) {
	// 		certLE, err := pemToCertificate([]byte(v))
	// 		if err != nil {
	// 			return errors.New("Error in conversion to x509 certificate: " + err.Error())
	// 		}
	// 		certPath := cfg.CertPath[k]
	// 		if certPath == "" {
	// 			certPath = "Not defined"
	// 		}
	// 		public, chain := splitCertPublicChain(v)
	// 		log.Infof("%s - Expiry Date: %s - Certificate directory: %s", k, certLE.NotAfter, certPath)
	// 		if c.Bool("v") || c.Bool("p") && c.Bool("c") {
	// 			log.Infof("%s", v)
	// 		} else if c.Bool("p") {
	// 			log.Info("Certificate of the domain")
	// 			log.Infof("%s", public)
	// 		} else if c.Bool("c") {
	// 			log.Info("Chain certificate")
	// 			log.Infof("%s", chain)
	// 		}
	// 	}
	// }

	// return cfg.saveConfig(c)
}

// Store a non-requested certificate by giving as argument the key this one will
// correspond to the key stored in the skipchain and the .pem file corresponding
// to the certificate file
func certStore(c *cli.Context) error {
	return errors.New("no longer implemented")
	// cfg := loadConfigOrFail(c)
	// if c.NArg() < 2 {
	// 	return errors.New("Please give a key certificate pair")
	// }
	// domain := c.Args().Get(0)
	// path := c.Args().Get(1)

	// // Check the validity of the certificate
	// cert, err := ioutil.ReadFile(path)

	// if err != nil {
	// 	return err
	// }

	// if !isCert(cert) {
	// 	return errors.New("Please give a certificate")
	// }
	// log.Info("Verify the validity of the cert:")
	// if !isValid(cert) {
	// 	return errors.New("Certificate is not valid, can't add it to proposal storage ")
	// }

	// id, err := cfg.findSCOrList(c, c.Args().Get(2))
	// if err != nil {
	// 	return err
	// }

	// prop := id.GetProposed()
	// log.Info("Valid Certificate, added to proposal storage")
	// prop.Storage[domain] = string(cert)
	// cfg.proposeSendVoteUpdate(id, prop)
	// return cfg.saveConfig(c)
}

// Verify the validity of a certificate by giving as argument the key
// corresponding to this latter
func certVerify(c *cli.Context) error {
	return errors.New("no longer implemented")
	// if c.NArg() < 1 {
	// 	return errors.New("Please give the certificate key for verification")
	// }
	// cfg := loadConfigOrFail(c)
	// id, err := cfg.findSCOrList(c, c.Args().Get(1))
	// if err != nil {
	// 	return err
	// }

	// k := c.Args().Get(0)

	// cert := []byte(id.Data.Storage[k])

	// if !isCert(cert) {
	// 	return errors.New("The values do not correspond to a certificate")
	// }
	// log.Info("Verify the validity of the cert:")
	// if !isValid(cert) {
	// 	return errors.New("Certificate is not valid")
	// }
	// return cfg.saveConfig(c)
}

// Renew a certificate by giving the domain name/key
func certRenew(c *cli.Context) error {
	return errors.New("no longer implemented")
	// cfg := loadConfigOrFail(c)
	// if c.NArg() < 1 {
	// 	return errors.New("Please give a domain name")
	// }

	// domain := c.Args().Get(0)
	// id, err := cfg.findSCOrList(c, c.Args().Get(1))
	// if err != nil {
	// 	return err
	// }

	// log.Info("Checking the certificate")

	// if _, ok := id.Data.Storage[domain]; !ok {
	// 	return errors.New("Didn't find key " + domain + " in the config")
	// }

	// cert := []byte(id.Data.Storage[domain])
	// if !isCert(cert) {
	// 	return errors.New("The values do not correspond to a certificate")
	// }

	// // Renew the cert (see certificate.go)
	// newcert, err := renewCert(cert)
	// if err != nil {
	// 	return errors.New("Error while renewing certificate: " + err.Error())
	// }

	// err = ioutil.WriteFile(path.Join(cfg.CertPath[domain], "fullchain.pem"), newcert, 0644)
	// if err != nil {
	// 	return errors.New("Can't create fullchain.pem" + err.Error())
	// }
	// log.Info("Certificate successfully renewed")

	// // Check the certificate
	// log.Info("Verify the validity of the cert:")
	// if !isValid(newcert) {
	// 	return errors.New("Certificate is not valid, can't add it to proposal storage ")
	// }
	// prop := id.GetProposed()
	// log.Info("Valid Certificate, added to proposal storage")
	// prop.Storage[domain] = string(newcert)
	// cfg.proposeSendVoteUpdate(id, prop)
	// return cfg.saveConfig(c)
}

// Revoke a certificate by giving the key corresponding to the certificate and
// the register key of this certificate. This certificate will be then deleted
// from the skipchain
func certRevoke(c *cli.Context) error {
	return errors.New("no longer implemented")
	// cfg := loadConfigOrFail(c)
	// if c.NArg() < 2 {
	// 	return errors.New("Please give the certificate to delete and the register key of the certificate")
	// }

	// id, err := cfg.findSCOrList(c, c.Args().Get(2))
	// if err != nil {
	// 	return err
	// }

	// key := c.Args().First()
	// prop := id.GetProposed()
	// if _, ok := prop.Storage[key]; !ok {
	// 	return errors.New("Didn't find key " + key + " in the config")
	// }
	// cert := []byte(prop.Storage[key])
	// if !isCert(cert) {
	// 	return errors.New("The values are not a certificate")
	// }

	// // Revoke the certificate (see certificate.go)
	// err = revokeCert(c.Args().Get(1), cert)
	// if err != nil {
	// 	return errors.New("Error revoking the certificate: " + err.Error())
	// }
	// delete(prop.Storage, key)
	// log.Info("Succesfully revoked")
	// cfg.proposeSendVoteUpdate(id, prop)
	// return cfg.saveConfig(c)
}

// Retrieve the fullchain.pem by giving the key corresponding to the certificate
// in the skipchain, you can give optionally a directory to write the certificates on it
func certRetrieve(c *cli.Context) error {
	return errors.New("no longer implemented")
	// if c.NArg() < 1 {
	// 	return errors.New("Please give the key of the certificate")
	// }
	// k := c.Args().Get(0)
	// cfg := loadConfigOrFail(c)
	// id, err := cfg.findSCOrList(c, c.Args().Get(1))
	// if err != nil {
	// 	return err
	// }

	// cert := []byte(id.Data.Storage[k])
	// public, chain := splitCertPublicChain(string(cert))

	// if cert == nil {
	// 	return errors.New("Cisc do not store a certificate for this key")
	// }
	// if !isCert(cert) {
	// 	return errors.New("The value corresponding to the key is not a certificate")
	// }
	// log.Info("Verify the validity of the cert:")
	// if !isValid(cert) {
	// 	return errors.New("Certificate is not valid")
	// }
	// log.Info("Valid certificate")
	// if c.String("d") != "" {
	// 	if _, err = os.Stat(c.String("d")); os.IsNotExist(err) {
	// 		os.MkdirAll(c.String("d"), 0777)
	// 	}
	// }
	// log.Info("Retrieves the domain certificate to: " + path.Join(c.String("directory"), k+".pem"))
	// err = ioutil.WriteFile(path.Join(c.String("d"), k+".pem"), []byte(public), 0644)
	// if err != nil {
	// 	return err
	// }
	// if chain != "" {
	// 	log.Info("Retrieve the fullchain certificate to: " + path.Join(c.String("d"), k+"_fullchain.pem"))
	// 	err = ioutil.WriteFile(path.Join(c.String("d"), k+"_fullchain.pem"), cert, 0644)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	// return cfg.saveConfig(c)
}
