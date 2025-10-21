package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"registry-auth/handlers"
	"registry-auth/models"
)

func main() {
	// Get configuration from environment variables
	port := getEnv("PORT", "8080")
	jwtSecret := getEnv("JWT_SECRET", "your-default-jwt-secret-change-in-production")
	externalURL := getEnv("EXTERNAL_URL", "http://localhost:8080")
	registryURL := getEnv("REGISTRY_URL", "http://localhost:5000")
	dbPath := getEnv("DATABASE_PATH", "./auth.db")

	// Initialize database
	db := initDatabase(dbPath)

	// Create admin user if no users exist
	createAdminUser(db)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db, jwtSecret, externalURL, registryURL)

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Public routes (no authentication required)
	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", authHandler.Login)
		r.Post("/register", authHandler.Register)
		r.Post("/docker", authHandler.DockerAuth)
		r.Get("/validate", authHandler.ValidateToken)
	})

	// Protected routes (authentication required)
	r.Route("/api", func(r chi.Router) {
		// Apply authentication middleware
		r.Use(handlers.AuthMiddleware(db, jwtSecret))

		// User routes
		r.Route("/users", func(r chi.Router) {
			r.Get("/me", handlers.GetProfile(db))
			r.Put("/me", handlers.UpdateProfile(db))
			r.Post("/me/password", handlers.ChangePassword(db))
			r.Get("/me/dashboard", handlers.GetDashboard(db))

			// Admin only routes
			r.Get("/", handlers.ListUsers(db))
			r.Get("/{userID}", handlers.GetUser(db))
		})

		// Token management routes
		r.Route("/tokens", func(r chi.Router) {
			r.Get("/", handlers.ListTokens(db))
			r.Post("/", handlers.CreateToken(db))
			r.Get("/{tokenID}", handlers.GetToken(db))
			r.Put("/{tokenID}", handlers.UpdateToken(db))
			r.Post("/{tokenID}/rotate", handlers.RotateToken(db))
			r.Delete("/{tokenID}", handlers.DeleteToken(db))
		})

		// Organization routes
		r.Route("/organizations", func(r chi.Router) {
			r.Get("/", handlers.ListOrganizations(db))
			r.Post("/", handlers.CreateOrganization(db))
			r.Get("/{organization}", handlers.GetOrganization(db))
			r.Put("/{organization}", handlers.UpdateOrganization(db))
			r.Delete("/{organization}", handlers.DeleteOrganization(db))

			// Member management
			r.Get("/{organization}/members", handlers.ListMembers(db))
			r.Post("/{organization}/members", handlers.AddMember(db))
			r.Put("/{organization}/members/{username}", handlers.UpdateMember(db))
			r.Delete("/{organization}/members/{username}", handlers.RemoveMember(db))
		})

		// Registry and repository routes
		r.Route("/registries", func(r chi.Router) {
			// Registry management (basic endpoints - full registry functionality would be in separate service)
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				// This would typically proxy to the main registry service
				w.WriteHeader(http.StatusNotImplemented)
				w.Write([]byte("Registry management is handled by the main registry service"))
			})

			// Repository routes
			r.Route("/{registry}/repositories", func(r chi.Router) {
				r.Get("/", handlers.ListRepositories(db))
				r.Post("/", handlers.CreateRepository(db))
				r.Get("/{repository}", handlers.GetRepository(db))
				r.Put("/{repository}", handlers.UpdateRepository(db))
				r.Delete("/{repository}", handlers.DeleteRepository(db))

				// Repository actions
				r.Post("/{repository}/star", handlers.StarRepository(db))
				r.Delete("/{repository}/star", handlers.UnstarRepository(db))
				r.Post("/{repository}/push", handlers.HandlePush(db))
				r.Get("/{repository}/pull", handlers.HandlePull(db))
			})
		})
	})

	// Docker Registry V2 API compatibility routes (for authentication)
	r.Route("/v2", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"name": "casreg-auth", "version": "1.0.0"}`))
		})
		// Add other Docker Registry V2 routes as needed
	})

	// Start server
	log.Printf("Starting auth service on port %s", port)
	log.Printf("External URL: %s", externalURL)
	log.Printf("Registry URL: %s", registryURL)

	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func initDatabase(dbPath string) *gorm.DB {
	// Configure GORM logger to show SQL queries in development
	gormLogger := logger.Default.LogMode(logger.Info)
	if os.Getenv("ENV") == "production" {
		gormLogger = logger.Default.LogMode(logger.Error)
	}

	// Open database connection
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get database instance:", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// Auto-migrate database schema
	log.Println("Running database migrations...")
	err = db.AutoMigrate(
		&models.User{},
		&models.Organization{},
		&models.OrganizationMembership{},
		&models.Registry{},
		&models.Repository{},
		&models.Tag{},
		&models.RepositoryStar{},
		&models.PersonalToken{},
		&models.AuditLog{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	log.Println("Database migrations completed successfully")
	return db
}

func createAdminUser(db *gorm.DB) {
	// Check if any users exist
	var userCount int64
	db.Model(&models.User{}).Count(&userCount)

	if userCount == 0 {
		log.Println("No users found, creating default admin user...")

		// Default admin credentials
		adminUsername := getEnv("ADMIN_USERNAME", "admin")
		adminPassword := getEnv("ADMIN_PASSWORD", "admin123")
		adminEmail := getEnv("ADMIN_EMAIL", "admin@casreg.local")

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Fatal("Failed to hash admin password:", err)
		}

		// Create admin user
		adminUser := &models.User{
			Username:  adminUsername,
			Email:     adminEmail,
			Password:  string(hashedPassword),
			FirstName: "System",
			LastName:  "Administrator",
			Role:      "admin",
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := db.Create(adminUser).Error; err != nil {
			log.Fatal("Failed to create admin user:", err)
		}

		log.Printf("Admin user created successfully:")
		log.Printf("  Username: %s", adminUsername)
		log.Printf("  Email: %s", adminEmail)
		log.Printf("  Password: %s", adminPassword)
		log.Printf("  Please change the password after first login!")
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}