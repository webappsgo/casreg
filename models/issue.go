package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// Issue represents a repository issue
type Issue struct {
	Base
	RepositoryID uint       `gorm:"not null;index" json:"repository_id"`
	Number       int        `gorm:"not null;index:idx_repo_issue,unique" json:"number"`
	Title        string     `gorm:"not null" json:"title"`
	Description  string     `gorm:"type:text" json:"description"`
	State        string     `gorm:"not null;default:'open';index" json:"state"`
	AuthorID     uint       `gorm:"not null;index" json:"author_id"`
	AssignedTo   *uint      `gorm:"index" json:"assigned_to,omitempty"`
	ClosedAt     *time.Time `json:"closed_at,omitempty"`
	ClosedBy     *uint      `json:"closed_by,omitempty"`
	DueDate      *time.Time `json:"due_date,omitempty"`
	Milestone    string     `json:"milestone,omitempty"`

	// Relationships
	Repository Repository     `gorm:"foreignKey:RepositoryID" json:"-"`
	Author     User           `gorm:"foreignKey:AuthorID" json:"-"`
	Assignee   *User          `gorm:"foreignKey:AssignedTo" json:"-"`
	Closer     *User          `gorm:"foreignKey:ClosedBy" json:"-"`
	Comments   []IssueComment `gorm:"foreignKey:IssueID;constraint:OnDelete:CASCADE" json:"-"`
	Labels     []IssueLabel   `gorm:"many2many:issue_label_relations;" json:"-"`
}

// IssueComment represents a comment on an issue
type IssueComment struct {
	Base
	IssueID  uint   `gorm:"not null;index" json:"issue_id"`
	UserID   uint   `gorm:"not null;index" json:"user_id"`
	Comment  string `gorm:"type:text;not null" json:"comment"`
	EditedAt *time.Time `json:"edited_at,omitempty"`

	// Relationships
	Issue Issue `gorm:"foreignKey:IssueID" json:"-"`
	User  User  `gorm:"foreignKey:UserID" json:"-"`
}

// IssueLabel represents a label that can be applied to issues
type IssueLabel struct {
	Base
	RepositoryID uint   `gorm:"not null;index:idx_repo_label,unique" json:"repository_id"`
	Name         string `gorm:"not null;index:idx_repo_label,unique" json:"name"`
	Color        string `gorm:"not null" json:"color"`
	Description  string `gorm:"type:text" json:"description,omitempty"`

	// Relationships
	Repository Repository `gorm:"foreignKey:RepositoryID" json:"-"`
	Issues     []Issue    `gorm:"many2many:issue_label_relations;" json:"-"`
}

// Issue states
const (
	IssueStateOpen       = "open"
	IssueStateInProgress = "in-progress"
	IssueStateResolved   = "resolved"
	IssueStateClosed     = "closed"
)

// Default issue labels
const (
	LabelBug             = "bug"
	LabelEnhancement     = "enhancement"
	LabelDocumentation   = "documentation"
	LabelQuestion        = "question"
	LabelHelpWanted      = "help-wanted"
	LabelGoodFirstIssue  = "good-first-issue"
)

// BeforeCreate validates issue data and assigns issue number
func (i *Issue) BeforeCreate(tx *gorm.DB) error {
	if i.Title == "" {
		return errors.New("issue title is required")
	}

	// Validate state
	validStates := map[string]bool{
		IssueStateOpen:       true,
		IssueStateInProgress: true,
		IssueStateResolved:   true,
		IssueStateClosed:     true,
	}

	if !validStates[i.State] {
		i.State = IssueStateOpen
	}

	// Assign issue number (next sequential number for the repository)
	if i.Number == 0 {
		var maxNumber int
		tx.Model(&Issue{}).
			Where("repository_id = ?", i.RepositoryID).
			Select("COALESCE(MAX(number), 0)").
			Scan(&maxNumber)
		i.Number = maxNumber + 1
	}

	return nil
}

// BeforeCreate validates comment data before creation
func (ic *IssueComment) BeforeCreate(tx *gorm.DB) error {
	if ic.Comment == "" {
		return errors.New("comment cannot be empty")
	}
	return nil
}

// BeforeCreate validates label data before creation
func (il *IssueLabel) BeforeCreate(tx *gorm.DB) error {
	if il.Name == "" {
		return errors.New("label name is required")
	}

	if il.Color == "" {
		il.Color = "#6200ea" // default purple color
	}

	return nil
}

