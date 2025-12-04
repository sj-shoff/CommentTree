package router

import (
	"comments-system/internal/http-server/handler/comments"
	"comments-system/internal/http-server/middleware"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	CommentsHandler *comments.CommentsHandler
}

func SetupRouter(h *Handler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RecoveryMiddleware)
	r.Use(middleware.LoggingMiddleware)
	r.Use(middleware.JSONMiddleware)

	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "static"))
	FileServer(r, "/static", filesDir)

	r.Post("/comments", h.CommentsHandler.CreateComment)
	r.Get("/comments", h.CommentsHandler.GetComments)
	r.Delete("/comments/{id}", h.CommentsHandler.DeleteComment)

	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		indexPath := filepath.Join(workDir, "templates", "index.html")
		http.ServeFile(w, r, indexPath)
	})

	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		indexPath := filepath.Join(workDir, "templates", "index.html")
		http.ServeFile(w, r, indexPath)
	})

	return r
}

func FileServer(r chi.Router, path string, root http.FileSystem) {
	if string(path[len(path)-1]) != "/" {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}
