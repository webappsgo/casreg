package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    int64
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.written += int64(n)
	return n, err
}

// Logging middleware logs all HTTP requests with structured logging
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		ww := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Get request ID from context (set by chi middleware)
		requestID := middleware.GetReqID(r.Context())

		// Process request
		next.ServeHTTP(ww, r)

		// Calculate request duration
		duration := time.Since(start)

		// Get user from context if authenticated
		var username string
		user := GetUserFromContext(r.Context())
		if user != nil {
			username = user.Username
		}

		// Get real IP from context (set by proxy middleware)
		realIP := r.RemoteAddr
		if ip := r.Context().Value("real_ip"); ip != nil {
			if ipStr, ok := ip.(string); ok {
				realIP = ipStr
			}
		}

		// Log with structured fields
		logrus.WithFields(logrus.Fields{
			"request_id":   requestID,
			"method":       r.Method,
			"uri":          r.RequestURI,
			"path":         r.URL.Path,
			"query":        r.URL.RawQuery,
			"status":       ww.statusCode,
			"duration_ms":  duration.Milliseconds(),
			"bytes":        ww.written,
			"remote_addr":  realIP,
			"user_agent":   r.UserAgent(),
			"referer":      r.Referer(),
			"username":     username,
			"proto":        r.Proto,
		}).Info("HTTP request")

		// Log warnings for slow requests (>1s)
		if duration > time.Second {
			logrus.WithFields(logrus.Fields{
				"request_id":  requestID,
				"method":      r.Method,
				"uri":         r.RequestURI,
				"duration_ms": duration.Milliseconds(),
			}).Warn("Slow request detected")
		}

		// Log errors for 5xx responses
		if ww.statusCode >= 500 {
			logrus.WithFields(logrus.Fields{
				"request_id": requestID,
				"method":     r.Method,
				"uri":        r.RequestURI,
				"status":     ww.statusCode,
			}).Error("Server error response")
		}
	})
}

// LogEntry logs a custom message with request context
func LogEntry(r *http.Request) *logrus.Entry {
	requestID := middleware.GetReqID(r.Context())

	fields := logrus.Fields{
		"request_id": requestID,
		"method":     r.Method,
		"uri":        r.RequestURI,
	}

	user := GetUserFromContext(r.Context())
	if user != nil {
		fields["username"] = user.Username
		fields["user_id"] = user.ID
	}

	return logrus.WithFields(fields)
}
