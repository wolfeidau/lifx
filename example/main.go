package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/wolfeidau/lifx"
)

func main() {
	os.Exit(realMain())
}

func realMain() int {
	c := lifx.NewClient()

	err := c.StartDiscovery()

	if err != nil {
		log.Fatalf("Woops %s", err)
	}

	for {
		time.Sleep(10 * time.Second)

		//c.LightsOn()

		// time.Sleep(10 * time.Second)

		// c.LightsOff()
		// func (b *Client) LightColour(bulb *Bulb, hue uint16, sat uint16, lum uint16, kelvin uint16, timing uint32) error {

		for _, bulb := range c.GetBulbs() {

			log.Printf("send to bulb %v", bulb)

			time.Sleep(1 * time.Second)
			c.LightColour(bulb, 0xcc15, 0xffff, 0x1f4, 0, 0x1f4)

			// bright white
			time.Sleep(1 * time.Second)
			c.LightColour(bulb, 0, 0, 0x8000, 0x0af0, 0x1f4)
		}

	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill)

	// Block until a signal is received.
	s := <-sigChan
	fmt.Println("Got signal:", s)

	return 0
}
