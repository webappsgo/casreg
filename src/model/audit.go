package model

import (
	"encoding/json"
	"errors"

	"gorm.io/gorm"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	Base
	UserID       uint   `gorm:"index" json:"user_id"` // 0 for system actions
	Action       string `gorm:"not null;index" json:"action"`
	EntityType   string `gorm:"not null;index" json:"entity_type"`
	EntityID     uint   `gorm:"index" json:"entity_id"`
	EntityName   string `json:"entity_name,omitempty"`
	Description  string `gorm:"type:text" json:"description"`
	IPAddress    string `gorm:"index" json:"ip_address"`
	UserAgent    string `json:"user_agent,omitempty"`
	Method       string `json:"method,omitempty"` // HTTP method
	Path         string `json:"path,omitempty"`   // Request path
	StatusCode   int    `json:"status_code,omitempty"`
	Severity     string `gorm:"not null;default:'info';index" json:"severity"`
	Category     string `gorm:"not null;index" json:"category"`
	Metadata     string `gorm:"type:jsonb" json:"metadata,omitempty"` // JSON metadata
	ChangesBefore string `gorm:"type:jsonb" json:"changes_before,omitempty"` // Previous state
	ChangesAfter  string `gorm:"type:jsonb" json:"changes_after,omitempty"`  // New state

	// Relationships
	User *User `gorm:"foreignKey:UserID" json:"-"`
}

// Audit actions
const (
	AuditActionCreate = "create"
	AuditActionRead   = "read"
	AuditActionUpdate = "update"
	AuditActionDelete = "delete"
	AuditActionLogin  = "login"
	AuditActionLogout = "logout"
	AuditActionAccess = "access"
	AuditActionDenied = "denied"
	AuditActionPush   = "push"
	AuditActionPull   = "pull"
	AuditActionScan   = "scan"
	AuditActionSign   = "sign"
	AuditActionVerify = "verify"
)

// Audit entity types
const (
	AuditEntityUser         = "user"
	AuditEntityOrganization = "organization"
	AuditEntityRegistry     = "registry"
	AuditEntityRepository   = "repository"
	AuditEntityTag          = "tag"
	AuditEntityToken        = "token"
	AuditEntityTicket       = "ticket"
	AuditEntityIssue        = "issue"
	AuditEntityConfig       = "configuration"
	AuditEntitySystem       = "system"
)

// Audit severities
const (
	AuditSeverityInfo     = "info"
	AuditSeverityWarning  = "warning"
	AuditSeverityError    = "error"
	AuditSeverityCritical = "critical"
)

// Audit categories
const (
	AuditCategoryAuthentication = "authentication"
	AuditCategoryAuthorization  = "authorization"
	AuditCategoryData           = "data"
	AuditCategorySecurity       = "security"
	AuditCategoryConfiguration  = "configuration"
	AuditCategoryRegistry       = "registry"
	AuditCategorySystem         = "system"
	AuditCategoryCompliance     = "compliance"
)

// BeforeCreate validates audit log data before creation
func (a *AuditLog) BeforeCreate(tx *gorm.DB) error {
	if a.Action == "" {
		return errors.New("audit action is required")
	}

	if a.EntityType == "" {
		return errors.New("audit entity type is required")
	}

	// Validate severity
	validSeverities := map[string]bool{
		AuditSeverityInfo:     true,
		AuditSeverityWarning:  true,
		AuditSeverityError:    true,
		AuditSeverityCritical: true,
	}

	if !validSeverities[a.Severity] {
		a.Severity = AuditSeverityInfo
	}

	// Validate category
	validCategories := map[string]bool{
		AuditCategoryAuthentication: true,
		AuditCategoryAuthorization:  true,
		AuditCategoryData:           true,
		AuditCategorySecurity:       true,
		AuditCategoryConfiguration:  true,
		AuditCategoryRegistry:       true,
		AuditCategorySystem:         true,
		AuditCategoryCompliance:     true,
	}

	if !validCategories[a.Category] {
		a.Category = AuditCategorySystem
	}

	return nil
}

// SetMetadata sets the audit log metadata from a map
func (a *AuditLog) SetMetadata(data map[string]interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	a.Metadata = string(jsonData)
	return nil
}

// GetMetadata retrieves the audit log metadata as a map
func (a *AuditLog) GetMetadata() (map[string]interface{}, error) {
	if a.Metadata == "" {
		return make(map[string]interface{}), nil
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(a.Metadata), &data); err != nil {
		return nil, err
	}

	return data, nil
}

