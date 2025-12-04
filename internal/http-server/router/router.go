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

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(r.URL.Path, "/static/") {
				middleware.LoggingMiddleware(next).ServeHTTP(w, r)
			} else {
				next.ServeHTTP(w, r)
			}
		})
	})

	workDir, _ := os.Getwd()

	fs := http.FileServer(http.Dir(filepath.Join(workDir, "static")))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	r.Route("/api", func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				next.ServeHTTP(w, r)
			})
		})

		r.Route("/comments", func(r chi.Router) {
			r.Post("/", h.CommentsHandler.CreateComment)
			r.Get("/", h.CommentsHandler.GetComments)
			r.Delete("/{id}", h.CommentsHandler.DeleteComment)
		})

		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"status":"ok"}`))
		})
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		serveHTML(w, r, workDir)
	})

	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/static/") && !strings.HasPrefix(r.URL.Path, "/api/") {
			serveHTML(w, r, workDir)
		} else {
			http.NotFound(w, r)
		}
	})

	return r
}

func serveHTML(w http.ResponseWriter, r *http.Request, workDir string) {
	indexPath := filepath.Join(workDir, "templates", "index.html")

	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		http.Error(w, "HTML template not found", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	http.ServeFile(w, r, indexPath)
}
