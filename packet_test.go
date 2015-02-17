package lifx

import (
	"bytes"
	"encoding/hex"
	"reflect"
	"testing"
)

func TestPacketEncodeGetPANgateway(t *testing.T) {
	p := newPacketHeader(PktGetPANgateway)
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

// func TestPacketDecodingPANgateway(t *testing.T) {
// 	buf := panGatewayMsg()
// 	p, err := DecodePacketHeader(buf)
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	if p.Packet_type != PANgateway {
// 		t.Fatalf("expected % x, got: % x", PANgateway, p.Packet_type)
// 	}

// 	payload, err := NewPANgatewayPayload(buf[36:])
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	log.Printf("msg %+v", payload)

// }

// Get PAN Gateway
func getPANgatewayMsg() []byte {
	buf, _ := hex.DecodeString("240000340000000000000000000000000000000000000000000000000000000002000000")
	return buf
}

// PAN Gateway
func panGatewayMsg() []byte {
	buf, _ := hex.DecodeString("2900003400000000d073d50035f70000d073d50035f70000000000000000000003000000017cdd0000")
	return buf
}

// Get Light Status
func getLightStatusMsg() []byte {
	buf, _ := hex.DecodeString("24000014000000000000000000000000d073d50035f70000000000000000000065000000")
	return buf
}

// Light status
func lightStatusMsg() []byte {
	buf, _ := hex.DecodeString("5800005400000000d073d50035f70000d073d50035f7000000000000000000006b00000000000000ffffac0d0000ffff00000000000000000000000000000000000000000000000000000000000000000000000000000000")
	return buf
}

// set power state
func setPowerStateMsg() []byte {
	buf, _ := hex.DecodeString("26000014000000000000000000000000d073d50035f700000000000000000000150000000001")
	return buf
}

// power state
func powerStateMsg() []byte {
	buf, _ := hex.DecodeString("2600005400000000d073d50035f70000d073d50035f70000000000000000000016000000ffff")
	return buf
}
