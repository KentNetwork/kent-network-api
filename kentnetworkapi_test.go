package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testConfig = runtimeConfig{
		influxUser: `river`,
		influxPwd:  `NCQxM3Socdc2K4nEwS`,
		serverBind: ":80",
		influxHost: `https://influxdb.kent.network`,
		couchHost:  `https://couchdb.kent.network`,
	}
)

func TestDevicesRoute(t *testing.T) {
	router := setupRouter(testConfig)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/devices", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.JSONEq(t, `{"meta":{"publisher":"Kent Network","license":"Creative Commons","version":"0.1","resultLimit":0},"devices":[{"@id":"kent.network/devices/280bc7e6-5313-4764-880f-0b2131ce0589","location":{"nearestTown":"canterbury","catchmentName":"","associatedWith":"","lat":51.28325,"lon":1.080233,"altitude":10,"easting":"","northing":""},"ttn":{"appId":"","devId":"","hardwareSerial":""},"hardwareRef":"0.0.1","batteryType":"eneloop"},{"@id":"kent.network/devices/2cd66aaf-c920-4268-a5d9-7360a15877b6","location":{"nearestTown":"canterbury","catchmentName":"","associatedWith":"","lat":51.28325,"lon":1.080233,"altitude":10,"easting":"","northing":""},"ttn":{"appId":"","devId":"","hardwareSerial":""},"hardwareRef":"0.0.1","batteryType":"eneloop"},{"@id":"kent.network/devices/device:testsen1","location":{"nearestTown":"ashford","catchmentName":"","associatedWith":"","lat":0,"lon":0,"altitude":0,"easting":"","northing":""},"ttn":{"appId":"","devId":"","hardwareSerial":""},"hardwareRef":"0.0.1","batteryType":"eneloop"}]}`, w.Body.String())
}

func TestDeviceRoute(t *testing.T) {
	router := setupRouter(testConfig)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/devices/device:testsen1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.JSONEq(t, `{"meta":{"publisher":"Kent Network","license":"Creative Commons","version":"0.1","resultLimit":0},"device":{"@id":"kent.network/devices/device:testsen1","type":"device","location":{"nearestTown":"ashford","catchmentName":"","associatedWith":"","lat":0,"lon":0,"altitude":0,"easting":"","northing":""},"ttn":{"appId":"","devId":"","hardwareSerial":""},"hardwareRef":"0.0.1","batteryType":"eneloop"}}`, w.Body.String())
}

func TestDeviceSensorsRoute(t *testing.T) {
	router := setupRouter(testConfig)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/devices/device:testsen1/sensors", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.JSONEq(t, `{"message": "{"meta":{"publisher":"Kent Network","license":"Creative Commons","version":"0.1","resultLimit":0},"sensors":[{"@id":"kent.network/sensors/device:testsen1:sensorid:2","updateInterval":10,"sensorType":"waterTemperature","unit":"c"}]}"}`, w.Body.String())
}

func TestSensorsRoute(t *testing.T) {
	router := setupRouter(testConfig)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/sensors", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.JSONEq(t, `{"meta":{"publisher":"Kent Network","license":"Creative Commons","version":"0.1","resultLimit":0},"sensors":[{"@id":"kent.network/sensors/280bc7e6-5313-4764-880f-0b2131ce0589:01","updateInterval":15,"sensorType":"waterTemperature","unit":"c"},{"@id":"kent.network/sensors/2cd66aaf-c920-4268-a5d9-7360a15877b6:1","updateInterval":15,"sensorType":"waterTemperature","unit":"c"},{"@id":"kent.network/sensors/2cd66aaf-c920-4268-a5d9-7360a15877b6:2","updateInterval":15,"sensorType":"waterFlow","unit":"cfm"},{"@id":"kent.network/sensors/device:testsen1:sensorid:2","updateInterval":10,"sensorType":"waterTemperature","unit":"c"}]}`, w.Body.String())
}

func TestSensorRoute(t *testing.T) {
	router := setupRouter(testConfig)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/sensors/device:testsen1:sensorid:2", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.JSONEq(t, `{"meta":{"publisher":"Kent Network","license":"Creative Commons","version":"0.1","resultLimit":0},"sensor":{"@id":"kent.network/sensors/device:testsen1:sensorid:2","updateInterval":10,"sensorType":"waterTemperature","unit":"c"}}`, w.Body.String())
}

func TestDataReadingsRoute(t *testing.T) {
	router := setupRouter(testConfig)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/data/readings", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.JSONEq(t, `{"message": "Here is all the readings from all the devices"}`, w.Body.String())
}
