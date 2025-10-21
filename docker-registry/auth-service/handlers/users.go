package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"registry-auth/models"
)

// ProfileUpdateRequest represents the request body for updating user profile
type ProfileUpdateRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

// PasswordChangeRequest represents the request body for changing password
type PasswordChangeRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// DashboardStats represents dashboard statistics
type DashboardStats struct {
	TotalRegistries      int                    `json:"total_registries"`
	TotalRepositories    int                    `json:"total_repositories"`
	TotalOrganizations   int                    `json:"total_organizations"`
	TotalImages          int                    `json:"total_images"`
	TotalPulls           int64                  `json:"total_pulls"`
	RecentActivity       []ActivityItem         `json:"recent_activity"`
	PopularRepositories  []RepositoryStats      `json:"popular_repositories"`
	StorageUsed          int64                  `json:"storage_used"`
}

// ActivityItem represents a single activity item
type ActivityItem struct {
	ID        uint      `json:"id"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource"`
	Details   string    `json:"details"`
	CreatedAt time.Time `json:"created_at"`
}

// RepositoryStats represents repository statistics
type RepositoryStats struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Stars       int       `json:"stars"`
	Pulls       int64     `json:"pulls"`
	LastPushed  time.Time `json:"last_pushed"`
}

// UserListResponse represents a user in the admin user list
type UserListResponse struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetProfile returns the current user's profile
func GetProfile(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value("userID").(uint)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var user models.User
		if err := db.First(&user, userID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				http.Error(w, "User not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Clear password from response
		user.Password = ""

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	}
}

// UpdateProfile updates the current user's profile
func UpdateProfile(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value("userID").(uint)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req ProfileUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Check if email is already taken by another user
		if req.Email != "" {
			var existingUser models.User
			if err := db.Where("email = ? AND id != ?", req.Email, userID).First(&existingUser).Error; err == nil {
				http.Error(w, "Email already in use", http.StatusConflict)
				return
			}
		}

		// Get current user
		var user models.User
		if err := db.First(&user, userID).Error; err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		// Update fields if provided
		if req.FirstName != "" {
			user.FirstName = req.FirstName
		}
		if req.LastName != "" {
			user.LastName = req.LastName
		}
		if req.Email != "" {
			user.Email = req.Email
		}
		user.UpdatedAt = time.Now()

		// Save updated user
		if err := db.Save(&user).Error; err != nil {
			http.Error(w, "Failed to update profile", http.StatusInternalServerError)
			return
		}

		// Clear password from response
		user.Password = ""

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	}
}

// ChangePassword changes the current user's password
func ChangePassword(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value("userID").(uint)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req PasswordChangeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Get current user
		var user models.User
		if err := db.First(&user, userID).Error; err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		// Verify current password
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword)); err != nil {
			http.Error(w, "Current password is incorrect", http.StatusUnauthorized)
			return
		}

		// Hash new password
		newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Failed to process password", http.StatusInternalServerError)
			return
		}

		// Update password
		user.Password = string(newHash)
		user.UpdatedAt = time.Now()

		if err := db.Save(&user).Error; err != nil {
			http.Error(w, "Failed to update password", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Password updated successfully"})
	}
}

// GetDashboard returns dashboard statistics for the current user
func GetDashboard(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value("userID").(uint)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userRole, _ := r.Context().Value("role").(string)

		stats := DashboardStats{}

		// Get registry count
		if userRole == "admin" {
			var count int64
			db.Model(&models.Registry{}).Count(&count)
			stats.TotalRegistries = int(count)
		} else {
			// Count personal and organizational registries
			var personalCount int64
			db.Model(&models.Registry{}).Where("user_id = ?", userID).Count(&personalCount)

			var orgCount int64
			db.Table("registries").
				Joins("JOIN organization_memberships ON registries.organization_id = organization_memberships.organization_id").
				Where("organization_memberships.user_id = ?", userID).
				Count(&orgCount)

			stats.TotalRegistries = int(personalCount + orgCount)
		}

		// Get repository count
		if userRole == "admin" {
			var count int64
			db.Model(&models.Repository{}).Count(&count)
			stats.TotalRepositories = int(count)
		} else {
			// Count repositories in accessible registries
			var count int64
			db.Table("repositories").
				Joins("JOIN registries ON repositories.registry_id = registries.id").
				Joins("LEFT JOIN organization_memberships ON registries.organization_id = organization_memberships.organization_id").
				Where("registries.user_id = ? OR organization_memberships.user_id = ?", userID, userID).
				Count(&count)
			stats.TotalRepositories = int(count)
		}

		// Get organization count (where user is a member)
		var orgCount int64
		db.Model(&models.OrganizationMembership{}).Where("user_id = ?", userID).Count(&orgCount)
		stats.TotalOrganizations = int(orgCount)

		// Get total images (tags)
		if userRole == "admin" {
			var count int64
			db.Model(&models.Tag{}).Count(&count)
			stats.TotalImages = int(count)
		} else {
			var count int64
			db.Table("tags").
				Joins("JOIN repositories ON tags.repository_id = repositories.id").
				Joins("JOIN registries ON repositories.registry_id = registries.id").
				Joins("LEFT JOIN organization_memberships ON registries.organization_id = organization_memberships.organization_id").
				Where("registries.user_id = ? OR organization_memberships.user_id = ?", userID, userID).
				Count(&count)
			stats.TotalImages = int(count)
		}

		// Get total pulls
		if userRole == "admin" {
			var totalPulls sql.NullInt64
			db.Model(&models.Repository{}).Select("COALESCE(SUM(pull_count), 0)").Scan(&totalPulls)
			stats.TotalPulls = totalPulls.Int64
		} else {
			var totalPulls sql.NullInt64
			db.Table("repositories").
				Select("COALESCE(SUM(pull_count), 0)").
				Joins("JOIN registries ON repositories.registry_id = registries.id").
				Joins("LEFT JOIN organization_memberships ON registries.organization_id = organization_memberships.organization_id").
				Where("registries.user_id = ? OR organization_memberships.user_id = ?", userID, userID).
				Scan(&totalPulls)
			stats.TotalPulls = totalPulls.Int64
		}

		// Get recent activity
		stats.RecentActivity = getRecentActivity(db, userID, userRole)

		// Get popular repositories
		stats.PopularRepositories = getPopularRepositories(db, userID, userRole)

		// Get storage used (sum of tag sizes)
		if userRole == "admin" {
			var totalSize sql.NullInt64
			db.Model(&models.Tag{}).Select("COALESCE(SUM(size), 0)").Scan(&totalSize)
			stats.StorageUsed = totalSize.Int64
		} else {
			var totalSize sql.NullInt64
			db.Table("tags").
				Select("COALESCE(SUM(size), 0)").
				Joins("JOIN repositories ON tags.repository_id = repositories.id").
				Joins("JOIN registries ON repositories.registry_id = registries.id").
				Joins("LEFT JOIN organization_memberships ON registries.organization_id = organization_memberships.organization_id").
				Where("registries.user_id = ? OR organization_memberships.user_id = ?", userID, userID).
				Scan(&totalSize)
			stats.StorageUsed = totalSize.Int64
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	}
}

