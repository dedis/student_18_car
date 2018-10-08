package byzcoin

import (
	"testing"
	"time"

	"github.com/dedis/cothority"
	"github.com/dedis/cothority/byzcoin/darc"
	"github.com/dedis/onet"
	"github.com/stretchr/testify/require"
)

func TestClient_GetProof(t *testing.T) {
	l := onet.NewTCPTest(cothority.Suite)
	servers, roster, _ := l.GenTree(3, true)
	registerDummy(servers)
	defer l.CloseAll()

	// Initialise the genesis message and send it to the service.
	signer := darc.NewSignerEd25519(nil, nil)
	msg, err := DefaultGenesisMsg(CurrentVersion, roster, []string{"spawn:dummy"}, signer.Identity())
	msg.BlockInterval = 100 * time.Millisecond
	require.Nil(t, err)

	// The darc inside it should be valid.
	d := msg.GenesisDarc
	require.Nil(t, d.Verify(true))

	c, csr, err := NewLedger(msg, false)
	require.Nil(t, err)

	// Create a new transaction.
	value := []byte{5, 6, 7, 8}
	kind := "dummy"
	tx, err := createOneClientTx(d.GetBaseID(), kind, value, signer)
	require.Nil(t, err)
	_, err = c.AddTransaction(tx)
	require.Nil(t, err)

	// We should have a proof of our transaction in the skipchain.
	newId := tx.Instructions[0].Hash()
	var p *GetProofResponse
	var i int
	for i = 0; i < 10; i++ {
		time.Sleep(4 * msg.BlockInterval)
		var err error
		p, err = c.GetProof(newId)
		if err != nil {
			continue
		}
		if p.Proof.InclusionProof.Match() {
			break
		}
	}
	require.NotEqual(t, 10, i, "didn't get proof in time")
	require.Nil(t, p.Proof.Verify(csr.Skipblock.SkipChainID()))
	k, vs, err := p.Proof.KeyValue()
	require.Nil(t, err)
	require.Equal(t, k, newId)
	require.Equal(t, value, vs[0])
}
