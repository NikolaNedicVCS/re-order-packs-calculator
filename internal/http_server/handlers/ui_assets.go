package handlers

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed ui
var uiFS embed.FS

var uiAssetsFS = func() fs.FS {
	sub, err := fs.Sub(uiFS, "ui")
	if err != nil {
		panic(err)
	}
	return sub
}()

func UIIndexHandler(w http.ResponseWriter, r *http.Request) {
	b, err := fs.ReadFile(uiAssetsFS, "index.html")
	if err != nil {
		http.Error(w, "ui not available", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(b)
}

func UIAssetsHandler(w http.ResponseWriter, r *http.Request) {
	// chi pattern is "/assets/*" so we strip "/assets/" prefix.
	p := strings.TrimPrefix(r.URL.Path, "/assets/")
	if p == "" || strings.Contains(p, "..") {
		http.NotFound(w, r)
		return
	}
	r2 := r.Clone(r.Context())
	r2.URL.Path = "/" + p
	http.FileServer(http.FS(uiAssetsFS)).ServeHTTP(w, r2)
}
