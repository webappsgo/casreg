package model

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// Ticket represents a support ticket
type Ticket struct {
	Base
	UserID      uint       `gorm:"not null;index" json:"user_id"`
	Title       string     `gorm:"not null" json:"title"`
	Description string     `gorm:"type:text;not null" json:"description"`
	Status      string     `gorm:"not null;default:'open'" json:"status"`
	Priority    string     `gorm:"not null;default:'medium'" json:"priority"`
	Category    string     `json:"category,omitempty"`
	AssignedTo  *uint      `gorm:"index" json:"assigned_to,omitempty"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
	ClosedAt    *time.Time `json:"closed_at,omitempty"`

	// Relationships
	User     User            `gorm:"foreignKey:UserID" json:"-"`
	Assignee *User           `gorm:"foreignKey:AssignedTo" json:"-"`
	Comments []TicketComment `gorm:"foreignKey:TicketID;constraint:OnDelete:CASCADE" json:"-"`
}

// TicketComment represents a comment on a ticket
type TicketComment struct {
	Base
	TicketID   uint   `gorm:"not null;index" json:"ticket_id"`
	UserID     uint   `gorm:"not null;index" json:"user_id"`
	Comment    string `gorm:"type:text;not null" json:"comment"`
	IsInternal bool   `gorm:"not null;default:false" json:"is_internal"` // staff-only comments
	IsResolution bool `gorm:"not null;default:false" json:"is_resolution"` // marks the resolving comment

	// Relationships
	Ticket Ticket `gorm:"foreignKey:TicketID" json:"-"`
	User   User   `gorm:"foreignKey:UserID" json:"-"`
}

// Ticket categories
const (
	TicketCategoryTechnical = "technical-issue"
	TicketCategoryFeature   = "feature-request"
	TicketCategoryAccount   = "account-management"
	TicketCategoryBilling   = "billing-inquiry"
	TicketCategorySecurity  = "security-concern"
	TicketCategoryGeneral   = "general-question"
)

// BeforeCreate validates ticket data before creation
func (t *Ticket) BeforeCreate(tx *gorm.DB) error {
	if t.Title == "" {
		return errors.New("ticket title is required")
	}

	if t.Description == "" {
		return errors.New("ticket description is required")
	}

	// Validate status
	validStatuses := map[string]bool{
		StatusOpen:       true,
		StatusInProgress: true,
		StatusResolved:   true,
		StatusClosed:     true,
	}

	if !validStatuses[t.Status] {
		t.Status = StatusOpen
	}

	// Validate priority
	validPriorities := map[string]bool{
		PriorityLow:      true,
		PriorityMedium:   true,
		PriorityHigh:     true,
		PriorityCritical: true,
	}

	if !validPriorities[t.Priority] {
		t.Priority = PriorityMedium
	}

	// Validate category
	validCategories := map[string]bool{
		TicketCategoryTechnical: true,
		TicketCategoryFeature:   true,
		TicketCategoryAccount:   true,
		TicketCategoryBilling:   true,
		TicketCategorySecurity:  true,
		TicketCategoryGeneral:   true,
	}

	if t.Category != "" && !validCategories[t.Category] {
		t.Category = TicketCategoryGeneral
	}

	return nil
}

// BeforeCreate validates comment data before creation
func (tc *TicketComment) BeforeCreate(tx *gorm.DB) error {
	if tc.Comment == "" {
		return errors.New("comment cannot be empty")
	}
	return nil
}

// UpdateStatus updates the ticket status and related timestamps
func (t *Ticket) UpdateStatus(db *gorm.DB, newStatus string) error {
	validStatuses := map[string]bool{
		StatusOpen:       true,
		StatusInProgress: true,
		StatusResolved:   true,
		StatusClosed:     true,
	}

	if !validStatuses[newStatus] {
		return errors.New("invalid status")
	}

	t.Status = newStatus

	now := time.Now()
	if newStatus == StatusResolved && t.ResolvedAt == nil {
		t.ResolvedAt = &now
	}

	if newStatus == StatusClosed && t.ClosedAt == nil {
		t.ClosedAt = &now
	}

	return db.Save(t).Error
}

// Assign assigns the ticket to a user
func (t *Ticket) Assign(db *gorm.DB, userID uint) error {
	// Verify user exists
	var user User
	if err := db.First(&user, userID).Error; err != nil {
		return errors.New("user not found")
	}

	t.AssignedTo = &userID

	// Update status to in-progress if currently open
	if t.Status == StatusOpen {
		t.Status = StatusInProgress
	}

	return db.Save(t).Error
}

// Unassign removes the assignment from the ticket
func (t *Ticket) Unassign(db *gorm.DB) error {
	t.AssignedTo = nil
	return db.Save(t).Error
}

// AddComment adds a comment to the ticket
func (t *Ticket) AddComment(db *gorm.DB, userID uint, comment string, isInternal bool) (*TicketComment, error) {
	tc := &TicketComment{
		TicketID:   t.ID,
		UserID:     userID,
		Comment:    comment,
		IsInternal: isInternal,
	}

	if err := db.Create(tc).Error; err != nil {
		return nil, err
	}

	return tc, nil
}

// Resolve marks the ticket as resolved with a resolution comment
func (t *Ticket) Resolve(db *gorm.DB, userID uint, resolutionComment string) error {
	// Add resolution comment
	tc := &TicketComment{
		TicketID:     t.ID,
		UserID:       userID,
		Comment:      resolutionComment,
		IsInternal:   false,
		IsResolution: true,
	}

	if err := db.Create(tc).Error; err != nil {
		return err
	}

	// Update ticket status
	return t.UpdateStatus(db, StatusResolved)
}

// Close marks the ticket as closed
func (t *Ticket) Close(db *gorm.DB) error {
	return t.UpdateStatus(db, StatusClosed)
}

// Reopen reopens a closed or resolved ticket
func (t *Ticket) Reopen(db *gorm.DB) error {
	t.ResolvedAt = nil
	t.ClosedAt = nil
	return t.UpdateStatus(db, StatusOpen)
}

// GetComments retrieves all comments for the ticket
func (t *Ticket) GetComments(db *gorm.DB, includeInternal bool) ([]TicketComment, error) {
	var comments []TicketComment
	query := db.Where("ticket_id = ?", t.ID).Order("created_at ASC")

	if !includeInternal {
		query = query.Where("is_internal = ?", false)
	}

	err := query.Find(&comments).Error
	return comments, err
}

// IsOpen returns true if the ticket is open or in progress
func (t *Ticket) IsOpen() bool {
	return t.Status == StatusOpen || t.Status == StatusInProgress
}

// IsClosed returns true if the ticket is closed or resolved
func (t *Ticket) IsClosed() bool {
	return t.Status == StatusClosed || t.Status == StatusResolved
}
