package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/constants"
	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/http_server/response"
	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/log"
	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/models"
	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/repository"
	"github.com/go-chi/chi/v5"
)

func ListPackSizesHandler(w http.ResponseWriter, r *http.Request) {
	packs, err := repository.PackSizes().List(r.Context())
	if err != nil {
		log.Error("error listing pack sizes", "err", err)
		response.WriteError(w, http.StatusInternalServerError, constants.InternalServerErrorMsg)
		return
	}

	response.WriteSuccess(w, http.StatusOK, models.ListPackSizesResponse{Packs: packs})
}

func CreatePackSizeHandler(w http.ResponseWriter, r *http.Request) {
	var req models.CreatePackSizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Size <= 0 {
		response.WriteError(w, http.StatusBadRequest, "size must be > 0")
		return
	}

	created, err := repository.PackSizes().Create(r.Context(), req.Size)
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			response.WriteError(w, http.StatusConflict, "pack size already exists")
			return
		}
		log.Error("error creating pack size", "err", err)
		response.WriteError(w, http.StatusInternalServerError, constants.InternalServerErrorMsg)
		return
	}
	response.WriteSuccess(w, http.StatusCreated, created)
}

func UpdatePackSizeHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		response.WriteError(w, http.StatusBadRequest, "invalid pack size ID")
		return
	}

	var req models.UpdatePackSizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Size <= 0 {
		response.WriteError(w, http.StatusBadRequest, "size must be > 0")
		return
	}

	updated, err := repository.PackSizes().Update(r.Context(), id, req.Size)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.WriteError(w, http.StatusNotFound, "not found")
			return
		}
		if errors.Is(err, repository.ErrConflict) {
			response.WriteError(w, http.StatusConflict, "pack size already exists")
			return
		}
		log.Error("error updating pack size", "err", err)
		response.WriteError(w, http.StatusInternalServerError, constants.InternalServerErrorMsg)
		return
	}
	response.WriteSuccess(w, http.StatusOK, updated)
}

func DeletePackSizeHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		response.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := repository.PackSizes().Delete(r.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.WriteError(w, http.StatusNotFound, "not found")
			return
		}
		log.Error("error deleting pack size", "err", err)
		response.WriteError(w, http.StatusInternalServerError, constants.InternalServerErrorMsg)
		return
	}

	// 204 must not include a response body; use 200 with empty data for consistency.
	response.WriteSuccess(w, http.StatusOK, struct{}{})
}

func ResetPackSizesHandler(w http.ResponseWriter, r *http.Request) {
	sizes, err := repository.PackSizes().ResetToDefault(r.Context())
	if err != nil {
		log.Error("error resetting pack sizes", "err", err)
		response.WriteError(w, http.StatusInternalServerError, constants.InternalServerErrorMsg)
		return
	}

	response.WriteSuccess(w, http.StatusOK, models.ResetPackSizesResponse{Sizes: sizes})
}
