package middleware

import (
	"net/http"
	"time"

	"github.com/wb-go/wbf/zlog"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)

		zlog.Logger.Info().
			Str("method", r.Method).
			Str("url", r.URL.String()).
			Dur("duration", duration).
			Msg("Request completed")
	})
}