// Close closes the issue
func (i *Issue) Close(db *gorm.DB, closerID uint) error {
	now := time.Now()
	i.State = IssueStateClosed
	i.ClosedAt = &now
	i.ClosedBy = &closerID
	return db.Save(i).Error
}

// Reopen reopens a closed issue
func (i *Issue) Reopen(db *gorm.DB) error {
	i.State = IssueStateOpen
	i.ClosedAt = nil
	i.ClosedBy = nil
	return db.Save(i).Error
}

// Assign assigns the issue to a user
func (i *Issue) Assign(db *gorm.DB, userID uint) error {
	// Verify user exists
	var user User
	if err := db.First(&user, userID).Error; err != nil {
		return errors.New("user not found")
	}

	i.AssignedTo = &userID

	// Update state to in-progress if currently open
	if i.State == IssueStateOpen {
		i.State = IssueStateInProgress
	}

	return db.Save(i).Error
}

// Unassign removes the assignment from the issue
func (i *Issue) Unassign(db *gorm.DB) error {
	i.AssignedTo = nil
	return db.Save(i).Error
}

// AddComment adds a comment to the issue
func (i *Issue) AddComment(db *gorm.DB, userID uint, comment string) (*IssueComment, error) {
	ic := &IssueComment{
		IssueID: i.ID,
		UserID:  userID,
		Comment: comment,
	}

	if err := db.Create(ic).Error; err != nil {
		return nil, err
	}

	return ic, nil
}

// UpdateComment updates a comment
func (ic *IssueComment) UpdateComment(db *gorm.DB, newComment string) error {
	ic.Comment = newComment
	now := time.Now()
	ic.EditedAt = &now
	return db.Save(ic).Error
}

// AddLabel adds a label to the issue
func (i *Issue) AddLabel(db *gorm.DB, labelID uint) error {
	var label IssueLabel
	if err := db.First(&label, labelID).Error; err != nil {
		return errors.New("label not found")
	}

	return db.Model(i).Association("Labels").Append(&label)
}

// RemoveLabel removes a label from the issue
func (i *Issue) RemoveLabel(db *gorm.DB, labelID uint) error {
	var label IssueLabel
	if err := db.First(&label, labelID).Error; err != nil {
		return errors.New("label not found")
	}

	return db.Model(i).Association("Labels").Delete(&label)
}

// HasLabel checks if the issue has a specific label
func (i *Issue) HasLabel(db *gorm.DB, labelName string) bool {
	var count int64
	db.Model(i).
		Joins("JOIN issue_label_relations ON issue_label_relations.issue_id = issues.id").
		Joins("JOIN issue_labels ON issue_labels.id = issue_label_relations.issue_label_id").
		Where("issue_labels.name = ?", labelName).
		Count(&count)
	return count > 0
}

// GetComments retrieves all comments for the issue
func (i *Issue) GetComments(db *gorm.DB) ([]IssueComment, error) {
	var comments []IssueComment
	err := db.Where("issue_id = ?", i.ID).Order("created_at ASC").Find(&comments).Error
	return comments, err
}

// IsOpen returns true if the issue is open or in progress
func (i *Issue) IsOpen() bool {
	return i.State == IssueStateOpen || i.State == IssueStateInProgress
}

// IsClosed returns true if the issue is closed or resolved
func (i *Issue) IsClosed() bool {
	return i.State == IssueStateClosed || i.State == IssueStateResolved
}

// CreateDefaultLabels creates default labels for a repository
func CreateDefaultLabels(db *gorm.DB, repositoryID uint) error {
	defaultLabels := []IssueLabel{
		{RepositoryID: repositoryID, Name: LabelBug, Color: "#ff5555", Description: "Something isn't working"},
		{RepositoryID: repositoryID, Name: LabelEnhancement, Color: "#50fa7b", Description: "New feature or request"},
		{RepositoryID: repositoryID, Name: LabelDocumentation, Color: "#8be9fd", Description: "Improvements or additions to documentation"},
		{RepositoryID: repositoryID, Name: LabelQuestion, Color: "#ffb86c", Description: "Further information is requested"},
		{RepositoryID: repositoryID, Name: LabelHelpWanted, Color: "#bd93f9", Description: "Extra attention is needed"},
		{RepositoryID: repositoryID, Name: LabelGoodFirstIssue, Color: "#ff79c6", Description: "Good for newcomers"},
	}

	for _, label := range defaultLabels {
		if err := db.Create(&label).Error; err != nil {
			return err
		}
	}

	return nil
}