// ListUsers returns a list of all users (admin only)
func ListUsers(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userRole, ok := r.Context().Value("role").(string)
		if !ok || userRole != "admin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Get pagination parameters
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		if limit <= 0 || limit > 100 {
			limit = 20
		}

		var users []models.User
		var total int64

		query := db.Model(&models.User{})

		// Add search filter if provided
		if search := r.URL.Query().Get("search"); search != "" {
			query = query.Where("username ILIKE ? OR email ILIKE ? OR first_name ILIKE ? OR last_name ILIKE ?",
				"%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%")
		}

		// Add role filter if provided
		if role := r.URL.Query().Get("role"); role != "" {
			query = query.Where("role = ?", role)
		}

		// Add active filter if provided
		if active := r.URL.Query().Get("active"); active != "" {
			isActive := active == "true"
			query = query.Where("is_active = ?", isActive)
		}

		// Get total count
		query.Count(&total)

		// Get users with pagination
		if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&users).Error; err != nil {
			http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
			return
		}

		// Convert to response format (without passwords)
		var userResponses []UserListResponse
		for _, user := range users {
			userResponses = append(userResponses, UserListResponse{
				ID:        user.ID,
				Username:  user.Username,
				Email:     user.Email,
				FirstName: user.FirstName,
				LastName:  user.LastName,
				Role:      user.Role,
				IsActive:  user.IsActive,
				CreatedAt: user.CreatedAt,
				UpdatedAt: user.UpdatedAt,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"users":  userResponses,
			"total":  total,
			"limit":  limit,
			"offset": offset,
		})
	}
}

// GetUser returns a specific user (admin only)
func GetUser(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userRole, ok := r.Context().Value("role").(string)
		if !ok || userRole != "admin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		userIDStr := chi.URLParam(r, "userID")
		userID, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		var user models.User
		if err := db.First(&user, uint(userID)).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				http.Error(w, "User not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Clear password from response
		user.Password = ""

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	}
}

// Helper functions

func getRecentActivity(db *gorm.DB, userID uint, role string) []ActivityItem {
	var activities []ActivityItem

	query := db.Model(&models.AuditLog{}).
		Select("id, action, resource, resource_id as details, created_at").
		Order("created_at DESC").
		Limit(10)

	if role != "admin" {
		query = query.Where("user_id = ?", userID)
	}

	var auditLogs []struct {
		ID        uint      `json:"id"`
		Action    string    `json:"action"`
		Resource  string    `json:"resource"`
		Details   uint      `json:"details"`
		CreatedAt time.Time `json:"created_at"`
	}

	if err := query.Scan(&auditLogs).Error; err != nil {
		return activities
	}

	for _, log := range auditLogs {
		activities = append(activities, ActivityItem{
			ID:        log.ID,
			Action:    log.Action,
			Resource:  log.Resource,
			Details:   fmt.Sprintf("ID: %d", log.Details),
			CreatedAt: log.CreatedAt,
		})
	}

	return activities
}

func getPopularRepositories(db *gorm.DB, userID uint, role string) []RepositoryStats {
	var repos []RepositoryStats

	query := db.Table("repositories").
		Select("repositories.id, repositories.name, repositories.pull_count as pulls, repositories.created_at as last_pushed").
		Order("repositories.pull_count DESC").
		Limit(5)

	if role != "admin" {
		query = query.
			Joins("JOIN registries ON repositories.registry_id = registries.id").
			Joins("LEFT JOIN organization_memberships ON registries.organization_id = organization_memberships.organization_id").
			Where("registries.user_id = ? OR organization_memberships.user_id = ?", userID, userID)
	}

	var repoData []struct {
		ID         uint      `json:"id"`
		Name       string    `json:"name"`
		Pulls      int64     `json:"pulls"`
		LastPushed time.Time `json:"last_pushed"`
	}

	if err := query.Scan(&repoData).Error; err != nil {
		return repos
	}

	for _, repo := range repoData {
		// Get star count for each repository
		var starCount int64
		db.Model(&models.RepositoryStar{}).Where("repository_id = ?", repo.ID).Count(&starCount)

		repos = append(repos, RepositoryStats{
			ID:         repo.ID,
			Name:       repo.Name,
			Stars:      int(starCount),
			Pulls:      repo.Pulls,
			LastPushed: repo.LastPushed,
		})
	}

	return repos
}