package lifx

import (
	"fmt"
	"log"
	"net"
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
}

type Client struct {
	gateways      []Gateway
	bulbs         []Bulb
	intervalID    int
	DiscoInterval int

	peerSocket  net.Conn
	bcastSocket *net.UDPConn
	discoTicker *time.Ticker
}

func (c *Client) StartDiscovery() error {

	log.Printf("Listing for bcast :%d", BroadcastPort)

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

			log.Printf("Found a buffer %s % x", addr.String(), buf[:n])

		}
	}()

	// once you pop you can't stop
	go func() {

		for t := range c.discoTicker.C {
			fmt.Println("Tick at", t)
			socket, _ := net.DialUDP("udp4", nil, &net.UDPAddr{
				IP:   net.IPv4(255, 255, 255, 255),
				Port: BroadcastPort,
			})
			p := NewPacket(GetPANgateway)
			n, _ := p.Encode(socket)
			//n, _ := socket.Write()
			fmt.Println("Bcast sent %d", n)
		}

	}()

	return nil
}

func NewClient() *Client {
	return &Client{}
}
