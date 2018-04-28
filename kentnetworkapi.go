package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	jwt "github.com/appleboy/gin-jwt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	client "github.com/influxdata/influxdb/client/v2"
	"gopkg.in/yaml.v2"
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

	runtimeFlags := doFlags()
	config := importYmlConf(runtimeFlags.configFile)

	var err error
	influxClient, err = influxDBClient(config)
	if err != nil {
		log.Fatal(err)
	}
	r := setupRouter(config)

	// Listen and Server in 0.0.0.0:80
	r.Run(config.ServerBind)
}

func setupRouter(config runtimeConfig) *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	authMiddleware := getJWTMiddleware()

	r.Use(cors.Default())
	r.POST("/login", authMiddleware.LoginHandler)
	r.GET("/status", GET_status(config))
	r.GET("/login/token", GET_token(config))

	auth := r.Group("/")
	auth.Use(authMiddleware.MiddlewareFunc())
	auth.Use(cors.Default())

	auth.GET("/devices", GET_devices(config))
	auth.GET("/devices/:deviceId", GET_devices_id(config))
	auth.GET("/devices/:deviceId/sensors", GET_devices_id_sensors(config))
	auth.GET("/devices/:deviceId/readings", GET_device_id_readings(config))
	auth.GET("/sensors", GET_sensors(config))
	auth.GET("/sensors/:sensorId", GET_sensors_id(config))
	auth.GET("/sensors/:sensorId/readings", GET_sensors_id_readings(config))
	auth.GET("/data/readings", GET_data_readings(config))
	auth.GET("/gateways", GET_gateways(config))
	return r
}

func getJWTMiddleware() *jwt.GinJWTMiddleware {
	return &jwt.GinJWTMiddleware{
		Realm:      "test zone",
		Key:        []byte("secret key"),
		Timeout:    time.Hour,
		MaxRefresh: time.Hour,
		Authenticator: func(userId string, password string, c *gin.Context) (string, bool) {
			if (userId == "admin" && password == "admin") || (userId == "test" && password == "test") {
				return userId, true
			}

			return userId, false
		},
		Authorizator: func(userId string, c *gin.Context) bool {
			if userId == "admin" {
				return true
			}

			return false
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		},
		// TokenLookup is a string in the form of "<source>:<name>" that is used
		// to extract token from the request.
		// Optional. Default value "header:Authorization".
		// Possible values:
		// - "header:<name>"
		// - "query:<name>"
		// - "cookie:<name>"
		TokenLookup: "header:Authorization",
		// TokenLookup: "query:token",
		// TokenLookup: "cookie:token",

		// TokenHeadName is a string in the header. Default value is "Bearer"
		TokenHeadName: "Bearer",

		// TimeFunc provides the current time. You can override it to use another time value. This is useful for testing or if your server uses a different time zone than your tokens.
		TimeFunc: time.Now,
	}
}

func doFlags() runtimeFlags {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	var config runtimeFlags
	flag.StringVar(&config.configFile, `config`, `config.yaml`, "Enter path for yaml file")
	flag.Parse()

	return config

}

func importYmlConf(yamlFilePath string) runtimeConfig {
	var config runtimeConfig
	yamlFile, err := ioutil.ReadFile(yamlFilePath)
	if err != nil {
		panic("Error reading yaml config")
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		panic("Error unmarshalling yaml config")
	}
	return config
}

func influxDBClient(c runtimeConfig) (client.Client, error) {
	config := client.HTTPConfig{
		Addr:     c.InfluxHost,
		Username: c.InfluxUser,
		Password: c.InfluxPwd}

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

func getSensorData(sensorID string, latest bool, startDate time.Time, endDate time.Time, influxDb string) (readings []reading, err error) {
	var q string
	if latest {
		q = fmt.Sprintf("SELECT last(\"value\") FROM /.*/ WHERE (\"sensor_id\" = '%s') ORDER BY time DESC LIMIT %d ", sensorID, resultLimit)
	} else if (startDate != time.Time{}) && (endDate != time.Time{}) {
		q = fmt.Sprintf("SELECT \"value\" FROM /.*/ WHERE (\"sensor_id\" = '%s' AND time >= '"+startDate.Format(time.RFC3339)+"' AND time <= '"+endDate.Format(time.RFC3339)+"') ORDER BY time DESC LIMIT %d ", sensorID, resultLimit)
	} else {
		q = fmt.Sprintf("SELECT \"value\" FROM /.*/ WHERE (\"sensor_id\" = '%s') ORDER BY time DESC LIMIT %d ", sensorID, resultLimit)
	}
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

func getGatewaysMeta(influxDb string) (gateways []gateway, err error) {
	var q string

	q = "select last(lat) as lat,lon from stat group by gatewayMac"

	var response []client.Result
	if response, err = queryInfluxDB(influxClient, q, influxDb); err == nil {
		if response[0].Series == nil {
			return nil, nil
		}

		for i := range response[0].Series {
			r := response[0].Series[i].Tags["gatewayMac"]
			s, sErr := response[0].Series[i].Values[0][1].(json.Number).Float64()
			t, tErr := response[0].Series[i].Values[0][2].(json.Number).Float64()
			if sErr == nil && tErr == nil {
				var k gateway
				k.GatewayMac = r
				k.Lat = s
				k.Lon = t
				gateways = append(gateways, k)
			}
		}
		return gateways, nil
	}
	return gateways, err
}
