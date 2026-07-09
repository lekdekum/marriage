package controllers

import (
	"net/http"

	"marriage/internal/response"
	"marriage/internal/services"
)

type HealthController struct {
	service services.HealthService
}

func NewHealthController() HealthController {
	return HealthController{
		service: services.NewHealthService(),
	}
}

func (controller HealthController) Show(writer http.ResponseWriter, request *http.Request) {
	health := controller.service.Check(request.Context())

	response.JSON(writer, http.StatusOK, health)
}
