package response

import (
	"encoding/json"
	"net/http"

	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/app/log"
)

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Message string `json:"message"`
}

type SuccessResponse struct {
	Data any `json:"data"`
}

func WriteError(w http.ResponseWriter, statusCode int, message string) {
	WriteJSON(w, statusCode, ErrorResponse{
		Error: ErrorBody{Message: message},
	})
}

func WriteSuccess(w http.ResponseWriter, statusCode int, data any) {
	WriteJSON(w, statusCode, SuccessResponse{Data: data})
}

func WriteJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Error("failed to write json response", "err", err)
	}
}
