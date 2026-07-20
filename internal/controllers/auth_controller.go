package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"marriage/internal/response"
	"marriage/internal/services"
)

type AuthController struct {
	service services.AuthService
}

func NewAuthController(service services.AuthService) AuthController {
	return AuthController{service: service}
}

func (controller AuthController) Login(writer http.ResponseWriter, request *http.Request) {
	var loginRequest services.LoginRequest
	if err := json.NewDecoder(request.Body).Decode(&loginRequest); err != nil {
		response.JSON(writer, http.StatusBadRequest, response.Error{
			Error: "invalid JSON body",
		})
		return
	}

	result, err := controller.service.Login(loginRequest)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			response.JSON(writer, http.StatusUnauthorized, response.Error{
				Error: http.StatusText(http.StatusUnauthorized),
			})
			return
		}

		response.JSON(writer, http.StatusInternalServerError, response.Error{
			Error: http.StatusText(http.StatusInternalServerError),
		})
		return
	}

	response.JSON(writer, http.StatusOK, result)
}
