package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	client "github.com/influxdata/influxdb/client/v2"
)

const (
	resultLimit = 100
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
	Location    location `json:"location"`
	Ttn         ttn      `json:"ttn"`
	HardwareRef string   `json:"hardwareRef"`
	BatteryType string   `json:"batteryType"`
}

type runtimeConfig struct {
	influxUser string
	influxPwd  string
	influxDb   string
	serverBind string
	influxHost string
	couchHost  string
}

var influxClient client.Client

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

func metaConsuctor(limit int) meta {
	metaData := meta{}
	metaData.License = "Creative Commons"
	metaData.Publisher = "Kent Network"
	metaData.Version = "0.1"
	return metaData
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
		//lat := c.Query("loc-lat")
		//lon := c.Query("loc-lon")
		//radius := c.Query("loc-radius")

		type okResponse struct {
			Meta    meta     `json:"meta"`
			Devices []device `json:"devices"`
		}

		type couchView struct {
			TotalRows int `json:"total_rows"`
			Offset    int `json:"offset"`
			Rows      []struct {
				ID     string      `json:"id"`
				Key    interface{} `json:"key"`
				Value  interface{} `json:"value"`
				Device device      `json:"doc"`
			} `json:"rows"`
		}

		code, resp, err := queryCouchdb(config.couchHost + "/kentnetwork/_design/devices/_view/getDevices?include_docs=true")
		if err != nil && code != 200 {
			c.String(500, "Internal server error")
			return
		}

		var couchResp couchView
		if err = json.Unmarshal(resp, &couchResp); err != nil {
			c.String(500, "Internal server error")
			return
		}

		// Build OK response
		var a okResponse
		a.Meta = metaConsuctor(resultLimit)
		for i := range couchResp.Rows {
			a.Devices = append(a.Devices, couchResp.Rows[i].Device)
		}

		c.JSON(http.StatusOK, a)
	})

	// A device
	r.GET("/devices/:deviceId", func(c *gin.Context) {

		type okResponse struct {
			Meta   meta   `json:"meta"`
			Device device `json:"device"`
		}

		code, resp, err := queryCouchdb(config.couchHost + "/kentnetwork/" + c.Param("deviceId"))
		if err != nil || code == 500 {
			c.String(500, "Internal server error")
			return
		}
		if code == 404 {
			c.String(404, "Device not found")
			return
		}

		var returnedDevice device
		if err = json.Unmarshal(resp, &returnedDevice); err != nil {
			c.String(500, "Internal server error")
			return
		}

		// Build OK response
		var a okResponse
		a.Device = returnedDevice
		a.Meta = metaConsuctor(resultLimit)

		c.JSON(http.StatusOK, a)

	})

	// All sensors for a device
	r.GET("/devices/:deviceId/sensors", func(c *gin.Context) {

		type okResponse struct {
			Meta    meta     `json:"meta"`
			Sensors []sensor `json:"sensors"`
		}

		type couchView struct {
			TotalRows int `json:"total_rows"`
			Offset    int `json:"offset"`
			Rows      []struct {
				ID     string      `json:"id"`
				Key    string      `json:"key"`
				Value  interface{} `json:"value"`
				Sensor sensor      `json:"doc"`
			} `json:"rows"`
		}

		code, resp, err := queryCouchdb(config.couchHost + "/kentnetwork/_design/sensors/_view/getByDeviceID?include_docs=true&startkey=\"" + c.Param("deviceId") + "\"&endkey=\"" + c.Param("deviceId") + "\ufff0\"")
		if err != nil && code != 200 {
			c.String(500, "Internal server error")
			return
		}

		var couchResp couchView
		if err = json.Unmarshal(resp, &couchResp); err != nil {
			c.String(500, "Internal server error")
			return
		}

		if len(couchResp.Rows) == 0 {
			c.String(404, "Device not found or device currently has no sensors")
			return
		}

		// Build OK response
		var a okResponse
		a.Meta = metaConsuctor(resultLimit)
		for i := range couchResp.Rows {
			a.Sensors = append(a.Sensors, couchResp.Rows[i].Sensor)
		}

		c.JSON(http.StatusOK, a)

	})

	// Return all readings for a device
	r.GET("/devices/:deviceId/readings", func(c *gin.Context) {
		// latest := c.Query("latest")       // latest values
		// today := c.Query("today")         // values for date
		// date := c.Query("date")           //values on date
		// since := c.Query("since")         // values since date
		// startDate := c.Query("startDate") // values from start_date until end_date
		// endDate := c.Query("endDate")     // values from start_date until end_date
		// if !(((startDate != "") && (endDate != "")) ||
		// 	((startDate == "") && (endDate == ""))) {
		// 	c.String(400, "Invalid parameters")
		// 	return
		// }

		c.JSON(200, gin.H{
			"message": "Here are all the readings for this device",
		})
	})

	// All sensors for all devices
	r.GET("/sensors", func(c *gin.Context) {
		type okResponse struct {
			Meta    meta     `json:"meta"`
			Sensors []sensor `json:"sensors"`
		}

		type couchView struct {
			TotalRows int `json:"total_rows"`
			Offset    int `json:"offset"`
			Rows      []struct {
				ID     string      `json:"id"`
				Key    interface{} `json:"key"`
				Value  interface{} `json:"value"`
				Sensor sensor      `json:"doc"`
			} `json:"rows"`
		}

		code, resp, err := queryCouchdb(config.couchHost + "/kentnetwork/_design/sensors/_view/getSensors?include_docs=true")
		if err != nil && code != 200 {
			c.String(500, "Internal server error")
			return
		}

		var couchResp couchView
		if err = json.Unmarshal(resp, &couchResp); err != nil {
			c.String(500, "Internal server error")
			return
		}

		// Build OK response
		var a okResponse
		a.Meta = metaConsuctor(resultLimit)
		for i := range couchResp.Rows {
			a.Sensors = append(a.Sensors, couchResp.Rows[i].Sensor)
		}

		c.JSON(http.StatusOK, a)
	})

	// A sensor
	r.GET("/sensors/:sensorId", func(c *gin.Context) {
		type okResponse struct {
			Meta   meta   `json:"meta"`
			Sensor sensor `json:"sensor"`
		}

		code, resp, err := queryCouchdb(config.couchHost + "/kentnetwork/" + c.Param("sensorId"))
		if err != nil || code == 500 {
			c.String(500, "Internal server error")
			return
		}

		if code == 404 {
			c.String(404, "Sensor not found")
			return
		}

		var returnedSensor sensor
		if err = json.Unmarshal(resp, &returnedSensor); err != nil {
			c.String(500, "Internal server error")
			return
		}

		// Build OK response
		var a okResponse
		a.Sensor = returnedSensor
		a.Meta = metaConsuctor(resultLimit)

		c.JSON(http.StatusOK, a)
	})

	// All readings of a sensor
	r.GET("/sensors/:sensorId/readings", func(c *gin.Context) {

		q := fmt.Sprintf("SELECT \"value\" FROM /.*/ WHERE (\"sensor_id\" = '%s') LIMIT %d ", c.Param("sensorId"), resultLimit)

		if response, err := queryInfluxDB(influxClient, q, config.influxDb); err == nil {
			if response[0].Series == nil {
				c.String(404, "Sensor not found or sensor has no readings")
				return
			}
			c.JSON(http.StatusOK, response)
		} else {
			c.String(500, "Internal server error")
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
	var err error
	influxClient, err = influxDBClient(config)
	if err != nil {
		log.Fatal(err)
	}
	r := setupRouter(config)
	// Listen and Server in 0.0.0.0:80
	r.Run(config.serverBind)
}

func doFlags() runtimeConfig {
	var config runtimeConfig
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.StringVar(&config.influxHost, "influxserver", `https://influxdb.kent.network`, "Influx server to connect to.")
	flag.StringVar(&config.influxUser, "influxuser", `reader`, "Influx user to connect with.")
	// TODO: Passwords shoudln't be read from command line when possible, as this leaves passwords in the shell history"
	flag.StringVar(&config.influxPwd, "influxpwd", `asij8X3rNU8U`, "Influx password user to connect with.")
	flag.StringVar(&config.influxDb, "influxdb", `readings`, "Influx database to use.")
	flag.StringVar(&config.serverBind, "bind", ":80", "Port Bind definition eg, \":80\"")
	flag.StringVar(&config.couchHost, "couchserver", `https://couchdb.kent.network`, "Couchdb server to connect to.")

	flag.Parse()

	return config

}

func influxDBClient(c runtimeConfig) (client.Client, error) {
	config := client.HTTPConfig{
		Addr:     c.influxHost,
		Username: c.influxUser,
		Password: c.influxPwd}

	client, err := client.NewHTTPClient(config)
	return client, err
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

func queryCouchdb(request string) (code int, response []byte, err error) {
	resp, err := http.Get(request)
	if err != nil {
		return 500, nil, err
	}
	defer resp.Body.Close()
	response, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return 500, nil, err
	}
	code = resp.StatusCode
	return code, response, err
}
