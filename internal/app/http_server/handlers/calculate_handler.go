package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/app/http_server/response"
	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/app/log"
	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/app/models"
	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/app/packcalc"
	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/app/repository"
	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/constants"
)

func CalculateHandler(w http.ResponseWriter, r *http.Request) {
	var req models.CalculateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Quantity <= 0 {
		response.WriteError(w, http.StatusBadRequest, "quantity must be > 0")
		return
	}
	if req.Quantity > 50_000_000 {
		response.WriteError(w, http.StatusBadRequest, "quantity too large")
		return
	}

	packs, err := repository.PackSizes().List(r.Context())
	if err != nil {
		log.Error("error listing pack sizes for calculate", "err", err)
		response.WriteError(w, http.StatusInternalServerError, constants.InternalServerErrorMsg)
		return
	}

	allocations, err := packcalc.Calculate(req.Quantity, packs)
	if err != nil {
		switch err {
		case packcalc.ErrInvalidQuantity:
			response.WriteError(w, http.StatusBadRequest, "quantity must be > 0")
			return
		case packcalc.ErrQuantityTooLarge:
			response.WriteError(w, http.StatusBadRequest, "quantity too large")
			return
		case packcalc.ErrNoPackSizes:
			response.WriteError(w, http.StatusBadRequest, "no pack sizes configured")
			return
		case packcalc.ErrInvalidPackSizes:
			response.WriteError(w, http.StatusBadRequest, "invalid pack sizes configured")
			return
		default:
			log.Error("error calculating pack allocation", "err", err)
			response.WriteError(w, http.StatusInternalServerError, constants.InternalServerErrorMsg)
			return
		}
	}

	response.WriteSuccess(w, http.StatusOK, models.CalculateResponse{Packs: allocations})
}
