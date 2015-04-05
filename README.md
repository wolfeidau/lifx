# lifx [![GoDoc](https://img.shields.io/badge/godoc-Reference-brightgreen.svg?style=flat)](http://godoc.org/github.com/wolfeidau/lifx) [![Build Status](https://travis-ci.org/wolfeidau/lifx.svg?branch=master)](https://travis-ci.org/wolfeidau/lifx)

Hacking on a client for the [lifx](http://lifx.co) light bulb, this is based on the work already done in the [lifxjs](https://github.com/magicmonkey/lifxjs) and [go-lifx](https://github.com/bjeanes/go-lifx).

The aim of this project is to keep things simple and just provide a very thin API to the lifx bulbs with a view to focusing on packet decoding, coordination and discovery.

*Note:* This library works with 2.x firmware!

# Usage

Below is a simple example illustrating how to observe discovery and changes, as well as control of bulbs.

``` go
package main

import (
    "log"
    "os"
    "time"

    "gopkg.in/wolfeidau/lifx.v1"
)

func main() {
    c := lifx.NewClient()

    err := c.StartDiscovery()

    if err != nil {
        log.Fatalf("Woops %s", err)
    }

    go func() {

        sub := c.Subscribe()

        for {
            event := <-sub.Events

            switch event := event.(type) {
            case *lifx.Gateway:
                log.Printf("Gateway Update %v", event)
            case *lifx.Bulb:
                log.Printf("Bulb Update %v", event.GetState())
            case *lifx.LightSensorState:
                log.Printf("Light Sensor Update %s %f", event.GetLifxAddress(), event.Lux)
            default:
                log.Printf("Event %v", event)
            }

        }
    }()

    log.Printf("LightsOn")
    c.LightsOn()

    time.Sleep(10 * time.Second)

    for _, bulb := range c.GetBulbs() {

        time.Sleep(5 * time.Second)

        // transition to a dull purple
        c.LightColour(bulb, 0xcc15, 0xffff, 0x1f4, 0, 0x0513)

        time.Sleep(5 * time.Second)

        // transition to a bright white
        c.LightColour(bulb, 0, 0, 0x8000, 0x0af0, 0x0513)
    }

}
```

# Disclaimer

This is currently very early release, everything can and will change.

# License

Copyright (c) 2014 Mark Wolfe
Licensed under the MIT license.
