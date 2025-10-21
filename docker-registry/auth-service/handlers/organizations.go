package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"registry-auth/models"
)

// CreateOrganizationRequest represents the request body for creating an organization
type CreateOrganizationRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=50"`
	DisplayName string `json:"display_name" validate:"required,min=1,max=100"`
	Description string `json:"description,omitempty"`
	IsPublic    bool   `json:"is_public"`
}

// UpdateOrganizationRequest represents the request body for updating an organization
type UpdateOrganizationRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
	Description *string `json:"description,omitempty"`
	IsPublic    *bool   `json:"is_public,omitempty"`
}

// OrganizationResponse represents an organization with additional metadata
type OrganizationResponse struct {
	*models.Organization
	MemberCount    int    `json:"member_count"`
	RegistryCount  int    `json:"registry_count"`
	UserRole       string `json:"user_role,omitempty"`
	IsOwner        bool   `json:"is_owner"`
}

// AddMemberRequest represents the request body for adding a member to an organization
type AddMemberRequest struct {
	Username string `json:"username" validate:"required"`
	Role     string `json:"role" validate:"required,oneof=member admin owner"`
}

// UpdateMemberRequest represents the request body for updating a member's role
type UpdateMemberRequest struct {
	Role string `json:"role" validate:"required,oneof=member admin owner"`
}

