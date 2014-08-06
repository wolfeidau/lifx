package lifx

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
)

const (
	HeaderLen = 36

	GetPANgateway uint16 = 0x02
	PANgateway    uint16 = 0x03
)

type PacketHeader struct {
	Size               uint16
	Protocol           uint16
	Reserved1          uint32
	Target_mac_address [6]byte
	Reserved2          uint16
	Site               [6]byte
	Reserved3          uint16
	Timestamp          uint64
	Packet_type        uint16
	Reserved4          uint16
}

func NewPacketHeader(packetType uint16) *PacketHeader {
	p := &PacketHeader{}
	p.Size = 36
	p.Protocol = 21504
	p.Reserved1 = 0x0000
	p.Reserved2 = 0x00
	p.Reserved3 = 0x00
	p.Packet_type = packetType
	p.Reserved4 = 0x00
	return p
}

func DecodePacketHeader(buf []byte) (*PacketHeader, error) {
	p := &PacketHeader{}
	r := bytes.NewBuffer(buf)
	err := binary.Read(r, binary.LittleEndian, p)

	if err != nil {
		return nil, err
	}

	return p, err
}

func (p *PacketHeader) Encode(wr io.Writer) (int, error) {

	buf := new(bytes.Buffer)

	err := binary.Write(buf, binary.LittleEndian, p)

	if err != nil {
		log.Fatalf("Woops %s", err)
	}

	log.Printf("encode % x", buf.Bytes())

	return wr.Write(buf.Bytes())
}

type PANgatewayPayload struct {
	Service uint8
	Port    uint16
}

func NewPANgatewayPayload(buf []byte) (*PANgatewayPayload, error) {
	p := &PANgatewayPayload{}
	r := bytes.NewBuffer(buf)
	err := binary.Read(r, binary.LittleEndian, p)
	if err != nil {
		return nil, err
	}
	return p, err
}
