package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	_ "github.com/go-kivik/couchdb" // The CouchDB driver
	"github.com/go-kivik/kivik"     // Development version of Kivik

	client "github.com/influxdata/influxdb/client/v2"
)

const (
	influxQueryLimit = 100
)

// Meta - Most json responses contain a metadata object
type meta struct {
	Publisher   string `json:"publisher"`
	License     string `json:"license"`
	Version     string `json:"version"`
	ResultLimit uint32 `json:"resultLimit"`
}

// Reading - A sensor takes readings which consists of a timestamp and values
type reading struct {
	DateTime string  `json:"dateTime"`
	Sensor   string  `json:"sensor"` // URI of sensor
	Value    float32 `json:"value"`
}

// Sensor - A device contains one or more sensors that can take readings
type sensor struct {
	ID               string  `json:"@id"` // URI of sensor
	UpdateInterval   uint32  `json:"updateInterval"`
	Value            float32 `json:"value"`
	Type             string  `json:"type"`
	Unit             string  `json:"unit"`
	MaxOnRecord      float32 `json:"maxOnRecord"`
	MinOnRecord      float32 `json:"minOnRecord"`
	HighestRecent    float32 `json:"highestRecent"`
	TypicalRangeLow  float32 `json:"typicalRangeLow"`
	TypicalRangeHigh float32 `json:"typicalRangeHigh"`
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
	AppID          string `json:"appId"`
	DevID          string `json:"devId"`
	HardwareSerial string `json:"hardwareSerial"`
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
	ID          string   `json:"@id"` // URI of device
	Type        string   `json:"type"`
	Location    location `json:"location"`
	Ttn         ttn      `json:"ttn"`
	HardwareRef string   `json:"hardwareRef"`
	BatteryType string   `json:"batteryType"`
}

type runtimeConfig struct {
	influxUser string
	influxPwd  string
	serverBind string
	influxHost string
}

var influxClient client.Client
var couchClient *kivik.Client

var events = [...]string{
	"Unseen",
	"Active",
	"Decommisioned",
	"Fault",
	"Maintenance",
}

// String() function will return the english name
// that we want out constant events be recognized as
func (event eventType) String() string {
	return events[event-1]
}

