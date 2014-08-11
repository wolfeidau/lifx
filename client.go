package lifx

import (
	"log"
	"net"
	"reflect"
	"time"
)

const (
	BroadcastPort = 56700
	PeerPort      = 56750

	bulbOff uint16 = 0
	bulbOn  uint16 = 1
)

var emptyAddr = [6]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0}

type Bulb struct {
	LifxAddress [6]byte // incoming messages are desimanated by lifx address
	Label       string
	State       BulbState
	LastSeen    time.Time
}

type BulbState struct {
	Hue        *uint16
	Saturation *uint16
	Brightness *uint16
	Kelvin     *uint16
	Dim        *uint16
	Power      *uint16
}

func newBulb(lifxAddress [6]byte) *Bulb {
	return &Bulb{LifxAddress: lifxAddress}
}

func (b *Bulb) setLabel(label string) {
	b.Label = label
}

func (b *Bulb) updateState(hue, saturation, brightness, kelvin, dim, power *uint16) {
	if hue != nil {
		b.State.Hue = hue
	}
	if saturation != nil {
		b.State.Saturation = saturation
	}
	if brightness != nil {
		b.State.Brightness = brightness
	}
	if kelvin != nil {
		b.State.Kelvin = kelvin
	}
	if dim != nil {
		b.State.Dim = dim
	}
	if power != nil {
		b.State.Power = power
	}
}

type gateway struct {
	lifxAddress [6]byte
	hostAddress string
	Port        uint16
	Site        [6]byte // incoming messages are desimanated by site
	lastSeen    time.Time
	socket      *net.UDPConn
}

func newGateway(lifxAddress [6]byte, hostAddress string, port uint16, site [6]byte) (*gateway, error) {

	gw := &gateway{lifxAddress: lifxAddress, hostAddress: hostAddress, Port: port, Site: site}

	// can we connect to the gw
	addr, err := net.ResolveUDPAddr("udp4", gw.hostAddress)

	if err != nil {
		return nil, err
	}

	// open the connection, which we retain for all peer -> globe comms
	gw.socket, err = net.DialUDP("udp4", nil, addr)

	if err != nil {
		return nil, err
	}

	return gw, nil
}

func (g *gateway) sendTo(cmd command) error {

	// send to globe
	_, err := cmd.WriteTo(g.socket)

	if err != nil {
		return err
	}

	log.Printf("Sent command to gateway %s", reflect.TypeOf(cmd))

	return nil
}

func (g *gateway) findBulbs() error {

	// get Light State
	lcmd := newGetLightStateCommand(g.Site)

	err := g.sendTo(lcmd)

	if err != nil {
		return err
	}

	tcmd := newGetTagsCommand(g.Site)

	err = g.sendTo(tcmd)

	if err != nil {
		return err
	}

	return nil
}

type Client struct {
	gateways      []*gateway
	bulbs         []*Bulb
	intervalID    int
	DiscoInterval int

	peerSocket  net.Conn
	bcastSocket *net.UDPConn
	discoTicker *time.Ticker
}

func (c *Client) StartDiscovery() (err error) {

	log.Printf("Listening for bcast :%d", BroadcastPort)

	// this socket will recieve broadcast packets on this socket
	c.bcastSocket, err = net.ListenUDP("udp4", &net.UDPAddr{Port: BroadcastPort})

	if err != nil {
		return
	}

	c.discoTicker = time.NewTicker(time.Second * 10)

	// once you pop you can't stop
	go c.startMainEventLoop()

	// once you pop you can't stop
	go func() {

		c.sendDiscovery(time.Now())

		for t := range c.discoTicker.C {
			c.sendDiscovery(t)
		}

	}()

	return
}

func (b *Client) LightsOn() error {

	cmd := newSetPowerStateCommand(bulbOn)

	return b.sendToAll(cmd)
}

func (b *Client) LightsOff() error {

	cmd := newSetPowerStateCommand(bulbOff)

	return b.sendToAll(cmd)
}

func (b *Client) LightsColour(hue uint16, sat uint16, lum uint16, kelvin uint16, timing uint32) error {
	cmd := newSetLightColour(hue, sat, lum, kelvin, timing)

	return b.sendToAll(cmd)
}

func (b *Client) LightOn(bulb *Bulb) error {

	cmd := newSetPowerStateCommand(bulbOn)

	return b.sendTo(bulb, cmd)
}

func (b *Client) LightOff(bulb *Bulb) error {

	cmd := newSetPowerStateCommand(bulbOff)

	return b.sendTo(bulb, cmd)
}

// func (b *Bulb) updateState(hue, saturation, brightness, kelvin, dim, power *uint16, label *string) {

func (b *Client) LightColour(bulb *Bulb, hue *uint16, sat *uint16, lum *uint16, kelvin *uint16, timing *uint32) error {
	bulb.updateState(hue, sat, lum, kelvin, nil, nil)
	cmd := newSetLightColour(*hue, *sat, *lum, *kelvin, *timing)
	return b.sendTo(bulb, cmd)
}

