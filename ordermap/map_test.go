package ordermap

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMap(t *testing.T) {
	omEmpty := Map{}
	omEmptyJ, err := json.Marshal(omEmpty)
	require.Nil(t, err)
	require.Equal(t, []byte("{}"), omEmptyJ)

	om := Map{
		{
			Key: "bbb",
			Val: "bv",
		},
		{
			Key: "aaa",
			Val: "av",
		},
	}
	omJ, err := json.Marshal(om)
	require.Nil(t, err)
	require.Equal(t, []byte("{\"bbb\":\"bv\",\"aaa\":\"av\"}"), omJ)
}