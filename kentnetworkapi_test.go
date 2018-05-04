package main

import (
	"testing"

	"net/http"
	"net/http/httptest"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	badTestConfig = runtimeConfig{
		Influx:     influxConfig{},
		ServerBind: `:80`,
		Couch:      couchConfig{},
	}
)

func TestRoutes(t *testing.T) {
	testConfig := importYmlConf("config.yaml")
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

		})

		Convey("Test: /devices/device_ID/readings responds appropiately:", func() {

			Convey("When a valid HTTP request is made to it", func() {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/devices/device:testsen1/readings", nil)
				router.ServeHTTP(w, req)
				Convey("Then the response code should be 200", nil)
				So(w.Code, ShouldEqual, 200)
			})

			Convey("When an invalid device_id is supplied", func() {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/devices/badrobot/readings", nil)
				router.ServeHTTP(w, req)
				Convey("Then the response code should be 404", nil)
				So(w.Code, ShouldEqual, 404)
				Convey("With the msg \"Device not found...\"", nil)
				So(w.Body.String(), ShouldEqual, "Device not found or device has sensors with no readings")
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

		})

		Convey("Test: /sensors/sensor_id/readings responds appropriately:", func() {

			Convey("When a valid HTTP request is made to it", func() {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/sensors/device:testsen1:sensorid:2/readings", nil)
				router.ServeHTTP(w, req)
				Convey("Then the response code should be 200", nil)
				So(w.Code, ShouldEqual, 200)
			})

			Convey("When an invalid device_id is supplied", func() {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/sensors/badrobot/readings", nil)
				router.ServeHTTP(w, req)
				Convey("Then the response code should be 404", nil)
				So(w.Code, ShouldEqual, 404)
				Convey("With the msg \"Device not found\"", nil)
				So(w.Body.String(), ShouldEqual, "Sensor not found or sensor has no readings")
			})

		})
	})
	Convey("Subject: Test data based routes", t, func() {

		Convey("Test: /data/readings responds appropriately:", func() {

			Convey("When a valid HTTP request is made to it", func() {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/data/readings", nil)
				router.ServeHTTP(w, req)
				Convey("Then the response code should be 200", nil)
				So(w.Code, ShouldEqual, 200)
			})

		})
	})
}

func Test500Handling(t *testing.T) {
	router := setupRouter(badTestConfig)
	Convey("Subject: Test device based routes", t, func() {

		Convey("Test: /devices responds appropriately:", func() {

			Convey("When an internal server error occurs", func() {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/devices", nil)
				router.ServeHTTP(w, req)
				Convey("Then the response code should be 500", nil)
				So(w.Code, ShouldEqual, 500)
				Convey("With the msg \"Internal server error\"", nil)
				So(w.Body.String(), ShouldEqual, "Internal server error")
			})

		})

		Convey("Test: /devices/device_ID responds appropriately:", func() {

			Convey("When an internal server error occurs", func() {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/devices/device:testsen1", nil)
				router.ServeHTTP(w, req)
				Convey("Then the response code should be 500", nil)
				So(w.Code, ShouldEqual, 500)
				Convey("With the msg \"Internal server error\"", nil)
				So(w.Body.String(), ShouldEqual, "Internal server error")
			})

		})

		Convey("Test: /devices/device_ID/sensors responds appropiately:", func() {

			Convey("When an internal server error occurs", func() {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/devices/testsen1/sensors", nil)
				router.ServeHTTP(w, req)
				Convey("Then the response code should be 500", nil)
				So(w.Code, ShouldEqual, 500)
				Convey("With the msg \"Internal server error\"", nil)
				So(w.Body.String(), ShouldEqual, "Internal server error")
			})

		})

		Convey("Test: /devices/device_ID/readings responds appropiately:", func() {

			Convey("When an internal server error occurs", func() {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/devices/testsen1/readings", nil)
				router.ServeHTTP(w, req)
				Convey("Then the response code should be 500", nil)
				So(w.Code, ShouldEqual, 500)
				Convey("With the msg \"Internal server error\"", nil)
				So(w.Body.String(), ShouldEqual, "Internal server error")
			})

		})

	})

	Convey("Subject: Test sensor based routes", t, func() {

		Convey("Test: /sensors responds appropriately:", func() {

			Convey("When an internal server error occurs", func() {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/sensors", nil)
				router.ServeHTTP(w, req)
				Convey("Then the response code should be 500", nil)
				So(w.Code, ShouldEqual, 500)
				Convey("With the msg \"Internal server error\"", nil)
				So(w.Body.String(), ShouldEqual, "Internal server error")
			})

		})

		Convey("Test: /sensors/sensor_id responds appropriately:", func() {

			Convey("When an internal server error occurs", func() {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/sensors/device:testsen1:sensorid:2", nil)
				router.ServeHTTP(w, req)
				Convey("Then the response code should be 500", nil)
				So(w.Code, ShouldEqual, 500)
				Convey("With the msg \"Internal server error\"", nil)
				So(w.Body.String(), ShouldEqual, "Internal server error")
			})

		})

		Convey("Test: /sensors/sensor_id/readings responds appropriately:", func() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/sensors/device:testsen1:sensorid:2/readings", nil)
			router.ServeHTTP(w, req)
			Convey("When an internal server error occurs", func() {
				Convey("Then the response code should be 500", nil)
				So(w.Code, ShouldEqual, 500)
				Convey("With the msg \"Internal server error\"", nil)
				So(w.Body.String(), ShouldEqual, "Internal server error")
			})

		})

	})

	Convey("Subject: Test data based routes", t, func() {

		Convey("Test: /data/readings responds appropriately:", func() {

			Convey("When an internal server error occurs", func() {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/data/readings", nil)
				router.ServeHTTP(w, req)
				Convey("Then the response code should be 500", nil)
				So(w.Code, ShouldEqual, 500)
				Convey("With the msg \"Internal server error\"", nil)
				So(w.Body.String(), ShouldEqual, "Internal server error")
			})

		})
	})
}
