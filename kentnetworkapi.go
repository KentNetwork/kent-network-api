package main

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	auth0 "github.com/auth0-community/go-auth0"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	client "github.com/influxdata/influxdb/client/v2"
	jose "gopkg.in/square/go-jose.v2"
)

const (
	resultLimit = 100
)

var validator *auth0.JWTValidator

var events = [...]string{
	"Unseen",
	"Active",
	"Decommisioned",
	"Fault",
	"Maintenance",
}

func main() {

	runtimeFlags := doFlags()

	var config runtimeConfig
	if runtimeFlags.configFile == "" {
		config = importEnvConf()
	} else {
		config = importYmlConf(runtimeFlags.configFile)
	}

	configErr := validConfig(config)
	if configErr != nil {
		panic(configErr)
	}

	if config.Auth0.Key != "" {
		setupAuth0(config)
	}

	var err error
	config, err = config.influxDBClient()

	if err != nil {
		log.Fatal(err)
	}

	r := setupRouter(config)

	// CORS -- update
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"PUT", "PATCH", "DELETE", "GET", "POST"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Listen and Server in 0.0.0.0:80
	r.Run(config.ServerBind)
}

func setupRouter(config runtimeConfig) *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	r.GET("/status", GET_status(config))

	// If an auth0 key is defined use this for endpoints
	if config.Auth0.Key != "" {
		r.GET("/devices", Auth0Groups(), GET_devices(config))
		r.PUT("/devices", Auth0Groups(), PUT_devices(config))
		r.GET("/devices/:deviceId", Auth0Groups(), GET_devices_id(config))
		r.GET("/devices/:deviceId/sensors", Auth0Groups(), GET_devices_id_sensors(config))
		r.GET("/devices/:deviceId/readings", Auth0Groups(), GET_device_id_readings(config))
		r.GET("/sensors", Auth0Groups(), GET_sensors(config))
		r.GET("/sensors/:sensorId", Auth0Groups(), GET_sensors_id(config))
		r.GET("/sensors/:sensorId/readings", Auth0Groups(), GET_sensors_id_readings(config))
		r.GET("/data/readings", Auth0Groups(), GET_data_readings(config))
		r.GET("/gateways", Auth0Groups(), GET_gateways(config))
	} else {
		// Else no middleware -- this causes exception.
		r.GET("/devices", GET_devices(config))
		r.PUT("/devices", PUT_devices(config))
		r.GET("/devices/:deviceId", GET_devices_id(config))
		r.GET("/devices/:deviceId/sensors", GET_devices_id_sensors(config))
		r.GET("/devices/:deviceId/readings", GET_device_id_readings(config))
		r.GET("/sensors", GET_sensors(config))
		r.GET("/sensors/:sensorId", GET_sensors_id(config))
		r.GET("/sensors/:sensorId/readings", GET_sensors_id_readings(config))
		r.GET("/data/readings", GET_data_readings(config))
		r.GET("/gateways", GET_gateways(config))
	}

	return r
}

// LoadPublicKey loads a public key from PEM/DER-encoded data for jwt verifying
func LoadPublicKey(data []byte) (interface{}, error) {
	input := data

	block, _ := pem.Decode(data)
	if block != nil {
		input = block.Bytes
	}

	// Try to load SubjectPublicKeyInfo
	pub, err0 := x509.ParsePKIXPublicKey(input)
	if err0 == nil {
		return pub, nil
	}

	cert, err1 := x509.ParseCertificate(input)
	if err1 == nil {
		return cert.PublicKey, nil
	}

	return nil, fmt.Errorf("square/go-jose: parse error, got '%s' and '%s'", err0, err1)
}

func setupAuth0(config runtimeConfig) {
	publicKeyLocation := config.Auth0.Key
	//Creates a configuration with the Auth0 information
	data, err := ioutil.ReadFile(publicKeyLocation)
	if err != nil {
		panic(fmt.Sprintf("Unable to read public key from disk (%s)", publicKeyLocation))
	}

	secret, err := LoadPublicKey(data)
	if err != nil {
		panic("Invalid public key")
	}
	secretProvider := auth0.NewKeyProvider(secret)
	configuration := auth0.NewConfiguration(secretProvider, []string{"kentnetwork"}, "https://kentnetworkuk.eu.auth0.com/", jose.RS256)
	validator = auth0.NewValidator(configuration, nil)
}

func Auth0Groups(validGroups ...string) gin.HandlerFunc {

	return gin.HandlerFunc(func(c *gin.Context) {

		tok, err := validator.ValidateRequest(c.Request)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			log.Println("Invalid token:", err)
			return
		}

		claims := map[string]interface{}{}
		err = validator.Claims(c.Request, tok, &claims)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			c.Abort()
			log.Println("Invalid claims:", err)
			return
		}

		c.Next()
	})
}

func doFlags() runtimeFlags {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	var config runtimeFlags
	flag.StringVar(&config.configFile, `config`, ``, "Enter path for yaml file")
	flag.Parse()

	return config

}

func getSensorData(influx influxConfig, sensorID string, latest bool, startDate time.Time, endDate time.Time, influxDb string) (readings []reading, err error) {
	var q string
	if latest {
		q = fmt.Sprintf("SELECT last(\"value\") FROM /.*/ WHERE (\"sensor_id\" = '%s') ORDER BY time DESC LIMIT %d ", sensorID, resultLimit)
	} else if (startDate != time.Time{}) && (endDate != time.Time{}) {
		q = fmt.Sprintf("SELECT \"value\" FROM /.*/ WHERE (\"sensor_id\" = '%s' AND time >= '"+startDate.Format(time.RFC3339)+"' AND time <= '"+endDate.Format(time.RFC3339)+"') ORDER BY time DESC LIMIT %d ", sensorID, resultLimit)
	} else {
		q = fmt.Sprintf("SELECT \"value\" FROM /.*/ WHERE (\"sensor_id\" = '%s') ORDER BY time DESC LIMIT %d ", sensorID, resultLimit)
	}
	var response []client.Result
	if response, err = influx.queryInfluxDB(q, influxDb); err == nil {
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

func getGatewaysMeta(influx influxConfig, influxDb string) (gateways []gateway, err error) {
	var q string

	q = "select last(lat) as lat,lon from stat group by gatewayMac"

	var response []client.Result
	if response, err = influx.queryInfluxDB(q, influxDb); err == nil {
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
