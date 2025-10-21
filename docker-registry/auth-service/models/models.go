package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// User represents a user account
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Username  string    `json:"username" gorm:"unique;not null;size:50"`
	Email     string    `json:"email" gorm:"unique;not null;size:255"`
	Password  string    `json:"-" gorm:"not null;size:255"`
	FirstName string    `json:"first_name" gorm:"size:100"`
	LastName  string    `json:"last_name" gorm:"size:100"`
	Role      string    `json:"role" gorm:"not null;default:'user';size:20"`
	IsActive  bool      `json:"is_active" gorm:"not null;default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	OwnedRegistries   []Registry                  `json:"-" gorm:"foreignKey:UserID"`
	PersonalTokens    []PersonalToken             `json:"-" gorm:"foreignKey:UserID"`
	OrganizationMemberships []OrganizationMembership `json:"-" gorm:"foreignKey:UserID"`
	RepositoryStars   []RepositoryStar            `json:"-" gorm:"foreignKey:UserID"`
	AuditLogs         []AuditLog                  `json:"-" gorm:"foreignKey:UserID"`
}

// Organization represents an organization
type Organization struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string    `json:"name" gorm:"unique;not null;size:50;index"`
	DisplayName string    `json:"display_name" gorm:"not null;size:100"`
	Description string    `json:"description" gorm:"type:text"`
	IsPublic    bool      `json:"is_public" gorm:"not null;default:false"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	Members     []OrganizationMembership `json:"-" gorm:"foreignKey:OrganizationID"`
	Registries  []Registry               `json:"-" gorm:"foreignKey:OrganizationID"`
}

// OrganizationMembership represents membership in an organization
type OrganizationMembership struct {
	ID             uint         `json:"id" gorm:"primaryKey;autoIncrement"`
	OrganizationID uint         `json:"organization_id" gorm:"not null;index"`
	UserID         uint         `json:"user_id" gorm:"not null;index"`
	Role           string       `json:"role" gorm:"not null;default:'member';size:20"`
	CreatedAt      time.Time    `json:"created_at"`

	// Relationships
	Organization Organization `json:"-" gorm:"constraint:OnDelete:CASCADE"`
	User         User         `json:"-" gorm:"constraint:OnDelete:CASCADE"`
}

// Registry represents a Docker registry
type Registry struct {
	ID             uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	Name           string `json:"name" gorm:"not null;size:100;index"`
	DisplayName    string `json:"display_name" gorm:"size:100"`
	Description    string `json:"description" gorm:"type:text"`
	IsPublic       bool   `json:"is_public" gorm:"not null;default:false"`
	UserID         *uint  `json:"user_id,omitempty" gorm:"index"`
	OrganizationID *uint  `json:"organization_id,omitempty" gorm:"index"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// Relationships
	User         *User         `json:"-" gorm:"constraint:OnDelete:CASCADE"`
	Organization *Organization `json:"-" gorm:"constraint:OnDelete:CASCADE"`
	Repositories []Repository  `json:"-" gorm:"foreignKey:RegistryID"`
}

// Repository represents a Docker repository within a registry
type Repository struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	RegistryID  uint      `json:"registry_id" gorm:"not null;index"`
	Name        string    `json:"name" gorm:"not null;size:100;index"`
	Description string    `json:"description" gorm:"type:text"`
	IsPublic    bool      `json:"is_public" gorm:"not null;default:false"`
	PullCount   int64     `json:"pull_count" gorm:"not null;default:0"`
	PushCount   int64     `json:"push_count" gorm:"not null;default:0"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	Registry Registry         `json:"-" gorm:"constraint:OnDelete:CASCADE"`
	Tags     []Tag            `json:"-" gorm:"foreignKey:RepositoryID"`
	Stars    []RepositoryStar `json:"-" gorm:"foreignKey:RepositoryID"`
}

