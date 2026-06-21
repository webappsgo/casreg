package model

import (
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// ScanResult represents a vulnerability scan result
type ScanResult struct {
	Base
	TagID           uint       `gorm:"uniqueIndex;not null" json:"tag_id"`
	ScannerName     string     `gorm:"not null;default:'trivy'" json:"scanner_name"`
	ScannerVersion  string     `json:"scanner_version"`
	Status          string     `gorm:"not null;default:'pending';index" json:"status"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	CriticalCount   int        `gorm:"default:0;index" json:"critical_count"`
	HighCount       int        `gorm:"default:0;index" json:"high_count"`
	MediumCount     int        `gorm:"default:0;index" json:"medium_count"`
	LowCount        int        `gorm:"default:0;index" json:"low_count"`
	UnknownCount    int        `gorm:"default:0" json:"unknown_count"`
	TotalCount      int        `gorm:"default:0" json:"total_count"`
	FixableCount    int        `gorm:"default:0" json:"fixable_count"`
	RawResult       string     `gorm:"type:jsonb" json:"-"` // Full JSON scan result
	ErrorMessage    string     `gorm:"type:text" json:"error_message,omitempty"`
	DatabaseVersion string     `json:"database_version,omitempty"`

	// Relationships
	Tag             Tag               `gorm:"foreignKey:TagID" json:"-"`
	Vulnerabilities []Vulnerability   `gorm:"foreignKey:ScanResultID;constraint:OnDelete:CASCADE" json:"-"`
}

// Vulnerability represents a specific vulnerability found in a scan
type Vulnerability struct {
	Base
	ScanResultID    uint   `gorm:"not null;index" json:"scan_result_id"`
	VulnerabilityID string `gorm:"not null;index" json:"vulnerability_id"` // CVE ID
	PackageName     string `gorm:"not null;index" json:"package_name"`
	InstalledVersion string `gorm:"not null" json:"installed_version"`
	FixedVersion    string `json:"fixed_version,omitempty"`
	Severity        string `gorm:"not null;index" json:"severity"`
	Title           string `json:"title"`
	Description     string `gorm:"type:text" json:"description"`
	CVSS            float64 `json:"cvss,omitempty"`
	CVSSVector      string `json:"cvss_vector,omitempty"`
	References      string `gorm:"type:jsonb" json:"references,omitempty"` // JSON array of URLs
	PublishedDate   *time.Time `json:"published_date,omitempty"`
	LastModified    *time.Time `json:"last_modified,omitempty"`

	// Relationships
	ScanResult ScanResult `gorm:"foreignKey:ScanResultID" json:"-"`
}

// SignatureVerification represents a signature verification result
type SignatureVerification struct {
	Base
	TagID          uint       `gorm:"uniqueIndex;not null" json:"tag_id"`
	SignatureType  string     `gorm:"not null;default:'cosign'" json:"signature_type"`
	Status         string     `gorm:"not null;default:'pending';index" json:"status"`
	VerifiedAt     *time.Time `json:"verified_at,omitempty"`
	SignerIdentity string     `json:"signer_identity,omitempty"`
	SignerIssuer   string     `json:"signer_issuer,omitempty"`
	SignatureDigest string    `json:"signature_digest,omitempty"`
	PublicKey      string     `gorm:"type:text" json:"public_key,omitempty"`
	Certificate    string     `gorm:"type:text" json:"certificate,omitempty"`
	CertChain      string     `gorm:"type:jsonb" json:"cert_chain,omitempty"`
	RekorEntry     string     `gorm:"type:jsonb" json:"rekor_entry,omitempty"` // Transparency log entry
	ErrorMessage   string     `gorm:"type:text" json:"error_message,omitempty"`

	// Relationships
	Tag Tag `gorm:"foreignKey:TagID" json:"-"`
}

// Scan statuses
const (
	ScanStatusPending   = "pending"
	ScanStatusRunning   = "running"
	ScanStatusCompleted = "completed"
	ScanStatusFailed    = "failed"
)

// Signature verification statuses
const (
	SignatureStatusPending  = "pending"
	SignatureStatusVerified = "verified"
	SignatureStatusFailed   = "failed"
	SignatureStatusNotSigned = "not_signed"
)

// Vulnerability severities
const (
	SeverityCritical = "CRITICAL"
	SeverityHigh     = "HIGH"
	SeverityMedium   = "MEDIUM"
	SeverityLow      = "LOW"
	SeverityUnknown  = "UNKNOWN"
)

// BeforeCreate validates scan result data before creation
func (s *ScanResult) BeforeCreate(tx *gorm.DB) error {
	// Validate status
	validStatuses := map[string]bool{
		ScanStatusPending:   true,
		ScanStatusRunning:   true,
		ScanStatusCompleted: true,
		ScanStatusFailed:    true,
	}

	if !validStatuses[s.Status] {
		s.Status = ScanStatusPending
	}

	return nil
}

// BeforeCreate validates signature verification data before creation
func (sv *SignatureVerification) BeforeCreate(tx *gorm.DB) error {
	// Validate status
	validStatuses := map[string]bool{
		SignatureStatusPending:  true,
		SignatureStatusVerified: true,
		SignatureStatusFailed:   true,
		SignatureStatusNotSigned: true,
	}

	if !validStatuses[sv.Status] {
		sv.Status = SignatureStatusPending
	}

	return nil
}

// Start marks the scan as started
func (s *ScanResult) Start(db *gorm.DB) error {
	now := time.Now()
	s.Status = ScanStatusRunning
	s.StartedAt = &now
	return db.Save(s).Error
}

// Complete marks the scan as completed
func (s *ScanResult) Complete(db *gorm.DB) error {
	now := time.Now()
	s.Status = ScanStatusCompleted
	s.CompletedAt = &now
	s.TotalCount = s.CriticalCount + s.HighCount + s.MediumCount + s.LowCount + s.UnknownCount
	return db.Save(s).Error
}

// Fail marks the scan as failed
func (s *ScanResult) Fail(db *gorm.DB, errorMessage string) error {
	now := time.Now()
	s.Status = ScanStatusFailed
	s.CompletedAt = &now
	s.ErrorMessage = errorMessage
	return db.Save(s).Error
}

// SetRawResult sets the raw scan result from a map
func (s *ScanResult) SetRawResult(data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	s.RawResult = string(jsonData)
	return nil
}

// GetRawResult retrieves the raw scan result as a map
func (s *ScanResult) GetRawResult() (map[string]interface{}, error) {
	if s.RawResult == "" {
		return make(map[string]interface{}), nil
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(s.RawResult), &data); err != nil {
		return nil, err
	}

	return data, nil
}

// HasCriticalVulnerabilities checks if the scan found critical vulnerabilities
func (s *ScanResult) HasCriticalVulnerabilities() bool {
	return s.CriticalCount > 0
}

// HasHighVulnerabilities checks if the scan found high severity vulnerabilities
func (s *ScanResult) HasHighVulnerabilities() bool {
	return s.HighCount > 0
}

// GetHighestSeverity returns the highest severity found in the scan
func (s *ScanResult) GetHighestSeverity() string {
	if s.CriticalCount > 0 {
		return SeverityCritical
	}
	if s.HighCount > 0 {
		return SeverityHigh
	}
	if s.MediumCount > 0 {
		return SeverityMedium
	}
	if s.LowCount > 0 {
		return SeverityLow
	}
	return SeverityUnknown
}

// IsClean returns true if no vulnerabilities were found
func (s *ScanResult) IsClean() bool {
	return s.TotalCount == 0
}

// IsPassed returns true if the scan passed (no critical or high vulnerabilities)
func (s *ScanResult) IsPassed() bool {
	return s.CriticalCount == 0 && s.HighCount == 0
}

// Verify marks the signature as verified
func (sv *SignatureVerification) Verify(db *gorm.DB, signerIdentity, signerIssuer, signatureDigest string) error {
	now := time.Now()
	sv.Status = SignatureStatusVerified
	sv.VerifiedAt = &now
	sv.SignerIdentity = signerIdentity
	sv.SignerIssuer = signerIssuer
	sv.SignatureDigest = signatureDigest
	return db.Save(sv).Error
}

// FailVerification marks the signature verification as failed
func (sv *SignatureVerification) FailVerification(db *gorm.DB, errorMessage string) error {
	sv.Status = SignatureStatusFailed
	sv.ErrorMessage = errorMessage
	return db.Save(sv).Error
}

// MarkAsNotSigned marks the image as not signed
func (sv *SignatureVerification) MarkAsNotSigned(db *gorm.DB) error {
	sv.Status = SignatureStatusNotSigned
	return db.Save(sv).Error
}

// IsVerified returns true if the signature is verified
func (sv *SignatureVerification) IsVerified() bool {
	return sv.Status == SignatureStatusVerified
}

// AddVulnerability adds a vulnerability to the scan result
func (s *ScanResult) AddVulnerability(db *gorm.DB, vuln *Vulnerability) error {
	vuln.ScanResultID = s.ID

	// Validate severity
	validSeverities := map[string]bool{
		SeverityCritical: true,
		SeverityHigh:     true,
		SeverityMedium:   true,
		SeverityLow:      true,
		SeverityUnknown:  true,
	}

	if !validSeverities[vuln.Severity] {
		return errors.New("invalid severity")
	}

	if err := db.Create(vuln).Error; err != nil {
		return err
	}

	// Update counts
	switch vuln.Severity {
	case SeverityCritical:
		s.CriticalCount++
	case SeverityHigh:
		s.HighCount++
	case SeverityMedium:
		s.MediumCount++
	case SeverityLow:
		s.LowCount++
	default:
		s.UnknownCount++
	}

	if vuln.FixedVersion != "" {
		s.FixableCount++
	}

	return db.Save(s).Error
}

// GetVulnerabilities retrieves all vulnerabilities for the scan
func (s *ScanResult) GetVulnerabilities(db *gorm.DB) ([]Vulnerability, error) {
	var vulnerabilities []Vulnerability
	err := db.Where("scan_result_id = ?", s.ID).
		Order("severity DESC, vulnerability_id ASC").
		Find(&vulnerabilities).Error
	return vulnerabilities, err
}

// GetVulnerabilitiesBySeverity retrieves vulnerabilities filtered by severity
func (s *ScanResult) GetVulnerabilitiesBySeverity(db *gorm.DB, severity string) ([]Vulnerability, error) {
	var vulnerabilities []Vulnerability
	err := db.Where("scan_result_id = ? AND severity = ?", s.ID, severity).
		Order("vulnerability_id ASC").
		Find(&vulnerabilities).Error
	return vulnerabilities, err
}
