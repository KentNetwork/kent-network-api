package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func GET_gateways(config runtimeConfig) func(c *gin.Context) {
	return func(c *gin.Context) {

		type okResponse struct {
			Meta     meta      `json:"meta"`
			Gateways []gateway `json:"items"`
		}

		gateways, err := getGatewaysMeta(config.Influx, "gatewayrxpkts")
		if err != nil {
			c.String(500, "Internal server error")
			return
		}

		// Build OK response
		var a okResponse
		a.Meta = newMeta(resultLimit)
		a.Gateways = gateways
		c.JSON(http.StatusOK, a)
	}
}

func GET_sensors(config runtimeConfig) func(*gin.Context) {
	return func(c *gin.Context) {
		type okResponse struct {
			Meta    meta     `json:"meta"`
			Sensors []sensor `json:"items"`
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

		code, resp, err := config.Couch.query("/kentnetwork/_design/sensors/_view/getSensors?include_docs=true")
		if err != nil && code != 200 {
			c.String(500, "Couchdb connection error")
			return
		}

		var couchResp couchView
		if err = json.Unmarshal(resp, &couchResp); err != nil {
			c.String(500, "Unmarshalling error")
			return
		}

		// Build OK response
		var a okResponse
		a.Meta = newMeta(resultLimit)
		for i := range couchResp.Rows {
			a.Sensors = append(a.Sensors, couchResp.Rows[i].Sensor)
		}

		c.JSON(http.StatusOK, a)
	}
}

func GET_sensors_id(config runtimeConfig) func(*gin.Context) {
	return func(c *gin.Context) {
		type okResponse struct {
			Meta   meta   `json:"meta"`
			Sensor sensor `json:"items"`
		}

		code, resp, err := config.Couch.query("/kentnetwork/" + c.Param("sensorId"))
		if err != nil || code == 500 {
			c.String(500, "Couchdb connection error")
			return
		}

		if code == 404 {
			c.String(404, "Sensor not found")
			return
		}

		var returnedSensor sensor
		if err = json.Unmarshal(resp, &returnedSensor); err != nil {
			c.String(500, "Unmarshalling error")
			return
		}

		// Build OK response
		var a okResponse
		a.Sensor = returnedSensor
		a.Meta = newMeta(resultLimit)

		c.JSON(http.StatusOK, a)
	}
}

func GET_sensors_id_readings(config runtimeConfig) func(*gin.Context) {
	return func(c *gin.Context) {
		type okResponse struct {
			Meta     meta      `json:"meta"`
			Readings []reading `json:"items"`
		}

		var err error
		latest := false
		validDate := false
		var startDate time.Time
		var endDate time.Time

		if c.Query("latest") != "" {
			latest, err = strconv.ParseBool(c.Query("latest"))
		} else if c.Query("startDate") != "" {
			startDate, err = time.Parse("2006-01-02T15:04:05.999Z07:00", c.Query("startDate"))
			if c.Query("endDate") != "" {
				endDate, err = time.Parse("2006-01-02T15:04:05.999Z07:00", c.Query("endDate"))
			} else {
				endDate = time.Now()
			}
			if err == nil {
				validDate = true
			}
		}

		if err != nil {
			c.String(400, "User supplied parameter error")
			return
		}

		var readings []reading
		if latest == false && validDate == false {
			readings, err = getSensorData(config.Influx, c.Param("sensorId"), false, time.Time{}, time.Time{}, config.Influx.Db)
		} else if latest {
			readings, err = getSensorData(config.Influx, c.Param("sensorId"), true, time.Time{}, time.Time{}, config.Influx.Db)
		} else if validDate {
			readings, err = getSensorData(config.Influx, c.Param("sensorId"), false, startDate, endDate, config.Influx.Db)
		}

		if err != nil {
			c.String(500, "Influxdb connection error")
			return
		}

		if readings == nil {
			c.String(404, "Sensor not found or sensor has no readings")
			return
		}

		// Build OK response
		var a okResponse
		a.Meta = newMeta(resultLimit)
		a.Readings = readings
		c.JSON(http.StatusOK, a)

	}
}

func GET_data_readings(config runtimeConfig) func(*gin.Context) {
	return func(c *gin.Context) {
		// latest := c.Query("latest")       // latest values
		// today := c.Query("today")         // values for date
		// date := c.Query("date")           //values on date
		// since := c.Query("since")         // values since date
		//startDate := c.Query("startDate") // values from start_date until end_date
		//endDate := c.Query("endDate")     // values from start_date until end_date
		type okResponse struct {
			Meta     meta      `json:"meta"`
			Readings []reading `json:"items"`
		}

		type couchView struct {
			TotalRows int `json:"total_rows"`
			Offset    int `json:"offset"`
			Rows      []struct {
				ID    string      `json:"id"`
				Key   string      `json:"key"`
				Value interface{} `json:"value"`
			} `json:"rows"`
		}

		var err error
		latest := false
		validDate := false
		var startDate time.Time
		var endDate time.Time

		if c.Query("latest") != "" {
			latest, err = strconv.ParseBool(c.Query("latest"))
		} else if c.Query("startDate") != "" {
			startDate, err = time.Parse("2006-01-02T15:04:05.999Z07:00", c.Query("startDate"))
			if c.Query("endDate") != "" {
				endDate, err = time.Parse("2006-01-02T15:04:05.999Z07:00", c.Query("endDate"))
			} else {
				endDate = time.Now()
			}
			if err == nil {
				validDate = true
			}
		}

		if err != nil {
			c.String(400, "User supplied parameter error")
			return
		}

		code, resp, err := config.Couch.query("/kentnetwork/_design/sensors/_view/getSensors")
		if err != nil && code != 200 {
			c.String(500, "Couchdb connection error")
			return
		}

		var couchResp couchView
		if err = json.Unmarshal(resp, &couchResp); err != nil {
			c.String(500, "Unmarshalling error")
			return
		}

		if len(couchResp.Rows) == 0 {
			c.String(404, "No sensors found or system has sensors with no readings")
			return
		}

		// Build OK response
		var a okResponse
		a.Meta = newMeta(resultLimit)

		for i := range couchResp.Rows {

			var readings []reading
			if latest == false && validDate == false {
				readings, err = getSensorData(config.Influx, couchResp.Rows[i].ID, false, time.Time{}, time.Time{}, config.Influx.Db)
			} else if latest {
				readings, err = getSensorData(config.Influx, couchResp.Rows[i].ID, true, time.Time{}, time.Time{}, config.Influx.Db)
			} else if validDate {
				readings, err = getSensorData(config.Influx, couchResp.Rows[i].ID, false, startDate, endDate, config.Influx.Db)
			}

			if err != nil {
				c.String(500, "Influxdb connection error")
				return
			}

			if readings != nil {
				for i := range readings {
					a.Readings = append(a.Readings, readings[i])
				}
			}
		}

		if a.Readings == nil {
			c.String(404, "No sensors found or system has sensors with no readings")
			return
		}

		c.JSON(http.StatusOK, a)
	}
}
