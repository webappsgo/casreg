package model

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Token represents an API token for authentication
type Token struct {
	Base
	UserID      uint       `gorm:"not null;index" json:"user_id"`
	Name        string     `gorm:"not null" json:"name"`
	TokenHash   string     `gorm:"uniqueIndex;not null" json:"-"`
	Scopes      string     `gorm:"type:text" json:"scopes"` // comma-separated scopes
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	LastUsed    *time.Time `json:"last_used,omitempty"`
	UsageCount  int64      `gorm:"default:0" json:"usage_count"`
	IsActive    bool       `gorm:"not null;default:true" json:"is_active"`
	Description string     `gorm:"type:text" json:"description,omitempty"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"-"`
}

// GenerateToken generates a new cryptographically secure token
func GenerateToken() (string, error) {
	b := make([]byte, 48) // 48 bytes = 64 characters in base64
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// BeforeCreate generates a token hash before creating
func (t *Token) BeforeCreate(tx *gorm.DB) error {
	if t.Name == "" {
		return errors.New("token name is required")
	}

	// Validate scopes
	if t.Scopes == "" {
		t.Scopes = ScopeGlobal
	}

	scopes := t.GetScopesArray()
	validScopes := map[string]bool{
		ScopeGlobal:        true,
		ScopeRegistryRead:  true,
		ScopeRegistryWrite: true,
		ScopeRegistryAdmin: true,
		ScopeOrgRead:       true,
		ScopeOrgWrite:      true,
		ScopeOrgAdmin:      true,
		ScopeUserProfile:   true,
		ScopeAPIReadonly:   true,
	}

	for _, scope := range scopes {
		if !validScopes[scope] {
			return errors.New("invalid scope: " + scope)
		}
	}

	return nil
}

// SetScopes sets the token scopes from a slice
func (t *Token) SetScopes(scopes []string) {
	t.Scopes = strings.Join(scopes, ",")
}

// GetScopesArray returns the token scopes as a slice
func (t *Token) GetScopesArray() []string {
	if t.Scopes == "" {
		return []string{}
	}
	scopes := strings.Split(t.Scopes, ",")
	// Trim whitespace
	for i, s := range scopes {
		scopes[i] = strings.TrimSpace(s)
	}
	return scopes
}

// HasScope checks if the token has a specific scope
func (t *Token) HasScope(scope string) bool {
	scopes := t.GetScopesArray()
	for _, s := range scopes {
		if s == scope || s == ScopeGlobal {
			return true
		}
	}
	return false
}

// IsExpired checks if the token has expired
func (t *Token) IsExpired() bool {
	if t.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*t.ExpiresAt)
}

// IsValid checks if the token is valid for use
func (t *Token) IsValid() bool {
	return t.IsActive && !t.IsExpired()
}

// RecordUsage updates the last used timestamp and usage count
func (t *Token) RecordUsage(db *gorm.DB) error {
	now := time.Now()
	t.LastUsed = &now
	t.UsageCount++
	return db.Save(t).Error
}

// Revoke deactivates the token
func (t *Token) Revoke(db *gorm.DB) error {
	t.IsActive = false
	return db.Save(t).Error
}

// Rotate generates a new token value while keeping the same permissions
func (t *Token) Rotate(db *gorm.DB) (string, error) {
	// Generate new token
	newToken, err := GenerateToken()
	if err != nil {
		return "", err
	}

	// Update token hash
	t.TokenHash = HashToken(newToken)
	t.UsageCount = 0
	now := time.Now()
	t.LastUsed = &now

	if err := db.Save(t).Error; err != nil {
		return "", err
	}

	return newToken, nil
}

// HashToken creates a hash of the token for storage
func HashToken(token string) string {
	// In production, use a proper cryptographic hash like SHA256
	// For now, we'll use the token as-is for simplicity
	return token
}

// ValidateTokenScopes validates that the requested scopes are valid
func ValidateTokenScopes(scopes []string) error {
	validScopes := map[string]bool{
		ScopeGlobal:        true,
		ScopeRegistryRead:  true,
		ScopeRegistryWrite: true,
		ScopeRegistryAdmin: true,
		ScopeOrgRead:       true,
		ScopeOrgWrite:      true,
		ScopeOrgAdmin:      true,
		ScopeUserProfile:   true,
		ScopeAPIReadonly:   true,
	}

	for _, scope := range scopes {
		if !validScopes[scope] {
			return errors.New("invalid scope: " + scope)
		}
	}

	return nil
}
