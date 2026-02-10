package v2

import (
	"encoding/json"
	"testing"
)

func TestUnicastPeer(t *testing.T) {

	for _, b := range [][]byte{
		[]byte(`["a","b","c"]`),
		[]byte(`{"10.10.14.216": "","10.10.14.217": ""}`),
	} {
		var up unicastPeer
		if err := json.Unmarshal(b, &up); err != nil {
			t.Fatal(err)
		}

		t.Logf("%#v", up)

		if j, err := json.MarshalIndent(up, "", "  "); err != nil {
			t.Fatal(err)
		} else {
			t.Log(string(j))
		}
	}
}
