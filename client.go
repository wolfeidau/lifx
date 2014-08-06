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
)

type Bulb struct {
	lifxAddress string
	name        string
	address     []byte
}

type Gateway struct {
	lifxAddress [6]byte
	Port        uint16
	Site        [6]byte
}

func NewGateway(addr [6]byte, port uint16, site [6]byte) *Gateway {
	return &Gateway{lifxAddress: addr, Port: port, Site: site}
}

type Client struct {
	gateways      []*Gateway
	bulbs         []*Bulb
	intervalID    int
	DiscoInterval int

	peerSocket  net.Conn
	bcastSocket *net.UDPConn
	discoTicker *time.Ticker
}

func (c *Client) StartDiscovery() error {

	log.Printf("Listening for bcast :%d", BroadcastPort)

	// this socket will recieve unicast and broadcast packets on this socket
	bcast, err := net.ListenUDP("udp4", &net.UDPAddr{Port: BroadcastPort})

	if err != nil {
		return err
	}

	c.bcastSocket = bcast

	c.discoTicker = time.NewTicker(time.Second * 10)

	// once you pop you can't stop
	go func() {

		buf := make([]byte, 1024)

		for {
			n, addr, err := c.bcastSocket.ReadFrom(buf)

			if err != nil {
				log.Fatalf("Woops %s", err)
			}

			log.Printf("Received buffer from %s of %x", addr.String(), buf[:n])

			ph, err := DecodePacketHeader(buf)

			if err != nil {
				log.Fatalf("Woops %s", err)
			}

			if ph.Packet_type == PANgateway {

				pl, err := NewPANgatewayPayload(buf[HeaderLen:])
				if err != nil {
					log.Fatalf("Woops %s", err)
				}
				if pl.Service == 1 {
					gw := NewGateway(ph.Target_mac_address, pl.Port, ph.Site)
					c.AddGateway(gw)
				}
			}

		}
	}()

	// once you pop you can't stop
	go func() {

		for t := range c.discoTicker.C {
			log.Println("Discovery sent at", t)
			socket, _ := net.DialUDP("udp4", nil, &net.UDPAddr{
				IP:   net.IPv4(255, 255, 255, 255),
				Port: BroadcastPort,
			})
			p := NewPacketHeader(GetPANgateway)
			n, _ := p.Encode(socket)

			log.Printf("Bcast sent %d", n)
		}

	}()

	return nil
}

func (c *Client) AddGateway(gw *Gateway) {
	if !gatewayInSlice(gw, c.gateways) {
		log.Printf("Added gw %v", gw)
		c.gateways = append(c.gateways, gw)
	}
}

func gatewayInSlice(a *Gateway, list []*Gateway) bool {
	for _, b := range list {
		if reflect.DeepEqual(a, b) {
			return true
		}
	}
	return false
}

func NewClient() *Client {
	return &Client{}
}
