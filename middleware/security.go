package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"
	"time"

	"github.com/casapps/casreg/config"
	"github.com/sirupsen/logrus"
)

// CSRF token storage
type csrfStore struct {
	mu     sync.RWMutex
	tokens map[string]time.Time
}

var globalCSRFStore = &csrfStore{
	tokens: make(map[string]time.Time),
}

func init() {
	// Start cleanup goroutine for expired CSRF tokens
	go cleanupCSRFTokens()
}

// Security middleware adds security headers and CSRF protection
func Security(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set security headers
			setSecurityHeaders(w, r)

			// CSRF protection for state-changing requests
			if isStateChangingRequest(r) {
				if !validateCSRFToken(r) {
					logrus.WithField("method", r.Method).Warn("CSRF validation failed")
					http.Error(w, `{"error":{"code":"CSRF_VALIDATION_FAILED","message":"CSRF token validation failed"}}`, http.StatusForbidden)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// setSecurityHeaders sets various security headers
func setSecurityHeaders(w http.ResponseWriter, r *http.Request) {
	// Prevent MIME type sniffing
	w.Header().Set("X-Content-Type-Options", "nosniff")

	// Enable XSS protection
	w.Header().Set("X-XSS-Protection", "1; mode=block")

	// Control framing
	w.Header().Set("X-Frame-Options", "SAMEORIGIN")

	// Referrer policy
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

	// Content Security Policy
	csp := []string{
		"default-src 'self'",
		"script-src 'self' 'unsafe-inline' 'unsafe-eval'", // For Svelte/Vue
		"style-src 'self' 'unsafe-inline'",
		"img-src 'self' data: https:",
		"font-src 'self' data:",
		"connect-src 'self'",
		"frame-ancestors 'self'",
		"base-uri 'self'",
		"form-action 'self'",
	}
	w.Header().Set("Content-Security-Policy", joinCSP(csp))

	// Strict Transport Security (HSTS) - only set if using HTTPS
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	}

	// Permissions Policy (formerly Feature Policy)
	permissions := []string{
		"geolocation=()",
		"microphone=()",
		"camera=()",
		"payment=()",
		"usb=()",
		"magnetometer=()",
	}
	w.Header().Set("Permissions-Policy", joinPermissions(permissions))
}

// isStateChangingRequest checks if the request is a state-changing operation
func isStateChangingRequest(r *http.Request) bool {
	// Only check CSRF for authenticated state-changing requests
	method := r.Method
	return method == http.MethodPost || method == http.MethodPut ||
		method == http.MethodPatch || method == http.MethodDelete
}

// validateCSRFToken validates the CSRF token from request
func validateCSRFToken(r *http.Request) bool {
	// Get token from header
	token := r.Header.Get("X-CSRF-Token")
	if token == "" {
		// Try to get from form value
		token = r.FormValue("csrf_token")
	}

	if token == "" {
		// For Docker Registry API, skip CSRF validation
		// Docker clients don't support CSRF tokens
		if r.Header.Get("User-Agent") != "" &&
		   (r.Header.Get("User-Agent")[:6] == "docker" || r.URL.Path[:4] == "/v2/") {
			return true
		}
		return false
	}

	// Validate token exists and is not expired
	globalCSRFStore.mu.RLock()
	expiresAt, exists := globalCSRFStore.tokens[token]
	globalCSRFStore.mu.RUnlock()

	if !exists {
		return false
	}

	if time.Now().After(expiresAt) {
		// Token expired, remove it
		globalCSRFStore.mu.Lock()
		delete(globalCSRFStore.tokens, token)
		globalCSRFStore.mu.Unlock()
		return false
	}

	return true
}

// GenerateCSRFToken generates a new CSRF token
func GenerateCSRFToken() (string, error) {
	// Generate random token
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	token := base64.URLEncoding.EncodeToString(b)

	// Store token with 1 hour expiration
	globalCSRFStore.mu.Lock()
	globalCSRFStore.tokens[token] = time.Now().Add(1 * time.Hour)
	globalCSRFStore.mu.Unlock()

	return token, nil
}

// cleanupCSRFTokens removes expired CSRF tokens
func cleanupCSRFTokens() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		globalCSRFStore.mu.Lock()

		for token, expiresAt := range globalCSRFStore.tokens {
			if now.After(expiresAt) {
				delete(globalCSRFStore.tokens, token)
			}
		}

		globalCSRFStore.mu.Unlock()
	}
}

// Helper functions

func joinCSP(policies []string) string {
	result := ""
	for i, policy := range policies {
		if i > 0 {
			result += "; "
		}
		result += policy
	}
	return result
}

func joinPermissions(permissions []string) string {
	result := ""
	for i, perm := range permissions {
		if i > 0 {
			result += ", "
		}
		result += perm
	}
	return result
}

// CSRFTokenHandler returns a handler that provides CSRF tokens
func CSRFTokenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := GenerateCSRFToken()
		if err != nil {
			logrus.WithError(err).Error("Failed to generate CSRF token")
			http.Error(w, `{"error":{"code":"INTERNAL_ERROR","message":"Failed to generate CSRF token"}}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"csrf_token":"` + token + `"}`))
	}
}
