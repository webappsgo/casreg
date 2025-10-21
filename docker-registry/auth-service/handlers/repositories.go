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

// CreateRepositoryRequest represents the request body for creating a repository
type CreateRepositoryRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Description string `json:"description,omitempty"`
	IsPublic    bool   `json:"is_public"`
}

// UpdateRepositoryRequest represents the request body for updating a repository
type UpdateRepositoryRequest struct {
	Description *string `json:"description,omitempty"`
	IsPublic    *bool   `json:"is_public,omitempty"`
}

// RepositoryResponse represents a repository with additional metadata
type RepositoryResponse struct {
	*models.Repository
	TagCount     int       `json:"tag_count"`
	LastPushedAt *time.Time `json:"last_pushed_at"`
	PullCount    int64     `json:"pull_count"`
	Stars        int       `json:"stars"`
	IsStarred    bool      `json:"is_starred"`
}

// ListRepositories returns repositories for a registry
func ListRepositories(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		registryName := chi.URLParam(r, "registry")
		if registryName == "" {
			http.Error(w, "Registry name is required", http.StatusBadRequest)
			return
		}

		// Get user from context for access control
		userID, _ := r.Context().Value("userID").(uint)

		// Find the registry first
		var registry models.Registry
		if err := db.Where("name = ?", registryName).First(&registry).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				http.Error(w, "Registry not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Check access permissions
		if !registry.IsPublic && (registry.UserID == nil || *registry.UserID != userID) {
			// Check if user has access through organization membership
			if registry.OrganizationID != nil {
				var membership models.OrganizationMembership
				if err := db.Where("organization_id = ? AND user_id = ?", *registry.OrganizationID, userID).First(&membership).Error; err != nil {
					http.Error(w, "Access denied", http.StatusForbidden)
					return
				}
			} else {
				http.Error(w, "Access denied", http.StatusForbidden)
				return
			}
		}

		// Get pagination parameters
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		if limit <= 0 || limit > 100 {
			limit = 20
		}

		// Get repositories
		var repositories []models.Repository
		query := db.Where("registry_id = ?", registry.ID)

		// Add search filter if provided
		if search := r.URL.Query().Get("search"); search != "" {
			query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
		}

		if err := query.Offset(offset).Limit(limit).Find(&repositories).Error; err != nil {
			http.Error(w, "Failed to fetch repositories", http.StatusInternalServerError)
			return
		}

		// Get total count
		var total int64
		query.Model(&models.Repository{}).Count(&total)

		// Enhance repositories with additional data
		var enrichedRepos []RepositoryResponse
		for _, repo := range repositories {
			// Get tag count
			var tagCount int64
			db.Model(&models.Tag{}).Where("repository_id = ?", repo.ID).Count(&tagCount)

			// Get last pushed tag
			var lastTag models.Tag
			var lastPushedAt *time.Time
			if err := db.Where("repository_id = ?", repo.ID).Order("created_at DESC").First(&lastTag).Error; err == nil {
				lastPushedAt = &lastTag.CreatedAt
			}

			// Get star count
			var starCount int64
			db.Model(&models.RepositoryStar{}).Where("repository_id = ?", repo.ID).Count(&starCount)

			// Check if current user starred this repo
			isStarred := false
			if userID > 0 {
				var star models.RepositoryStar
				if err := db.Where("repository_id = ? AND user_id = ?", repo.ID, userID).First(&star).Error; err == nil {
					isStarred = true
				}
			}

			enrichedRepos = append(enrichedRepos, RepositoryResponse{
				Repository:   &repo,
				TagCount:     int(tagCount),
				LastPushedAt: lastPushedAt,
				PullCount:    repo.PullCount,
				Stars:        int(starCount),
				IsStarred:    isStarred,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"repositories": enrichedRepos,
			"total":        total,
			"limit":        limit,
			"offset":       offset,
		})
	}
}

// CreateRepository creates a new repository in a registry
func CreateRepository(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		registryName := chi.URLParam(r, "registry")
		if registryName == "" {
			http.Error(w, "Registry name is required", http.StatusBadRequest)
			return
		}

		userID, ok := r.Context().Value("userID").(uint)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req CreateRepositoryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Validate repository name
		if !isValidRepositoryName(req.Name) {
			http.Error(w, "Invalid repository name. Must contain only lowercase letters, numbers, hyphens, and underscores", http.StatusBadRequest)
			return
		}

		// Find the registry
		var registry models.Registry
		if err := db.Where("name = ?", registryName).First(&registry).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				http.Error(w, "Registry not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Check write permissions
		if !hasRegistryWriteAccess(db, &registry, userID) {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		// Check if repository already exists
		var existingRepo models.Repository
		if err := db.Where("registry_id = ? AND name = ?", registry.ID, req.Name).First(&existingRepo).Error; err == nil {
			http.Error(w, "Repository already exists", http.StatusConflict)
			return
		}

		// Create repository
		repository := &models.Repository{
			RegistryID:  registry.ID,
			Name:        req.Name,
			Description: req.Description,
			IsPublic:    req.IsPublic,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := db.Create(repository).Error; err != nil {
			http.Error(w, "Failed to create repository", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(repository)
	}
}

// GetRepository returns a specific repository
func GetRepository(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		registryName := chi.URLParam(r, "registry")
		repositoryName := chi.URLParam(r, "repository")

		if registryName == "" || repositoryName == "" {
			http.Error(w, "Registry and repository names are required", http.StatusBadRequest)
			return
		}

		userID, _ := r.Context().Value("userID").(uint)

		// Find repository with registry
		var repository models.Repository
		if err := db.Joins("JOIN registries ON repositories.registry_id = registries.id").
			Where("registries.name = ? AND repositories.name = ?", registryName, repositoryName).
			First(&repository).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				http.Error(w, "Repository not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Get registry for access control
		var registry models.Registry
		if err := db.First(&registry, repository.RegistryID).Error; err != nil {
			http.Error(w, "Registry not found", http.StatusNotFound)
			return
		}

		// Check access permissions
		if !repository.IsPublic && !registry.IsPublic && (registry.UserID == nil || *registry.UserID != userID) {
			if registry.OrganizationID != nil {
				var membership models.OrganizationMembership
				if err := db.Where("organization_id = ? AND user_id = ?", *registry.OrganizationID, userID).First(&membership).Error; err != nil {
					http.Error(w, "Access denied", http.StatusForbidden)
					return
				}
			} else {
				http.Error(w, "Access denied", http.StatusForbidden)
				return
			}
		}

		// Get additional repository data
		var tagCount int64
		db.Model(&models.Tag{}).Where("repository_id = ?", repository.ID).Count(&tagCount)

		var starCount int64
		db.Model(&models.RepositoryStar{}).Where("repository_id = ?", repository.ID).Count(&starCount)

		isStarred := false
		if userID > 0 {
			var star models.RepositoryStar
			if err := db.Where("repository_id = ? AND user_id = ?", repository.ID, userID).First(&star).Error; err == nil {
				isStarred = true
			}
		}

		// Get last pushed tag
		var lastTag models.Tag
		var lastPushedAt *time.Time
		if err := db.Where("repository_id = ?", repository.ID).Order("created_at DESC").First(&lastTag).Error; err == nil {
			lastPushedAt = &lastTag.CreatedAt
		}

		response := RepositoryResponse{
			Repository:   &repository,
			TagCount:     int(tagCount),
			LastPushedAt: lastPushedAt,
			PullCount:    repository.PullCount,
			Stars:        int(starCount),
			IsStarred:    isStarred,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// UpdateRepository updates a repository
func UpdateRepository(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		registryName := chi.URLParam(r, "registry")
		repositoryName := chi.URLParam(r, "repository")

		if registryName == "" || repositoryName == "" {
			http.Error(w, "Registry and repository names are required", http.StatusBadRequest)
			return
		}

		userID, ok := r.Context().Value("userID").(uint)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req UpdateRepositoryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Find repository with registry
		var repository models.Repository
		if err := db.Joins("JOIN registries ON repositories.registry_id = registries.id").
			Where("registries.name = ? AND repositories.name = ?", registryName, repositoryName).
			First(&repository).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				http.Error(w, "Repository not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Get registry for access control
		var registry models.Registry
		if err := db.First(&registry, repository.RegistryID).Error; err != nil {
			http.Error(w, "Registry not found", http.StatusNotFound)
			return
		}

		// Check write permissions
		if !hasRegistryWriteAccess(db, &registry, userID) {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		// Update fields if provided
		if req.Description != nil {
			repository.Description = *req.Description
		}
		if req.IsPublic != nil {
			repository.IsPublic = *req.IsPublic
		}
		repository.UpdatedAt = time.Now()

		if err := db.Save(&repository).Error; err != nil {
			http.Error(w, "Failed to update repository", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(repository)
	}
}

// DeleteRepository deletes a repository and all its tags
func DeleteRepository(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		registryName := chi.URLParam(r, "registry")
		repositoryName := chi.URLParam(r, "repository")

		if registryName == "" || repositoryName == "" {
			http.Error(w, "Registry and repository names are required", http.StatusBadRequest)
			return
		}

		userID, ok := r.Context().Value("userID").(uint)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Find repository with registry
		var repository models.Repository
		if err := db.Joins("JOIN registries ON repositories.registry_id = registries.id").
			Where("registries.name = ? AND repositories.name = ?", registryName, repositoryName).
			First(&repository).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				http.Error(w, "Repository not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Get registry for access control
		var registry models.Registry
		if err := db.First(&registry, repository.RegistryID).Error; err != nil {
			http.Error(w, "Registry not found", http.StatusNotFound)
			return
		}

		// Check admin permissions (deletion requires admin access)
		if !hasRegistryAdminAccess(db, &registry, userID) {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		// Delete repository and cascade delete tags, stars, etc.
		if err := db.Delete(&repository).Error; err != nil {
			http.Error(w, "Failed to delete repository", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Repository deleted successfully",
		})
	}
}

// StarRepository adds a star to a repository
func StarRepository(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		registryName := chi.URLParam(r, "registry")
		repositoryName := chi.URLParam(r, "repository")

		userID, ok := r.Context().Value("userID").(uint)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Find repository
		var repository models.Repository
		if err := db.Joins("JOIN registries ON repositories.registry_id = registries.id").
			Where("registries.name = ? AND repositories.name = ?", registryName, repositoryName).
			First(&repository).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				http.Error(w, "Repository not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Check if already starred
		var existingStar models.RepositoryStar
		if err := db.Where("repository_id = ? AND user_id = ?", repository.ID, userID).First(&existingStar).Error; err == nil {
			http.Error(w, "Repository already starred", http.StatusConflict)
			return
		}

		// Create star
		star := &models.RepositoryStar{
			RepositoryID: repository.ID,
			UserID:       userID,
			CreatedAt:    time.Now(),
		}

		if err := db.Create(star).Error; err != nil {
			http.Error(w, "Failed to star repository", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Repository starred successfully",
		})
	}
}

// UnstarRepository removes a star from a repository
func UnstarRepository(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		registryName := chi.URLParam(r, "registry")
		repositoryName := chi.URLParam(r, "repository")

		userID, ok := r.Context().Value("userID").(uint)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Find repository
		var repository models.Repository
		if err := db.Joins("JOIN registries ON repositories.registry_id = registries.id").
			Where("registries.name = ? AND repositories.name = ?", registryName, repositoryName).
			First(&repository).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				http.Error(w, "Repository not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Remove star
		if err := db.Where("repository_id = ? AND user_id = ?", repository.ID, userID).Delete(&models.RepositoryStar{}).Error; err != nil {
			http.Error(w, "Failed to unstar repository", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Repository unstarred successfully",
		})
	}
}

// HandlePush handles Docker registry push operations
func HandlePush(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		registryName := chi.URLParam(r, "registry")
		repositoryName := chi.URLParam(r, "repository")
		tag := r.URL.Query().Get("tag")

		if tag == "" {
			tag = "latest"
		}

		userID, ok := r.Context().Value("userID").(uint)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Find repository
		var repository models.Repository
		if err := db.Joins("JOIN registries ON repositories.registry_id = registries.id").
			Where("registries.name = ? AND repositories.name = ?", registryName, repositoryName).
			First(&repository).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				http.Error(w, "Repository not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Get registry for access control
		var registry models.Registry
		if err := db.First(&registry, repository.RegistryID).Error; err != nil {
			http.Error(w, "Registry not found", http.StatusNotFound)
			return
		}

		// Check write permissions
		if !hasRegistryWriteAccess(db, &registry, userID) {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		// Create or update tag record
		var tagRecord models.Tag
		result := db.Where("repository_id = ? AND name = ?", repository.ID, tag).First(&tagRecord)
		if result.Error == gorm.ErrRecordNotFound {
			// Create new tag
			tagRecord = models.Tag{
				RepositoryID: repository.ID,
				Name:         tag,
				Digest:       "", // This would be set by actual registry implementation
				Size:         0,  // This would be set by actual registry implementation
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}
			db.Create(&tagRecord)
		} else {
			// Update existing tag
			tagRecord.UpdatedAt = time.Now()
			db.Save(&tagRecord)
		}

		// Update repository push count and last pushed time
		db.Model(&repository).Updates(map[string]interface{}{
			"push_count":  repository.PushCount + 1,
			"updated_at":  time.Now(),
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Push successful",
			"tag":     tag,
		})
	}
}

// HandlePull handles Docker registry pull operations
func HandlePull(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		registryName := chi.URLParam(r, "registry")
		repositoryName := chi.URLParam(r, "repository")
		tag := r.URL.Query().Get("tag")

		if tag == "" {
			tag = "latest"
		}

		userID, _ := r.Context().Value("userID").(uint)

		// Find repository
		var repository models.Repository
		if err := db.Joins("JOIN registries ON repositories.registry_id = registries.id").
			Where("registries.name = ? AND repositories.name = ?", registryName, repositoryName).
			First(&repository).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				http.Error(w, "Repository not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Get registry for access control
		var registry models.Registry
		if err := db.First(&registry, repository.RegistryID).Error; err != nil {
			http.Error(w, "Registry not found", http.StatusNotFound)
			return
		}

		// Check read permissions
		if !hasRegistryReadAccess(db, &registry, userID) {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		// Find tag
		var tagRecord models.Tag
		if err := db.Where("repository_id = ? AND name = ?", repository.ID, tag).First(&tagRecord).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				http.Error(w, "Tag not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Update pull counts
		db.Model(&repository).Update("pull_count", repository.PullCount+1)
		db.Model(&tagRecord).Update("pull_count", tagRecord.PullCount+1)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Pull authorized",
			"tag":     tagRecord,
		})
	}
}

// Utility functions

func isValidRepositoryName(name string) bool {
	if len(name) < 1 || len(name) > 100 {
		return false
	}
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') ||
			 (char >= '0' && char <= '9') ||
			 char == '-' || char == '_') {
			return false
		}
	}
	return true
}

func hasRegistryReadAccess(db *gorm.DB, registry *models.Registry, userID uint) bool {
	if registry.IsPublic {
		return true
	}

	if registry.UserID != nil && *registry.UserID == userID {
		return true
	}

	if registry.OrganizationID != nil {
		var membership models.OrganizationMembership
		if err := db.Where("organization_id = ? AND user_id = ?", *registry.OrganizationID, userID).First(&membership).Error; err == nil {
			return true
		}
	}

	return false
}

func hasRegistryWriteAccess(db *gorm.DB, registry *models.Registry, userID uint) bool {
	if registry.UserID != nil && *registry.UserID == userID {
		return true
	}

	if registry.OrganizationID != nil {
		var membership models.OrganizationMembership
		if err := db.Where("organization_id = ? AND user_id = ?", *registry.OrganizationID, userID).First(&membership).Error; err == nil {
			return membership.Role == "admin" || membership.Role == "owner"
		}
	}

	return false
}

func hasRegistryAdminAccess(db *gorm.DB, registry *models.Registry, userID uint) bool {
	if registry.UserID != nil && *registry.UserID == userID {
		return true
	}

	if registry.OrganizationID != nil {
		var membership models.OrganizationMembership
		if err := db.Where("organization_id = ? AND user_id = ?", *registry.OrganizationID, userID).First(&membership).Error; err == nil {
			return membership.Role == "owner"
		}
	}

	return false
}