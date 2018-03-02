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
	assert.JSONEq(t, `{"message": "Here are all the devices"}`, w.Body.String())
}

func TestDevicesRouteMissingParamField(t *testing.T) {
	router := setupRouter(testConfig)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/devices?loc-lat=51.23", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
	assert.Equal(t, `Invalid parameters`, w.Body.String())
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
	req, _ := http.NewRequest("GET", "/devices/boing/sensors", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.JSONEq(t, `{"message": "Here are all the sensors for a device"}`, w.Body.String())
}

func TestDeviceReadingsRoute(t *testing.T) {
	router := setupRouter(testConfig)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/devices/boing/readings", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.JSONEq(t, `{"message": "Here are all the readings for this device"}`, w.Body.String())
}

func TestDeviceReadingsRouteMissingParamField(t *testing.T) {
	router := setupRouter(testConfig)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/devices/boing/readings?startDate=blah", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
	assert.Equal(t, `Invalid parameters`, w.Body.String())
}

func TestSensorsRoute(t *testing.T) {
	router := setupRouter(testConfig)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/sensors", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.JSONEq(t, `{"message": "Here are all the sensors"}`, w.Body.String())
}

func TestSensorRoute(t *testing.T) {
	router := setupRouter(testConfig)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/sensors/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.JSONEq(t, `{"message": "Here is a sensor"}`, w.Body.String())
}

func TestSensorReadingsRouteMissingParamField(t *testing.T) {
	router := setupRouter(testConfig)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/sensors/R_T_test/readings?startDate=adate", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
	assert.Equal(t, `Invalid parameters`, w.Body.String())
}

func TestSensorReadingsRouteBadSensorId(t *testing.T) {
	router := setupRouter(testConfig)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/sensors/test/readings?startDate=adate", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 404, w.Code)
	assert.Equal(t, `Sensor not found`, w.Body.String())
}

func TestDataReadingsRoute(t *testing.T) {
	router := setupRouter(testConfig)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/data/readings", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.JSONEq(t, `{"message": "Here is all the readings from all the devices"}`, w.Body.String())
}

func TestDataReadingsRouteMissingParamField(t *testing.T) {
	router := setupRouter(testConfig)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/data/readings?startDate=adate", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
	assert.Equal(t, `Invalid parameters`, w.Body.String())
}
