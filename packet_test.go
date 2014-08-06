package lifx

import (
	"bytes"
	"encoding/hex"
	"reflect"
	"testing"
)

func TestPacketEncodeGetPANgateway(t *testing.T) {
	p := NewPacket(GetPANgateway)
	buf := new(bytes.Buffer)

	n, err := p.Encode(buf)

	if err != nil {
		t.Error(err)
	}

	if n != 36 {
		t.Fatalf("expected %d, got: %d", 36, n)
	}
	expBuf := getPANgatewayMsg()

	if !reflect.DeepEqual(expBuf, buf.Bytes()) {
		t.Fatalf("expected % x, got: % x", expBuf, buf.Bytes())

	}

}

func getPANgatewayMsg() []byte {
	buf, _ := hex.DecodeString("240000540000000000000000000000000000000000000000000000000000000002000000")
	return buf
}
