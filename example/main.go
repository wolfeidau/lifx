package main

import (
	"log"
	"os"
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

	go func() {
		log.Printf("Subscribing to changes")

		sub := c.Subscribe()
		for {
			event := <-sub.Events

			switch event := event.(type) {
			case lifx.Gateway:
				log.Printf("Gateway Update %v", event)
			case lifx.Bulb:
				log.Printf("Bulb Update %v", event.GetState())
			default:
				log.Printf("Event %v", event)
			}

		}
	}()

	log.Printf("Looping")

	for {
		time.Sleep(10 * time.Second)

		log.Printf("LightsOn")
		c.LightsOn()

		time.Sleep(10 * time.Second)

		for _, bulb := range c.GetBulbs() {

			time.Sleep(5 * time.Second)

			log.Printf("purple %v", bulb.LifxAddress)
			c.LightColour(bulb, 0xcc15, 0xffff, 0x1f4, 0, 0x0513)

			time.Sleep(5 * time.Second)

			// bright white
			log.Printf("white %v", bulb.LifxAddress)
			c.LightColour(bulb, 0, 0, 0x8000, 0x0af0, 0x0513)
		}

	}

	return 0
}
