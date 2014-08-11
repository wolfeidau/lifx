# lifx 

Hacking on a client for the [lifx](http://lifx.co) light globe, this is based on the work already done in the [lifxjs](https://github.com/magicmonkey/lifxjs) and [go-lifx](https://github.com/bjeanes/go-lifx).

The aim of this project is to keep things simple and just provide a very thin API to the lifx globes with a view to focusing on packet decoding, coordination and discovery.

# Usage

    import "github.com/wolfeidau/lifx"


## type Bulb
``` go
type Bulb struct {
    LifxAddress [6]byte

    LastSeen time.Time
    // contains filtered or unexported fields
}
```

### func (\*Bulb) GetLabel
``` go
func (b *Bulb) GetLabel() string
```

### func (\*Bulb) GetPower
``` go
func (b *Bulb) GetPower() uint16
```

### func (\*Bulb) GetState
``` go
func (b *Bulb) GetState() *BulbState
```

## type BulbState
``` go
type BulbState struct {
    Hue        uint16
    Saturation uint16
    Brightness uint16
    Kelvin     uint16
    Dim        uint16
}
```

## type Client
``` go
type Client struct {
    DiscoInterval int
    // contains filtered or unexported fields
}
```

### func NewClient
``` go
func NewClient() *Client
```

### func (\*Client) GetBulbs
``` go
func (b *Client) GetBulbs() []*Bulb
```

### func (\*Client) LightColour
``` go
func (b *Client) LightColour(bulb *Bulb, hue uint16, sat uint16, lum uint16, kelvin uint16, timing uint32) error
```

### func (\*Client) LightOff
``` go
func (b *Client) LightOff(bulb *Bulb) error
```

### func (\*Client) LightOn
``` go
func (b *Client) LightOn(bulb *Bulb) error
```

### func (\*Client) LightsColour
``` go
func (b *Client) LightsColour(hue uint16, sat uint16, lum uint16, kelvin uint16, timing uint32) error
```

### func (\*Client) LightsOff
``` go
func (b *Client) LightsOff() error
```

### func (\*Client) LightsOn
``` go
func (b *Client) LightsOn() error
```

### func (\*Client) StartDiscovery
``` go
func (c *Client) StartDiscovery() (err error)
```

# Disclaimer

This is currently very early release, everything can and will change.

# License

Copyright (c) 2014 Mark Wolfe
Licensed under the MIT license.