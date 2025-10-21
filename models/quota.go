package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// Quota represents resource quota management
type Quota struct {
	Base
	EntityType  string     `gorm:"not null;index:idx_quota_entity,unique" json:"entity_type"` // user, organization, registry, repository
	EntityID    uint       `gorm:"not null;index:idx_quota_entity,unique" json:"entity_id"`
	ResourceType string    `gorm:"not null;index:idx_quota_entity,unique" json:"resource_type"` // storage, bandwidth, repositories, tags, api_calls
	Limit       int64      `gorm:"not null" json:"limit"` // 0 = unlimited
	Used        int64      `gorm:"default:0" json:"used"`
	SoftLimit   int64      `gorm:"default:0" json:"soft_limit"` // warning threshold
	ResetPeriod string     `json:"reset_period,omitempty"` // daily, weekly, monthly, never
	LastReset   *time.Time `json:"last_reset,omitempty"`
	NextReset   *time.Time `json:"next_reset,omitempty"`
	IsActive    bool       `gorm:"not null;default:true" json:"is_active"`
}

// Entity types
const (
	EntityTypeUser         = "user"
	EntityTypeOrganization = "organization"
	EntityTypeRegistry     = "registry"
	EntityTypeRepository   = "repository"
)

// Resource types
const (
	ResourceTypeStorage      = "storage"
	ResourceTypeBandwidth    = "bandwidth"
	ResourceTypeRepositories = "repositories"
	ResourceTypeTags         = "tags"
	ResourceTypeAPICalls     = "api_calls"
	ResourceTypePushOps      = "push_operations"
	ResourceTypePullOps      = "pull_operations"
)

// Reset periods
const (
	ResetPeriodNever   = "never"
	ResetPeriodDaily   = "daily"
	ResetPeriodWeekly  = "weekly"
	ResetPeriodMonthly = "monthly"
	ResetPeriodYearly  = "yearly"
)

// BeforeCreate validates quota data before creation
func (q *Quota) BeforeCreate(tx *gorm.DB) error {
	// Validate entity type
	validEntityTypes := map[string]bool{
		EntityTypeUser:         true,
		EntityTypeOrganization: true,
		EntityTypeRegistry:     true,
		EntityTypeRepository:   true,
	}

	if !validEntityTypes[q.EntityType] {
		return errors.New("invalid entity type")
	}

	// Validate resource type
	validResourceTypes := map[string]bool{
		ResourceTypeStorage:      true,
		ResourceTypeBandwidth:    true,
		ResourceTypeRepositories: true,
		ResourceTypeTags:         true,
		ResourceTypeAPICalls:     true,
		ResourceTypePushOps:      true,
		ResourceTypePullOps:      true,
	}

	if !validResourceTypes[q.ResourceType] {
		return errors.New("invalid resource type")
	}

	// Validate reset period
	validResetPeriods := map[string]bool{
		ResetPeriodNever:   true,
		ResetPeriodDaily:   true,
		ResetPeriodWeekly:  true,
		ResetPeriodMonthly: true,
		ResetPeriodYearly:  true,
	}

	if q.ResetPeriod != "" && !validResetPeriods[q.ResetPeriod] {
		q.ResetPeriod = ResetPeriodNever
	}

	// Set initial reset time
	if q.ResetPeriod != "" && q.ResetPeriod != ResetPeriodNever {
		now := time.Now()
		q.LastReset = &now
		nextReset := q.calculateNextReset()
		q.NextReset = &nextReset
	}

	return nil
}

// HasSpace checks if there is available quota space
func (q *Quota) HasSpace(additionalUsage int64) bool {
	if !q.IsActive {
		return true // quota not enforced
	}

	if q.Limit == 0 {
		return true // unlimited
	}

	return q.Used+additionalUsage <= q.Limit
}

// IsSoftLimitExceeded checks if the soft limit has been exceeded
func (q *Quota) IsSoftLimitExceeded() bool {
	if q.SoftLimit == 0 {
		return false
	}

	return q.Used >= q.SoftLimit
}

// IsHardLimitExceeded checks if the hard limit has been exceeded
func (q *Quota) IsHardLimitExceeded() bool {
	if q.Limit == 0 {
		return false // unlimited
	}

	return q.Used >= q.Limit
}

