package handlers

import (
	"net/http"

	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/app/http_server/response"
	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/app/log"
	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/app/models"
	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/app/repository"
	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/constants"
)

func ResetPackSizesHandler(w http.ResponseWriter, r *http.Request) {
	sizes, err := repository.PackSizes().ResetToDefault(r.Context())
	if err != nil {
		log.Error("error resetting pack sizes", "err", err)
		response.WriteError(w, http.StatusInternalServerError, constants.InternalServerErrorMsg)
		return
	}

	response.WriteSuccess(w, http.StatusOK, models.ResetPackSizesResponse{Sizes: sizes})
}
