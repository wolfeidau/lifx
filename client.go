package lifx

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"reflect"
	"time"
)

const (
	// BroadcastPort port used for broadcasting messages to lifx globes
	BroadcastPort = 56700

	// PeerPort port used for peer to peer messages to lifx globes
	PeerPort = 56750

	bulbOff uint16 = 0
	bulbOn  uint16 = 1
)

var emptyAddr = [6]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0}

// StateHandler this is called when there is a change in the state of a bulb
type StateHandler func(newState *BulbState)

// Bulb Holds the state for a lifx bulb
type Bulb struct {
	LifxAddress      [6]byte // incoming messages are desimanated by lifx address
	bulbState        *BulbState
	lightSensorState *LightSensorState
	stateHandler     StateHandler

	lastLightState *lightStateCommand
	LastSeen       time.Time
}

func newBulb(lifxAddress [6]byte) *Bulb {
	return &Bulb{LifxAddress: lifxAddress}
}

// GetState Get a *snapshot* of the state for the bulb
func (b *Bulb) GetState() BulbState {
	return *b.bulbState
}

// GetLifxAddress returns the unique lifx bulb address
func (b *Bulb) GetLifxAddress() string {
	return fmt.Sprintf("%x", b.LifxAddress)
}

// GetPower Is the globe powered on or off
func (b *Bulb) GetPower() uint16 {
	return b.bulbState.Power
}

// GetLabel Get the label from the globe
func (b *Bulb) GetLabel() string {
	return string(bytes.Trim(b.lastLightState.Payload.BulbLabel[:], "\x00"))
}

// SetStateHandler add a handler which is invoked each time a state change comes through
func (b *Bulb) SetStateHandler(handler StateHandler) {
	//log.Printf("bulb %s", b)
	b.stateHandler = handler
}

func (b *Bulb) update(bulb *Bulb) bool {
	if !reflect.DeepEqual(b.bulbState, bulb.bulbState) {

		// log.Printf("Updated bulb %v", lbulb)
		b.LastSeen = time.Now()

		// update the state
		b.bulbState = bulb.bulbState

		if b.stateHandler != nil {
			b.stateHandler(b.bulbState)
		}

		return true
	}
	return false
}

// BulbState a snapshot of the bulbs last state
type BulbState struct {
	Hue        uint16
	Saturation uint16
	Brightness uint16
	Kelvin     uint16
	Dim        uint16
	Power      uint16
}

func newBulbState(hue, saturation, brightness, kelvin, dim, power uint16) *BulbState {
	return &BulbState{
		Hue:        hue,
		Saturation: saturation,
		Brightness: brightness,
		Kelvin:     kelvin,
		Dim:        dim,
		Power:      power,
	}
}

// LightSensorState a snapshot of the bulbs ambient light sensor read
type LightSensorState struct {
	Lux float32
}

// Gateway Lifx bulb which is acting as a gateway to the mesh
type Gateway struct {
	lifxAddress [6]byte
	hostAddress string
	Port        uint16
	Site        [6]byte // incoming messages are desimanated by site
	lastSeen    time.Time
}

// GetLifxAddress returns the unique lifx address of the gateway
func (g *Gateway) GetLifxAddress() string {
	return fmt.Sprintf("%x", g.lifxAddress)
}

// GetSite returns the unique site identifier for the gateway
func (g *Gateway) GetSite() string {
	return fmt.Sprintf("%x", g.Site)
}

func newGateway(lifxAddress [6]byte, hostAddress string, port uint16, site [6]byte) *Gateway {

	gw := &Gateway{lifxAddress: lifxAddress, hostAddress: hostAddress, Port: port, Site: site}

	return gw
}

func (g *Gateway) sendTo(cmd command) error {

	// can we connect to the gw
	addr, err := net.ResolveUDPAddr("udp4", g.hostAddress)

	if err != nil {
		return err
	}

	// open the connection, which we retain for all peer -> globe comms
	socket, err := net.DialUDP("udp4", nil, addr)

	if err != nil {
		return err
	}

	defer socket.Close()

	// send to globe
	_, err = cmd.WriteTo(socket)

	if err != nil {
		return err
	}

	//log.Printf("Sent command to gateway %s", reflect.TypeOf(cmd))

	return nil
}

