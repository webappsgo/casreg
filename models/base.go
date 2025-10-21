package models

import (
	"time"

	"gorm.io/gorm"
)

// Base contains common columns for all models
type Base struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// Visibility levels
const (
	VisibilityPublic   = "public"
	VisibilityPrivate  = "private"
	VisibilityInternal = "internal"
	VisibilityHidden   = "hidden"
)

// User roles
const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

// Organization roles
const (
	OrgRoleOwner  = "owner"
	OrgRoleAdmin  = "admin"
	OrgRoleMember = "member"
)

// Ticket priorities
const (
	PriorityLow      = "low"
	PriorityMedium   = "medium"
	PriorityHigh     = "high"
	PriorityCritical = "critical"
)

// Ticket statuses
const (
	StatusOpen       = "open"
	StatusInProgress = "in-progress"
	StatusResolved   = "resolved"
	StatusClosed     = "closed"
)

// Token scopes
const (
	ScopeGlobal        = "global"
	ScopeRegistryRead  = "registry:read"
	ScopeRegistryWrite = "registry:write"
	ScopeRegistryAdmin = "registry:admin"
	ScopeOrgRead       = "org:read"
	ScopeOrgWrite      = "org:write"
	ScopeOrgAdmin      = "org:admin"
	ScopeUserProfile   = "user:profile"
	ScopeAPIReadonly   = "api:readonly"
)
