package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/juju/loggo"
	"github.com/wolfeidau/lifx"
)

var logger = loggo.GetLogger("")

func init() {
	logger.SetLogLevel(loggo.DEBUG)
}

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
		logger.Infof("Subscribing to changes")

		sub := c.Subscribe()
		for {
			event := <-sub.Events

			switch event := event.(type) {
			case lifx.Gateway:
				logger.Infof("Gateway Update %v", event)
			case lifx.Bulb:
				logger.Infof("Bulb Update %v", event.GetState())
			default:
				logger.Infof("Event %v", event)
			}

		}
	}()

	logger.Infof("Looping")

	for {
		time.Sleep(10 * time.Second)

		log.Printf("LightsOn")
		c.LightsOn()

		time.Sleep(10 * time.Second)

		for _, bulb := range c.GetBulbs() {

			time.Sleep(5 * time.Second)

			logger.Infof("purple %v", bulb.LifxAddress)
			c.LightColour(bulb, 0xcc15, 0xffff, 0x1f4, 0, 0x0513)

			time.Sleep(5 * time.Second)

			// bright white
			logger.Infof("white %v", bulb.LifxAddress)
			c.LightColour(bulb, 0, 0, 0x8000, 0x0af0, 0x0513)
		}

	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill)

	// Block until a signal is received.
	s := <-sigChan
	fmt.Println("Got signal:", s)

	return 0
}
