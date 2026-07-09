package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"marriage/internal/response"
	"marriage/internal/services"
)

type ConfirmationController struct {
	service services.ConfirmationService
}

func NewConfirmationController() ConfirmationController {
	return ConfirmationController{
		service: services.NewConfirmationService(),
	}
}

func (controller ConfirmationController) Create(writer http.ResponseWriter, request *http.Request) {
	var confirmations []services.ConfirmationRequest
	if err := json.NewDecoder(request.Body).Decode(&confirmations); err != nil {
		response.JSON(writer, http.StatusBadRequest, response.Error{
			Error: "invalid JSON body",
		})
		return
	}

	result, err := controller.service.Create(request.Context(), confirmations)
	if err != nil {
		var validationError services.ValidationError
		if errors.As(err, &validationError) {
			response.JSON(writer, http.StatusBadRequest, response.ValidationError{
				Error:   "invalid confirmation request",
				Details: validationError.Details,
			})
			return
		}

		response.JSON(writer, http.StatusInternalServerError, response.Error{
			Error: http.StatusText(http.StatusInternalServerError),
		})
		return
	}

	response.JSON(writer, http.StatusAccepted, result)
}
