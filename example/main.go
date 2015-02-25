package main

import (
	"log"
	"os"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/wolfeidau/lifx"
)

func main() {
	os.Exit(realMain())
}

func buildHandler(lifxAddress string) lifx.StateHandler {
	return func(newState *lifx.BulbState) {
		log.Printf("changing state %++v", newState)
	}
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
			case *lifx.Gateway:
				log.Printf("Gateway Update %s", event.GetLifxAddress())
			case *lifx.Bulb:
				log.Printf("Bulb Update %s", event.GetLifxAddress())
				log.Printf(spew.Sprintf("%+v", event))
				event.SetStateHandler(buildHandler(event.GetLifxAddress()))
			case *lifx.LightSensorState:
				log.Printf("Light Sensor Update %s %f", event.GetLifxAddress(), event.Lux)
			default:
				log.Printf("Event %+v", event)
			}

		}
	}()

	log.Printf("Looping")

	for {
		time.Sleep(5 * time.Second)

		log.Printf("LightsOn")
		c.LightsOn()

		time.Sleep(5 * time.Second)

		for _, bulb := range c.GetBulbs() {

			time.Sleep(5 * time.Second)

			log.Printf("purple %s", bulb.GetLifxAddress())
			c.LightColour(bulb, 0xcc15, 0xffff, 0x1f4, 0, 0x0513)
			time.Sleep(200 * time.Millisecond)
			c.GetBulbState(bulb)
			c.GetAmbientLight(bulb)

			time.Sleep(5 * time.Second)

			// bright white

			temps := []int{0, 1000, 2000, 3000, 4000, 5000, 6000}

			for _, val := range temps {
				time.Sleep(5 * time.Second)
				log.Printf("white %s %d", bulb.GetLifxAddress(), val)

				c.LightColour(bulb, 0, 0, 0x8000, uint16(val), 0x0513)
				time.Sleep(200 * time.Millisecond)
				c.GetBulbState(bulb)
				c.GetAmbientLight(bulb)

			}

		}

		time.Sleep(5 * time.Second)

		log.Printf("LightsOff")
		c.LightsOff()
	}

	return 0
}