func setupRouter(config runtimeConfig) *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	// All devices
	r.GET("/devices", func(c *gin.Context) {
		// catchmentName := c.Query("catchmentName")
		// associatedWith := c.Query("associatedWith")
		// status := c.Query("status")
		// town := c.Query("town")
		lat := c.Query("loc-lat")
		lon := c.Query("loc-lon")
		radius := c.Query("loc-radius")
		if !(((lat != "") && (lon != "") && (radius != "")) ||
			((lat == "") && (lon == "") && (radius == ""))) {
			c.String(400, "Invalid parameters")
			return
		}
		c.JSON(200, gin.H{
			"message": "Here are all the devices",
		})
	})

	// A device
	r.GET("/devices/:deviceId", func(c *gin.Context) {
		db, err := couchClient.DB(context.TODO(), "kentnetwork")
		if err != nil {
			c.String(500, "Server Error")
			return
		}
		row := db.Get(context.TODO(), c.Param("deviceId"))
		if err != nil {
			c.String(404, "Device not found")
			return
		}
		var aDevice device
		if err = row.ScanDoc(&aDevice); err != nil {
			c.String(404, "Device not found")
			return
		}

		c.JSON(200, gin.H{
			"Device": aDevice.ID,
		})
	})

	// All sensors for a device
	r.GET("/devices/:deviceId/sensors", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Here are all the sensors for a device",
		})
	})

	// Return all readings for a device
	r.GET("/devices/:deviceId/readings", func(c *gin.Context) {
		// latest := c.Query("latest")       // latest values
		// today := c.Query("today")         // values for date
		// date := c.Query("date")           //values on date
		// since := c.Query("since")         // values since date
		startDate := c.Query("startDate") // values from start_date until end_date
		endDate := c.Query("endDate")     // values from start_date until end_date
		if !(((startDate != "") && (endDate != "")) ||
			((startDate == "") && (endDate == ""))) {
			c.String(400, "Invalid parameters")
			return
		}
		c.JSON(200, gin.H{
			"message": "Here are all the readings for this device",
		})
	})

	// All sensors for all devices
	r.GET("/sensors", func(c *gin.Context) {
		// deviceId := c.Query("deviceId")
		c.JSON(200, gin.H{
			"message": "Here are all the sensors",
		})
	})

	// A sensor
	r.GET("/sensors/:sensorId", func(c *gin.Context) {
		// deviceId := c.Query("deviceId")
		c.JSON(200, gin.H{
			"message": "Here is a sensor",
		})
	})

	// All readings of a sensor
	r.GET("/sensors/:sensorId/readings", func(c *gin.Context) {

		// latest := c.Query("latest")       // latest values
		// today := c.Query("today")         // values for date
		// date := c.Query("date")           //values on date
		// since := c.Query("since")         // values since date
		s := strings.Split(c.Param("sensorId"), "_")
		if len(s) != 3 {
			c.String(404, "Sensor not found")
			return
		}
		var db, measure string
		switch s[0] {
		case "R":
			db = "rivers"
		case "A":
			db = "air"
		default:
			c.String(404, "Sensor not found")
			return
		}
		switch s[1] {
		case "T":
			measure = "temperature"
		case "F":
			measure = "flow"
		default:
			c.String(404, "Sensor not found")
			return
		}
		sensorID := s[2]
		startDate := c.Query("startDate") // values from start_date until end_date
		endDate := c.Query("endDate")     // values from start_date until end_date
		if !(((startDate != "") && (endDate != "")) ||
			((startDate == "") && (endDate == ""))) {
			c.String(400, "Invalid parameters")
			return
		}

		q := fmt.Sprintf("SELECT \"value\" FROM %s WHERE (\"sensor_id\" = '%s') LIMIT %d ", measure, sensorID, 100)

		if response, err := queryInfluxDB(influxClient, q, db); err == nil {
			byteSlice, err := json.Marshal(response[0].Series)

			if err != nil {
				c.JSON(500, gin.H{
					"Error": "Marshalling error",
				})
				return
			}
			c.Writer.Header().Set("Content-Type", "application/json")
			c.Writer.WriteHeader(200)
			c.Writer.Write(byteSlice)
		} else {
			c.JSON(500, gin.H{
				"Error": "Database Query Error",
			})
			return
		}

	})

	// Return all readings from all sensors
	r.GET("/data/readings", func(c *gin.Context) {
		// latest := c.Query("latest")       // latest values
		// today := c.Query("today")         // values for date
		// date := c.Query("date")           //values on date
		// since := c.Query("since")         // values since date
		startDate := c.Query("startDate") // values from start_date until end_date
		endDate := c.Query("endDate")     // values from start_date until end_date
		if !(((startDate != "") && (endDate != "")) ||
			((startDate == "") && (endDate == ""))) {
			c.String(400, "Invalid parameters")
			return
		}
		c.JSON(200, gin.H{
			"message": "Here is all the readings from all the devices",
		})
	})

	return r
}

func main() {

	config := doFlags()
	influxClient = influxDBClient(config)
	var couchErr error
	couchClient, couchErr = kivik.New(context.TODO(), "couch", "https://couchdb.kent.network/")
	if couchErr != nil {
		panic(couchErr)
	}

	r := setupRouter(config)
	// Listen and Server in 0.0.0.0:80
	r.Run(config.serverBind)
}

func doFlags() runtimeConfig {
	var config runtimeConfig
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.StringVar(&config.influxHost, "influxserver", `https://influxdb.kent.network`, "Influx server to connect to.")
	flag.StringVar(&config.influxUser, "influxuser", `river`, "Influx user to connect with.")
	// TODO: Passwords shoudln't be read from command line when possible, as this leaves passwords in the shell history"
	flag.StringVar(&config.influxPwd, "influxpwd", `NCQxM3Socdc2K4nEwS`, "Influx password user to connect with.")
	flag.StringVar(&config.serverBind, "bind", ":80", "Port Bind definition eg, \":80\"")

	flag.Parse()

	return config

}

func influxDBClient(c runtimeConfig) client.Client {
	config := client.HTTPConfig{
		Addr:     c.influxHost,
		Username: c.influxUser,
		Password: c.influxPwd}

	client, err := client.NewHTTPClient(config)
	if err != nil {
		log.Fatalln("Error: ", err)
	}
	return client
}

// queryInfluxDB convenience function to query the influx database
func queryInfluxDB(clnt client.Client, cmd string, database string) (res []client.Result, err error) {
	q := client.Query{
		Command:  cmd,
		Database: database,
	}
	if response, err := clnt.Query(q); err == nil {
		if response.Error() != nil {
			return res, response.Error()
		}
		res = response.Results
	} else {
		return res, err
	}
	return res, nil
}
