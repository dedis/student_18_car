package lib

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/dedis/cothority"
	"github.com/dedis/cothority/byzcoin"
	"github.com/dedis/cothority/byzcoin/darc"
	"github.com/dedis/cothority/skipchain"
	"github.com/dedis/onet"
	"github.com/dedis/onet/network"
	"github.com/dedis/protobuf"
)

// ConfigPath points to where the files will be stored by default.
var ConfigPath = "."

// Config is the structure used by ol to save its configuration. It holds everything
// necessary to talk to a ByzCoin instance. The GenesisDarc and AdminIdentity
// can change over the lifetime of a ledger.
type Config struct {
	Roster        onet.Roster
	ByzCoinID     skipchain.SkipBlockID
	GenesisDarc   darc.Darc
	AdminIdentity darc.Identity
}

// LoadKey returns the signer of a given identity. It searches it in the ConfigPath.
func LoadKey(id darc.Identity) (*darc.Signer, error) {
	// Find private key file.
	fn := fmt.Sprintf("key-%s.cfg", id)
	fn = filepath.Join(ConfigPath, fn)
	return LoadSigner(fn)
}

// LoadSigner loads a signer from a file given by fn.
func LoadSigner(fn string) (*darc.Signer, error) {
	buf, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, err
	}

	var signer darc.Signer
	err = protobuf.DecodeWithConstructors(buf, &signer,
		network.DefaultConstructors(cothority.Suite))
	return &signer, err
}

// SaveKey stores a signer in a file.
func SaveKey(signer darc.Signer) error {
	os.MkdirAll(ConfigPath, 0755)

	fn := fmt.Sprintf("key-%s.cfg", signer.Identity())
	fn = filepath.Join(ConfigPath, fn)

	// perms = 0400 because there is key material inside this file.
	f, err := os.OpenFile(fn, os.O_RDWR|os.O_CREATE, 0400)
	if err != nil {
		return fmt.Errorf("could not write %v: %v", fn, err)
	}

	buf, err := protobuf.Encode(&signer)
	if err != nil {
		return err
	}
	_, err = f.Write(buf)
	if err != nil {
		return err
	}
	return f.Close()
}

// SaveConfig stores the config in the ConfigPath directory. It returns the
// pathname of the stored file.
func SaveConfig(cfg Config) (string, error) {
	os.MkdirAll(ConfigPath, 0755)

	fn := fmt.Sprintf("bc-%x.cfg", cfg.ByzCoinID)
	fn = filepath.Join(ConfigPath, fn)

	buf, err := protobuf.Encode(&cfg)
	if err != nil {
		return fn, err
	}
	err = ioutil.WriteFile(fn, buf, 0644)
	if err != nil {
		return fn, err
	}

	return fn, nil
}

// LoadConfig returns a config read from the file and an initialized
// Client that can be used to communicate with ByzCoin.
func LoadConfig(file string) (cfg Config, cl *byzcoin.Client, err error) {
	var cfgBuf []byte
	cfgBuf, err = ioutil.ReadFile(file)
	if err != nil {
		return
	}
	err = protobuf.DecodeWithConstructors(cfgBuf, &cfg,
		network.DefaultConstructors(cothority.Suite))
	if err != nil {
		return
	}
	cl = byzcoin.NewClient(cfg.ByzCoinID, cfg.Roster)
	return
}
