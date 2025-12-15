package handlers

import (
	"net/http"
	"time"
)

func ExampleHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	time.Sleep(5 * time.Second)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"message":"example"}`))
}
