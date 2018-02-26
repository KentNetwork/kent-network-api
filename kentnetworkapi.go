package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	client "github.com/influxdata/influxdb/client/v2"
)

const (
	influxHost       = "https://influxdb.kent.network"
	influxUser       = "river"
	influxPwd        = "NCQxM3Socdc2K4nEwS"
	influxQueryLimit = 100
)

type Meta struct {
	Publisher   string
	License     string
	Version     string
	ResultLimit uint32
}

type Reading struct {
	DateTime string
	Sensor   string // URI of sensor
	Value    float32
}

type Sensor struct {
	ID               string // URI of sensor
	UpdateInterval   uint32
	Value            float32
	Type             string
	Unit             string
	MaxOnRecord      float32
	MinOnRecord      float32
	HighestRecent    float32
	TypicalRangeLow  float32
	TypicalRangeHigh float32
}

type EventType int

const (
	Unseen        EventType = iota + 1
	Active        EventType = iota + 1
	Decommisioned EventType = iota + 1
	Fault         EventType = iota + 1
	Maintenance   EventType = iota + 1
)

type Status struct {
	Type   EventType
	Reason string
	Date   string
}

type Ttn struct {
	AppID          string
	DevID          string
	HardwareSerial string
}

type Location struct {
	NearestTown    string
	CatchmentName  string
	AssociatedWith string
	Lat            float32
	Lon            float32
	Alititude      float32
	Easting        string
	Northing       string
}

type Device struct {
	ID          string // URI of device
	Type        string
	Status      []Status
	Location    Location
	Ttn         Ttn
	Lat         float32
	HardwareRef string
	BatteryType string
}

var influxClient = influxDBClient()

var events = [...]string{
	"Unseen",
	"Active",
	"Decommisioned",
	"Fault",
	"Maintenance",
}

// String() function will return the english name
// that we want out constant events be recognized as
func (event EventType) String() string {
	return events[event-1]
}

func setupRouter() *gin.Engine {
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
		c.JSON(200, gin.H{
			"message": "Here is a device",
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
	r := setupRouter()
	// Listen and Server in 0.0.0.0:80
	r.Run(":80")
}

func influxDBClient() client.Client {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     influxHost,
		Username: influxUser,
		Password: influxPwd,
	})
	if err != nil {
		log.Fatalln("Error: ", err)
	}
	return c
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
