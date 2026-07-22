package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"marriage/internal/response"
	"marriage/internal/services"
)

type ReservedGiftController struct {
	service services.ReservedGiftService
}

func NewReservedGiftController(service services.ReservedGiftService) ReservedGiftController {
	return ReservedGiftController{
		service: service,
	}
}

func (controller ReservedGiftController) Create(writer http.ResponseWriter, request *http.Request) {
	var reservedGiftRequest services.ReservedGiftRequest
	if err := json.NewDecoder(request.Body).Decode(&reservedGiftRequest); err != nil {
		response.JSON(writer, http.StatusBadRequest, response.Error{
			Error: "invalid JSON body",
		})
		return
	}

	reservedGift, err := controller.service.Create(request.Context(), reservedGiftRequest)
	if err != nil {
		var validationError services.ValidationError
		if errors.As(err, &validationError) {
			response.JSON(writer, http.StatusBadRequest, response.ValidationError{
				Error:   "invalid reserved gift request",
				Details: validationError.Details,
			})
			return
		}

		var conflictError services.ConflictError
		if errors.As(err, &conflictError) {
			response.JSON(writer, http.StatusConflict, response.Error{
				Error: "reserved gift already exists",
			})
			return
		}

		response.JSON(writer, http.StatusInternalServerError, response.Error{
			Error: http.StatusText(http.StatusInternalServerError),
		})
		return
	}

	response.JSON(writer, http.StatusCreated, reservedGift)
}

func (controller ReservedGiftController) List(writer http.ResponseWriter, request *http.Request) {
	reservedGifts, err := controller.service.List(request.Context())
	if err != nil {
		response.JSON(writer, http.StatusInternalServerError, response.Error{
			Error: http.StatusText(http.StatusInternalServerError),
		})
		return
	}

	response.JSON(writer, http.StatusOK, reservedGifts)
}
