package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"context"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"registry-auth/models"
)

type AuthHandler struct {
	DB          *gorm.DB
	JWTSecret   []byte
	ExternalURL string
	RegistryURL string
}

func NewAuthHandler(db *gorm.DB, jwtSecret, externalURL, registryURL string) *AuthHandler {
	return &AuthHandler{
		DB:          db,
		JWTSecret:   []byte(jwtSecret),
		ExternalURL: externalURL,
		RegistryURL: registryURL,
	}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type AuthResponse struct {
	Token     string       `json:"token"`
	User      *models.User `json:"user"`
	ExpiresIn int          `json:"expires_in"`
}

// Login handles user authentication
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get user from database
	var user models.User
	if err := h.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Check if user is active
	if !user.IsActive {
		http.Error(w, "Account is disabled", http.StatusForbidden)
		return
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	})

	tokenString, err := token.SignedString(h.JWTSecret)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Update last login
	h.DB.Model(&user).Update("updated_at", time.Now())

	// Log the action
	h.logAudit(user.ID, "login", "user", user.ID, r.RemoteAddr, r.UserAgent())

	// Clear password from response
	user.Password = ""

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{
		Token:     tokenString,
		User:      &user,
		ExpiresIn: 86400, // 24 hours
	})
}

// Register creates a new user account
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Username == "" || req.Email == "" || req.Password == "" {
		http.Error(w, "Username, email and password are required", http.StatusBadRequest)
		return
	}

	// Check if username exists
	var existingUser models.User
	if err := h.DB.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		http.Error(w, "Username already exists", http.StatusConflict)
		return
	}

	// Check if email exists
	if err := h.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		http.Error(w, "Email already exists", http.StatusConflict)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to process password", http.StatusInternalServerError)
		return
	}

	// Determine role (first user is admin)
	var userCount int64
	h.DB.Model(&models.User{}).Count(&userCount)
	role := "user"
	if userCount == 0 {
		role = "admin"
	}

	// Create user
	user := &models.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  string(hashedPassword),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      role,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.DB.Create(user).Error; err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Log the action
	h.logAudit(user.ID, "register", "user", user.ID, r.RemoteAddr, r.UserAgent())

	// Generate JWT token for auto-login
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	})

	tokenString, _ := token.SignedString(h.JWTSecret)

	// Clear password from response
	user.Password = ""

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(AuthResponse{
		Token:     tokenString,
		User:      user,
		ExpiresIn: 86400,
	})
}

// ValidateToken validates a JWT token for Docker Registry
func (h *AuthHandler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	// Get token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "No authorization header", http.StatusUnauthorized)
		return
	}

	// Parse Bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
		return
	}

	// Validate JWT token
	token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return h.JWTSecret, nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Get claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}

	// Check token expiration
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			http.Error(w, "Token expired", http.StatusUnauthorized)
			return
		}
	}

	// Get user info
	userID := uint(claims["user_id"].(float64))
	username := claims["username"].(string)
	role := claims["role"].(string)

	// Set headers for registry
	w.Header().Set("X-Auth-User-Id", fmt.Sprintf("%d", userID))
	w.Header().Set("X-Auth-Username", username)
	w.Header().Set("X-Auth-Role", role)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// DockerAuth handles Docker client authentication requests
func (h *AuthHandler) DockerAuth(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	service := r.URL.Query().Get("service")
	scope := r.URL.Query().Get("scope")

	// Get credentials from Authorization header (Basic auth for docker login)
	authHeader := r.Header.Get("Authorization")
	var username, password string

	if authHeader != "" && strings.HasPrefix(authHeader, "Basic ") {
		payload, _ := base64.StdEncoding.DecodeString(authHeader[6:])
		pair := strings.SplitN(string(payload), ":", 2)
		if len(pair) == 2 {
			username = pair[0]
			password = pair[1]
		}
	}

	// For anonymous pulls of public repos
	if username == "" && strings.Contains(scope, ":pull") {
		// Check if the repository is public
		repoName := extractRepoName(scope)
		var repo models.Repository
		if err := h.DB.Where("name = ?", repoName).First(&repo).Error; err == nil && repo.IsPublic {
			// Generate anonymous token for public repo
			token := h.generateDockerToken("", service, scope, 300) // 5 minutes
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"token":      token,
				"expires_in": 300,
				"issued_at":  time.Now().Format(time.RFC3339),
			})
			return
		}
	}

	// Authenticate user
	if username != "" && password != "" {
		var user models.User
		if err := h.DB.Where("username = ?", username).First(&user).Error; err == nil {
			if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) == nil && user.IsActive {
				// Generate Docker token
				token := h.generateDockerToken(user.Username, service, scope, 3600) // 1 hour

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"token":      token,
					"expires_in": 3600,
					"issued_at":  time.Now().Format(time.RFC3339),
				})
				return
			}
		}
	}

	// Authentication failed
	w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Bearer realm="%s/auth/docker",service="%s"`, h.ExternalURL, service))
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

// generateDockerToken creates a JWT token for Docker Registry
func (h *AuthHandler) generateDockerToken(username, service, scope string, expiresIn int) string {
	claims := jwt.MapClaims{
		"iss": h.ExternalURL,
		"sub": username,
		"aud": service,
		"exp": time.Now().Add(time.Duration(expiresIn) * time.Second).Unix(),
		"nbf": time.Now().Unix(),
		"iat": time.Now().Unix(),
		"jti": generateRandomID(),
	}

	if scope != "" {
		claims["access"] = []map[string]interface{}{
			{
				"type":    extractResourceType(scope),
				"name":    extractResourceName(scope),
				"actions": extractActions(scope),
			},
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(h.JWTSecret)
	return tokenString
}

// AuthMiddleware validates JWT tokens and adds user info to context
func AuthMiddleware(db *gorm.DB, jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "No authorization header", http.StatusUnauthorized)
				return
			}

			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				authHeader = authHeader[7:]
			}

			token, err := jwt.Parse(authHeader, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method")
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}

			// Add user info to context
			ctx := context.WithValue(r.Context(), "userID", uint(claims["user_id"].(float64)))
			ctx = context.WithValue(ctx, "username", claims["username"].(string))
			ctx = context.WithValue(ctx, "role", claims["role"].(string))

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Helper functions
func extractRepoName(scope string) string {
	// scope format: repository:namespace/name:action
	parts := strings.Split(scope, ":")
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

func extractResourceType(scope string) string {
	parts := strings.Split(scope, ":")
	if len(parts) > 0 {
		return parts[0]
	}
	return "repository"
}

func extractResourceName(scope string) string {
	parts := strings.Split(scope, ":")
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

func extractActions(scope string) []string {
	parts := strings.Split(scope, ":")
	if len(parts) >= 3 {
		return strings.Split(parts[2], ",")
	}
	return []string{}
}

func generateRandomID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func (h *AuthHandler) logAudit(userID uint, action, resource string, resourceID uint, ip, userAgent string) {
	auditLog := &models.AuditLog{
		UserID:     userID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		IP:         ip,
		UserAgent:  userAgent,
		CreatedAt:  time.Now(),
	}
	h.DB.Create(auditLog)
}