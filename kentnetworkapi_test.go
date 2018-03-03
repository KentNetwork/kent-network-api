package main

import (
	"testing"

	"net/http"
	"net/http/httptest"

	. "github.com/smartystreets/goconvey/convey"
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

func TestGetDevices(t *testing.T) {
	router := setupRouter(testConfig)
	Convey("Subject: Test device based routes", t, func() {

		Convey("Test: /devices responds appropriately:", func() {

			Convey("When a valid HTTP request is made to it", func() {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/devices", nil)
				router.ServeHTTP(w, req)
				Convey("Then the response code should be 200", nil)
				So(w.Code, ShouldEqual, 200)
			})

			Convey("When an internal server error occurs", func() {
				Convey("Then the response code should be 500", nil)
				Convey("With the msg \"Internal server error\"", nil)

			})

		})

		Convey("Test: /devices/device_ID responds appropriately:", func() {

			Convey("When a valid HTTP request is made to it", func() {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/devices/device:testsen1", nil)
				router.ServeHTTP(w, req)
				Convey("Then the response code should be 200", nil)
				So(w.Code, ShouldEqual, 200)
			})

			Convey("When an invalid device_id is supplied", func() {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/devices/badrobot", nil)
				router.ServeHTTP(w, req)
				Convey("Then the response code should be 404", nil)
				So(w.Code, ShouldEqual, 404)
				Convey("With the msg \"Device not found\"", nil)
				So(w.Body.String(), ShouldEqual, "Device not found")
			})

			Convey("When an internal server error occurs", func() {

				Convey("Then the response code should be 500", nil)

				Convey("With the msg \"Internal server error\"", nil)

			})

		})

		Convey("Test: /devices/device_ID/sensors responds appropiately:", func() {

			Convey("When a valid HTTP request is made to it", func() {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/devices/device:testsen1/sensors", nil)
				router.ServeHTTP(w, req)
				Convey("Then the response code should be 200", nil)
				So(w.Code, ShouldEqual, 200)
			})

			Convey("When an invalid device_id is supplied", func() {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/devices/badrobot/sensors", nil)
				router.ServeHTTP(w, req)
				Convey("Then the response code should be 404", nil)
				So(w.Code, ShouldEqual, 404)
				Convey("With the msg \"Device not found\"", nil)
				So(w.Body.String(), ShouldEqual, "Device not found or device currently has no sensors")
			})

			Convey("When an internal server error occurs", func() {

				Convey("Then the response code should be 500", nil)

				Convey("With the msg \"Internal server error\"", nil)

			})

		})

	})

	Convey("Subject: Test sensor based routes", t, func() {

		Convey("Test: /sensors responds appropriately:", func() {

			Convey("When a valid HTTP request is made to it", func() {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/sensors", nil)
				router.ServeHTTP(w, req)
				Convey("Then the response code should be 200", nil)
				So(w.Code, ShouldEqual, 200)
			})

			Convey("When an internal server error occurs", func() {

				Convey("Then the response code should be 500", nil)

				Convey("With the msg \"Internal server error\"", nil)

			})

		})

		Convey("Test: /sensors/sensor_id responds appropriately:", func() {

			Convey("When a valid HTTP request is made to it", func() {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/sensors/device:testsen1:sensorid:2", nil)
				router.ServeHTTP(w, req)
				Convey("Then the response code should be 200", nil)
				So(w.Code, ShouldEqual, 200)
			})

			Convey("When an invalid device_id is supplied", func() {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/sensors/badrobot", nil)
				router.ServeHTTP(w, req)
				Convey("Then the response code should be 404", nil)
				So(w.Code, ShouldEqual, 404)
				Convey("With the msg \"Device not found\"", nil)
				So(w.Body.String(), ShouldEqual, "Sensor not found")
			})

			Convey("When an internal server error occurs", func() {

				Convey("Then the response code should be 500", nil)

				Convey("With the msg \"Internal server error\"", nil)

			})

		})

	})
}