// GetUsagePercentage returns the usage percentage
func (q *Quota) GetUsagePercentage() float64 {
	if q.Limit == 0 {
		return 0.0 // unlimited
	}

	return (float64(q.Used) / float64(q.Limit)) * 100.0
}

// GetAvailable returns the available quota
func (q *Quota) GetAvailable() int64 {
	if q.Limit == 0 {
		return -1 // unlimited
	}

	available := q.Limit - q.Used
	if available < 0 {
		return 0
	}

	return available
}

// Increment increments the quota usage
func (q *Quota) Increment(db *gorm.DB, amount int64) error {
	if !q.IsActive {
		return nil // quota not enforced
	}

	// Check if reset is needed
	if q.NeedsReset() {
		if err := q.Reset(db); err != nil {
			return err
		}
	}

	// Check if hard limit would be exceeded
	if q.Limit > 0 && q.Used+amount > q.Limit {
		return errors.New("quota limit exceeded")
	}

	q.Used += amount
	return db.Save(q).Error
}

// Decrement decrements the quota usage
func (q *Quota) Decrement(db *gorm.DB, amount int64) error {
	q.Used -= amount
	if q.Used < 0 {
		q.Used = 0
	}
	return db.Save(q).Error
}

// SetUsage sets the quota usage to a specific value
func (q *Quota) SetUsage(db *gorm.DB, usage int64) error {
	q.Used = usage
	if q.Used < 0 {
		q.Used = 0
	}
	return db.Save(q).Error
}

// NeedsReset checks if the quota needs to be reset based on the reset period
func (q *Quota) NeedsReset() bool {
	if q.ResetPeriod == "" || q.ResetPeriod == ResetPeriodNever {
		return false
	}

	if q.NextReset == nil {
		return false
	}

	return time.Now().After(*q.NextReset)
}

// Reset resets the quota usage
func (q *Quota) Reset(db *gorm.DB) error {
	q.Used = 0
	now := time.Now()
	q.LastReset = &now

	if q.ResetPeriod != "" && q.ResetPeriod != ResetPeriodNever {
		nextReset := q.calculateNextReset()
		q.NextReset = &nextReset
	}

	return db.Save(q).Error
}

// calculateNextReset calculates the next reset time based on the reset period
func (q *Quota) calculateNextReset() time.Time {
	now := time.Now()

	switch q.ResetPeriod {
	case ResetPeriodDaily:
		return now.Add(24 * time.Hour)
	case ResetPeriodWeekly:
		return now.Add(7 * 24 * time.Hour)
	case ResetPeriodMonthly:
		return now.AddDate(0, 1, 0)
	case ResetPeriodYearly:
		return now.AddDate(1, 0, 0)
	default:
		return now
	}
}

// GetQuota retrieves a quota for a specific entity and resource type
func GetQuota(db *gorm.DB, entityType string, entityID uint, resourceType string) (*Quota, error) {
	var quota Quota
	err := db.Where("entity_type = ? AND entity_id = ? AND resource_type = ?",
		entityType, entityID, resourceType).First(&quota).Error
	return &quota, err
}

// CreateOrUpdateQuota creates or updates a quota
func CreateOrUpdateQuota(db *gorm.DB, entityType string, entityID uint, resourceType string, limit int64, softLimit int64, resetPeriod string) (*Quota, error) {
	quota, err := GetQuota(db, entityType, entityID, resourceType)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new quota
			quota = &Quota{
				EntityType:   entityType,
				EntityID:     entityID,
				ResourceType: resourceType,
				Limit:        limit,
				SoftLimit:    softLimit,
				ResetPeriod:  resetPeriod,
				IsActive:     true,
			}
			return quota, db.Create(quota).Error
		}
		return nil, err
	}

	// Update existing quota
	quota.Limit = limit
	quota.SoftLimit = softLimit
	quota.ResetPeriod = resetPeriod
	return quota, db.Save(quota).Error
}

// CheckAndEnforceQuota checks if a quota would be exceeded and returns an error if so
func CheckAndEnforceQuota(db *gorm.DB, entityType string, entityID uint, resourceType string, additionalUsage int64) error {
	quota, err := GetQuota(db, entityType, entityID, resourceType)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// No quota defined, allow operation
			return nil
		}
		return err
	}

	if !quota.HasSpace(additionalUsage) {
		return errors.New("quota limit exceeded")
	}

	return nil
}
