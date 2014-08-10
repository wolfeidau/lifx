package lifx

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"log"

	"github.com/davecgh/go-spew/spew"
)

type command interface {
	SetSiteAddr(site [6]byte)
	SetLifxAddr(addr [6]byte)
	WriteTo(wr io.Writer) (int, error)
}

func decodeCommand(buf []byte) (command, error) {

	// read and validate the packet header
	ph, err := decodePacketHeader(buf)

	if err != nil {
		return nil, err
	}

	switch ph.PacketType {
	case PktPANgateway:
		return decodePANGatewayCommand(ph, buf[HeaderLen:])
	case PktLightState:
		return decodeLightStateCommand(ph, buf[HeaderLen:])
	case PktPowerState:
		return decodePowerStateCommand(ph, buf[HeaderLen:])
	case PktTags:
		return decodeTagsCommand(ph, buf[HeaderLen:])
	}

	return nil, errors.New("command not found")
}

type commandPacket struct {
	Header *packetHeader
}

func (c *commandPacket) SetSiteAddr(site [6]byte) {
	c.Header.Site = site
}

func (c *commandPacket) SetLifxAddr(addr [6]byte) {
	c.Header.TargetMacAddress = addr
}

func (c *commandPacket) WriteTo(wr io.Writer) (int, error) {
	return writeHeaderOnly(c.Header, wr)
}

// GetPANGatewayCommand 0x02
type getPANGatewayCommand struct {
	commandPacket
}

func newGetPANGatewayCommand() *getPANGatewayCommand {
	cmd := &getPANGatewayCommand{}
	cmd.Header = newPacketHeader(PktGetPANgateway)
	return cmd
}

type panGatewayCommand struct {
	commandPacket
	Payload struct {
		Service uint8
		Port    uint16
	}
}

func decodePANGatewayCommand(ph *packetHeader, payload []byte) (*panGatewayCommand, error) {
	cmd := &panGatewayCommand{}

	cmd.Header = ph

	// decode payload
	log.Printf("payload len : %d", len(payload))
	decodePayload(payload, &cmd.Payload)

	log.Printf("Command: \n %s", spew.Sdump(cmd))

	return cmd, nil
}

// GetLightStateCommand 0x65
type getLightStateCommand struct {
	commandPacket
}

func newGetLightStateCommand(site [6]byte) *getLightStateCommand {
	ph := newPacketHeader(PktGetLightState)
	ph.Protocol = 13312
	ph.Site = site

	cmd := &getLightStateCommand{}
	cmd.Header = ph
	return cmd
}

// LightStateCommand 0x6b
type lightStateCommand struct {
	commandPacket
	Payload struct {
		Hue        uint16
		Saturation uint16
		Brightness uint16
		Kelvin     uint16
		Dim        uint16
		Power      uint16
		BulbLabel  [32]byte
		Tags       uint64
	}
}

func decodeLightStateCommand(ph *packetHeader, payload []byte) (*lightStateCommand, error) {
	cmd := &lightStateCommand{}
	cmd.Header = ph

	// decode payload
	log.Printf("payload len : %d", len(payload))
	decodePayload(payload, &cmd.Payload)

	log.Printf("Command: \n %s", spew.Sdump(cmd))

	return cmd, nil
}

// SetLightColour 0x66
type setLightColour struct {
	commandPacket
	Payload struct {
		Stream     uint8
		Hue        uint16
		Saturation uint16
		Brightness uint16
		Kelvin     uint16
		Dim        uint32
	}
}

func newSetLightColour(hue uint16, sat uint16, lum uint16, kelvin uint16, timing uint32) *setLightColour {
	ph := newPacketHeader(PktSetLightColour)
	ph.Protocol = 13312
	ph.Size = 49

	cmd := &setLightColour{}

	cmd.Header = ph

	cmd.Payload.Hue = hue
	cmd.Payload.Saturation = sat
	cmd.Payload.Brightness = lum
	cmd.Payload.Kelvin = kelvin
	cmd.Payload.Dim = timing

	return cmd
}

func (c *setLightColour) WriteTo(wr io.Writer) (int, error) {

	buf := new(bytes.Buffer)

	err := binary.Write(buf, binary.LittleEndian, &c.Payload)

	if err != nil {
		return 0, err
	}

	return writeHeaderAndPayload(c.Header, buf.Bytes(), wr)
}

// GetPowerStateCommand 0x14
type getPowerStateCommand struct {
	commandPacket
}

func newGetPowerStateCommand(site [6]byte, lifxAddress [6]byte) *getPowerStateCommand {
	ph := newPacketHeader(PktGetPowerState)
	ph.Protocol = 13312
	ph.Site = site
	ph.TargetMacAddress = lifxAddress

	cmd := &getPowerStateCommand{}
	cmd.Header = ph
	return cmd
}

// SetPowerStateCommand 0x15
type setPowerStateCommand struct {
	commandPacket
	Payload struct {
		OnOff uint16
	}
}

func newSetPowerStateCommand(onoff uint16) *setPowerStateCommand {
	ph := newPacketHeader(PktSetPowerState)

	ph.Size = 38
	ph.Protocol = 13312

	cmd := &setPowerStateCommand{}
	cmd.Header = ph
	cmd.Payload.OnOff = onoff

	return cmd
}

func (c *setPowerStateCommand) WriteTo(wr io.Writer) (int, error) {

	buf := []byte{0x0, 0x0}

	binary.BigEndian.PutUint16(buf, c.Payload.OnOff)

	return writeHeaderAndPayload(c.Header, buf, wr)
}

// PowerStateCommand 0x16
type powerStateCommand struct {
	commandPacket
	Payload struct {
		OnOff uint16
	}
}

func decodePowerStateCommand(ph *packetHeader, payload []byte) (*powerStateCommand, error) {
	cmd := &powerStateCommand{}
	cmd.Header = ph

	// decode payload
	log.Printf("payload len : %d", len(payload))
	decodePayload(payload, &cmd.Payload)

	log.Printf("Command: \n %s", spew.Sdump(cmd))

	return cmd, nil
}

// GetTagsCommand 0x1a
type getTagsCommand struct {
	commandPacket
}

func newGetTagsCommand(site [6]byte) *getTagsCommand {
	ph := newPacketHeader(PktGetTags)
	ph.Protocol = 13312
	ph.Site = site
	cmd := &getTagsCommand{}
	cmd.Header = ph
	return cmd
}

// TagsCommand 0x1c
type tagsCommand struct {
	commandPacket
}

func decodeTagsCommand(ph *packetHeader, payload []byte) (*tagsCommand, error) {

	cmd := &tagsCommand{}
	cmd.Header = ph

	// decode payload
	log.Printf("payload len : %d", len(payload))

	return cmd, nil
}

func writeHeaderOnly(h *packetHeader, wr io.Writer) (int, error) {
	buf := new(bytes.Buffer)
	n, err := h.Encode(buf)

	if err != nil {
		return n, err
	}

	return wr.Write(buf.Bytes())
}

func writeHeaderAndPayload(h *packetHeader, payload []byte, wr io.Writer) (int, error) {
	buf := new(bytes.Buffer)
	n, err := h.Encode(buf)

	if err != nil {
		return n, err
	}

	_, err = buf.Write(payload)

	if err != nil {
		return n, err
	}

	return wr.Write(buf.Bytes())
}
