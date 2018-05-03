package main

import (
	"github.com/TheThingsNetwork/go-app-sdk"
	client "github.com/influxdata/influxdb/client/v2"
	"io/ioutil"
	"net/http"
)

type ttnConfig struct {
	AppID         string `yaml:"appID"`
	AppAccessKey  string `yaml:"appAccessKey"`
	SdkClientName string `yaml:"sdkClientName"`
	init          bool
	client        ttnsdk.Client
}

func (ttn ttnConfig) connect() ttnsdk.Client {
	/*
		if ttn.init {
			return ttn.client
		}*/

	ttn_config := ttnsdk.NewCommunityConfig(ttn.SdkClientName)
	ttn_config.ClientVersion = "2.0.5"

	ttn.client = ttn_config.NewClient(ttn.AppID, ttn.AppAccessKey)
	ttn.init = true
	return ttn.client
}

type influxConfig struct {
	Host   string `yaml:"host"`
	User   string `yaml:"user"`
	Pwd    string `yaml:"password"`
	Db     string `yaml:"db"`
	client client.Client
}

func (c influxConfig) influxDBClient() error {
	config := client.HTTPConfig{
		Addr:     c.Host,
		Username: c.User,
		Password: c.Pwd}

	client, err := client.NewHTTPClient(config)
	c.client = client
	return err
}

// queryInfluxDB convenience function to query the influx database
func (c influxConfig) queryInfluxDB(cmd string, database string) (res []client.Result, err error) {
	q := client.Query{
		Command:  cmd,
		Database: database,
	}
	if response, err := c.client.Query(q); err == nil {
		if response.Error() != nil {
			return res, response.Error()
		}
		res = response.Results
	} else {
		return res, err
	}
	return res, nil
}

type auth0Config struct {
	Key string `yaml:"key,omitempty"`
}

type couchConfig struct {
	Host string `yaml:"couchhost"`
}

func (c couchConfig) query(request string) (code int, response []byte, err error) {
	request = c.Host + request
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

type runtimeConfig struct {
	ServerBind string       `yaml:"serverbind"`
	Couch      couchConfig  `yaml:"couch"`
	Auth0      auth0Config  `yaml:"auth0,omitempty"`
	Influx     influxConfig `yaml:"influx"`
	TTN        ttnConfig    `yaml:"ttn"`
}

type runtimeFlags struct {
	configFile string
}
