package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/casapps/casreg/config"
	"github.com/casapps/casreg/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// UserContextKey is the context key for the authenticated user
type contextKey string

const UserContextKey contextKey = "user"

// Claims represents JWT claims for authentication
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	TokenID  uint   `json:"token_id,omitempty"` // For API tokens
	Type     string `json:"type"`               // "session", "api", "docker"
	jwt.RegisteredClaims
}

// Authenticate middleware validates JWT tokens and sets user context
func Authenticate(cfg *config.Config, db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := authenticateRequest(r, cfg, db)
			if err != nil {
				logrus.WithError(err).Warn("Authentication failed")
				http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"Authentication required"}}`, http.StatusUnauthorized)
				return
			}

			// Set user in context
			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuth middleware attempts authentication but allows unauthenticated requests
func OptionalAuth(cfg *config.Config, db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := authenticateRequest(r, cfg, db)
			if err == nil && user != nil {
				// Set user in context if authentication succeeded
				ctx := context.WithValue(r.Context(), UserContextKey, user)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Continue without authentication
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAdmin middleware ensures the user has admin role
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"Authentication required"}}`, http.StatusUnauthorized)
			return
		}

		if !user.IsAdmin() {
			http.Error(w, `{"error":{"code":"FORBIDDEN","message":"Admin access required"}}`, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// authenticateRequest extracts and validates authentication credentials from request
func authenticateRequest(r *http.Request, cfg *config.Config, db *gorm.DB) (*models.User, error) {
	// Try Bearer token authentication first
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return validateBearerToken(parts[1], cfg, db)
		}
	}

	// Try Docker Registry basic auth (used by docker client)
	username, password, ok := r.BasicAuth()
	if ok {
		return validateBasicAuth(username, password, cfg, db)
	}

	return nil, fmt.Errorf("no valid authentication credentials provided")
}

// validateBearerToken validates a JWT bearer token
func validateBearerToken(tokenString string, cfg *config.Config, db *gorm.DB) (*models.User, error) {
	// Parse and validate JWT token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(cfg.Security.JWTSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Check token expiration
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("token expired")
	}

	// Load user from database
	var user models.User
	if err := db.First(&user, claims.UserID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Check if user is active
	if !user.IsActive {
		return nil, fmt.Errorf("user account is inactive")
	}

	// Check if user is locked
	if user.IsLocked() {
		return nil, fmt.Errorf("user account is locked")
	}

	// If this is an API token, verify it still exists and is valid
	if claims.TokenID > 0 {
		var apiToken models.Token
		if err := db.First(&apiToken, claims.TokenID).Error; err != nil {
			return nil, fmt.Errorf("api token not found or revoked")
		}

		// Update last used timestamp
		now := time.Now()
		apiToken.LastUsed = &now
		db.Save(&apiToken)
	}

	return &user, nil
}

// validateBasicAuth validates username/password authentication (for Docker registry)
func validateBasicAuth(username, password string, cfg *config.Config, db *gorm.DB) (*models.User, error) {
	// First try API token authentication (username can be token name, password is token value)
	user, err := validateAPIToken(password, cfg, db)
	if err == nil {
		return user, nil
	}

	// Fall back to username/password authentication
	var user2 models.User
	if err := db.Where("username = ? OR email = ?", username, username).First(&user2).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invalid credentials")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Check if user is active
	if !user2.IsActive {
		return nil, fmt.Errorf("user account is inactive")
	}

	// Check if user is locked
	if user2.IsLocked() {
		return nil, fmt.Errorf("user account is locked")
	}

	// Verify password
	if err := user2.CheckPassword(password); err != nil {
		// Increment failed login counter
		user2.IncrementFailedLogins(db)
		return nil, fmt.Errorf("invalid credentials")
	}

	// Reset failed login counter on successful login
	user2.ResetFailedLogins(db)

	return &user2, nil
}

// validateAPIToken validates an API token
func validateAPIToken(tokenString string, cfg *config.Config, db *gorm.DB) (*models.User, error) {
	var apiToken models.Token
	if err := db.Where("token = ?", tokenString).First(&apiToken).Error; err != nil {
		return nil, fmt.Errorf("invalid token")
	}

	// Check if token is expired
	if apiToken.ExpiresAt != nil && apiToken.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("token expired")
	}

	// Load user
	var user models.User
	if err := db.First(&user, apiToken.UserID).Error; err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, fmt.Errorf("user account is inactive")
	}

	// Update last used timestamp
	now := time.Now()
	apiToken.LastUsed = &now
	db.Save(&apiToken)

	return &user, nil
}

// GetUserFromContext retrieves the authenticated user from request context
func GetUserFromContext(ctx context.Context) *models.User {
	user, ok := ctx.Value(UserContextKey).(*models.User)
	if !ok {
		return nil
	}
	return user
}

// GenerateJWT generates a JWT token for a user
func GenerateJWT(user *models.User, cfg *config.Config, tokenType string, tokenID uint) (string, error) {
	now := time.Now()
	expiresAt := now.Add(cfg.Security.JWTExpiration)

	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		TokenID:  tokenID,
		Type:     tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "casreg",
			Subject:   user.Username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Security.JWTSecret))
}
