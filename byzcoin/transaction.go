package byzcoin

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"

	"github.com/dedis/onet/log"
	"github.com/dedis/onet/network"

	"github.com/dedis/cothority/byzcoin/collection"
	"github.com/dedis/cothority/byzcoin/darc"
	"github.com/dedis/protobuf"
)

// An InstanceID is a unique identifier for one instance of a contract.
type InstanceID [32]byte

func (iID InstanceID) String() string {
	return fmt.Sprintf("%x", iID.Slice())
}

// Nonce is used to prevent replay attacks in instructions.
type Nonce [32]byte

func init() {
	network.RegisterMessages(Instruction{}, TxResult{},
		StateChange{})
}

// NewNonce returns a nonce given a slice of bytes.
func NewNonce(buf []byte) Nonce {
	if len(buf) != 32 {
		return Nonce{}
	}
	n := Nonce{}
	copy(n[:], buf)
	return n
}

// NewInstanceID converts the first 32 bytes of in into an InstanceID.
// Giving nil as in results in the zero InstanceID, which is the special
// key that holds the ledger config.
func NewInstanceID(in []byte) InstanceID {
	var i InstanceID
	copy(i[:], in)
	return i
}

// Equal returns if both InstanceIDs point to the same instance.
func (iID InstanceID) Equal(other InstanceID) bool {
	return bytes.Equal(iID[:], other[:])
}

// Slice returns the InstanceID as a []byte.
func (iID InstanceID) Slice() []byte {
	return iID[:]
}

// Arguments is a searchable list of arguments.
type Arguments []Argument

// Search returns the value of a given argument. If it is not found, nil
// is returned.
// TODO: An argument with nil value cannot be distinguished from
// a missing argument!
func (args Arguments) Search(name string) []byte {
	for _, arg := range args {
		if arg.Name == name {
			return arg.Value
		}
	}
	return nil
}

// Hash computes the digest of the hash function
func (instr Instruction) Hash() []byte {
	h := sha256.New()
	h.Write(instr.InstanceID[:])
	h.Write(instr.Nonce[:])
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(instr.Index))
	h.Write(b)
	binary.LittleEndian.PutUint32(b, uint32(instr.Length))
	h.Write(b)
	var args []Argument
	switch instr.GetType() {
	case SpawnType:
		h.Write([]byte{0})
		h.Write([]byte(instr.Spawn.ContractID))
		args = instr.Spawn.Args
	case InvokeType:
		h.Write([]byte{1})
		args = instr.Invoke.Args
	case DeleteType:
		h.Write([]byte{2})
	}
	for _, a := range args {
		h.Write([]byte(a.Name))
		h.Write(a.Value)
	}
	return h.Sum(nil)
}

// DeriveID derives a new InstanceID from the hash of the instruction, its signatures,
// and the given string.
//
// DeriveID is used inside of contracts that need to create additional keys in
// the collection. By convention newly spawned instances should have their
// InstanceID derived via inst.DeriveID("").
func (instr Instruction) DeriveID(what string) InstanceID {
	var b [4]byte

	// Une petit primer on domain separation in hashing:
	// Domain separation is required when the input has variable lengths,
	// because an attacker could try to construct two messages resulting in the
	// same hash by moving bytes from one neighboring input to the
	// other. With fixed-length inputs, moving bytes is not possible, so
	// no domain separation is needed.

	h := sha256.New()
	h.Write(instr.Hash())

	binary.LittleEndian.PutUint32(b[:], uint32(len(instr.Signatures)))
	h.Write(b[:])

	for _, s := range instr.Signatures {
		binary.LittleEndian.PutUint32(b[:], uint32(len(s.Signature)))
		h.Write(b[:])
		h.Write(s.Signature)
		// TODO: Why not h.Write(s.Signer)
	}
	// Because there is no attacker-controlled input after what, we do not need
	// domain separation here.
	h.Write([]byte(what))

	return NewInstanceID(h.Sum(nil))

	// Addendum:
	//
	// While considering this we also considered the possibility that
	// allowing the attackers to mess with the signatures in order to
	// attempt to create InstanceID collisions is not a risk, since moving
	// a byte from sig[1] over to sig[0] would invalidate both signatures.
	// This is true for the Schnorr sigs we use today, but if there's some
	// other kind of data in the Signature field in the future, it might
	// be tolerant of mutations, meaning that what seems unrisky today could
	// be leaving a trap for later. So to be conservative, we are implementing
	// strict domain separation now.
}

