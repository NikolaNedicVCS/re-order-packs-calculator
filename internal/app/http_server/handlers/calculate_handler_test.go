package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/app/http_server"
	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/app/models"
	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/app/packcalc"
	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/app/repository"
	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/constants"
)

func TestCalculateHandler(t *testing.T) {
	fake := &fakePackSizesRepo{
		listFn: func(ctx context.Context) ([]models.PackSize, error) {
			_ = ctx
			return []models.PackSize{{ID: 1, Size: 250}, {ID: 2, Size: 500}}, nil
		},
		createFn: func(ctx context.Context, size int) (*models.PackSize, error) { _ = ctx; _ = size; return nil, nil },
		updateFn: func(ctx context.Context, id int64, size int) (*models.PackSize, error) {
			_ = ctx
			_ = id
			_ = size
			return nil, nil
		},
		deleteFn: func(ctx context.Context, id int64) error { _ = ctx; _ = id; return nil },
		resetFn:  func(ctx context.Context) ([]int, error) { _ = ctx; return nil, nil },
	}
	repository.SetPackSizesRepository(fake)

	h := http_server.NewHTTPHandler()

	rr := doJSON(t, h, http.MethodPost, "/api/calculate", models.CalculateRequest{Quantity: 1})
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	mustJSONEqual(t, rr, `{"data":{"packs":[{"size":250,"count":1}]}}`)
}

func TestCalculateHandler_InvalidQuantity(t *testing.T) {
	h := http_server.NewHTTPHandler()
	rr := doJSON(t, h, http.MethodPost, "/api/calculate", models.CalculateRequest{Quantity: 0})
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rr.Code, rr.Body.String())
	}
	mustJSONEqual(t, rr, `{"error":{"message":"quantity must be > 0"}}`)
}

type fakeCalculator struct {
	calcFn func(quantity int, packSizes []models.PackSize) ([]models.PackAllocation, error)
}

func (f fakeCalculator) Calculate(quantity int, packSizes []models.PackSize) ([]models.PackAllocation, error) {
	return f.calcFn(quantity, packSizes)
}

func TestCalculateHandler_ErrorMappingFromPackcalc(t *testing.T) {
	origRepo := repository.PackSizes()
	t.Cleanup(func() {
		repository.SetPackSizesRepository(origRepo)
	})
	origCalc := packcalc.CalculatorImpl()
	t.Cleanup(func() {
		packcalc.SetCalculator(origCalc)
	})

	fakeRepo := &fakePackSizesRepo{
		listFn: func(ctx context.Context) ([]models.PackSize, error) {
			_ = ctx
			return []models.PackSize{{ID: 1, Size: 250}}, nil
		},
		createFn: func(ctx context.Context, size int) (*models.PackSize, error) { _ = ctx; _ = size; return nil, nil },
		updateFn: func(ctx context.Context, id int64, size int) (*models.PackSize, error) {
			_ = ctx
			_ = id
			_ = size
			return nil, nil
		},
		deleteFn: func(ctx context.Context, id int64) error { _ = ctx; _ = id; return nil },
		resetFn:  func(ctx context.Context) ([]int, error) { _ = ctx; return nil, nil },
	}
	repository.SetPackSizesRepository(fakeRepo)

	cases := []struct {
		name        string
		calcErr     error
		wantStatus  int
		wantMessage string
	}{
		{"ErrInvalidQuantity -> 400", packcalc.ErrInvalidQuantity, http.StatusBadRequest, "quantity must be > 0"},
		{"ErrNoPackSizes -> 400", packcalc.ErrNoPackSizes, http.StatusBadRequest, "no pack sizes configured"},
		{"ErrInvalidPackSizes -> 400", packcalc.ErrInvalidPackSizes, http.StatusBadRequest, "invalid pack sizes configured"},
		{"default -> 500", errors.New("boom"), http.StatusInternalServerError, constants.InternalServerErrorMsg},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			packcalc.SetCalculator(fakeCalculator{
				calcFn: func(quantity int, packSizes []models.PackSize) ([]models.PackAllocation, error) {
					_ = quantity
					_ = packSizes
					return nil, tc.calcErr
				},
			})

			h := http_server.NewHTTPHandler()
			rr := doJSON(t, h, http.MethodPost, "/api/calculate", models.CalculateRequest{Quantity: 1})
			if rr.Code != tc.wantStatus {
				t.Fatalf("expected %d, got %d body=%s", tc.wantStatus, rr.Code, rr.Body.String())
			}

			if tc.wantStatus >= 400 {
				mustJSONEqual(t, rr, `{"error":{"message":"`+tc.wantMessage+`"}}`)
			}
		})
	}
}
