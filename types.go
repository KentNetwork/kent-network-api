package main

import (
	ttnsdk "github.com/TheThingsNetwork/go-app-sdk"
)

// Reading - A sensor takes readings which consists of a timestamp and values
type reading struct {
	DateTime string  `json:"dateTime"`
	Sensor   string  `json:"sensor"` // URI of sensor
	Value    float64 `json:"value"`
}

// Sensor - A device contains one or more sensors that can take readings
type sensor struct {
	ID             string `json:"@id"` // URI of sensor
	UpdateInterval uint32 `json:"updateInterval"`
	ParentDevice   string `json:"parentDevice"`
	SensorType     string `json:"sensorType"`
	Unit           string `json:"unit"`
}

// EventType - event type enum for device status
type eventType int

// event type enum for device status
const (
	Unseen        eventType = iota + 1
	Active        eventType = iota + 1
	Decommisioned eventType = iota + 1
	Fault         eventType = iota + 1
	Maintenance   eventType = iota + 1
)

// Status - A device contains an array of different status events
type status struct {
	Type     eventType `json:"type"`
	Reason   string    `json:"reason"`
	DateTime string    `json:"date"`
}

// Ttn - A device contains an object with things network metadata
type ttn struct {
	AppEUI string `json:"appEUI"`
	DevID  string `json:"devId"`
	AppKey string `json:"appKey"`
}

func TtnFromTtnsdkDevice(d ttnsdk.Device) ttn {
	return ttn{
		AppEUI: d.AppEUI.String(),
		DevID:  d.DevID,
		AppKey: d.AppKey.String(),
	}
}

// Location - A device contains an object with location metadata
type location struct {
	NearestTown    string  `json:"nearestTown"`
	CatchmentName  string  `json:"catchmentName"`
	AssociatedWith string  `json:"associatedWith"`
	Lat            float32 `json:"lat"`
	Lon            float32 `json:"lon"`
	Altitude       float32 `json:"altitude"`
	Easting        string  `json:"easting"`
	Northing       string  `json:"northing"`
}

// Device represents a physical device
type device struct {
	ID          string    `json:"@id"` // URI of device
	Location    *location `json:"location,omitempty"`
	Ttn         *ttn      `json:"ttn,omitempty"`
	HardwareRef string    `json:"hardwareRef"`
	BatteryType string    `json:"batteryType"`
	Owner       string    `json:"owner"`
}

// Gateway represents metadata about a gateway
type gateway struct {
	GatewayMac string  `json:"gatewayMac"` // Mac address of gateway
	Lat        float64 `json:"lat"`        // Lat cord of gateway
	Lon        float64 `json:"lon"`        // Lon cord of gateway
}
