package models

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User represents a system user
type User struct {
	Base
	Username      string    `gorm:"uniqueIndex;not null" json:"username"`
	Email         string    `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash  string    `gorm:"not null" json:"-"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	Role          string    `gorm:"not null;default:'user'" json:"role"`
	IsActive      bool      `gorm:"not null;default:true" json:"is_active"`
	Theme         string    `gorm:"default:'dracula'" json:"theme"`
	QuotaLimit    int64     `gorm:"default:0" json:"quota_limit"` // 0 = unlimited
	QuotaUsed     int64     `gorm:"default:0" json:"quota_used"`
	LastLogin     *time.Time `json:"last_login,omitempty"`
	FailedLogins  int       `gorm:"default:0" json:"-"`
	LockedUntil   *time.Time `json:"-"`

	// Relationships
	Tokens        []Token        `gorm:"foreignKey:UserID" json:"-"`
	Organizations []Organization `gorm:"many2many:organization_members;" json:"-"`
	Registries    []Registry     `gorm:"foreignKey:OwnerID" json:"-"`
	Tickets       []Ticket       `gorm:"foreignKey:UserID" json:"-"`
}

// SetPassword hashes and sets the user's password
func (u *User) SetPassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.PasswordHash = string(hash)
	return nil
}

// CheckPassword verifies the provided password against the stored hash
func (u *User) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
}

// IsAdmin returns true if the user has admin role
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// IsLocked returns true if the user account is currently locked
func (u *User) IsLocked() bool {
	if u.LockedUntil == nil {
		return false
	}
	return time.Now().Before(*u.LockedUntil)
}

// IncrementFailedLogins increments the failed login counter and locks account if threshold exceeded
func (u *User) IncrementFailedLogins(db *gorm.DB) error {
	u.FailedLogins++

	// Lock account for 15 minutes after 5 failed attempts
	if u.FailedLogins >= 5 {
		lockUntil := time.Now().Add(15 * time.Minute)
		u.LockedUntil = &lockUntil
	}

	return db.Save(u).Error
}

// ResetFailedLogins resets the failed login counter
func (u *User) ResetFailedLogins(db *gorm.DB) error {
	u.FailedLogins = 0
	u.LockedUntil = nil
	now := time.Now()
	u.LastLogin = &now
	return db.Save(u).Error
}

// HasQuotaSpace returns true if the user has quota space available
func (u *User) HasQuotaSpace(additionalSize int64) bool {
	if u.QuotaLimit == 0 {
		return true // unlimited
	}
	return u.QuotaUsed+additionalSize <= u.QuotaLimit
}

// UpdateQuotaUsage updates the user's quota usage
func (u *User) UpdateQuotaUsage(db *gorm.DB, delta int64) error {
	u.QuotaUsed += delta
	if u.QuotaUsed < 0 {
		u.QuotaUsed = 0
	}
	return db.Save(u).Error
}
