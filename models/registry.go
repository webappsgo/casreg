package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// Registry represents a Docker registry
type Registry struct {
	Base
	Name           string `gorm:"uniqueIndex;not null" json:"name"`
	DisplayName    string `json:"display_name"`
	Description    string `gorm:"type:text" json:"description"`
	Visibility     string `gorm:"not null;default:'private'" json:"visibility"`
	IsPublic       bool   `gorm:"not null;default:false" json:"is_public"`
	OwnerType      string `gorm:"not null;default:'user'" json:"owner_type"` // user or organization
	OwnerID        uint   `gorm:"not null;index" json:"owner_id"`
	EnableScanning bool   `gorm:"not null;default:true" json:"enable_scanning"`
	EnableSigning  bool   `gorm:"not null;default:false" json:"enable_signing"`
	QuotaLimit     int64  `gorm:"default:0" json:"quota_limit"` // 0 = unlimited
	QuotaUsed      int64  `gorm:"default:0" json:"quota_used"`

	// Relationships
	Repositories []Repository `gorm:"foreignKey:RegistryID;constraint:OnDelete:CASCADE" json:"-"`
}

// Repository represents a repository within a registry
type Repository struct {
	Base
	RegistryID      uint       `gorm:"not null;index:idx_registry_repo,unique" json:"registry_id"`
	Name            string     `gorm:"not null;index:idx_registry_repo,unique" json:"name"`
	Description     string     `gorm:"type:text" json:"description"`
	Visibility      string     `gorm:"not null;default:'inherit'" json:"visibility"`
	IsPublic        bool       `gorm:"not null;default:false" json:"is_public"`
	EnableScanning  bool       `gorm:"not null;default:true" json:"enable_scanning"`
	EnableSigning   bool       `gorm:"not null;default:false" json:"enable_signing"`
	LastPushed      *time.Time `json:"last_pushed,omitempty"`
	PullCount       int64      `gorm:"default:0" json:"pull_count"`
	PushCount       int64      `gorm:"default:0" json:"push_count"`
	Size            int64      `gorm:"default:0" json:"size"`
	TagCount        int        `gorm:"default:0" json:"tag_count"`
	README          string     `gorm:"type:text" json:"readme,omitempty"`

	// Relationships
	Registry  Registry   `gorm:"foreignKey:RegistryID" json:"-"`
	Tags      []Tag      `gorm:"foreignKey:RepositoryID;constraint:OnDelete:CASCADE" json:"-"`
	Manifests []Manifest `gorm:"foreignKey:RepositoryID;constraint:OnDelete:CASCADE" json:"-"`
	Issues    []Issue    `gorm:"foreignKey:RepositoryID;constraint:OnDelete:CASCADE" json:"-"`
}

// Tag represents an image tag
type Tag struct {
	Base
	RepositoryID uint       `gorm:"not null;index:idx_repo_tag,unique" json:"repository_id"`
	Name         string     `gorm:"not null;index:idx_repo_tag,unique" json:"name"`
	ManifestID   uint       `gorm:"not null;index" json:"manifest_id"`
	IsProtected  bool       `gorm:"not null;default:false" json:"is_protected"`
	LastPulled   *time.Time `json:"last_pulled,omitempty"`
	PullCount    int64      `gorm:"default:0" json:"pull_count"`
	Size         int64      `gorm:"default:0" json:"size"`

	// Relationships
	Repository Repository          `gorm:"foreignKey:RepositoryID" json:"-"`
	Manifest   Manifest            `gorm:"foreignKey:ManifestID" json:"-"`
	ScanResult *ScanResult         `gorm:"foreignKey:TagID" json:"-"`
	Signature  *SignatureVerification `gorm:"foreignKey:TagID" json:"-"`
}

// Manifest represents an OCI/Docker manifest
type Manifest struct {
	Base
	RepositoryID   uint   `gorm:"not null;index:idx_repo_digest,unique" json:"repository_id"`
	Digest         string `gorm:"not null;index:idx_repo_digest,unique" json:"digest"`
	SchemaVersion  int    `gorm:"not null" json:"schema_version"`
	MediaType      string `gorm:"not null" json:"media_type"`
	ConfigDigest   string `json:"config_digest"`
	ConfigMediaType string `json:"config_media_type"`
	ConfigSize     int64  `json:"config_size"`
	TotalSize      int64  `json:"total_size"`
	Architecture   string `json:"architecture"`
	OS             string `json:"os"`
	Variant        string `json:"variant,omitempty"`
	ManifestData   []byte `gorm:"type:bytea" json:"-"`

	// Relationships
	Repository Repository `gorm:"foreignKey:RepositoryID" json:"-"`
	Layers     []Layer    `gorm:"foreignKey:ManifestID;constraint:OnDelete:CASCADE" json:"-"`
	Blobs      []Blob     `gorm:"many2many:manifest_blobs;" json:"-"`
}

