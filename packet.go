package lifx

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
)

const (
	HeaderLen = 36

	PktGetPANgateway uint16 = 0x0002
	PktPANgateway    uint16 = 0x0003

	PktGetTime   uint16 = 0x0004
	PktSetTime   uint16 = 0x0005
	PktTimeState uint16 = 0x0006

	PktGetPowerState uint16 = 0x0014
	PktSetPowerState uint16 = 0x0015
	PktPowerState    uint16 = 0x0016

	PktGetLightState  uint16 = 0x0065
	PktSetLightColour uint16 = 0x0066
	PktLightState     uint16 = 0x006b

	PktGetAmbientLight   uint16 = 0x0191
	PktAmbientLightState uint16 = 0x0192

	PktGetTags uint16 = 0x001a
	PktSetTags uint16 = 0x001b
	PktTags    uint16 = 0x001c

	PktGetTagLabels uint16 = 0x001d
	PktSetTagLabels uint16 = 0x001e
	PktTagLabels    uint16 = 0x001f
)

type packetHeader struct {
	Size             uint16
	Protocol         uint16
	Reserved1        uint32
	TargetMacAddress [6]byte
	Reserved2        uint16
	Site             [6]byte
	Reserved3        uint16
	Timestamp        uint64
	PacketType       uint16
	Reserved4        uint16
}

func newPacketHeader(packetType uint16) *packetHeader {
	p := &packetHeader{}
	p.Size = 36
	p.Protocol = 0x3400
	p.Reserved1 = 0x0000
	p.Reserved2 = 0x00
	p.Reserved3 = 0x00
	p.PacketType = packetType
	p.Reserved4 = 0x00
	return p
}

func decodePacketHeader(buf []byte) (*packetHeader, error) {
	p := &packetHeader{}
	r := bytes.NewBuffer(buf)
	err := binary.Read(r, binary.LittleEndian, p)

	if err != nil {
		return nil, err
	}

	return p, err
}

func (p *packetHeader) Encode(wr io.Writer) (int, error) {
	buf := new(bytes.Buffer)

	err := binary.Write(buf, binary.LittleEndian, p)

	if err != nil {
		log.Fatalf("Woops %s", err)
	}

	//log.Printf("encode % x", buf.Bytes())

	return wr.Write(buf.Bytes())
}

func decodePayload(buf []byte, payload interface{}) error {
	r := bytes.NewBuffer(buf)
	return binary.Read(r, binary.LittleEndian, payload)
}
