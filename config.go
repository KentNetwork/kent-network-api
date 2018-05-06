package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/TheThingsNetwork/go-app-sdk"
	client "github.com/influxdata/influxdb/client/v2"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type ttnConfig struct {
	AppID         string `yaml:"appID"`
	AppAccessKey  string `yaml:"appAccessKey"`
	SdkClientName string `yaml:"sdkClientName"`
	init          bool
	client        ttnsdk.Client
}

func validConfig(config runtimeConfig) error {
	log.Printf("%S", config)
	if config.Couch.Host == "" {
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

func (c runtimeConfig) influxDBClient() (runtimeConfig, error) {
	i := c.Influx
	config := client.HTTPConfig{
		Addr:     i.Host,
		Username: i.User,
		Password: i.Pwd}

	client, err := client.NewHTTPClient(config)
	c.Influx.client = client
	return c, err
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
	Host string `yaml:"host"` //Host to connect to for CouchDB e.g. "http://couch.example.com"
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

func (c couchConfig) put(request string, body interface{}) (code int, response []byte, err error) {
	request = c.Host + request

	client := &http.Client{}
	data, err := json.Marshal(body)
	if err != nil {
		// handle error
		log.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPut, request, bytes.NewReader(data))
	if err != nil {
		// handle error
		log.Fatal(err)
	}
	resp, err := client.Do(req)

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

// Runtime configuration. This should be considdered immutable and all methods that modify it should return a new copy.
type runtimeConfig struct {
	ServerBind string       `yaml:"serverbind"`
	Couch      couchConfig  `yaml:"couch"`
	Auth0      auth0Config  `yaml:"auth0,omitempty"`
	Influx     influxConfig `yaml:"influx"`
	TTN        ttnConfig    `yaml:"ttn"`
}

// Configuration options that can be set by "flags"
type runtimeFlags struct {
	configFile string // Location of the config File. If this is null the application was assume config is being passed in by ENVARGS
}

func importYmlConf(yamlFilePath string) runtimeConfig {
	var config runtimeConfig
	yamlFile, err := ioutil.ReadFile(yamlFilePath)
	if err != nil {
		panic(fmt.Sprintf("Error reading yaml config (%s)", err.Error()))
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		panic(fmt.Sprintf("Error unmarshalling yaml config (%s:%s)", yamlFilePath, err.Error()))
	}
	config = config.init()
	return config
}

func importEnvConf() runtimeConfig {
	var config runtimeConfig

	config.Couch.Host = os.Getenv("COUCHHOST")
	config.Influx = influxConfig{
		Db:   os.Getenv("INFLUXDB"),
		Host: os.Getenv("INFLUXHOST"),
		Pwd:  os.Getenv("INFLUXPWD"),
		User: os.Getenv("INFLUXUSER"),
	}

	config.TTN = ttnConfig{
		AppAccessKey:  os.Getenv("TTNAPPKEY"),
		AppID:         os.Getenv("TTNAPPID"),
		SdkClientName: os.Getenv("TTNSDKCLIENTNAME"),
	}

	config.ServerBind = os.Getenv("SERVERBIND")
	config.Auth0.Key = os.Getenv("AUTH0KEY")

	config = config.init()

	return config
}

// Check config and init clients.
func (c runtimeConfig) init() runtimeConfig {
	if err := validConfig(c); err != nil {
		panic(err)
	}

	var err error
	if c, err = c.influxDBClient(); err != nil {
		panic(err)
	}

	return c
}