func (g *Gateway) findBulbs() error {

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

// Client holds all the state and connections for the lifx client.
type Client struct {
	gateways      []*Gateway
	bulbs         []*Bulb
	intervalID    int
	DiscoInterval int

	peerSocket  net.Conn
	bcastSocket *net.UDPConn
	discoTicker *time.Ticker
	in          chan interface{}
	subs        []*Sub
}

// NewClient make a new lifx client
func NewClient() *Client {
	return &Client{}
}

// StartDiscovery Begin searching for lifx globes on the local LAN
func (c *Client) StartDiscovery() (err error) {

	//log.Printf("Listening for bcast :%d", BroadcastPort)

	// this socket will recieve broadcast packets on this socket
	c.bcastSocket, err = net.ListenUDP("udp4", &net.UDPAddr{Port: BroadcastPort})

	if err != nil {
		return
	}

	c.discoTicker = time.NewTicker(time.Second * 3)

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

// LightsOn turn all lifx bulbs on
func (c *Client) LightsOn() error {

	cmd := newSetPowerStateCommand(bulbOn)

	return c.sendToAll(cmd)
}

// LightsOff turn all lifx bulbs off
func (c *Client) LightsOff() error {

	cmd := newSetPowerStateCommand(bulbOff)

	return c.sendToAll(cmd)
}

// LightsColour changes the color of all lifx bulbs
func (c *Client) LightsColour(hue uint16, sat uint16, lum uint16, kelvin uint16, timing uint32) error {

	cmd := newSetLightColour(hue, sat, lum, kelvin, timing)

	return c.sendToAll(cmd)
}

// LightOn turn on a bulb
func (c *Client) LightOn(bulb *Bulb) error {

	cmd := newSetPowerStateCommand(bulbOn)

	return c.sendTo(bulb, cmd)
}

// LightOff turn off a bulb
func (c *Client) LightOff(bulb *Bulb) error {

	cmd := newSetPowerStateCommand(bulbOff)

	return c.sendTo(bulb, cmd)
}

// LightColour change the color of a bulb
func (c *Client) LightColour(bulb *Bulb, hue uint16, sat uint16, lum uint16, kelvin uint16, timing uint32) error {

	cmd := newSetLightColour(hue, sat, lum, kelvin, timing)

	return c.sendTo(bulb, cmd)
}

// GetBulbs get a list of the bulbs found by the client
func (c *Client) GetBulbs() []*Bulb {
	return c.bulbs
}

// GetBulbState send a notification to the bulb to emit it's current state
func (c *Client) GetBulbState(bulb *Bulb) error {
	log.Printf("GetBulbState sent to %s", bulb.GetLifxAddress())
	cmd := newGetLightStateCommandFromBulb(bulb.LifxAddress)
	return c.sendTo(bulb, cmd)
}

// GetAmbientLight send a notification to the bulb to emit the current ambient light
func (c *Client) GetAmbientLight(bulb *Bulb) error {
	log.Printf("GetAmbientLight sent to %s", bulb.GetLifxAddress())
	cmd := newGetAmbientLightCommandFromBulb(bulb.LifxAddress)
	return c.sendTo(bulb, cmd)
}

// Subscribe listen for new bulbs or gateways, note this is a pointer to the actual value.
func (c *Client) Subscribe() *Sub {
	sub := newSub()
	c.subs = append(c.subs, sub)
	return sub
}

func (c *Client) sendTo(bulb *Bulb, cmd command) error {

	cmd.SetLifxAddr(bulb.LifxAddress) // ensure the message is addressed to the correct bulb

	for _, gw := range c.gateways {
		//log.Printf("sending command to %s", gw.hostAddress)
		cmd.SetSiteAddr(gw.Site) // update the site address for each gateway
		err := gw.sendTo(cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) sendToAll(cmd command) error {
	for _, gw := range c.gateways {
		//log.Printf("sending command to %s", gw.hostAddress)
		cmd.SetSiteAddr(gw.Site) // update the site address so all globes change
		err := gw.sendTo(cmd)
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

		//log.Printf("Received buffer from %+v of %x", addr, buf[:n])

		cmd, err := decodeCommand(buf[:n])

		if err != nil {
			//log.Printf("Error processing command: %v", err)
			continue
		}
		//log.Printf("Recieved command: %s", reflect.TypeOf(cmd))

		switch cmd := cmd.(type) {
		case *panGatewayCommand:

			// found a gw
			if cmd.Payload.Service == 1 {
				gw := newGateway(cmd.Header.TargetMacAddress, addr.String(), cmd.Payload.Port, cmd.Header.Site)
				if err != nil {
					//log.Printf("failed to setup peer connection to gw")
				} else {
					c.addGateway(gw)
				}
			}

		case *lightStateCommand:

			// found a bulb
			bulb := newBulb(cmd.Header.TargetMacAddress)
			bulb.lastLightState = cmd

			bulb.bulbState = newBulbState(cmd.Payload.Hue, cmd.Payload.Saturation, cmd.Payload.Brightness, cmd.Payload.Kelvin, cmd.Payload.Dim, cmd.Payload.Power)

			c.addBulb(bulb)

		case *powerStateCommand:

			c.updateBulbPowerState(cmd.Header.TargetMacAddress, cmd.Payload.OnOff)

		case *ambientStateCommand:
			//log.Printf("Recieved lux: %f", cmd.Payload.Lux)

			c.updateAmbientLightState(cmd.Header.TargetMacAddress, cmd.Payload.Lux)

		default:
			//log.Printf("Recieved command: %s", reflect.TypeOf(cmd))
		}

	}
}

func (c *Client) sendDiscovery(t time.Time) {

	//log.Println("Discovery packet sent at", t)
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

	//log.Printf("Bcast sent %d", n)

	//log.Printf("gateways %v", c.gateways)
	//log.Printf("bulbs %v", c.bulbs)

}

func (c *Client) addGateway(gw *Gateway) {
	if !gatewayInSlice(gw, c.gateways) {
		//log.Printf("Added gw %v", gw)
		gw.lastSeen = time.Now()
		c.gateways = append(c.gateways, gw)

		// notify subscribers
		go c.notifySubsGwNew(gw)

	} else {
		for _, lgw := range c.gateways {
			if gw.lifxAddress == lgw.lifxAddress && gw.Port == lgw.Port && gw.hostAddress == lgw.hostAddress {
				gw.lastSeen = time.Now()
				//log.Printf("update last seen for %v %s", gw.hostAddress, gw.lastSeen)
			}
		}
	}

	gw.findBulbs()
}

func (c *Client) addBulb(bulb *Bulb) {
	if !bulbInSlice(bulb, c.bulbs) {
		bulb.LastSeen = time.Now()
		c.bulbs = append(c.bulbs, bulb)

		// log.Printf("Added bulb %x state %v", bulb.LifxAddress, bulb.bulbState)

		// notify subscribers
		go c.notifySubsBulbNew(bulb)
	}
	for _, lbulb := range c.bulbs {
		if bulb.LifxAddress == lbulb.LifxAddress {
			lbulb.update(bulb)
		}
	}
}

func (c *Client) updateBulbPowerState(lifxAddress [6]byte, onoff uint16) {
	for _, b := range c.bulbs {
		// this needs further investigation
		if lifxAddress == b.LifxAddress {
			b.bulbState.Power = onoff
			// log.Printf("Updated bulb %v", b)

			// notify subscribers
			go c.notifySubsBulbNew(b)
		}
	}
}

func (c *Client) updateAmbientLightState(lifxAddress [6]byte, lux float32) {
	for _, b := range c.bulbs {
		// this needs further investigation
		if lifxAddress == b.LifxAddress {
			b.lightSensorState = &LightSensorState{lux}

			// notify subscribers
			go c.notifySubsBulbNew(b)
		}
	}
}

func (c *Client) notifySubsGwNew(gw *Gateway) {
	for _, sub := range c.subs {
		// check if it is open
		sub.Events <- gw
	}
}

// dereference bulb and pass it to the subscriber via the out channel
func (c *Client) notifySubsBulbNew(bulb *Bulb) {
	for _, sub := range c.subs {
		// check if it is open
		sub.Events <- bulb
	}
}

// Sub subscription of changes
type Sub struct {
	Events chan interface{}
}

func newSub() *Sub {
	sub := &Sub{}
	sub.Events = make(chan interface{}, 1)
	return sub
}

func gatewayInSlice(a *Gateway, list []*Gateway) bool {
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
