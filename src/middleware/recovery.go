package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
)

// Recovery middleware recovers from panics and logs the error
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Get request ID for correlation
				requestID := middleware.GetReqID(r.Context())

				// Get stack trace
				stack := debug.Stack()

				// Log the panic with full context
				logrus.WithFields(logrus.Fields{
					"request_id":   requestID,
					"method":       r.Method,
					"uri":          r.RequestURI,
					"remote_addr":  r.RemoteAddr,
					"panic":        err,
					"stack_trace":  string(stack),
				}).Error("Panic recovered")

				// Get user context if available
				user := GetUserFromContext(r.Context())
				if user != nil {
					logrus.WithFields(logrus.Fields{
						"request_id": requestID,
						"user_id":    user.ID,
						"username":   user.Username,
					}).Error("Panic occurred for authenticated user")
				}

				// Send error response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)

				errorResponse := fmt.Sprintf(
					`{"error":{"code":"INTERNAL_SERVER_ERROR","message":"An internal server error occurred","request_id":"%s"}}`,
					requestID,
				)
				w.Write([]byte(errorResponse))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
