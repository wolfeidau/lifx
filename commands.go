package lifx

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"log"

	"github.com/davecgh/go-spew/spew"
)

type Command interface {
	SetSiteAddr(site [6]byte)
	SetLifxAddr(addr [6]byte)
	WriteTo(wr io.Writer) (int, error)
}

func DecodeCommand(buf []byte) (Command, error) {

	// read and validate the packet header
	ph, err := DecodePacketHeader(buf)

	if err != nil {
		return nil, err
	}

	switch ph.PacketType {
	case PktPANgateway:
		return DecodePANGatewayCommand(ph, buf[HeaderLen:])
	case PktLightState:
		return DecodeLightStateCommand(ph, buf[HeaderLen:])
	case PktPowerState:
		return DecodePowerStateCommand(ph, buf[HeaderLen:])
	case PktTags:
		return DecodeTagsCommand(ph, buf[HeaderLen:])
	}

	return nil, errors.New("command not found")
}

type CommandPacket struct {
	Header *PacketHeader
}

func (c *CommandPacket) SetSiteAddr(site [6]byte) {
	c.Header.Site = site
}

func (c *CommandPacket) SetLifxAddr(addr [6]byte) {
	c.Header.TargetMacAddress = addr
}

func (c *CommandPacket) WriteTo(wr io.Writer) (int, error) {
	return writeHeaderOnly(c.Header, wr)
}

// GetPANGatewayCommand 0x02
type GetPANGatewayCommand struct {
	CommandPacket
}

func NewGetPANGatewayCommand() *GetPANGatewayCommand {
	cmd := &GetPANGatewayCommand{}
	cmd.Header = NewPacketHeader(PktGetPANgateway)
	return cmd
}

type PANGatewayCommand struct {
	CommandPacket
	Payload struct {
		Service uint8
		Port    uint16
	}
}

func DecodePANGatewayCommand(ph *PacketHeader, payload []byte) (*PANGatewayCommand, error) {
	cmd := &PANGatewayCommand{}

	cmd.Header = ph

	// decode payload
	log.Printf("payload len : %d", len(payload))
	DecodePayload(payload, &cmd.Payload)

	log.Printf("Command: \n %s", spew.Sdump(cmd))

	return cmd, nil
}

// GetLightStateCommand 0x65
type GetLightStateCommand struct {
	CommandPacket
}

func NewGetLightStateCommand(site [6]byte) *GetLightStateCommand {
	ph := NewPacketHeader(PktGetLightState)
	ph.Protocol = 13312
	ph.Site = site

	cmd := &GetLightStateCommand{}
	cmd.Header = ph
	return cmd
}

// LightStateCommand 0x6b
type LightStateCommand struct {
	CommandPacket
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

func DecodeLightStateCommand(ph *PacketHeader, payload []byte) (*LightStateCommand, error) {
	cmd := &LightStateCommand{}
	cmd.Header = ph

	// decode payload
	log.Printf("payload len : %d", len(payload))
	DecodePayload(payload, &cmd.Payload)

	log.Printf("Command: \n %s", spew.Sdump(cmd))

	return cmd, nil
}

// SetLightColour 0x66
type SetLightColour struct {
	CommandPacket
	Payload struct {
		Stream     uint8
		Hue        uint16
		Saturation uint16
		Brightness uint16
		Kelvin     uint16
		Dim        uint32
	}
}

func NewSetLightColour(hue uint16, sat uint16, lum uint16, kelvin uint16, timing uint32) *SetLightColour {
	ph := NewPacketHeader(PktSetLightColour)
	ph.Protocol = 13312
	ph.Size = 49

	cmd := &SetLightColour{}

	cmd.Header = ph

	cmd.Payload.Hue = hue
	cmd.Payload.Saturation = sat
	cmd.Payload.Brightness = lum
	cmd.Payload.Kelvin = kelvin
	cmd.Payload.Dim = timing

	return cmd
}

func (c *SetLightColour) WriteTo(wr io.Writer) (int, error) {

	buf := new(bytes.Buffer)

	err := binary.Write(buf, binary.LittleEndian, &c.Payload)

	if err != nil {
		return 0, err
	}

	return writeHeaderAndPayload(c.Header, buf.Bytes(), wr)
}

// GetPowerStateCommand 0x14
type GetPowerStateCommand struct {
	CommandPacket
}

func NewGetPowerStateCommand(site [6]byte, lifxAddress [6]byte) *GetPowerStateCommand {
	ph := NewPacketHeader(PktGetPowerState)
	ph.Protocol = 13312
	ph.Site = site
	ph.TargetMacAddress = lifxAddress

	cmd := &GetPowerStateCommand{}
	cmd.Header = ph
	return cmd
}

// SetPowerStateCommand 0x15
type SetPowerStateCommand struct {
	CommandPacket
	Payload struct {
		OnOff uint16
	}
}

func NewSetPowerStateCommand(onoff uint16) *SetPowerStateCommand {
	ph := NewPacketHeader(PktSetPowerState)

	ph.Size = 38
	ph.Protocol = 13312

	cmd := &SetPowerStateCommand{}
	cmd.Header = ph
	cmd.Payload.OnOff = onoff

	return cmd
}

func (c *SetPowerStateCommand) WriteTo(wr io.Writer) (int, error) {

	buf := []byte{0x0, 0x0}

	binary.BigEndian.PutUint16(buf, c.Payload.OnOff)

	return writeHeaderAndPayload(c.Header, buf, wr)
}

// PowerStateCommand 0x16
type PowerStateCommand struct {
	CommandPacket
	Payload struct {
		OnOff uint16
	}
}

func DecodePowerStateCommand(ph *PacketHeader, payload []byte) (*PowerStateCommand, error) {
	cmd := &PowerStateCommand{}
	cmd.Header = ph

	// decode payload
	log.Printf("payload len : %d", len(payload))
	DecodePayload(payload, &cmd.Payload)

	log.Printf("Command: \n %s", spew.Sdump(cmd))

	return cmd, nil
}

// GetTagsCommand 0x1a
type GetTagsCommand struct {
	CommandPacket
}

func NewGetTagsCommand(site [6]byte) *GetTagsCommand {
	ph := NewPacketHeader(PktGetTags)
	ph.Protocol = 13312
	ph.Site = site
	cmd := &GetTagsCommand{}
	cmd.Header = ph
	return cmd
}

// TagsCommand 0x1c
type TagsCommand struct {
	CommandPacket
}

func DecodeTagsCommand(ph *PacketHeader, payload []byte) (*TagsCommand, error) {

	cmd := &TagsCommand{}
	cmd.Header = ph

	// decode payload
	log.Printf("payload len : %d", len(payload))

	return cmd, nil
}

func writeHeaderOnly(h *PacketHeader, wr io.Writer) (int, error) {
	buf := new(bytes.Buffer)
	n, err := h.Encode(buf)

	if err != nil {
		return n, err
	}

	return wr.Write(buf.Bytes())
}

func writeHeaderAndPayload(h *PacketHeader, payload []byte, wr io.Writer) (int, error) {
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