// Tag represents a Docker image tag within a repository
type Tag struct {
	ID           uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	RepositoryID uint      `json:"repository_id" gorm:"not null;index"`
	Name         string    `json:"name" gorm:"not null;size:255"`
	Digest       string    `json:"digest" gorm:"not null;size:255;index"`
	Size         int64     `json:"size" gorm:"not null;default:0"`
	PullCount    int64     `json:"pull_count" gorm:"not null;default:0"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relationships
	Repository Repository `json:"-" gorm:"constraint:OnDelete:CASCADE"`
}

// RepositoryStar represents a user starring a repository
type RepositoryStar struct {
	ID           uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	RepositoryID uint      `json:"repository_id" gorm:"not null;index"`
	UserID       uint      `json:"user_id" gorm:"not null;index"`
	CreatedAt    time.Time `json:"created_at"`

	// Relationships
	Repository Repository `json:"-" gorm:"constraint:OnDelete:CASCADE"`
	User       User       `json:"-" gorm:"constraint:OnDelete:CASCADE"`
}

// PersonalToken represents a user's API token
type PersonalToken struct {
	ID          uint       `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID      uint       `json:"user_id" gorm:"not null;index"`
	Name        string     `json:"name" gorm:"not null;size:100"`
	Token       string     `json:"token,omitempty" gorm:"unique;not null;size:255;index"`
	Scopes      StringSlice `json:"scopes" gorm:"type:text"`
	Description string     `json:"description" gorm:"type:text"`
	ExpiresAt   *time.Time `json:"expires_at"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// Relationships
	User User `json:"-" gorm:"constraint:OnDelete:CASCADE"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID         uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID     uint      `json:"user_id" gorm:"not null;index"`
	Action     string    `json:"action" gorm:"not null;size:100"`
	Resource   string    `json:"resource" gorm:"not null;size:100"`
	ResourceID uint      `json:"resource_id" gorm:"not null"`
	IP         string    `json:"ip" gorm:"size:45"`
	UserAgent  string    `json:"user_agent" gorm:"type:text"`
	CreatedAt  time.Time `json:"created_at" gorm:"index"`

	// Relationships
	User User `json:"-" gorm:"constraint:OnDelete:CASCADE"`
}

// StringSlice is a custom type for storing string slices in the database
type StringSlice []string

// Scan implements the sql.Scanner interface for reading from database
func (ss *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*ss = StringSlice{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, ss)
	case string:
		return json.Unmarshal([]byte(v), ss)
	}

	*ss = StringSlice{}
	return nil
}

// Value implements the driver.Valuer interface for writing to database
func (ss StringSlice) Value() (driver.Value, error) {
	if len(ss) == 0 {
		return "[]", nil
	}

	b, err := json.Marshal(ss)
	if err != nil {
		return nil, err
	}

	return string(b), nil
}

// GormDataType tells GORM what SQL data type to use
func (StringSlice) GormDataType() string {
	return "text"
}

// TableName methods to ensure proper table names
func (User) TableName() string {
	return "users"
}

func (Organization) TableName() string {
	return "organizations"
}

func (OrganizationMembership) TableName() string {
	return "organization_memberships"
}

func (Registry) TableName() string {
	return "registries"
}

func (Repository) TableName() string {
	return "repositories"
}

func (Tag) TableName() string {
	return "tags"
}

func (RepositoryStar) TableName() string {
	return "repository_stars"
}

func (PersonalToken) TableName() string {
	return "personal_tokens"
}

func (AuditLog) TableName() string {
	return "audit_logs"
}

// BeforeCreate hooks for GORM
func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	return nil
}

func (u *User) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = time.Now()
	return nil
}

func (o *Organization) BeforeCreate(tx *gorm.DB) error {
	o.CreatedAt = time.Now()
	o.UpdatedAt = time.Now()
	return nil
}

func (o *Organization) BeforeUpdate(tx *gorm.DB) error {
	o.UpdatedAt = time.Now()
	return nil
}