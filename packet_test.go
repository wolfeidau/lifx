package lifx

import (
	"bytes"
	"encoding/hex"
	"log"
	"reflect"
	"testing"
)

func TestPacketEncodeGetPANgateway(t *testing.T) {
	p := NewPacketHeader(GetPANgateway)
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

func TestPacketDecodingPANgateway(t *testing.T) {
	buf := PANgatewayMsg()
	p, err := DecodePacketHeader(buf)
	if err != nil {
		t.Error(err)
	}

	if p.Packet_type != PANgateway {
		t.Fatalf("expected % x, got: % x", PANgateway, p.Packet_type)
	}

	payload, err := NewPANgatewayPayload(buf[36:])
	if err != nil {
		t.Error(err)
	}

	log.Printf("msg %+v", payload)

}

func getPANgatewayMsg() []byte {
	buf, _ := hex.DecodeString("240000540000000000000000000000000000000000000000000000000000000002000000")
	return buf
}

func PANgatewayMsg() []byte {
	buf, _ := hex.DecodeString("2900005400000000d073d50035f70000d073d50035f70000000000000000000003000000017cdd0000")
	return buf
}