// GetContractState searches for the contract kind of this instruction and the
// attached state to it. It needs the collection to do so.
//
// TODO: Deprecate/remove this; the state return is almost always ignored.
func (instr Instruction) GetContractState(coll CollectionView) (contractID string, state []byte, err error) {
	// Look the kind up in our database to find the kind.
	kv := coll.Get(instr.InstanceID.Slice())
	var record collection.Record
	record, err = kv.Record()
	if err != nil {
		return
	}
	var cv []interface{}
	cv, err = record.Values()
	if err != nil {
		if ConfigInstanceID.Equal(instr.InstanceID) {
			// Special case: first time call to genesis-configuration must return
			// correct contract type.
			return ContractConfigID, nil, nil
		}
		return
	}
	var ok bool
	var contractIDBuf []byte
	contractIDBuf, ok = cv[1].([]byte)
	if !ok {
		err = errors.New("failed to cast value to bytes")
		return
	}
	contractID = string(contractIDBuf)
	state, ok = cv[0].([]byte)
	if !ok {
		err = errors.New("failed to cast value to bytes")
		return
	}
	return
}

// Action returns the action that the user wants to do with this
// instruction.
func (instr Instruction) Action() string {
	a := "invalid"
	switch instr.GetType() {
	case SpawnType:
		a = "spawn:" + instr.Spawn.ContractID
	case InvokeType:
		a = "invoke:" + instr.Invoke.Command
	case DeleteType:
		a = "delete"
	}
	return a
}

// String returns a human readable form of the instruction.
func (instr Instruction) String() string {
	var out string
	out += fmt.Sprintf("instr: %x\n", instr.Hash())
	out += fmt.Sprintf("\tinstID: %v\n", instr.InstanceID)
	out += fmt.Sprintf("\tnonce: %x\n", instr.Nonce)
	out += fmt.Sprintf("\tindex: %d\n\tlength: %d\n", instr.Index, instr.Length)
	out += fmt.Sprintf("\taction: %s\n", instr.Action())
	out += fmt.Sprintf("\tsignatures: %d\n", len(instr.Signatures))
	return out
}

// SignBy gets one signature from each of the given signers
// and adds them into the Instruction.
func (instr *Instruction) SignBy(darcID darc.ID, signers ...darc.Signer) error {
	// Create the request and populate it with the right identities.  We
	// need to do this prior to signing because identities are a part of
	// the digest of the Instruction.
	sigs := make([]darc.Signature, len(signers))
	for i, signer := range signers {
		sigs[i].Signer = signer.Identity()
	}
	instr.Signatures = sigs

	req, err := instr.ToDarcRequest(darcID)
	if err != nil {
		return err
	}
	req.Identities = make([]darc.Identity, len(signers))
	for i := range signers {
		req.Identities[i] = signers[i].Identity()
	}

	// Sign the instruction and write the signatures to it.
	digest := req.Hash()
	instr.Signatures = make([]darc.Signature, len(signers))
	for i := range signers {
		sig, err := signers[i].Sign(digest)
		if err != nil {
			return err
		}
		instr.Signatures[i] = darc.Signature{
			Signature: sig,
			Signer:    signers[i].Identity(),
		}
	}
	return nil
}

// ToDarcRequest converts the Instruction content into a darc.Request.
func (instr Instruction) ToDarcRequest(baseID darc.ID) (*darc.Request, error) {
	action := instr.Action()
	ids := make([]darc.Identity, len(instr.Signatures))
	sigs := make([][]byte, len(instr.Signatures))
	for i, sig := range instr.Signatures {
		ids[i] = sig.Signer
		sigs[i] = sig.Signature // TODO shallow copy is ok?
	}
	var req darc.Request
	if action == "_evolve" {
		// We make a special case for darcs evolution because the Msg
		// part of the request must be the darc ID for verification to
		// pass.
		darcBuf := instr.Invoke.Args.Search("darc")
		d, err := darc.NewFromProtobuf(darcBuf)
		if err != nil {
			return nil, err
		}
		req = darc.NewRequest(baseID, darc.Action(action), d.GetID(), ids, sigs)
	} else {
		req = darc.NewRequest(baseID, darc.Action(action), instr.Hash(), ids, sigs)
	}
	return &req, nil
}

// VerifyDarcSignature will look up the darc of the instance pointed to by
// the instruction and then verify if the signature on the instruction
// can satisfy the rules of the darc. It returns an error if it couldn't
// find the darc or if the signature is wrong.
func (instr Instruction) VerifyDarcSignature(coll CollectionView) error {
	d, err := getInstanceDarc(coll, instr.InstanceID)
	if err != nil {
		return errors.New("darc not found: " + err.Error())
	}
	req, err := instr.ToDarcRequest(d.GetBaseID())
	if err != nil {
		return errors.New("couldn't create darc request: " + err.Error())
	}
	// Verify the request is signed by appropriate identities.
	// A callback is required to get any delegated DARC(s) during
	// expression evaluation.
	err = req.VerifyWithCB(d, func(str string, latest bool) *darc.Darc {
		if len(str) < 5 || string(str[0:5]) != "darc:" {
			return nil
		}
		darcID, err := hex.DecodeString(str[5:])
		if err != nil {
			return nil
		}
		d, err := LoadDarcFromColl(coll, darcID)
		if err != nil {
			return nil
		}
		return d
	})
	if err != nil {
		return errors.New("request verification failed: " + err.Error())
	}
	return nil
}

// Instructions is a slice of Instruction
type Instructions []Instruction