// Layer represents a layer in an image manifest
type Layer struct {
	Base
	ManifestID uint   `gorm:"not null;index" json:"manifest_id"`
	BlobID     uint   `gorm:"not null;index" json:"blob_id"`
	MediaType  string `gorm:"not null" json:"media_type"`
	Size       int64  `gorm:"not null" json:"size"`
	Digest     string `gorm:"not null;index" json:"digest"`
	LayerIndex int    `gorm:"not null" json:"layer_index"`

	// Relationships
	Manifest Manifest `gorm:"foreignKey:ManifestID" json:"-"`
	Blob     Blob     `gorm:"foreignKey:BlobID" json:"-"`
}

// Blob represents a content-addressable blob
type Blob struct {
	Base
	Digest        string `gorm:"uniqueIndex;not null" json:"digest"`
	MediaType     string `gorm:"not null" json:"media_type"`
	Size          int64  `gorm:"not null" json:"size"`
	StoragePath   string `gorm:"not null" json:"storage_path"`
	CompressionType string `json:"compression_type,omitempty"`
	RefCount      int    `gorm:"default:1" json:"ref_count"`

	// Relationships
	Manifests []Manifest `gorm:"many2many:manifest_blobs;" json:"-"`
}

// BeforeCreate validates registry data before creation
func (r *Registry) BeforeCreate(tx *gorm.DB) error {
	if r.Name == "" {
		return errors.New("registry name is required")
	}

	// Validate owner type
	if r.OwnerType != "user" && r.OwnerType != "organization" {
		return errors.New("owner type must be 'user' or 'organization'")
	}

	// Validate visibility
	validVisibility := map[string]bool{
		VisibilityPublic:   true,
		VisibilityPrivate:  true,
		VisibilityInternal: true,
	}

	if !validVisibility[r.Visibility] {
		r.Visibility = VisibilityPrivate
	}

	return nil
}

// BeforeCreate validates repository data before creation
func (r *Repository) BeforeCreate(tx *gorm.DB) error {
	if r.Name == "" {
		return errors.New("repository name is required")
	}
	return nil
}

// HasQuotaSpace returns true if the registry has quota space available
func (r *Registry) HasQuotaSpace(additionalSize int64) bool {
	if r.QuotaLimit == 0 {
		return true // unlimited
	}
	return r.QuotaUsed+additionalSize <= r.QuotaLimit
}

// UpdateQuotaUsage updates the registry's quota usage
func (r *Registry) UpdateQuotaUsage(db *gorm.DB, delta int64) error {
	r.QuotaUsed += delta
	if r.QuotaUsed < 0 {
		r.QuotaUsed = 0
	}
	return db.Save(r).Error
}

// UpdateStats updates repository statistics
func (r *Repository) UpdateStats(db *gorm.DB) error {
	// Update tag count
	var tagCount int64
	if err := db.Model(&Tag{}).Where("repository_id = ?", r.ID).Count(&tagCount).Error; err != nil {
		return err
	}
	r.TagCount = int(tagCount)

	// Update size
	var totalSize int64
	db.Model(&Tag{}).Where("repository_id = ?", r.ID).Select("COALESCE(SUM(size), 0)").Scan(&totalSize)
	r.Size = totalSize

	return db.Save(r).Error
}

// IncrementPullCount increments the repository pull count
func (r *Repository) IncrementPullCount(db *gorm.DB) error {
	return db.Model(r).UpdateColumn("pull_count", gorm.Expr("pull_count + ?", 1)).Error
}

// IncrementPushCount increments the repository push count and updates last pushed time
func (r *Repository) IncrementPushCount(db *gorm.DB) error {
	now := time.Now()
	r.LastPushed = &now
	r.PushCount++
	return db.Save(r).Error
}

// IncrementPullCount increments the tag pull count
func (t *Tag) IncrementPullCount(db *gorm.DB) error {
	now := time.Now()
	t.LastPulled = &now
	t.PullCount++
	return db.Save(t).Error
}

// IncrementRefCount increments the blob reference count
func (b *Blob) IncrementRefCount(db *gorm.DB) error {
	return db.Model(b).UpdateColumn("ref_count", gorm.Expr("ref_count + ?", 1)).Error
}

// DecrementRefCount decrements the blob reference count
func (b *Blob) DecrementRefCount(db *gorm.DB) error {
	return db.Model(b).UpdateColumn("ref_count", gorm.Expr("ref_count - ?", 1)).Error
}

// IsOrphaned returns true if the blob has no references
func (b *Blob) IsOrphaned() bool {
	return b.RefCount <= 0
}
