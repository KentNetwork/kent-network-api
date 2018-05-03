package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type serviceMessage struct {
	Title       string    `json:"title"`
	Created     time.Time `json:"created"`
	LastUpdated time.Time `json:lastupdated"`
	Message     string    `json:"message"`
}

type serviceStatus struct {
	Service  string           `json:"service"`
	Status   string           `json:"status"`
	Messages []serviceMessage `json:"messages"`
}

func getCouchStatus(config runtimeConfig) serviceStatus {
	status := "ok"

	_, err := http.Get(config.Couch.Host + "/")
	if err != nil {
		status = "error"
	}

	return serviceStatus{
		Service:  "couchDB",
		Status:   status,
		Messages: []serviceMessage{},
	}

}

func getInfluxStatus(config runtimeConfig) serviceStatus {
	status := "ok"

	//TODO: goto warning when ping is slow...
	_, _, err := config.Influx.client.Ping(10)
	if err != nil {
		status = "error"
	}

	return serviceStatus{
		Service:  "influx",
		Status:   status,
		Messages: []serviceMessage{},
	}

}

func GET_status(config runtimeConfig) func(c *gin.Context) {
	return func(c *gin.Context) {
		type okResponse struct {
			Status   string           `json:"status"`
			Services []serviceStatus  `json:"services"`
			Messages []serviceMessage `json:"messages"`
		}

		services := []serviceStatus{
			getInfluxStatus(config),
			getCouchStatus(config),
		}

		// Build OK response
		var a okResponse
		a.Status = "ok"
		a.Services = services
		a.Messages = []serviceMessage{serviceMessage{
			Title:       "test message",
			Created:     time.Now(),
			LastUpdated: time.Now(),
			Message:     "I think it's ok",
		},
		}
		c.JSON(http.StatusOK, a)
	}
}
