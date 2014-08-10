package lifx

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
)

const (
	HeaderLen = 36

	PktGetPANgateway uint16 = 0x02
	PktPANgateway    uint16 = 0x03

	PktGetTime   uint16 = 0x04
	PktSetTime   uint16 = 0x05
	PktTimeState uint16 = 0x06

	PktGetPowerState uint16 = 0x14
	PktSetPowerState uint16 = 0x15
	PktPowerState    uint16 = 0x16

	PktGetLightState  uint16 = 0x65
	PktSetLightColour uint16 = 0x66
	PktLightState     uint16 = 0x6b

	PktGetTags uint16 = 0x1a
	PktSetTags uint16 = 0x1b
	PktTags    uint16 = 0x1c
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
	p.Protocol = 21504
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

	log.Printf("encode % x", buf.Bytes())

	return wr.Write(buf.Bytes())
}

func decodePayload(buf []byte, payload interface{}) error {
	r := bytes.NewBuffer(buf)
	return binary.Read(r, binary.LittleEndian, payload)
}
