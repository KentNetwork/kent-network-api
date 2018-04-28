
package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func GET_devices(config runtimeConfig) func(c *gin.Context) {
	return func(c *gin.Context) {
		// catchmentName := c.Query("catchmentName")
		// associatedWith := c.Query("associatedWith")
		// status := c.Query("status")
		// town := c.Query("town")
		//lat := c.Query("loc-lat")
		//lon := c.Query("loc-lon")
		//radius := c.Query("loc-radius")

		type okResponse struct {
			Meta    meta     `json:"meta"`
			Devices []device `json:"items"`
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

		code, resp, err := queryCouchdb(config.CouchHost + "/kentnetwork/_design/devices/_view/getDevices?include_docs=true")
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
			a.Devices = append(a.Devices, couchResp.Rows[i].Device)
		}

		c.JSON(http.StatusOK, a)
	}
}


func GET_devices_id(config runtimeConfig) func(c *gin.Context) {
	return func(c *gin.Context) {

		type okResponse struct {
			Meta   meta   `json:"meta"`
			Device device `json:"items"`
		}

		code, resp, err := queryCouchdb(config.CouchHost + "/kentnetwork/" + c.Param("deviceId"))
		if err != nil || code == 500 {
			c.String(500, "Couchdb connection error")
			return
		}
		if code == 404 {
			c.String(404, "Device not found")
			return
		}

		var returnedDevice device
		if err = json.Unmarshal(resp, &returnedDevice); err != nil {
			c.String(500, "Unmarshalling error")
			return
		}

		// Build OK response
		var a okResponse
		a.Device = returnedDevice
		a.Meta = newMeta(resultLimit)

		c.JSON(http.StatusOK, a)

	}
}
func GET_devices_id_sensors(config runtimeConfig) func(c *gin.Context) {
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
				Key    string      `json:"key"`
				Value  interface{} `json:"value"`
				Sensor sensor      `json:"doc"`
			} `json:"rows"`
		}

		code, resp, err := queryCouchdb(config.CouchHost + "/kentnetwork/_design/sensors/_view/getByDeviceID?include_docs=true&startkey=\"" + c.Param("deviceId") + "\"&endkey=\"" + c.Param("deviceId") + "\ufff0\"")
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
			c.String(404, "Device not found or device currently has no sensors")
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


func GET_device_id_readings(config runtimeConfig) func(*gin.Context) {
	return func(c *gin.Context) {
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

		var paramErr error
		latest := false
		validDate := false
		var startDate time.Time
		var endDate time.Time

		if c.Query("latest") != "" {
			latest, paramErr = strconv.ParseBool(c.Query("latest"))
		} else if c.Query("startDate") != "" {
			startDate, paramErr = time.Parse("2006-01-02T15:04:05.999Z07:00", c.Query("startDate"))
			if c.Query("endDate") != "" {
				endDate, paramErr = time.Parse("2006-01-02T15:04:05.999Z07:00", c.Query("endDate"))
			} else {
				endDate = time.Now()
			}
			if paramErr == nil {
				validDate = true
			}
		}

		if paramErr != nil {
			c.String(400, "User supplied parameter error")
			return
		}

		code, resp, err := queryCouchdb(config.CouchHost + "/kentnetwork/_design/sensors/_view/getByDeviceID?startkey=\"" + c.Param("deviceId") + "\"&endkey=\"" + c.Param("deviceId") + "\ufff0\"")
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
			c.String(404, "Device not found or device has sensors with no readings")
			return
		}

		// Build OK response
		var a okResponse
		a.Meta = newMeta(resultLimit)

		for i := range couchResp.Rows {

			var readings []reading
			if latest == false && validDate == false {
				readings, err = getSensorData(couchResp.Rows[i].ID, false, time.Time{}, time.Time{}, config.InfluxDb)
			} else if latest {
				readings, err = getSensorData(couchResp.Rows[i].ID, true, time.Time{}, time.Time{}, config.InfluxDb)
			} else if validDate {
				readings, err = getSensorData(couchResp.Rows[i].ID, false, startDate, endDate, config.InfluxDb)
			}

			if err != nil {
				c.String(500, "Influxdb connection error")
				return
			}

			if a.Readings == nil {
				c.String(404, "Device not found or device has sensors with no readings")
				return
			}

			for i := range readings {
				a.Readings = append(a.Readings, readings[i])
			}

		}

		c.JSON(http.StatusOK, a)
	}
}


func PUT_devices(config runtimeConfig) func(*gin.Context) {
	return func(c *gin.Context) {
		type putData struct {
			Name string `json:"name"`
		}

		type newDev struct {
			ID string `json:"id"`
		}
		
		data := putData{}
		if err := c.BindJSON(&data); err != nil  {
			//TODO
		}

		ret := newDev{}
		c.JSON(http.StatusOK, ret)	
	}
}
