package eventlog

import (
	"bytes"
	"errors"

	"github.com/dedis/cothority/byzcoin"
	"github.com/dedis/cothority/byzcoin/darc"
	"github.com/dedis/protobuf"

	"github.com/dedis/cothority"
	"github.com/dedis/onet"
)

// Client is a structure to communicate with the eventlog service
type Client struct {
	ByzCoin *byzcoin.Client
	// The DarcID with "invoke:eventlog" permission on it.
	DarcID darc.ID
	// Signers are the Darc signers that will sign transactions sent with this client.
	Signers  []darc.Signer
	Instance byzcoin.InstanceID
	c        *onet.Client
}

// NewClient creates a new client to talk to the eventlog service.
// Fields DarcID, Instance, and Signers must be filled in before use.
func NewClient(ol *byzcoin.Client) *Client {
	return &Client{
		ByzCoin: ol,
		c:       onet.NewClient(cothority.Suite, ServiceName),
	}
}

// Create creates a new event log. This method is synchronous: it will only
// return once the new eventlog has been committed into the ledger (or after
// a timeout). Upon non-error return, c.Instance will be correctly set.
func (c *Client) Create() error {
	instr := byzcoin.Instruction{
		InstanceID: byzcoin.NewInstanceID(c.DarcID),
		Index:      0,
		Length:     1,
		Spawn:      &byzcoin.Spawn{ContractID: contractName},
	}
	if err := instr.SignBy(c.DarcID, c.Signers...); err != nil {
		return err
	}
	tx := byzcoin.ClientTransaction{
		Instructions: []byzcoin.Instruction{instr},
	}
	if _, err := c.ByzCoin.AddTransactionAndWait(tx, 2); err != nil {
		return err
	}

	c.Instance = instr.DeriveID("")
	return nil
}

// A LogID is an opaque unique identifier useful to find a given log message later
// via GetEvent.
type LogID []byte

// Log asks the service to log events.
func (c *Client) Log(ev ...Event) ([]LogID, error) {
	tx, keys, err := makeTx(c.DarcID, c.Instance, ev, c.Signers)
	if err != nil {
		return nil, err
	}
	if _, err := c.ByzCoin.AddTransaction(*tx); err != nil {
		return nil, err
	}
	return keys, nil
}

// GetEvent asks the service to retrieve an event.
func (c *Client) GetEvent(key []byte) (*Event, error) {
	reply, err := c.ByzCoin.GetProof(key)
	if err != nil {
		return nil, err
	}
	if !reply.Proof.InclusionProof.Match() {
		return nil, errors.New("not an inclusion proof")
	}
	k, vs, err := reply.Proof.KeyValue()
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(k, key) {
		return nil, errors.New("wrong key")
	}
	if len(vs) < 2 {
		return nil, errors.New("not enough values")
	}
	e := Event{}
	err = protobuf.Decode(vs[0], &e)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func makeTx(darcID darc.ID, id byzcoin.InstanceID, msgs []Event, signers []darc.Signer) (*byzcoin.ClientTransaction, []LogID, error) {
	// We need the identity part of the signatures before
	// calling ToDarcRequest() below, because the identities
	// go into the message digest.
	sigs := make([]darc.Signature, len(signers))
	for i, x := range signers {
		sigs[i].Signer = x.Identity()
	}

	keys := make([]LogID, len(msgs))

	instrNonce := byzcoin.GenNonce()
	tx := byzcoin.ClientTransaction{
		Instructions: make([]byzcoin.Instruction, len(msgs)),
	}
	for i, msg := range msgs {
		eventBuf, err := protobuf.Encode(&msg)
		if err != nil {
			return nil, nil, err
		}
		argEvent := byzcoin.Argument{
			Name:  "event",
			Value: eventBuf,
		}
		tx.Instructions[i] = byzcoin.Instruction{
			InstanceID: id,
			Nonce:      instrNonce,
			Index:      i,
			Length:     len(msgs),
			Invoke: &byzcoin.Invoke{
				Command: contractName,
				Args:    []byzcoin.Argument{argEvent},
			},
			Signatures: append([]darc.Signature{}, sigs...),
		}
	}
	for i := range tx.Instructions {
		darcSigs := make([]darc.Signature, len(signers))
		for j, signer := range signers {
			dr, err := tx.Instructions[i].ToDarcRequest(darcID)
			if err != nil {
				return nil, nil, err
			}

			sig, err := signer.Sign(dr.Hash())
			if err != nil {
				return nil, nil, err
			}
			darcSigs[j] = darc.Signature{
				Signature: sig,
				Signer:    signer.Identity(),
			}
		}
		tx.Instructions[i].Signatures = darcSigs
		keys[i] = LogID(tx.Instructions[i].DeriveID("").Slice())
	}
	return &tx, keys, nil
}

// Search executes a search on the filter in req. See the definition of
// type SearchRequest for additional details about how the filter is interpreted.
// The ID and Instance fields of the SearchRequest will be filled in from c.
func (c *Client) Search(req *SearchRequest) (*SearchResponse, error) {
	req.ID = c.ByzCoin.ID
	req.Instance = c.Instance

	reply := &SearchResponse{}
	if err := c.c.SendProtobuf(c.ByzCoin.Roster.List[0], req, reply); err != nil {
		return nil, err
	}
	return reply, nil
}
