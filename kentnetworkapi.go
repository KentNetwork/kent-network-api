package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	client "github.com/influxdata/influxdb/client/v2"
)

const (
	influxHost       = "https://influxdb.kent.network"
	influxUser       = "*REMOVED*"
	influxPwd        = "*REMOVED*"
	influxQueryLimit = 500
)

var influxClient = influxDBClient()

func setupRouter() *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	// Return all sensors
	r.GET("/sensors", func(c *gin.Context) {
		// catchmentName := c.Query("catchmentName")
		// associatedWith := c.Query("associatedWith")
		// status := c.Query("status")
		// town := c.Query("town")
		lat := c.Query("lat")
		lon := c.Query("lon")
		dist := c.Query("dist")
		if !(((lat != "") && (lon != "") && (dist != "")) ||
			((lat == "") && (lon == "") && (dist == ""))) {
			c.String(400, "Error: lat,lon,dist mandatory if one of the fields is defined")
		}
		c.JSON(200, gin.H{
			"message": "Here are all the sensors",
		})
	})

	// All measures available from a particular sensor
	r.GET("/sensors/:sensorReference/measures", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Here are all the measures for a sensor",
		})
	})

	// Return all measures
	r.GET("/measures", func(c *gin.Context) {
		// sensorReference := c.Query("sensorReference")
		c.JSON(200, gin.H{
			"message": "Here are all the measures",
		})
	})

	// Return all readings for a particular measure reference
	r.GET("/measures/:measurementReference/readings", func(c *gin.Context) {
		// latest := c.Query("latest")       // latest values
		// today := c.Query("today")         // values for date
		// date := c.Query("date")           //values on date
		// since := c.Query("since")         // values since date
		startDate := c.Query("startDate") // values from start_date until end_date
		endDate := c.Query("endDate")     // values from start_date until end_date
		if !(((startDate != "") && (endDate != "")) ||
			((startDate == "") && (endDate == ""))) {
			c.String(400, "Error: start_date,end_date mandatory if one of the fields is defined")
		}
		// Line 71 needs to be improved -> first lookup measurement type based on
		// measure reference then pass sensor reference
		q := fmt.Sprintf("SELECT * FROM %s LIMIT %d", "temperature", influxQueryLimit)

		if response, err := queryInfluxDB(influxClient, q, "rivers"); err == nil {
			byteSlice, err := json.Marshal(response)
			if err != nil {
				panic(err)
			}
			c.Writer.Header().Set("Content-Type", "application/json")
			c.Writer.WriteHeader(200)
			c.Writer.Write(byteSlice)
		} else {
			log.Println("err: influx read error")
		}

	})

	// Return all readings for a particular sensor id
	r.GET("/sensors/:sensorReference/readings", func(c *gin.Context) {
		// latest := c.Query("latest")       // latest values
		// today := c.Query("today")         // values for date
		// date := c.Query("date")           //values on date
		// since := c.Query("since")         // values since date
		startDate := c.Query("startDate") // values from start_date until end_date
		endDate := c.Query("endDate")     // values from start_date until end_date
		if !(((startDate != "") && (endDate != "")) ||
			((startDate == "") && (endDate == ""))) {
			c.String(400, "Error: start_date,end_date mandatory if one of the fields is defined")
		}
		c.JSON(200, gin.H{
			"message": "Here are all the readings for this sensor",
		})
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
			c.String(400, "Error: start_date,end_date mandatory if one of the fields is defined")
		}
		c.JSON(200, gin.H{
			"message": "Here is all the readings from all the sensors",
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
