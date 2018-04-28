package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type serviceStatus struct {
	Service  string   `json:"service"`
	Status   string   `json:"status"`
	Messages []string `json:"messages"`
}

func getInfluxStatus(config runtimeConfig) serviceStatus {
	return serviceStatus{
		Service:  "influx",
		Status:   "unknown",
		Messages: []string{},
	}

}

func GET_status(config runtimeConfig) func(c *gin.Context) {
	return func(c *gin.Context) {
		type okResponse struct {
			Status   string          `json:"status"`
			Services []serviceStatus `json:"services"`
			Messages []string        `json:"messages"`
		}

		services := []serviceStatus{getInfluxStatus(config)}

		// Build OK response
		var a okResponse
		a.Status = "Operational"
		a.Services = services
		a.Messages = []string{"Ok... I think"}
		c.JSON(http.StatusOK, a)
	}
}