// Hash returns the sha256 hash of the hash of every instruction.
func (instrs Instructions) Hash() []byte {
	h := sha256.New()
	for _, instr := range instrs {
		h.Write(instr.Hash())
	}
	return h.Sum(nil)
}

// TxResults is a list of results from executed transactions.
type TxResults []TxResult

// NewTxResults takes a list of client transactions and wraps them up
// in a TxResults with Accepted set to false for each.
func NewTxResults(ct ...ClientTransaction) TxResults {
	out := make([]TxResult, len(ct))
	for i := range ct {
		out[i].ClientTransaction = ct[i]
	}
	return out
}

// Hash returns the sha256 hash of all of the transactions.
func (txr TxResults) Hash() []byte {
	one := []byte{1}
	zero := []byte{0}

	h := sha256.New()
	for _, tx := range txr {
		h.Write(tx.ClientTransaction.Instructions.Hash())
		if tx.Accepted {
			h.Write(one[:])
		} else {
			h.Write(zero[:])
		}
	}
	return h.Sum(nil)
}

// NewStateChange is a convenience function that fills out a StateChange
// structure.
func NewStateChange(sa StateAction, iID InstanceID, contractID string, value []byte, darcID darc.ID) StateChange {
	return StateChange{
		StateAction: sa,
		InstanceID:  iID.Slice(),
		ContractID:  []byte(contractID),
		Value:       value,
		DarcID:      darcID,
	}
}

func (sc StateChange) toString(withValue bool) string {
	var out string
	out += "\nstatechange\n"
	out += fmt.Sprintf("\taction: %s\n", sc.StateAction)
	out += fmt.Sprintf("\tcontractID: %s\n", string(sc.ContractID))
	out += fmt.Sprintf("\tkey: %x\n", sc.InstanceID)
	if withValue {
		out += fmt.Sprintf("\tvalue: %x", sc.Value)
	}
	return out
}

// String can be used in print.
func (sc StateChange) String() string {
	return sc.toString(true)
}

// ShortString is the same as String but excludes the value part.
func (sc StateChange) ShortString() string {
	return sc.toString(false)
}

// StateChanges hold a slice of StateChange
type StateChanges []StateChange

// Hash returns the sha256 of all stateChanges
func (scs StateChanges) Hash() []byte {
	h := sha256.New()
	for _, sc := range scs {
		scBuf, err := protobuf.Encode(&sc)
		if err != nil {
			log.Lvl2("Couldn't marshal transaction")
		}
		h.Write(scBuf)
	}
	return h.Sum(nil)
}

// ShortStrings outputs the ShortString of every state change.
func (scs StateChanges) ShortStrings() []string {
	out := make([]string, len(scs))
	for i, sc := range scs {
		out[i] = sc.ShortString()
	}
	return out
}

// StateAction describes how the collectionDB will be modified.
type StateAction int

const (
	// Create allows to insert a new key-value association.
	Create StateAction = iota + 1
	// Update allows to change the value of an existing key.
	Update
	// Remove allows to delete an existing key-value association.
	Remove
)

// String returns a readable output of the action.
func (sc StateAction) String() string {
	switch sc {
	case Create:
		return "Create"
	case Update:
		return "Update"
	case Remove:
		return "Remove"
	default:
		return "Invalid stateChange"
	}
}

// InstrType is the instruction type, which can be spawn, invoke or delete.
type InstrType int

const (
	// InvalidInstrType represents an error in the instruction type.
	InvalidInstrType InstrType = iota
	// SpawnType represents the spawn instruction type.
	SpawnType
	// InvokeType represents the invoke instruction type.
	InvokeType
	// DeleteType represents the delete instruction type.
	DeleteType
)

// GetType returns the type of the instruction.
func (instr Instruction) GetType() InstrType {
	if instr.Spawn != nil && instr.Invoke == nil && instr.Delete == nil {
		return SpawnType
	} else if instr.Spawn == nil && instr.Invoke != nil && instr.Delete == nil {
		return InvokeType
	} else if instr.Spawn == nil && instr.Invoke == nil && instr.Delete != nil {
		return DeleteType
	} else {
		return InvalidInstrType
	}
}

// txBuffer is thread-safe data structure that store client transactions.
type txBuffer struct {
	sync.Mutex
	txsMap map[string][]ClientTransaction
}

func newTxBuffer() txBuffer {
	return txBuffer{
		txsMap: make(map[string][]ClientTransaction),
	}
}

func (r *txBuffer) take(key string) []ClientTransaction {
	r.Lock()
	defer r.Unlock()

	txs, ok := r.txsMap[key]
	if !ok {
		return []ClientTransaction{}
	}
	delete(r.txsMap, key)
	return txs
}

func (r *txBuffer) add(key string, newTx ClientTransaction) {
	r.Lock()
	defer r.Unlock()

	if txs, ok := r.txsMap[key]; !ok {
		r.txsMap[key] = []ClientTransaction{newTx}
	} else {
		txs = append(txs, newTx)
		r.txsMap[key] = txs
	}
}
