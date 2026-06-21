package model

import (
	"errors"

	"gorm.io/gorm"
)

// Organization represents a group of users with shared registries
type Organization struct {
	Base
	Name        string `gorm:"uniqueIndex;not null" json:"name"`
	DisplayName string `json:"display_name"`
	Description string `gorm:"type:text" json:"description"`
	Visibility  string `gorm:"not null;default:'private'" json:"visibility"`
	IsPublic    bool   `gorm:"not null;default:false" json:"is_public"`
	QuotaLimit  int64  `gorm:"default:0" json:"quota_limit"` // 0 = unlimited
	QuotaUsed   int64  `gorm:"default:0" json:"quota_used"`

	// Relationships
	Members    []OrganizationMember `gorm:"foreignKey:OrganizationID" json:"-"`
	Registries []Registry           `gorm:"foreignKey:OwnerID;constraint:OnDelete:CASCADE" json:"-"`
}

// OrganizationMember represents a user's membership in an organization
type OrganizationMember struct {
	Base
	OrganizationID uint   `gorm:"not null;index:idx_org_user,unique" json:"organization_id"`
	UserID         uint   `gorm:"not null;index:idx_org_user,unique" json:"user_id"`
	Role           string `gorm:"not null;default:'member'" json:"role"`

	// Relationships
	Organization Organization `gorm:"foreignKey:OrganizationID" json:"-"`
	User         User         `gorm:"foreignKey:UserID" json:"-"`
}

// BeforeCreate validates organization data before creation
func (o *Organization) BeforeCreate(tx *gorm.DB) error {
	if o.Name == "" {
		return errors.New("organization name is required")
	}

	// Validate visibility
	validVisibility := map[string]bool{
		VisibilityPublic:   true,
		VisibilityPrivate:  true,
		VisibilityInternal: true,
		VisibilityHidden:   true,
	}

	if !validVisibility[o.Visibility] {
		o.Visibility = VisibilityPrivate
	}

	return nil
}

// HasQuotaSpace returns true if the organization has quota space available
func (o *Organization) HasQuotaSpace(additionalSize int64) bool {
	if o.QuotaLimit == 0 {
		return true // unlimited
	}
	return o.QuotaUsed+additionalSize <= o.QuotaLimit
}

// UpdateQuotaUsage updates the organization's quota usage
func (o *Organization) UpdateQuotaUsage(db *gorm.DB, delta int64) error {
	o.QuotaUsed += delta
	if o.QuotaUsed < 0 {
		o.QuotaUsed = 0
	}
	return db.Save(o).Error
}

// GetMember retrieves a specific member by user ID
func (o *Organization) GetMember(db *gorm.DB, userID uint) (*OrganizationMember, error) {
	var member OrganizationMember
	err := db.Where("organization_id = ? AND user_id = ?", o.ID, userID).First(&member).Error
	return &member, err
}

// IsMember checks if a user is a member of the organization
func (o *Organization) IsMember(db *gorm.DB, userID uint) bool {
	var count int64
	db.Model(&OrganizationMember{}).
		Where("organization_id = ? AND user_id = ?", o.ID, userID).
		Count(&count)
	return count > 0
}

// IsOwner checks if a user is an owner of the organization
func (o *Organization) IsOwner(db *gorm.DB, userID uint) bool {
	var count int64
	db.Model(&OrganizationMember{}).
		Where("organization_id = ? AND user_id = ? AND role = ?", o.ID, userID, OrgRoleOwner).
		Count(&count)
	return count > 0
}

// IsAdmin checks if a user is an admin or owner of the organization
func (o *Organization) IsAdmin(db *gorm.DB, userID uint) bool {
	var count int64
	db.Model(&OrganizationMember{}).
		Where("organization_id = ? AND user_id = ? AND role IN ?", o.ID, userID, []string{OrgRoleOwner, OrgRoleAdmin}).
		Count(&count)
	return count > 0
}

// AddMember adds a user to the organization with the specified role
func (o *Organization) AddMember(db *gorm.DB, userID uint, role string) error {
	// Validate role
	validRoles := map[string]bool{
		OrgRoleOwner:  true,
		OrgRoleAdmin:  true,
		OrgRoleMember: true,
	}

	if !validRoles[role] {
		role = OrgRoleMember
	}

	member := &OrganizationMember{
		OrganizationID: o.ID,
		UserID:         userID,
		Role:           role,
	}

	return db.Create(member).Error
}

// RemoveMember removes a user from the organization
func (o *Organization) RemoveMember(db *gorm.DB, userID uint) error {
	// Prevent removing the last owner
	if o.IsOwner(db, userID) {
		var ownerCount int64
		db.Model(&OrganizationMember{}).
			Where("organization_id = ? AND role = ?", o.ID, OrgRoleOwner).
			Count(&ownerCount)

		if ownerCount <= 1 {
			return errors.New("cannot remove the last owner from the organization")
		}
	}

	return db.Where("organization_id = ? AND user_id = ?", o.ID, userID).
		Delete(&OrganizationMember{}).Error
}

// UpdateMemberRole updates a member's role in the organization
func (o *Organization) UpdateMemberRole(db *gorm.DB, userID uint, newRole string) error {
	// Validate role
	validRoles := map[string]bool{
		OrgRoleOwner:  true,
		OrgRoleAdmin:  true,
		OrgRoleMember: true,
	}

	if !validRoles[newRole] {
		return errors.New("invalid role")
	}

	// Prevent demoting the last owner
	if newRole != OrgRoleOwner {
		member, err := o.GetMember(db, userID)
		if err != nil {
			return err
		}

		if member.Role == OrgRoleOwner {
			var ownerCount int64
			db.Model(&OrganizationMember{}).
				Where("organization_id = ? AND role = ?", o.ID, OrgRoleOwner).
				Count(&ownerCount)

			if ownerCount <= 1 {
				return errors.New("cannot demote the last owner")
			}
		}
	}

	return db.Model(&OrganizationMember{}).
		Where("organization_id = ? AND user_id = ?", o.ID, userID).
		Update("role", newRole).Error
}
