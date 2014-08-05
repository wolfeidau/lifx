package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

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

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill)

	// Block until a signal is received.
	s := <-sigChan
	fmt.Println("Got signal:", s)

	return 0
}
