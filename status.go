package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type serviceMessage struct {
	Title string `json:"title"`
	Created time.Time `json:"created"`
	LastUpdated time.Time `json:lastupdated"`
	Message string `json:"message"`
}

type serviceStatus struct {
	Service  string   `json:"service"`
	Status   string   `json:"status"`
	Messages []serviceMessage`json:"messages"`
}

func getInfluxStatus(config runtimeConfig) serviceStatus {
	return serviceStatus{
		Service:  "influx",
		Status:   "unknown",
		Messages: []serviceMessage{},
	}

}

func GET_status(config runtimeConfig) func(c *gin.Context) {
	return func(c *gin.Context) {
		type okResponse struct {
			Status   string          `json:"status"`
			Services []serviceStatus `json:"services"`
			Messages []serviceMessage`json:"messages"`
		}

		services := []serviceStatus{getInfluxStatus(config)}

		// Build OK response
		var a okResponse
		a.Status = "Operational"
		a.Services = services
		a.Messages = []serviceMessage{serviceMessage{
			Title: "test message",
			Created: time.Now(),
			LastUpdated: time.Now(),
			Message: "I think it's ok",
		},

		}
		c.JSON(http.StatusOK, a)
	}
}