// SetChangesBefore sets the changes before data from a map
func (a *AuditLog) SetChangesBefore(data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	a.ChangesBefore = string(jsonData)
	return nil
}

// SetChangesAfter sets the changes after data from a map
func (a *AuditLog) SetChangesAfter(data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	a.ChangesAfter = string(jsonData)
	return nil
}

// CreateAuditLog creates a new audit log entry
func CreateAuditLog(db *gorm.DB, userID uint, action, entityType string, entityID uint, entityName, description, ipAddress, category string) error {
	auditLog := &AuditLog{
		UserID:      userID,
		Action:      action,
		EntityType:  entityType,
		EntityID:    entityID,
		EntityName:  entityName,
		Description: description,
		IPAddress:   ipAddress,
		Severity:    AuditSeverityInfo,
		Category:    category,
	}

	return db.Create(auditLog).Error
}

// CreateSecurityAuditLog creates a security-related audit log entry
func CreateSecurityAuditLog(db *gorm.DB, userID uint, action, entityType string, entityID uint, description, ipAddress string, severity string) error {
	auditLog := &AuditLog{
		UserID:      userID,
		Action:      action,
		EntityType:  entityType,
		EntityID:    entityID,
		Description: description,
		IPAddress:   ipAddress,
		Severity:    severity,
		Category:    AuditCategorySecurity,
	}

	return db.Create(auditLog).Error
}

// CreateLoginAuditLog creates an audit log for login attempts
func CreateLoginAuditLog(db *gorm.DB, userID uint, username, ipAddress, userAgent string, success bool) error {
	severity := AuditSeverityInfo
	action := AuditActionLogin
	description := "Successful login"

	if !success {
		severity = AuditSeverityWarning
		action = AuditActionDenied
		description = "Failed login attempt"
	}

	auditLog := &AuditLog{
		UserID:      userID,
		Action:      action,
		EntityType:  AuditEntityUser,
		EntityID:    userID,
		EntityName:  username,
		Description: description,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Severity:    severity,
		Category:    AuditCategoryAuthentication,
	}

	return db.Create(auditLog).Error
}

// CreateRegistryAccessLog creates an audit log for registry access
func CreateRegistryAccessLog(db *gorm.DB, userID uint, action string, repositoryID uint, repositoryName, ipAddress string) error {
	auditLog := &AuditLog{
		UserID:      userID,
		Action:      action,
		EntityType:  AuditEntityRepository,
		EntityID:    repositoryID,
		EntityName:  repositoryName,
		Description: action + " operation on repository",
		IPAddress:   ipAddress,
		Severity:    AuditSeverityInfo,
		Category:    AuditCategoryRegistry,
	}

	return db.Create(auditLog).Error
}

// CreateConfigChangeLog creates an audit log for configuration changes
func CreateConfigChangeLog(db *gorm.DB, userID uint, configKey string, oldValue, newValue interface{}, ipAddress string) error {
	auditLog := &AuditLog{
		UserID:      userID,
		Action:      AuditActionUpdate,
		EntityType:  AuditEntityConfig,
		EntityName:  configKey,
		Description: "Configuration change: " + configKey,
		IPAddress:   ipAddress,
		Severity:    AuditSeverityInfo,
		Category:    AuditCategoryConfiguration,
	}

	// Set changes
	if err := auditLog.SetChangesBefore(oldValue); err != nil {
		return err
	}

	if err := auditLog.SetChangesAfter(newValue); err != nil {
		return err
	}

	return db.Create(auditLog).Error
}

// GetUserAuditLogs retrieves audit logs for a specific user
func GetUserAuditLogs(db *gorm.DB, userID uint, limit int, offset int) ([]AuditLog, error) {
	var logs []AuditLog
	err := db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error
	return logs, err
}

// GetSecurityAuditLogs retrieves security-related audit logs
func GetSecurityAuditLogs(db *gorm.DB, limit int, offset int) ([]AuditLog, error) {
	var logs []AuditLog
	err := db.Where("category = ?", AuditCategorySecurity).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error
	return logs, err
}

// GetFailedLoginAttempts retrieves failed login attempts
func GetFailedLoginAttempts(db *gorm.DB, ipAddress string, since int) (int64, error) {
	var count int64
	err := db.Model(&AuditLog{}).
		Where("action = ? AND ip_address = ? AND created_at > NOW() - INTERVAL '? minutes'",
			AuditActionDenied, ipAddress, since).
		Count(&count).Error
	return count, err
}
