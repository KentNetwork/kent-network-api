package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	client "github.com/influxdata/influxdb/client/v2"
)

const (
	resultLimit = 100
)

var influxClient client.Client

var events = [...]string{
	"Unseen",
	"Active",
	"Decommisioned",
	"Fault",
	"Maintenance",
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

// String() function will return the english name
// that we want out constant events be recognized as
func (event eventType) String() string {
	return events[event-1]
}

func setupRouter(config runtimeConfig) *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	r.GET("/devices", GET_devices(config))
	r.GET("/devices/:deviceId", GET_devices_id(config))
	r.GET("/devices/:deviceId/sensors", GET_devices_id_sensors(config))
	r.GET("/devices/:deviceId/readings", GET_device_id_readings(config))
	r.GET("/sensors", GET_sensors(config))
	r.GET("/sensors/:sensorId", GET_sensors_id(config))
	r.GET("/sensors/:sensorId/readings", GET_sensors_id_readings(config))
	r.GET("/data/readings", GET_data_readings(config))

	return r
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

func getSensorData(sensorID string, influxDb string) (readings []reading, err error) {

	q := fmt.Sprintf("SELECT \"value\" FROM /.*/ WHERE (\"sensor_id\" = '%s') ORDER BY time DESC LIMIT %d ", sensorID, resultLimit)
	var response []client.Result
	if response, err = queryInfluxDB(influxClient, q, influxDb); err == nil {
		if response[0].Series == nil {
			return nil, nil
		}

		for i := range response[0].Series[0].Values {
			s, sErr := response[0].Series[0].Values[i][1].(json.Number).Float64()
			t, tErr := time.Parse(time.RFC3339, response[0].Series[0].Values[i][0].(string))
			if sErr == nil && tErr == nil {
				var k reading
				k.Sensor = sensorID
				k.DateTime = t.Format("2006-01-02T15:04:05.999Z07:00")
				k.Value = s
				readings = append(readings, k)
			}
		}
		return readings, nil
	}
	return readings, err
}
