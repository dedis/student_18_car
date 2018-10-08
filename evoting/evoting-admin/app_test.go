package main

import (
	"testing"

	"github.com/dedis/cothority"
	"github.com/dedis/kyber/util/random"
	"github.com/dedis/onet/log"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	log.MainTest(m)
}

func TestParseKey(t *testing.T) {
	_, err := parseKey("r")
	assert.NotNil(t, err)

	_, err = parseKey("")
	assert.NotNil(t, err)

	p1 := cothority.Suite.Point().Pick(random.New())
	p2, _ := parseKey(p1.String())
	assert.True(t, p1.Equal(p2))
}

func TestParseAdmins(t *testing.T) {
	admins, err := parseAdmins("")
	assert.Nil(t, admins, err)

	_, err = parseAdmins("1,2,a,3")
	assert.NotNil(t, err)

	admins, _ = parseAdmins("1,2,3")
	assert.Equal(t, []uint32{1, 2, 3}, admins)
}