// MemberResponse represents an organization member with user details
type MemberResponse struct {
	UserID    uint      `json:"user_id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Role      string    `json:"role"`
	JoinedAt  time.Time `json:"joined_at"`
}

// ListOrganizations returns organizations accessible to the user
func ListOrganizations(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := r.Context().Value("userID").(uint)

		// Get pagination parameters
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		if limit <= 0 || limit > 100 {
			limit = 20
		}

		visibility := r.URL.Query().Get("visibility") // public, private, all
		member := r.URL.Query().Get("member") == "true"

		var organizations []models.Organization
		var total int64

		query := db.Model(&models.Organization{})

		if member && userID > 0 {
			// Get organizations where user is a member
			query = query.Joins("JOIN organization_memberships ON organizations.id = organization_memberships.organization_id").
				Where("organization_memberships.user_id = ?", userID)
		} else {
			// Filter by visibility
			switch visibility {
			case "public":
				query = query.Where("is_public = ?", true)
			case "private":
				if userID > 0 {
					// Show private orgs where user is a member
					query = query.Joins("LEFT JOIN organization_memberships ON organizations.id = organization_memberships.organization_id").
						Where("organizations.is_public = ? OR organization_memberships.user_id = ?", false, userID)
				} else {
					query = query.Where("is_public = ?", false)
				}
			default: // "all" or empty
				if userID == 0 {
					// Guest users can only see public organizations
					query = query.Where("is_public = ?", true)
				}
			}
		}

		// Add search filter if provided
		if search := r.URL.Query().Get("search"); search != "" {
			query = query.Where("name ILIKE ? OR display_name ILIKE ? OR description ILIKE ?",
				"%"+search+"%", "%"+search+"%", "%"+search+"%")
		}

		// Get total count
		query.Count(&total)

		// Get organizations with pagination
		if err := query.Distinct().Offset(offset).Limit(limit).Find(&organizations).Error; err != nil {
			http.Error(w, "Failed to fetch organizations", http.StatusInternalServerError)
			return
		}

		// Enhance organizations with additional data
		var enrichedOrgs []OrganizationResponse
		for _, org := range organizations {
			// Get member count
			var memberCount int64
			db.Model(&models.OrganizationMembership{}).Where("organization_id = ?", org.ID).Count(&memberCount)

			// Get registry count
			var registryCount int64
			db.Model(&models.Registry{}).Where("organization_id = ?", org.ID).Count(&registryCount)

			// Get user role if authenticated
			var userRole string
			var isOwner bool
			if userID > 0 {
				var membership models.OrganizationMembership
				if err := db.Where("organization_id = ? AND user_id = ?", org.ID, userID).First(&membership).Error; err == nil {
					userRole = membership.Role
					isOwner = (membership.Role == "owner")
				}
			}

			enrichedOrgs = append(enrichedOrgs, OrganizationResponse{
				Organization:  &org,
				MemberCount:   int(memberCount),
				RegistryCount: int(registryCount),
				UserRole:      userRole,
				IsOwner:       isOwner,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"organizations": enrichedOrgs,
			"total":         total,
			"limit":         limit,
			"offset":        offset,
		})
	}
}

// CreateOrganization creates a new organization
func CreateOrganization(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value("userID").(uint)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req CreateOrganizationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Validate organization name
		if !isValidOrganizationName(req.Name) {
			http.Error(w, "Invalid organization name. Must contain only lowercase letters, numbers, and hyphens", http.StatusBadRequest)
			return
		}

		// Check if organization already exists
		var existingOrg models.Organization
		if err := db.Where("name = ?", req.Name).First(&existingOrg).Error; err == nil {
			http.Error(w, "Organization name already exists", http.StatusConflict)
			return
		}

		// Create organization
		organization := &models.Organization{
			Name:        req.Name,
			DisplayName: req.DisplayName,
			Description: req.Description,
			IsPublic:    req.IsPublic,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := db.Create(organization).Error; err != nil {
			http.Error(w, "Failed to create organization", http.StatusInternalServerError)
			return
		}

		// Add creator as owner
		membership := &models.OrganizationMembership{
			OrganizationID: organization.ID,
			UserID:         userID,
			Role:           "owner",
			CreatedAt:      time.Now(),
		}

		if err := db.Create(membership).Error; err != nil {
			// Rollback organization creation
			db.Delete(organization)
			http.Error(w, "Failed to create organization membership", http.StatusInternalServerError)
			return
		}

		response := OrganizationResponse{
			Organization:  organization,
			MemberCount:   1,
			RegistryCount: 0,
			UserRole:      "owner",
			IsOwner:       true,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}

// GetOrganization returns a specific organization
func GetOrganization(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgName := chi.URLParam(r, "organization")
		if orgName == "" {
			http.Error(w, "Organization name is required", http.StatusBadRequest)
			return
		}

		userID, _ := r.Context().Value("userID").(uint)

		var organization models.Organization
		if err := db.Where("name = ?", orgName).First(&organization).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				http.Error(w, "Organization not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Check access permissions for private organizations
		if !organization.IsPublic && userID > 0 {
			var membership models.OrganizationMembership
			if err := db.Where("organization_id = ? AND user_id = ?", organization.ID, userID).First(&membership).Error; err != nil {
				http.Error(w, "Access denied", http.StatusForbidden)
				return
			}
		} else if !organization.IsPublic {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		// Get additional organization data
		var memberCount int64
		db.Model(&models.OrganizationMembership{}).Where("organization_id = ?", organization.ID).Count(&memberCount)

		var registryCount int64
		db.Model(&models.Registry{}).Where("organization_id = ?", organization.ID).Count(&registryCount)

		// Get user role if authenticated
		var userRole string
		var isOwner bool
		if userID > 0 {
			var membership models.OrganizationMembership
			if err := db.Where("organization_id = ? AND user_id = ?", organization.ID, userID).First(&membership).Error; err == nil {
				userRole = membership.Role
				isOwner = (membership.Role == "owner")
			}
		}

		response := OrganizationResponse{
			Organization:  &organization,
			MemberCount:   int(memberCount),
			RegistryCount: int(registryCount),
			UserRole:      userRole,
			IsOwner:       isOwner,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// UpdateOrganization updates an organization
func UpdateOrganization(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgName := chi.URLParam(r, "organization")
		if orgName == "" {
			http.Error(w, "Organization name is required", http.StatusBadRequest)
			return
		}

		userID, ok := r.Context().Value("userID").(uint)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req UpdateOrganizationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		var organization models.Organization
		if err := db.Where("name = ?", orgName).First(&organization).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				http.Error(w, "Organization not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Check if user has admin or owner permissions
		var membership models.OrganizationMembership
		if err := db.Where("organization_id = ? AND user_id = ?", organization.ID, userID).First(&membership).Error; err != nil {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		if membership.Role != "admin" && membership.Role != "owner" {
			http.Error(w, "Insufficient permissions", http.StatusForbidden)
			return
		}

		// Update fields if provided
		if req.DisplayName != nil {
			organization.DisplayName = *req.DisplayName
		}
		if req.Description != nil {
			organization.Description = *req.Description
		}
		if req.IsPublic != nil {
			organization.IsPublic = *req.IsPublic
		}
		organization.UpdatedAt = time.Now()

		if err := db.Save(&organization).Error; err != nil {
			http.Error(w, "Failed to update organization", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(organization)
	}
}

// DeleteOrganization deletes an organization
func DeleteOrganization(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgName := chi.URLParam(r, "organization")
		if orgName == "" {
			http.Error(w, "Organization name is required", http.StatusBadRequest)
			return
		}

		userID, ok := r.Context().Value("userID").(uint)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var organization models.Organization
		if err := db.Where("name = ?", orgName).First(&organization).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				http.Error(w, "Organization not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Check if user is owner
		var membership models.OrganizationMembership
		if err := db.Where("organization_id = ? AND user_id = ? AND role = ?", organization.ID, userID, "owner").First(&membership).Error; err != nil {
			http.Error(w, "Only organization owners can delete the organization", http.StatusForbidden)
			return
		}

		// Check if organization has registries
		var registryCount int64
		db.Model(&models.Registry{}).Where("organization_id = ?", organization.ID).Count(&registryCount)
		if registryCount > 0 {
			http.Error(w, "Cannot delete organization with existing registries", http.StatusConflict)
			return
		}

		// Delete organization (this will cascade delete memberships)
		if err := db.Delete(&organization).Error; err != nil {
			http.Error(w, "Failed to delete organization", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Organization deleted successfully",
		})
	}
}

// ListMembers returns all members of an organization
func ListMembers(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgName := chi.URLParam(r, "organization")
		if orgName == "" {
			http.Error(w, "Organization name is required", http.StatusBadRequest)
			return
		}

		userID, _ := r.Context().Value("userID").(uint)

		var organization models.Organization
		if err := db.Where("name = ?", orgName).First(&organization).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				http.Error(w, "Organization not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Check access permissions
		if !organization.IsPublic && userID > 0 {
			var membership models.OrganizationMembership
			if err := db.Where("organization_id = ? AND user_id = ?", organization.ID, userID).First(&membership).Error; err != nil {
				http.Error(w, "Access denied", http.StatusForbidden)
				return
			}
		} else if !organization.IsPublic {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		// Get members with user details
		var members []MemberResponse
		if err := db.Table("organization_memberships").
			Select("users.id as user_id, users.username, users.email, users.first_name, users.last_name, organization_memberships.role, organization_memberships.created_at as joined_at").
			Joins("JOIN users ON organization_memberships.user_id = users.id").
			Where("organization_memberships.organization_id = ?", organization.ID).
			Order("organization_memberships.role DESC, users.username ASC").
			Scan(&members).Error; err != nil {
			http.Error(w, "Failed to fetch members", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"members": members,
			"total":   len(members),
		})
	}
}

// AddMember adds a new member to an organization
func AddMember(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgName := chi.URLParam(r, "organization")
		if orgName == "" {
			http.Error(w, "Organization name is required", http.StatusBadRequest)
			return
		}

		userID, ok := r.Context().Value("userID").(uint)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req AddMemberRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Validate role
		if req.Role != "member" && req.Role != "admin" && req.Role != "owner" {
			http.Error(w, "Invalid role. Must be one of: member, admin, owner", http.StatusBadRequest)
			return
		}

		var organization models.Organization
		if err := db.Where("name = ?", orgName).First(&organization).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				http.Error(w, "Organization not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Check if user has admin or owner permissions
		var membership models.OrganizationMembership
		if err := db.Where("organization_id = ? AND user_id = ?", organization.ID, userID).First(&membership).Error; err != nil {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		if membership.Role != "admin" && membership.Role != "owner" {
			http.Error(w, "Insufficient permissions", http.StatusForbidden)
			return
		}

		// Only owners can add other owners
		if req.Role == "owner" && membership.Role != "owner" {
			http.Error(w, "Only owners can add other owners", http.StatusForbidden)
			return
		}

		// Find user to add
		var user models.User
		if err := db.Where("username = ?", req.Username).First(&user).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				http.Error(w, "User not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Check if user is already a member
		var existingMembership models.OrganizationMembership
		if err := db.Where("organization_id = ? AND user_id = ?", organization.ID, user.ID).First(&existingMembership).Error; err == nil {
			http.Error(w, "User is already a member of this organization", http.StatusConflict)
			return
		}

		// Add member
		newMembership := &models.OrganizationMembership{
			OrganizationID: organization.ID,
			UserID:         user.ID,
			Role:           req.Role,
			CreatedAt:      time.Now(),
		}

		if err := db.Create(newMembership).Error; err != nil {
			http.Error(w, "Failed to add member", http.StatusInternalServerError)
			return
		}

		memberResponse := MemberResponse{
			UserID:    user.ID,
			Username:  user.Username,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Role:      newMembership.Role,
			JoinedAt:  newMembership.CreatedAt,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(memberResponse)
	}
}

// UpdateMember updates a member's role in an organization
func UpdateMember(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgName := chi.URLParam(r, "organization")
		username := chi.URLParam(r, "username")

		if orgName == "" || username == "" {
			http.Error(w, "Organization name and username are required", http.StatusBadRequest)
			return
		}

		userID, ok := r.Context().Value("userID").(uint)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req UpdateMemberRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Validate role
		if req.Role != "member" && req.Role != "admin" && req.Role != "owner" {
			http.Error(w, "Invalid role. Must be one of: member, admin, owner", http.StatusBadRequest)
			return
		}

		var organization models.Organization
		if err := db.Where("name = ?", orgName).First(&organization).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				http.Error(w, "Organization not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Check if current user has admin or owner permissions
		var currentMembership models.OrganizationMembership
		if err := db.Where("organization_id = ? AND user_id = ?", organization.ID, userID).First(&currentMembership).Error; err != nil {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		if currentMembership.Role != "admin" && currentMembership.Role != "owner" {
			http.Error(w, "Insufficient permissions", http.StatusForbidden)
			return
		}

		// Find target user
		var targetUser models.User
		if err := db.Where("username = ?", username).First(&targetUser).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				http.Error(w, "User not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Find target membership
		var targetMembership models.OrganizationMembership
		if err := db.Where("organization_id = ? AND user_id = ?", organization.ID, targetUser.ID).First(&targetMembership).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				http.Error(w, "User is not a member of this organization", http.StatusNotFound)
				return
			}
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Only owners can modify owner roles or create new owners
		if (targetMembership.Role == "owner" || req.Role == "owner") && currentMembership.Role != "owner" {
			http.Error(w, "Only owners can modify owner roles", http.StatusForbidden)
			return
		}

		// Update role
		targetMembership.Role = req.Role
		if err := db.Save(&targetMembership).Error; err != nil {
			http.Error(w, "Failed to update member role", http.StatusInternalServerError)
			return
		}

		memberResponse := MemberResponse{
			UserID:    targetUser.ID,
			Username:  targetUser.Username,
			Email:     targetUser.Email,
			FirstName: targetUser.FirstName,
			LastName:  targetUser.LastName,
			Role:      targetMembership.Role,
			JoinedAt:  targetMembership.CreatedAt,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(memberResponse)
	}
}

// RemoveMember removes a member from an organization
func RemoveMember(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgName := chi.URLParam(r, "organization")
		username := chi.URLParam(r, "username")

		if orgName == "" || username == "" {
			http.Error(w, "Organization name and username are required", http.StatusBadRequest)
			return
		}

		userID, ok := r.Context().Value("userID").(uint)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var organization models.Organization
		if err := db.Where("name = ?", orgName).First(&organization).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				http.Error(w, "Organization not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Find target user
		var targetUser models.User
		if err := db.Where("username = ?", username).First(&targetUser).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				http.Error(w, "User not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Find target membership
		var targetMembership models.OrganizationMembership
		if err := db.Where("organization_id = ? AND user_id = ?", organization.ID, targetUser.ID).First(&targetMembership).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				http.Error(w, "User is not a member of this organization", http.StatusNotFound)
				return
			}
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Check permissions
		if targetUser.ID == userID {
			// Users can leave organizations, but owners cannot leave if they're the last owner
			if targetMembership.Role == "owner" {
				var ownerCount int64
				db.Model(&models.OrganizationMembership{}).Where("organization_id = ? AND role = ?", organization.ID, "owner").Count(&ownerCount)
				if ownerCount <= 1 {
					http.Error(w, "Cannot leave organization as the last owner", http.StatusConflict)
					return
				}
			}
		} else {
			// Check if current user has admin or owner permissions
			var currentMembership models.OrganizationMembership
			if err := db.Where("organization_id = ? AND user_id = ?", organization.ID, userID).First(&currentMembership).Error; err != nil {
				http.Error(w, "Access denied", http.StatusForbidden)
				return
			}

			if currentMembership.Role != "admin" && currentMembership.Role != "owner" {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			// Only owners can remove other owners
			if targetMembership.Role == "owner" && currentMembership.Role != "owner" {
				http.Error(w, "Only owners can remove other owners", http.StatusForbidden)
				return
			}
		}

		// Remove member
		if err := db.Delete(&targetMembership).Error; err != nil {
			http.Error(w, "Failed to remove member", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Member removed successfully",
		})
	}
}

// Utility functions

func isValidOrganizationName(name string) bool {
	if len(name) < 1 || len(name) > 50 {
		return false
	}
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') ||
			 (char >= '0' && char <= '9') ||
			 char == '-') {
			return false
		}
	}
	// Cannot start or end with hyphen
	if name[0] == '-' || name[len(name)-1] == '-' {
		return false
	}
	return true
}