func (b *Client) GetBulbs() []*Bulb {
	return b.bulbs
}

func (c *Client) sendTo(bulb *Bulb, cmd command) error {

	cmd.SetLifxAddr(bulb.LifxAddress) // ensure the message is addressed to the correct bulb

	for _, gw := range c.gateways {
		log.Printf("sending command to %s", gw.hostAddress)
		cmd.SetSiteAddr(gw.Site) // update the site address for each gateway
		_, err := cmd.WriteTo(gw.socket)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) sendToAll(cmd command) error {
	for _, gw := range c.gateways {
		log.Printf("sending command to %s", gw.hostAddress)
		cmd.SetSiteAddr(gw.Site) // update the site address so all globes change
		_, err := cmd.WriteTo(gw.socket)
		if err != nil {
			return err
		}
	}
	return nil
}

// This function handles all response messages and dispatches events subscribers
func (c *Client) startMainEventLoop() {

	buf := make([]byte, 1024)

	for {
		n, addr, err := c.bcastSocket.ReadFrom(buf)

		if err != nil {
			log.Fatalf("Woops %s", err)
		}

		//		log.Printf("Received buffer from %+v of %x", addr, buf[:n])

		cmd, err := decodeCommand(buf[:n])

		if err != nil {
			log.Printf("Error processing command: %v", err)
			continue
		}
		log.Printf("Recieved command: %s", reflect.TypeOf(cmd))

		switch cmd := cmd.(type) {
		case *panGatewayCommand:

			// found a gw
			if cmd.Payload.Service == 1 {
				gw, err := newGateway(cmd.Header.TargetMacAddress, addr.String(), cmd.Payload.Port, cmd.Header.Site)
				if err != nil {
					log.Printf("failed to setup peer connection to gw")
				} else {
					c.addGateway(gw)
				}
			}

		case *lightStateCommand:

			// found a bulb
			bulb := newBulb(cmd.Header.TargetMacAddress)

			label := string(cmd.Payload.BulbLabel[:])
			bulb.setLabel(label)
			bulb.updateState(&cmd.Payload.Hue, &cmd.Payload.Saturation, &cmd.Payload.Brightness, &cmd.Payload.Kelvin, &cmd.Payload.Dim, &cmd.Payload.Power)

			c.addBulb(bulb)

		case *powerStateCommand:

			c.updateBulbPowerState(cmd.Header.TargetMacAddress, cmd.Payload.OnOff)

		default:
		}

	}
}

func (c *Client) sendDiscovery(t time.Time) {

	log.Println("Discovery packet sent at", t)
	socket, err := net.DialUDP("udp4", nil, &net.UDPAddr{
		IP:   net.IPv4(255, 255, 255, 255),
		Port: BroadcastPort,
	})

	if err != nil {
		return
	}

	defer socket.Close()

	p := newPacketHeader(PktGetPANgateway)
	_, _ = p.Encode(socket)

	//	log.Printf("Bcast sent %d", n)

	log.Printf("gateways %v", c.gateways)
	log.Printf("bulbs %v", c.bulbs)

}

func (c *Client) addGateway(gw *gateway) {
	if !gatewayInSlice(gw, c.gateways) {
		log.Printf("Added gw %v", gw)
		gw.lastSeen = time.Now()
		c.gateways = append(c.gateways, gw)
	} else {
		for _, lgw := range c.gateways {
			if gw.lifxAddress == lgw.lifxAddress && gw.Port == lgw.Port && gw.hostAddress == lgw.hostAddress {
				gw.lastSeen = time.Now()
				log.Printf("update last seen for %v %s", gw.hostAddress, gw.lastSeen)
			}
		}
	}

	gw.findBulbs()
}

func (c *Client) addBulb(bulb *Bulb) {
	if !bulbInSlice(bulb, c.bulbs) {
		log.Printf("Added bulb %v", bulb)
		c.bulbs = append(c.bulbs, bulb)
		bulb.LastSeen = time.Now()
	}
	for _, lbulb := range c.bulbs {
		if bulb.LifxAddress == lbulb.LifxAddress {
			lbulb.LastSeen = time.Now()
		}
	}
}

func (c *Client) updateBulbPowerState(lifxAddress [6]byte, onoff uint16) {
	for _, b := range c.bulbs {
		// this needs further investigation
		if lifxAddress == b.LifxAddress {
			b.State.Power = &onoff
		}
	}
}

func gatewayInSlice(a *gateway, list []*gateway) bool {
	for _, b := range list {
		// this needs further investigation
		if a.lifxAddress == b.lifxAddress && a.Port == b.Port && a.hostAddress == b.hostAddress {
			return true
		}
	}
	return false
}

func bulbInSlice(a *Bulb, list []*Bulb) bool {
	for _, b := range list {
		// this needs further investigation
		if a.LifxAddress == b.LifxAddress {
			return true
		}
	}
	return false
}

func NewClient() *Client {
	return &Client{}
}
