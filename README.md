# lifx 

Hacking on a client for the [lifx](http://lifx.co) light globe, this is based on the work already done in the [lifxjs](https://github.com/magicmonkey/lifxjs) and [go-lifx](https://github.com/bjeanes/go-lifx).

The aim of this project is to keep things simple and just provide a very thin API to the lifx globes with a view to focusing on packet decoding, coordination and discovery.

# Usage

    import "github.com/wolfeidau/lifx"

## Constants
``` go
const (
    // BroadcastPort port used for broadcasting messages to lifx globes
    BroadcastPort = 56700

    // PeerPort port used for peer to peer messages to lifx globes
    PeerPort = 56750
)
```

## type Bulb
``` go
type Bulb struct {
    LifxAddress [6]byte

    LastSeen time.Time
    // contains filtered or unexported fields
}
```
Bulb Holds the state for a lifx bulb

### func (\*Bulb) GetLabel
``` go
func (b *Bulb) GetLabel() string
```
GetLabel get the label from the globe

### func (\*Bulb) GetPower
``` go
func (b *Bulb) GetPower() uint16
```
GetPower is the globe powered on or off

### func (\*Bulb) GetState
``` go
func (b *Bulb) GetState() *BulbState
```
GetState get a snapshot of the state for the bulb

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
BulbState a snapshot of the bulbs last state

## type Client
``` go
type Client struct {
    DiscoInterval int
    // contains filtered or unexported fields
}
```
Client holds all the state and connections for the lifx client.

### func NewClient
``` go
func NewClient() *Client
```
NewClient make a new lifx client




### func (\*Client) GetBulbs
``` go
func (c *Client) GetBulbs() []*Bulb
```
GetBulbs get a list of the bulbs found by the client



### func (\*Client) LightColour
``` go
func (c *Client) LightColour(bulb *Bulb, hue uint16, sat uint16, lum uint16, kelvin uint16, timing uint32) error
```
LightColour change the color of a bulb



### func (\*Client) LightOff
``` go
func (c *Client) LightOff(bulb *Bulb) error
```
LightOff turn off a bulb



### func (\*Client) LightOn
``` go
func (c *Client) LightOn(bulb *Bulb) error
```
LightOn turn on a bulb



### func (\*Client) LightsColour
``` go
func (c *Client) LightsColour(hue uint16, sat uint16, lum uint16, kelvin uint16, timing uint32) error
```
LightsColour changes the color of all lifx bulbs



### func (\*Client) LightsOff
``` go
func (c *Client) LightsOff() error
```
LightsOff turn all lifx bulbs off



### func (\*Client) LightsOn
``` go
func (c *Client) LightsOn() error
```
LightsOn turn all lifx bulbs on



### func (\*Client) StartDiscovery
``` go
func (c *Client) StartDiscovery() (err error)
```
StartDiscovery Begin searching for lifx globes on the local LAN

# Disclaimer

This is currently very early release, everything can and will change.

# License

Copyright (c) 2014 Mark Wolfe
Licensed under the MIT license.