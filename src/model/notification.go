package model

import (
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// Notification represents a notification message
type Notification struct {
	Base
	UserID      uint       `gorm:"index" json:"user_id"` // 0 for system-wide notifications
	Type        string     `gorm:"not null;index" json:"type"`
	Title       string     `gorm:"not null" json:"title"`
	Message     string     `gorm:"type:text;not null" json:"message"`
	Severity    string     `gorm:"not null;default:'info'" json:"severity"`
	Category    string     `gorm:"index" json:"category"`
	IsRead      bool       `gorm:"not null;default:false;index" json:"is_read"`
	ReadAt      *time.Time `json:"read_at,omitempty"`
	Link        string     `json:"link,omitempty"`
	Metadata    string     `gorm:"type:jsonb" json:"metadata,omitempty"` // JSON metadata
	ExpiresAt   *time.Time `gorm:"index" json:"expires_at,omitempty"`
	DeliveryChannels string `json:"delivery_channels"` // comma-separated: email,webhook,websocket,in-app

	// Relationships
	User *User `gorm:"foreignKey:UserID" json:"-"`
}

// Notification types
const (
	NotificationTypeSecurity     = "security"
	NotificationTypeSystem       = "system"
	NotificationTypeOperational  = "operational"
	NotificationTypeAdmin        = "administrative"
	NotificationTypeSupport      = "support"
	NotificationTypeVulnerability = "vulnerability"
	NotificationTypeQuota        = "quota"
	NotificationTypeAudit        = "audit"
)

// Notification severities
const (
	NotificationSeverityInfo     = "info"
	NotificationSeverityWarning  = "warning"
	NotificationSeverityError    = "error"
	NotificationSeverityCritical = "critical"
	NotificationSeveritySuccess  = "success"
)

// Notification categories
const (
	NotificationCategoryAlert       = "alert"
	NotificationCategoryUpdate      = "update"
	NotificationCategoryAnnouncement = "announcement"
	NotificationCategoryReminder    = "reminder"
)

// Delivery channels
const (
	DeliveryChannelEmail     = "email"
	DeliveryChannelWebhook   = "webhook"
	DeliveryChannelWebSocket = "websocket"
	DeliveryChannelInApp     = "in-app"
	DeliveryChannelPush      = "push"
)

// BeforeCreate validates notification data before creation
func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.Title == "" {
		return errors.New("notification title is required")
	}

	if n.Message == "" {
		return errors.New("notification message is required")
	}

	// Validate type
	validTypes := map[string]bool{
		NotificationTypeSecurity:     true,
		NotificationTypeSystem:       true,
		NotificationTypeOperational:  true,
		NotificationTypeAdmin:        true,
		NotificationTypeSupport:      true,
		NotificationTypeVulnerability: true,
		NotificationTypeQuota:        true,
		NotificationTypeAudit:        true,
	}

	if !validTypes[n.Type] {
		return errors.New("invalid notification type")
	}

	// Validate severity
	validSeverities := map[string]bool{
		NotificationSeverityInfo:     true,
		NotificationSeverityWarning:  true,
		NotificationSeverityError:    true,
		NotificationSeverityCritical: true,
		NotificationSeveritySuccess:  true,
	}

	if !validSeverities[n.Severity] {
		n.Severity = NotificationSeverityInfo
	}

	// Set default delivery channel if not specified
	if n.DeliveryChannels == "" {
		n.DeliveryChannels = DeliveryChannelInApp
	}

	return nil
}

// MarkAsRead marks the notification as read
func (n *Notification) MarkAsRead(db *gorm.DB) error {
	if n.IsRead {
		return nil // already read
	}

	now := time.Now()
	n.IsRead = true
	n.ReadAt = &now
	return db.Save(n).Error
}

// MarkAsUnread marks the notification as unread
func (n *Notification) MarkAsUnread(db *gorm.DB) error {
	n.IsRead = false
	n.ReadAt = nil
	return db.Save(n).Error
}

// IsExpired checks if the notification has expired
func (n *Notification) IsExpired() bool {
	if n.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*n.ExpiresAt)
}

// SetMetadata sets the notification metadata from a map
func (n *Notification) SetMetadata(data map[string]interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	n.Metadata = string(jsonData)
	return nil
}

// GetMetadata retrieves the notification metadata as a map
func (n *Notification) GetMetadata() (map[string]interface{}, error) {
	if n.Metadata == "" {
		return make(map[string]interface{}), nil
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(n.Metadata), &data); err != nil {
		return nil, err
	}

	return data, nil
}

// HasDeliveryChannel checks if the notification should be delivered via a specific channel
func (n *Notification) HasDeliveryChannel(channel string) bool {
	if n.DeliveryChannels == "" {
		return false
	}

	channels := n.GetDeliveryChannels()
	for _, ch := range channels {
		if ch == channel {
			return true
		}
	}

	return false
}

// GetDeliveryChannels returns the delivery channels as a slice
func (n *Notification) GetDeliveryChannels() []string {
	if n.DeliveryChannels == "" {
		return []string{}
	}

	channels := []string{}
	for _, ch := range splitAndTrim(n.DeliveryChannels, ",") {
		channels = append(channels, ch)
	}

	return channels
}

// CreateSystemNotification creates a system-wide notification
func CreateSystemNotification(db *gorm.DB, notifType, title, message, severity string) error {
	notification := &Notification{
		UserID:   0, // system-wide
		Type:     notifType,
		Title:    title,
		Message:  message,
		Severity: severity,
		Category: NotificationCategoryAnnouncement,
	}

	return db.Create(notification).Error
}

// CreateUserNotification creates a notification for a specific user
func CreateUserNotification(db *gorm.DB, userID uint, notifType, title, message, severity string) error {
	notification := &Notification{
		UserID:   userID,
		Type:     notifType,
		Title:    title,
		Message:  message,
		Severity: severity,
		Category: NotificationCategoryAlert,
	}

	return db.Create(notification).Error
}

// GetUnreadCount returns the count of unread notifications for a user
func GetUnreadCount(db *gorm.DB, userID uint) (int64, error) {
	var count int64
	err := db.Model(&Notification{}).
		Where("user_id IN (?, ?) AND is_read = ?", userID, 0, false).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Count(&count).Error
	return count, err
}

// MarkAllAsRead marks all notifications for a user as read
func MarkAllAsRead(db *gorm.DB, userID uint) error {
	now := time.Now()
	return db.Model(&Notification{}).
		Where("user_id IN (?, ?) AND is_read = ?", userID, 0, false).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": &now,
		}).Error
}

// DeleteExpiredNotifications deletes all expired notifications
func DeleteExpiredNotifications(db *gorm.DB) error {
	return db.Where("expires_at IS NOT NULL AND expires_at < ?", time.Now()).
		Delete(&Notification{}).Error
}

// Helper function to split and trim strings
func splitAndTrim(s, sep string) []string {
	parts := []string{}
	for _, part := range splitString(s, sep) {
		trimmed := trimString(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

func splitString(s, sep string) []string {
	result := []string{}
	current := ""
	for _, char := range s {
		if string(char) == sep {
			result = append(result, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func trimString(s string) string {
	// Simple trim implementation
	start := 0
	end := len(s)

	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}

	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}

	return s[start:end]
}
