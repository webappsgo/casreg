package model

import (
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// Webhook represents a webhook configuration
type Webhook struct {
	Base
	Name           string `gorm:"not null" json:"name"`
	URL            string `gorm:"not null" json:"url"`
	Secret         string `json:"-"` // HMAC secret for signature validation
	Events         string `gorm:"type:text;not null" json:"events"` // comma-separated event types
	EntityType     string `gorm:"index" json:"entity_type"` // user, organization, registry, repository
	EntityID       uint   `gorm:"index" json:"entity_id"`
	IsActive       bool   `gorm:"not null;default:true;index" json:"is_active"`
	SSLVerify      bool   `gorm:"not null;default:true" json:"ssl_verify"`
	ContentType    string `gorm:"default:'application/json'" json:"content_type"`
	CustomHeaders  string `gorm:"type:jsonb" json:"custom_headers,omitempty"` // JSON object
	Timeout        int    `gorm:"default:30" json:"timeout"` // seconds
	RetryCount     int    `gorm:"default:3" json:"retry_count"`
	LastDeliveryAt *time.Time `json:"last_delivery_at,omitempty"`
	LastStatus     string `json:"last_status,omitempty"`
	FailureCount   int    `gorm:"default:0" json:"failure_count"`

	// Relationships
	Deliveries []WebhookDelivery `gorm:"foreignKey:WebhookID;constraint:OnDelete:CASCADE" json:"-"`
}

// WebhookDelivery represents a webhook delivery attempt
type WebhookDelivery struct {
	Base
	WebhookID      uint       `gorm:"not null;index" json:"webhook_id"`
	Event          string     `gorm:"not null;index" json:"event"`
	PayloadHash    string     `gorm:"index" json:"payload_hash"`
	Payload        string     `gorm:"type:jsonb;not null" json:"-"`
	Status         string     `gorm:"not null;default:'pending';index" json:"status"`
	StatusCode     int        `json:"status_code,omitempty"`
	RequestHeaders string     `gorm:"type:jsonb" json:"request_headers,omitempty"`
	ResponseBody   string     `gorm:"type:text" json:"response_body,omitempty"`
	ResponseHeaders string    `gorm:"type:jsonb" json:"response_headers,omitempty"`
	ErrorMessage   string     `gorm:"type:text" json:"error_message,omitempty"`
	DeliveredAt    *time.Time `json:"delivered_at,omitempty"`
	Duration       int64      `json:"duration,omitempty"` // milliseconds
	AttemptNumber  int        `gorm:"default:1" json:"attempt_number"`

	// Relationships
	Webhook Webhook `gorm:"foreignKey:WebhookID" json:"-"`
}

// Webhook events
const (
	WebhookEventPush              = "push"
	WebhookEventPull              = "pull"
	WebhookEventTagCreate         = "tag:create"
	WebhookEventTagDelete         = "tag:delete"
	WebhookEventRepoCreate        = "repository:create"
	WebhookEventRepoDelete        = "repository:delete"
	WebhookEventRegistryCreate    = "registry:create"
	WebhookEventRegistryDelete    = "registry:delete"
	WebhookEventScanComplete      = "scan:complete"
	WebhookEventScanFailed        = "scan:failed"
	WebhookEventSignatureVerified = "signature:verified"
	WebhookEventSignatureFailed   = "signature:failed"
	WebhookEventVulnerability     = "vulnerability:found"
	WebhookEventQuotaWarning      = "quota:warning"
	WebhookEventQuotaExceeded     = "quota:exceeded"
	WebhookEventUserCreate        = "user:create"
	WebhookEventUserDelete        = "user:delete"
	WebhookEventOrgCreate         = "organization:create"
	WebhookEventOrgDelete         = "organization:delete"
)

// Delivery statuses
const (
	DeliveryStatusPending   = "pending"
	DeliveryStatusDelivered = "delivered"
	DeliveryStatusFailed    = "failed"
	DeliveryStatusRetrying  = "retrying"
)

// BeforeCreate validates webhook data before creation
func (w *Webhook) BeforeCreate(tx *gorm.DB) error {
	if w.Name == "" {
		return errors.New("webhook name is required")
	}

	if w.URL == "" {
		return errors.New("webhook URL is required")
	}

	if w.Events == "" {
		return errors.New("at least one event must be specified")
	}

	// Validate events
	events := w.GetEventsArray()
	for _, event := range events {
		if !IsValidWebhookEvent(event) {
			return errors.New("invalid event type: " + event)
		}
	}

	// Set defaults
	if w.ContentType == "" {
		w.ContentType = "application/json"
	}

	if w.Timeout == 0 {
		w.Timeout = 30
	}

	if w.RetryCount == 0 {
		w.RetryCount = 3
	}

	return nil
}

// GetEventsArray returns the webhook events as a slice
func (w *Webhook) GetEventsArray() []string {
	if w.Events == "" {
		return []string{}
	}

	events := []string{}
	current := ""
	for _, char := range w.Events {
		if char == ',' {
			if trimmed := trimWebhookString(current); trimmed != "" {
				events = append(events, trimmed)
			}
			current = ""
		} else {
			current += string(char)
		}
	}
	if trimmed := trimWebhookString(current); trimmed != "" {
		events = append(events, trimmed)
	}

	return events
}

// HasEvent checks if the webhook is configured for a specific event
func (w *Webhook) HasEvent(event string) bool {
	events := w.GetEventsArray()
	for _, e := range events {
		if e == event {
			return true
		}
	}
	return false
}

// SetCustomHeaders sets the custom headers from a map
func (w *Webhook) SetCustomHeaders(headers map[string]string) error {
	jsonData, err := json.Marshal(headers)
	if err != nil {
		return err
	}
	w.CustomHeaders = string(jsonData)
	return nil
}

// GetCustomHeaders retrieves the custom headers as a map
func (w *Webhook) GetCustomHeaders() (map[string]string, error) {
	if w.CustomHeaders == "" {
		return make(map[string]string), nil
	}

	var headers map[string]string
	if err := json.Unmarshal([]byte(w.CustomHeaders), &headers); err != nil {
		return nil, err
	}

	return headers, nil
}

// RecordDelivery records a successful delivery
func (w *Webhook) RecordDelivery(db *gorm.DB, status string) error {
	now := time.Now()
	w.LastDeliveryAt = &now
	w.LastStatus = status

	if status == DeliveryStatusFailed {
		w.FailureCount++
	} else if status == DeliveryStatusDelivered {
		w.FailureCount = 0
	}

	// Auto-disable webhook after too many failures
	if w.FailureCount >= 10 {
		w.IsActive = false
	}

	return db.Save(w).Error
}

// CreateDelivery creates a new webhook delivery
func (w *Webhook) CreateDelivery(db *gorm.DB, event string, payload interface{}) (*WebhookDelivery, error) {
	if !w.HasEvent(event) {
		return nil, errors.New("webhook not configured for this event")
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	delivery := &WebhookDelivery{
		WebhookID: w.ID,
		Event:     event,
		Payload:   string(jsonPayload),
		Status:    DeliveryStatusPending,
	}

	if err := db.Create(delivery).Error; err != nil {
		return nil, err
	}

	return delivery, nil
}

// MarkDelivered marks the delivery as successfully delivered
func (wd *WebhookDelivery) MarkDelivered(db *gorm.DB, statusCode int, responseBody string, duration int64) error {
	now := time.Now()
	wd.Status = DeliveryStatusDelivered
	wd.StatusCode = statusCode
	wd.ResponseBody = responseBody
	wd.DeliveredAt = &now
	wd.Duration = duration
	return db.Save(wd).Error
}

// MarkFailed marks the delivery as failed
func (wd *WebhookDelivery) MarkFailed(db *gorm.DB, statusCode int, errorMessage string, duration int64) error {
	wd.Status = DeliveryStatusFailed
	wd.StatusCode = statusCode
	wd.ErrorMessage = errorMessage
	wd.Duration = duration
	return db.Save(wd).Error
}

// Retry increments the attempt number for retry
func (wd *WebhookDelivery) Retry(db *gorm.DB) error {
	wd.AttemptNumber++
	wd.Status = DeliveryStatusRetrying
	return db.Save(wd).Error
}

// GetPayload retrieves the delivery payload as a map
func (wd *WebhookDelivery) GetPayload() (map[string]interface{}, error) {
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(wd.Payload), &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

// SetRequestHeaders sets the request headers from a map
func (wd *WebhookDelivery) SetRequestHeaders(headers map[string]string) error {
	jsonData, err := json.Marshal(headers)
	if err != nil {
		return err
	}
	wd.RequestHeaders = string(jsonData)
	return nil
}

// SetResponseHeaders sets the response headers from a map
func (wd *WebhookDelivery) SetResponseHeaders(headers map[string]string) error {
	jsonData, err := json.Marshal(headers)
	if err != nil {
		return err
	}
	wd.ResponseHeaders = string(jsonData)
	return nil
}

// IsValidWebhookEvent checks if an event type is valid
func IsValidWebhookEvent(event string) bool {
	validEvents := map[string]bool{
		WebhookEventPush:              true,
		WebhookEventPull:              true,
		WebhookEventTagCreate:         true,
		WebhookEventTagDelete:         true,
		WebhookEventRepoCreate:        true,
		WebhookEventRepoDelete:        true,
		WebhookEventRegistryCreate:    true,
		WebhookEventRegistryDelete:    true,
		WebhookEventScanComplete:      true,
		WebhookEventScanFailed:        true,
		WebhookEventSignatureVerified: true,
		WebhookEventSignatureFailed:   true,
		WebhookEventVulnerability:     true,
		WebhookEventQuotaWarning:      true,
		WebhookEventQuotaExceeded:     true,
		WebhookEventUserCreate:        true,
		WebhookEventUserDelete:        true,
		WebhookEventOrgCreate:         true,
		WebhookEventOrgDelete:         true,
	}

	return validEvents[event]
}

// GetPendingDeliveries retrieves all pending webhook deliveries
func GetPendingDeliveries(db *gorm.DB, limit int) ([]WebhookDelivery, error) {
	var deliveries []WebhookDelivery
	err := db.Where("status IN ?", []string{DeliveryStatusPending, DeliveryStatusRetrying}).
		Order("created_at ASC").
		Limit(limit).
		Find(&deliveries).Error
	return deliveries, err
}

// Helper function to trim strings
func trimWebhookString(s string) string {
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
