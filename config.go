package main

import (
	"errors"

	"github.com/TheThingsNetwork/go-app-sdk"
	client "github.com/influxdata/influxdb/client/v2"
)

type ttnConfig struct {
	AppID         string `yaml:"appID"`
	AppAccessKey  string `yaml:"appAccessKey"`
	SdkClientName string `yaml:"sdkClientName"`
	init          bool
	client        ttnsdk.Client
}

func validConfig(config runtimeConfig) error {
	if config.Auth0.Key == "" {
		return errors.New("Parameter: missing auth0 key")
	} else if config.CouchHost == "" {
		return errors.New("Parameter: missing couch host")
	} else if config.ServerBind == "" {
		return errors.New("Parameter: missing server bind")
	} else if config.Influx.Db == "" {
		return errors.New("Parameter: missing influx db")
	} else if config.Influx.Pwd == "" {
		return errors.New("Parameter: missing influx password")
	} else if config.Influx.User == "" {
		return errors.New("Parameter: missing influx user")
	} else if config.Influx.Host == "" {
		return errors.New("Parameter: missing influx host")
	} else if config.TTN.AppAccessKey == "" {
		return errors.New("Parameter: missing TTN app access key")
	} else if config.TTN.AppID == "" {
		return errors.New("Parameter: missing TTN app id")
	} else if config.TTN.SdkClientName == "" {
		return errors.New("Parameter: missing TTN sdk client name")
	}

	return nil
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

type runtimeConfig struct {
	ServerBind string        `yaml:"serverbind"`
	CouchHost  string        `yaml:"couchhost"`
	Auth0      auth0Config   `yaml:"auth0,omitempty"`
	Influx     *influxConfig `yaml:"influx"`
	TTN        *ttnConfig    `yaml:"ttn"`
}

type runtimeFlags struct {
	configFile string
}
