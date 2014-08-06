package lifx

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
)

const (
	GetPANgateway uint16 = 0x02
	PANgateway    uint16 = 0x03
)

type Packet struct {
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

func NewPacket(packetType uint16) *Packet {
	p := &Packet{}
	p.Size = 36
	p.Protocol = 21504
	p.Reserved1 = 0x0000
	p.Reserved2 = 0x00
	p.Reserved3 = 0x00
	p.Packet_type = packetType
	p.Reserved4 = 0x00

	return p
}

func (p *Packet) Encode(wr io.Writer) (int, error) {

	buf := new(bytes.Buffer)

	err := binary.Write(buf, binary.LittleEndian, p)

	if err != nil {
		log.Fatalf("Woops %s", err)
	}

	return wr.Write(buf.Bytes())
}
