package main

import (
	"encoding/json"
	"net/http"

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

		code, resp, err := queryCouchdb(config.couchHost + "/kentnetwork/_design/sensors/_view/getByDeviceID?startkey=\"" + c.Param("deviceId") + "\"&endkey=\"" + c.Param("deviceId") + "\ufff0\"")
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
			c.String(404, "Device not found or device has sensors with no readings")
			return
		}

		// Build OK response
		var a okResponse
		a.Meta = newMeta(resultLimit)

		for i := range couchResp.Rows {
			readings, err := getSensorData(couchResp.Rows[i].ID, config.influxDb)
			if err == nil && readings != nil {
				for i := range readings {
					a.Readings = append(a.Readings, readings[i])
				}
			}
		}

		if a.Readings == nil {
			c.String(404, "Device not found or device has sensors with no readings")
			return
		}

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

		// Check if parameters have been passed.
		if c.Param("today" == true) {

		}

		readings, err := getSensorData(c.Param("sensorId"), config.influxDb)

		if err != nil {
			c.String(500, "Internal server error")
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

		code, resp, err := queryCouchdb(config.couchHost + "/kentnetwork/_design/sensors/_view/getSensors")
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
			c.String(404, "No sensors found or system has sensors with no readings")
			return
		}

		// Build OK response
		var a okResponse
		a.Meta = newMeta(resultLimit)

		for i := range couchResp.Rows {
			readings, err := getSensorData(couchResp.Rows[i].ID, config.influxDb)
			if err == nil && readings != nil {
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
