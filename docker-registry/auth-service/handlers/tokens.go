package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"crypto/rand"
	"encoding/base64"
	"context"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"registry-auth/models"
)

// CreateTokenRequest represents the request body for creating a token
type CreateTokenRequest struct {
	Name        string    `json:"name" validate:"required,min=1,max=100"`
	Scopes      []string  `json:"scopes" validate:"required"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	Description string    `json:"description,omitempty"`
}

// CreateTokenResponse represents the response for creating a token
type CreateTokenResponse struct {
	Token     string                `json:"token"`
	TokenInfo *models.PersonalToken `json:"token_info"`
	Message   string                `json:"message"`
}

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// validateTokenScopes validates that the provided scopes are valid
func validateTokenScopes(scopes []string) bool {
	validScopes := map[string]bool{
		"global":         true,
		"registry:read":  true,
		"registry:write": true,
		"registry:admin": true,
		"org:read":       true,
		"org:write":      true,
		"org:admin":      true,
		"user:profile":   true,
		"api:readonly":   true,
	}

	for _, scope := range scopes {
		if !validScopes[scope] {
			return false
		}
	}
	return len(scopes) > 0
}

// CreateToken creates a new API token
func CreateToken(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user from context (set by auth middleware)
		userID, ok := r.Context().Value("userID").(uint)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req CreateTokenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Validate scopes
		if !validateTokenScopes(req.Scopes) {
			http.Error(w, "Invalid scopes provided", http.StatusBadRequest)
			return
		}

		// Generate secure token
		tokenValue, err := generateSecureToken(32)
		if err != nil {
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}

		// Create token record
		token := &models.PersonalToken{
			UserID:      userID,
			Name:        req.Name,
			Token:       tokenValue,
			Scopes:      req.Scopes,
			ExpiresAt:   req.ExpiresAt,
			Description: req.Description,
			LastUsedAt:  nil,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := db.Create(token).Error; err != nil {
			http.Error(w, "Failed to create token", http.StatusInternalServerError)
			return
		}

		response := CreateTokenResponse{
			Token:     tokenValue,
			TokenInfo: token,
			Message:   "Token created successfully. This is the only time you'll see the full token value.",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}

// ListTokens returns all tokens for the authenticated user (excluding token values)
func ListTokens(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user from context
		userID, ok := r.Context().Value("userID").(uint)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var tokens []models.PersonalToken
		if err := db.Where("user_id = ?", userID).Find(&tokens).Error; err != nil {
			http.Error(w, "Failed to fetch tokens", http.StatusInternalServerError)
			return
		}

		// Remove token values for security
		for i := range tokens {
			tokens[i].Token = "***hidden***"
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"tokens": tokens,
			"total":  len(tokens),
		})
	}
}

// GetToken returns a specific token (excluding token value)
func GetToken(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user from context
		userID, ok := r.Context().Value("userID").(uint)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tokenIDStr := chi.URLParam(r, "tokenID")
		tokenID, err := strconv.ParseUint(tokenIDStr, 10, 32)
		if err != nil {
			http.Error(w, "Invalid token ID", http.StatusBadRequest)
			return
		}

		var token models.PersonalToken
		if err := db.Where("id = ? AND user_id = ?", uint(tokenID), userID).First(&token).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				http.Error(w, "Token not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Failed to fetch token", http.StatusInternalServerError)
			return
		}

		// Hide token value for security
		token.Token = "***hidden***"

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(token)
	}
}

// UpdateToken updates a token's metadata
func UpdateToken(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user from context
		userID, ok := r.Context().Value("userID").(uint)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tokenIDStr := chi.URLParam(r, "tokenID")
		tokenID, err := strconv.ParseUint(tokenIDStr, 10, 32)
		if err != nil {
			http.Error(w, "Invalid token ID", http.StatusBadRequest)
			return
		}

		var updateReq struct {
			Name        string    `json:"name,omitempty"`
			Description string    `json:"description,omitempty"`
			ExpiresAt   *time.Time `json:"expires_at,omitempty"`
		}

		if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		var token models.PersonalToken
		if err := db.Where("id = ? AND user_id = ?", uint(tokenID), userID).First(&token).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				http.Error(w, "Token not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Failed to fetch token", http.StatusInternalServerError)
			return
		}

		// Update fields if provided
		if updateReq.Name != "" {
			token.Name = updateReq.Name
		}
		if updateReq.Description != "" {
			token.Description = updateReq.Description
		}
		if updateReq.ExpiresAt != nil {
			token.ExpiresAt = updateReq.ExpiresAt
		}

		token.UpdatedAt = time.Now()

		if err := db.Save(&token).Error; err != nil {
			http.Error(w, "Failed to update token", http.StatusInternalServerError)
			return
		}

		// Hide token value for security
		token.Token = "***hidden***"

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(token)
	}
}

// RotateToken generates a new token value while keeping the same metadata
func RotateToken(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user from context
		userID, ok := r.Context().Value("userID").(uint)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tokenIDStr := chi.URLParam(r, "tokenID")
		tokenID, err := strconv.ParseUint(tokenIDStr, 10, 32)
		if err != nil {
			http.Error(w, "Invalid token ID", http.StatusBadRequest)
			return
		}

		var token models.PersonalToken
		if err := db.Where("id = ? AND user_id = ?", uint(tokenID), userID).First(&token).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				http.Error(w, "Token not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Failed to fetch token", http.StatusInternalServerError)
			return
		}

		// Generate new secure token
		newTokenValue, err := generateSecureToken(32)
		if err != nil {
			http.Error(w, "Failed to generate new token", http.StatusInternalServerError)
			return
		}

		// Update token with new value
		token.Token = newTokenValue
		token.UpdatedAt = time.Now()
		token.LastUsedAt = nil // Reset last used since this is a new token

		if err := db.Save(&token).Error; err != nil {
			http.Error(w, "Failed to rotate token", http.StatusInternalServerError)
			return
		}

		response := CreateTokenResponse{
			Token:     newTokenValue,
			TokenInfo: &token,
			Message:   "Token rotated successfully. This is the only time you'll see the new token value.",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// DeleteToken deletes a token
func DeleteToken(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user from context
		userID, ok := r.Context().Value("userID").(uint)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tokenIDStr := chi.URLParam(r, "tokenID")
		tokenID, err := strconv.ParseUint(tokenIDStr, 10, 32)
		if err != nil {
			http.Error(w, "Invalid token ID", http.StatusBadRequest)
			return
		}

		// Check if token exists and belongs to user
		var token models.PersonalToken
		if err := db.Where("id = ? AND user_id = ?", uint(tokenID), userID).First(&token).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				http.Error(w, "Token not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Failed to fetch token", http.StatusInternalServerError)
			return
		}

		// Delete the token
		if err := db.Delete(&token).Error; err != nil {
			http.Error(w, "Failed to delete token", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Token deleted successfully",
		})
	}
}

// ValidateTokenMiddleware validates API tokens for authentication
func ValidateTokenMiddleware(db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			// Extract token from "Bearer <token>" format
			if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			tokenValue := authHeader[7:]

			// Find token in database
			var token models.PersonalToken
			if err := db.Where("token = ?", tokenValue).First(&token).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					http.Error(w, "Invalid token", http.StatusUnauthorized)
					return
				}
				http.Error(w, "Database error", http.StatusInternalServerError)
				return
			}

			// Check if token is expired
			if token.ExpiresAt != nil && token.ExpiresAt.Before(time.Now()) {
				http.Error(w, "Token expired", http.StatusUnauthorized)
				return
			}

			// Update last used timestamp
			now := time.Now()
			token.LastUsedAt = &now
			db.Save(&token)

			// Add user and token info to context
			ctx := context.WithValue(r.Context(), "userID", token.UserID)
			ctx = context.WithValue(ctx, "tokenID", token.ID)
			ctx = context.WithValue(ctx, "tokenScopes", token.Scopes)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}