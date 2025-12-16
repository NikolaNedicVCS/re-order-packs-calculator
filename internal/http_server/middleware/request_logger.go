package middleware

import (
	"net"
	"net/http"
	"time"

	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/log"
	"github.com/go-chi/chi/v5/middleware"
)

type responseRecorder struct {
	w      http.ResponseWriter
	status int
	bytes  int
}

func (r *responseRecorder) Header() http.Header { return r.w.Header() }

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.w.WriteHeader(status)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	n, err := r.w.Write(b)
	r.bytes += n
	return n, err
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rr := &responseRecorder{w: w}

		next.ServeHTTP(rr, r)

		remoteIP := r.RemoteAddr
		if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
			remoteIP = host
		}

		requestID := middleware.GetReqID(r.Context())
		log.Info("http_request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rr.status,
			"bytes", rr.bytes,
			"duration_ms", time.Since(start).Milliseconds(),
			"remote_ip", remoteIP,
			"request_id", requestID,
		)
	})
}
