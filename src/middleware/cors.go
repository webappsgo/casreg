package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/casapps/casreg/src/config"
)

// CORS middleware handles Cross-Origin Resource Sharing
func CORS(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Set CORS headers
			if origin != "" {
				// Allow all origins for now - can be made configurable via config
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			} else {
				// No origin header, allow wildcard
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}

			// Set allowed methods
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")

			// Set allowed headers
			allowedHeaders := []string{
				"Accept",
				"Authorization",
				"Content-Type",
				"X-CSRF-Token",
				"X-Requested-With",
				"X-Request-ID",
			}
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ", "))

			// Set exposed headers
			exposedHeaders := []string{
				"Link",
				"X-Total-Count",
				"X-Page-Count",
				"X-Current-Page",
				"X-RateLimit-Limit",
				"X-RateLimit-Remaining",
				"X-RateLimit-Reset",
				"Docker-Content-Digest",
				"Docker-Upload-UUID",
				"Location",
				"Range",
			}
			w.Header().Set("Access-Control-Expose-Headers", strings.Join(exposedHeaders, ", "))

			// Set max age for preflight cache
			w.Header().Set("Access-Control-Max-Age", strconv.Itoa(300)) // 5 minutes

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
