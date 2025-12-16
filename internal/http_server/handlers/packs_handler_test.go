package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/constants"
	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/http_server"
	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/models"
	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/repository"
)

type fakePackSizesRepo struct {
	listFn   func(ctx context.Context) ([]models.PackSize, error)
	createFn func(ctx context.Context, size int) (*models.PackSize, error)
	updateFn func(ctx context.Context, id int64, size int) (*models.PackSize, error)
	deleteFn func(ctx context.Context, id int64) error
	resetFn  func(ctx context.Context) ([]int, error)
}

func (f *fakePackSizesRepo) List(ctx context.Context) ([]models.PackSize, error) {
	return f.listFn(ctx)
}
func (f *fakePackSizesRepo) Create(ctx context.Context, size int) (*models.PackSize, error) {
	return f.createFn(ctx, size)
}
func (f *fakePackSizesRepo) Update(ctx context.Context, id int64, size int) (*models.PackSize, error) {
	return f.updateFn(ctx, id, size)
}
func (f *fakePackSizesRepo) Delete(ctx context.Context, id int64) error        { return f.deleteFn(ctx, id) }
func (f *fakePackSizesRepo) ResetToDefault(ctx context.Context) ([]int, error) { return f.resetFn(ctx) }

func doJSON(t *testing.T, h http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var b []byte
	if body != nil {
		var err error
		b, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

func mustDecodeJSON(t *testing.T, b []byte) any {
	t.Helper()

	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatalf("unmarshal json: %v; body=%s", err, string(b))
	}
	return v
}

func mustJSONEqual(t *testing.T, rr *httptest.ResponseRecorder, expectedJSON string) {
	t.Helper()

	actual := mustDecodeJSON(t, rr.Body.Bytes())
	expected := mustDecodeJSON(t, []byte(expectedJSON))

	actualB, err := json.Marshal(actual)
	if err != nil {
		t.Fatalf("marshal actual json: %v; body=%s", err, rr.Body.String())
	}
	expectedB, err := json.Marshal(expected)
	if err != nil {
		t.Fatalf("marshal expected json: %v; expected=%s", err, expectedJSON)
	}

	if !bytes.Equal(actualB, expectedB) {
		t.Fatalf("unexpected json.\nexpected=%s\nactual=%s", expectedJSON, rr.Body.String())
	}
}

func TestPackSizesCRUDAndReset(t *testing.T) {
	fake := &fakePackSizesRepo{
		listFn: func(ctx context.Context) ([]models.PackSize, error) {
			_ = ctx
			return []models.PackSize{{ID: 1, Size: 250}}, nil
		},
		createFn: func(ctx context.Context, size int) (*models.PackSize, error) {
			_ = ctx
			if size == 777 {
				return &models.PackSize{ID: 10, Size: 777}, nil
			}
			return nil, repository.ErrConflict
		},
		updateFn: func(ctx context.Context, id int64, size int) (*models.PackSize, error) {
			_ = ctx
			if id == 9999999 {
				return nil, repository.ErrNotFound
			}
			if id == 1 && size == 778 {
				return nil, repository.ErrConflict
			}
			return &models.PackSize{ID: id, Size: size}, nil
		},
		deleteFn: func(ctx context.Context, id int64) error {
			_ = ctx
			if id == 9999999 {
				return repository.ErrNotFound
			}
			return nil
		},
		resetFn: func(ctx context.Context) ([]int, error) {
			_ = ctx
			return []int{250, 500, 1000, 2000, 5000}, nil
		},
	}
	repository.SetPackSizesRepository(fake)

	h := http_server.NewHTTPHandler()

	t.Run("list ok", func(t *testing.T) {
		rr := doJSON(t, h, http.MethodGet, "/api/packs/", nil)
		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
		}
		mustJSONEqual(t, rr, `{"data":{"packs":[{"id":1,"size":250}]}}`)
	})

	t.Run("reset ok", func(t *testing.T) {
		rr := doJSON(t, h, http.MethodPost, "/api/packs/reset", nil)
		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
		}
		mustJSONEqual(t, rr, `{"data":{"sizes":[250,500,1000,2000,5000]}}`)
	})

	t.Run("create ok", func(t *testing.T) {
		rr := doJSON(t, h, http.MethodPost, "/api/packs/", models.CreatePackSizeRequest{Size: 777})
		if rr.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d body=%s", rr.Code, rr.Body.String())
		}
		mustJSONEqual(t, rr, `{"data":{"id":10,"size":777}}`)
	})

	t.Run("create conflict -> 409", func(t *testing.T) {
		rr := doJSON(t, h, http.MethodPost, "/api/packs/", models.CreatePackSizeRequest{Size: 123})
		if rr.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d body=%s", rr.Code, rr.Body.String())
		}
		mustJSONEqual(t, rr, `{"error":{"message":"pack size already exists"}}`)
	})

	t.Run("update ok", func(t *testing.T) {
		rr := doJSON(t, h, http.MethodPut, "/api/packs/10", models.UpdatePackSizeRequest{Size: 778})
		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
		}
		mustJSONEqual(t, rr, `{"data":{"id":10,"size":778}}`)
	})

	t.Run("update not found -> 404", func(t *testing.T) {
		rr := doJSON(t, h, http.MethodPut, "/api/packs/9999999", models.UpdatePackSizeRequest{Size: 12})
		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d body=%s", rr.Code, rr.Body.String())
		}
		mustJSONEqual(t, rr, `{"error":{"message":"not found"}}`)
	})

	t.Run("update conflict -> 409", func(t *testing.T) {
		rr := doJSON(t, h, http.MethodPut, "/api/packs/1", models.UpdatePackSizeRequest{Size: 778})
		if rr.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d body=%s", rr.Code, rr.Body.String())
		}
		mustJSONEqual(t, rr, `{"error":{"message":"pack size already exists"}}`)
	})

	t.Run("delete ok -> 200", func(t *testing.T) {
		rr := doJSON(t, h, http.MethodDelete, "/api/packs/10", nil)
		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
		}
		mustJSONEqual(t, rr, `{"data":{}}`)
	})

	t.Run("delete not found -> 404", func(t *testing.T) {
		rr := doJSON(t, h, http.MethodDelete, "/api/packs/9999999", nil)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d body=%s", rr.Code, rr.Body.String())
		}
		mustJSONEqual(t, rr, `{"error":{"message":"not found"}}`)
	})

	t.Run("create invalid body -> 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/packs/", bytes.NewBufferString("{"))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d body=%s", rr.Code, rr.Body.String())
		}
		mustJSONEqual(t, rr, `{"error":{"message":"invalid json"}}`)
	})

	t.Run("create invalid size -> 400", func(t *testing.T) {
		rr := doJSON(t, h, http.MethodPost, "/api/packs/", models.CreatePackSizeRequest{Size: 0})
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d body=%s", rr.Code, rr.Body.String())
		}
		mustJSONEqual(t, rr, `{"error":{"message":"size must be > 0"}}`)
	})

	t.Run("list internal error mapping", func(t *testing.T) {
		fake.listFn = func(ctx context.Context) ([]models.PackSize, error) {
			_ = ctx
			return nil, errors.New("db down")
		}
		rr := doJSON(t, h, http.MethodGet, "/api/packs/", nil)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d body=%s", rr.Code, rr.Body.String())
		}
		mustJSONEqual(t, rr, `{"error":{"message":"`+constants.InternalServerErrorMsg+`"}}`)
	})

	t.Run("update invalid id -> 400", func(t *testing.T) {
		rr := doJSON(t, h, http.MethodPut, "/api/packs/abc", models.UpdatePackSizeRequest{Size: 10})
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d body=%s", rr.Code, rr.Body.String())
		}
		mustJSONEqual(t, rr, `{"error":{"message":"invalid pack size ID"}}`)
	})

	t.Run("update invalid json -> 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/packs/10", bytes.NewBufferString("{"))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d body=%s", rr.Code, rr.Body.String())
		}
		mustJSONEqual(t, rr, `{"error":{"message":"invalid json"}}`)
	})

	t.Run("update invalid size -> 400", func(t *testing.T) {
		rr := doJSON(t, h, http.MethodPut, "/api/packs/10", models.UpdatePackSizeRequest{Size: 0})
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d body=%s", rr.Code, rr.Body.String())
		}
		mustJSONEqual(t, rr, `{"error":{"message":"size must be > 0"}}`)
	})

	t.Run("delete invalid id -> 400", func(t *testing.T) {
		rr := doJSON(t, h, http.MethodDelete, "/api/packs/abc", nil)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d body=%s", rr.Code, rr.Body.String())
		}
		mustJSONEqual(t, rr, `{"error":{"message":"invalid id"}}`)
	})

	t.Run("reset internal error -> 500", func(t *testing.T) {
		fake.resetFn = func(ctx context.Context) ([]int, error) {
			_ = ctx
			return nil, errors.New("db down")
		}
		rr := doJSON(t, h, http.MethodPost, "/api/packs/reset", nil)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d body=%s", rr.Code, rr.Body.String())
		}
		mustJSONEqual(t, rr, `{"error":{"message":"`+constants.InternalServerErrorMsg+`"}}`)
	})

	t.Run("create internal error -> 500", func(t *testing.T) {
		fake.createFn = func(ctx context.Context, size int) (*models.PackSize, error) {
			_ = ctx
			_ = size
			return nil, errors.New("db down")
		}
		rr := doJSON(t, h, http.MethodPost, "/api/packs/", models.CreatePackSizeRequest{Size: 123})
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d body=%s", rr.Code, rr.Body.String())
		}
		mustJSONEqual(t, rr, `{"error":{"message":"`+constants.InternalServerErrorMsg+`"}}`)
	})

	t.Run("update internal error -> 500", func(t *testing.T) {
		fake.updateFn = func(ctx context.Context, id int64, size int) (*models.PackSize, error) {
			_ = ctx
			_ = id
			_ = size
			return nil, errors.New("db down")
		}
		rr := doJSON(t, h, http.MethodPut, "/api/packs/10", models.UpdatePackSizeRequest{Size: 12})
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d body=%s", rr.Code, rr.Body.String())
		}
		mustJSONEqual(t, rr, `{"error":{"message":"`+constants.InternalServerErrorMsg+`"}}`)
	})

	t.Run("delete internal error -> 500", func(t *testing.T) {
		fake.deleteFn = func(ctx context.Context, id int64) error {
			_ = ctx
			_ = id
			return errors.New("db down")
		}
		rr := doJSON(t, h, http.MethodDelete, "/api/packs/10", nil)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d body=%s", rr.Code, rr.Body.String())
		}
		mustJSONEqual(t, rr, `{"error":{"message":"`+constants.InternalServerErrorMsg+`"}}`)
	})
}
