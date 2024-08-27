package state

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestContractsVersion(t *testing.T) {
	cases := []struct {
		in  ContractsVersion
		err bool
	}{
		{"local", false},
		{"latest", false},
		{"op-contracts/v1.2.3", false},
		{"3d166d5a3fba9a2fcba5b08c2b64636412e65f4b", false},
		{"op-contracts/v1.2.3.4", true},
		{"not-a-version", true},
	}
	for _, c := range cases {
		err := c.in.Check()
		if c.err {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}
