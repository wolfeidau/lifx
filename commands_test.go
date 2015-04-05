package lifx

import (
	"bytes"
	"reflect"
	"testing"
)

func TestGetPANGatewayCommandWrite(t *testing.T) {
	buf := new(bytes.Buffer)

	c := newGetPANGatewayCommand()

	n, err := c.WriteTo(buf)

	if err != nil {
		t.Error(err)
	}

	if n != HeaderLen {
		t.Fatalf("expected %d, got: %d", 36, n)
	}
	expBuf := getPANgatewayMsg()

	if !reflect.DeepEqual(expBuf, buf.Bytes()) {
		t.Fatalf("expected % x, got: % x", expBuf, buf.Bytes())
	}
}

func TestPANGatewayCommandDecode(t *testing.T) {
	buf := panGatewayMsg()

	cmd, err := decodeCommand(buf)

	if err != nil {
		t.Error(err)
	}

	expSite := [6]byte{0xd0, 0x73, 0xd5, 0x00, 0x35, 0xf7}

	if err != nil {
		t.Error(err)
	}

	switch cmd := cmd.(type) {
	case *panGatewayCommand:
		if !reflect.DeepEqual(expSite, cmd.Header.Site) {
			t.Fatalf("expected % x, got: % x", expSite, cmd.Header.Site)
		}
	default:
		t.Fatal("expected panGatewayCommand")
	}

}

func TestSetPowerStateCommandWrite(t *testing.T) {
	buf := new(bytes.Buffer)
	addr := [6]byte{0xd0, 0x73, 0xd5, 0x00, 0x35, 0xf7}
	//	site := [6]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	c := newSetPowerStateCommand(bulbOn)

	//c.SetLifxAddr(addr)
	c.SetSiteAddr(addr)

	n, err := c.WriteTo(buf)

	if err != nil {
		t.Error(err)
	}

	if n != 38 {
		t.Fatalf("expected %d, got: %d", 38, n)
	}
	expBuf := setPowerStateMsg()

	if !reflect.DeepEqual(expBuf, buf.Bytes()) {
		t.Fatalf("expected:\n % x, got:\n % x", expBuf, buf.Bytes())
	}
}

func TestPowerStateCommandDecode(t *testing.T) {
	buf := powerStateMsg()

	cmd, err := decodeCommand(buf)

	if err != nil {
		t.Error(err)
	}

	expSite := [6]byte{0xd0, 0x73, 0xd5, 0x00, 0x35, 0xf7}

	if err != nil {
		t.Error(err)
	}

	switch cmd := cmd.(type) {
	case *powerStateCommand:
		if !reflect.DeepEqual(expSite, cmd.Header.Site) {
			t.Fatalf("expected % x, got: % x", expSite, cmd.Header.Site)
		}
	default:
		t.Fatal("expected panGatewayCommand")
	}
}
