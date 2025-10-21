package main

import (
	"embed"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/casapps/casreg/cli"
	"github.com/casapps/casreg/config"
	"github.com/casapps/casreg/handlers"
	"github.com/casapps/casreg/middleware"
	"github.com/casapps/casreg/models"
	"github.com/casapps/casreg/scheduler"
	"github.com/casapps/casreg/storage"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
)

var (
	version   = "1.0.0"
	buildTime = "unknown"
	gitCommit = "unknown"

	//go:embed ui/dist/*
	uiAssets embed.FS

	//go:embed support/docs/*
	docsAssets embed.FS
)

func main() {
	// Detect if running as server or CLI
	if len(os.Args) > 1 && !strings.HasPrefix(os.Args[1], "-") && os.Args[1] != "serve" {
		// CLI mode
		cli.Execute()
		return
	}

	// Server mode
	fmt.Printf("casreg v%s (built: %s, commit: %s)\n", version, buildTime, gitCommit)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logrus.Fatalf("Failed to load configuration: %v", err)
	}

	// Setup logging
	setupLogging(cfg)

	// Initialize database
	db, err := models.InitDatabase(cfg)
	if err != nil {
		logrus.Fatalf("Failed to initialize database: %v", err)
	}

	// Run migrations
	if err := models.RunMigrations(db); err != nil {
		logrus.Fatalf("Failed to run migrations: %v", err)
	}

	// Create first admin user if none exists
	if err := models.CreateFirstAdmin(db); err != nil {
		logrus.Fatalf("Failed to create first admin: %v", err)
	}

	// Initialize storage backend
	storage, err := initStorage(cfg)
	if err != nil {
		logrus.Fatalf("Failed to initialize storage: %v", err)
	}

	// Initialize scheduler
	sched := scheduler.New(cfg, db, storage)
	sched.Start()
	defer sched.Stop()

	// Setup HTTP router
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.Logging)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.ProxyHeaders(cfg))

	// CORS configuration
	r.Use(middleware.CORS(cfg))

	// Rate limiting
	if cfg.RateLimit.Enabled {
		r.Use(middleware.RateLimit(cfg))
	}

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"healthy","version":"` + version + `"}`))
	})

	// Metrics endpoint (Prometheus compatible)
	r.Get("/metrics", handlers.MetricsHandler(cfg, db, storage))

	// API v1 routes
	r.Route("/v1", func(r chi.Router) {
		// Authentication routes (no auth required)
		r.Post("/auth/login", handlers.Login(cfg, db))
		r.Post("/auth/register", handlers.Register(cfg, db))
		r.Post("/auth/refresh", handlers.RefreshToken(cfg, db))

		// Public routes (no auth required)
		r.Get("/registries", handlers.ListPublicRegistries(db))
		r.Get("/support/docs", handlers.ListDocs(docsAssets))
		r.Get("/support/docs/{id}", handlers.GetDoc(docsAssets))

		// Authenticated routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate(cfg, db))

			// User management
			r.Get("/users/me", handlers.GetCurrentUser(db))
			r.Put("/users/me", handlers.UpdateCurrentUser(db))
			r.Post("/users/me/password", handlers.ChangePassword(cfg, db))
			r.Get("/users/{id}", handlers.GetUser(db))

			// Token management
			r.Get("/tokens", handlers.ListTokens(db))
			r.Post("/tokens", handlers.CreateToken(cfg, db))
			r.Delete("/tokens/{id}", handlers.DeleteToken(db))
			r.Post("/tokens/{id}/rotate", handlers.RotateToken(cfg, db))

			// Organization management
			r.Get("/organizations", handlers.ListOrganizations(db))
			r.Post("/organizations", handlers.CreateOrganization(db))
			r.Get("/organizations/{name}", handlers.GetOrganization(db))
			r.Put("/organizations/{name}", handlers.UpdateOrganization(db))
			r.Delete("/organizations/{name}", handlers.DeleteOrganization(db))
			r.Post("/organizations/{name}/members", handlers.AddOrgMember(db))
			r.Get("/organizations/{name}/members", handlers.ListOrgMembers(db))
			r.Delete("/organizations/{name}/members/{username}", handlers.RemoveOrgMember(db))

			// Registry management
			r.Get("/registries", handlers.ListRegistries(db))
			r.Post("/registries", handlers.CreateRegistry(db, storage))
			r.Get("/registries/{name}", handlers.GetRegistry(db))
			r.Put("/registries/{name}", handlers.UpdateRegistry(db))
			r.Delete("/registries/{name}", handlers.DeleteRegistry(db, storage))

			// Repository management
			r.Get("/registries/{registry}/repositories", handlers.ListRepositories(db))
			r.Post("/registries/{registry}/repositories", handlers.CreateRepository(db))
			r.Get("/registries/{registry}/repositories/{repo}", handlers.GetRepository(db))
			r.Put("/registries/{registry}/repositories/{repo}", handlers.UpdateRepository(db))
			r.Delete("/registries/{registry}/repositories/{repo}", handlers.DeleteRepository(db, storage))

			// Tag management
			r.Get("/registries/{registry}/repositories/{repo}/tags", handlers.ListTags(db))
			r.Get("/registries/{registry}/repositories/{repo}/tags/{tag}", handlers.GetTag(db))
			r.Delete("/registries/{registry}/repositories/{repo}/tags/{tag}", handlers.DeleteTag(db, storage))
			r.Post("/registries/{registry}/repositories/{repo}/tags/{tag}/scan", handlers.ScanTag(cfg, db, storage))
			r.Get("/registries/{registry}/repositories/{repo}/tags/{tag}/scan-results", handlers.GetScanResults(db))

			// Support system
			r.Get("/support/tickets", handlers.ListTickets(db))
			r.Post("/support/tickets", handlers.CreateTicket(db))
			r.Get("/support/tickets/{id}", handlers.GetTicket(db))
			r.Post("/support/tickets/{id}/comments", handlers.AddTicketComment(db))

			// Admin routes
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireAdmin)

				r.Get("/admin/users", handlers.AdminListUsers(db))
				r.Post("/admin/users", handlers.AdminCreateUser(cfg, db))
				r.Put("/admin/users/{id}", handlers.AdminUpdateUser(db))
				r.Delete("/admin/users/{id}", handlers.AdminDeleteUser(db))

				r.Get("/admin/organizations", handlers.AdminListOrganizations(db))
				r.Get("/admin/registries", handlers.AdminListRegistries(db))
				r.Get("/admin/system/stats", handlers.AdminSystemStats(cfg, db, storage))
				r.Post("/admin/system/cleanup", handlers.AdminCleanup(db, storage))
			})
		})
	})

	// Docker Registry V2 API
	r.Route("/v2", func(r chi.Router) {
		r.Get("/", handlers.RegistryVersion(version))

		// Optional auth for public registries
		r.Use(middleware.OptionalAuth(cfg, db))

		r.Get("/{registry}/{repository}/manifests/{reference}", handlers.GetManifest(db, storage))
		r.Put("/{registry}/{repository}/manifests/{reference}", handlers.PutManifest(db, storage))
		r.Delete("/{registry}/{repository}/manifests/{reference}", handlers.DeleteManifest(db, storage))

		r.Get("/{registry}/{repository}/blobs/{digest}", handlers.GetBlob(db, storage))
		r.Delete("/{registry}/{repository}/blobs/{digest}", handlers.DeleteBlob(db, storage))

		r.Post("/{registry}/{repository}/blobs/uploads/", handlers.StartBlobUpload(db, storage))
		r.Get("/{registry}/{repository}/blobs/uploads/{uuid}", handlers.GetBlobUploadStatus(db, storage))
		r.Patch("/{registry}/{repository}/blobs/uploads/{uuid}", handlers.UploadBlobChunk(db, storage))
		r.Put("/{registry}/{repository}/blobs/uploads/{uuid}", handlers.CompleteBlobUpload(db, storage))
		r.Delete("/{registry}/{repository}/blobs/uploads/{uuid}", handlers.CancelBlobUpload(db, storage))

		r.Get("/{registry}/{repository}/tags/list", handlers.ListTagsV2(db))
	})

	// Swagger UI
	r.Get("/swagger/*", handlers.SwaggerUI())

	// Serve embedded UI
	r.Get("/*", handlers.ServeUI(uiAssets))

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	logrus.Infof("Starting casreg server on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		logrus.Fatalf("Server failed: %v", err)
	}
}

func setupLogging(cfg *config.Config) {
	level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)
	logrus.SetFormatter(&logrus.JSONFormatter{})
}

func initStorage(cfg *config.Config) (storage.Storage, error) {
	switch cfg.Storage.Backend {
	case "s3":
		return storage.NewS3Storage(cfg)
	case "nfs":
		return storage.NewNFSStorage(cfg)
	default:
		return storage.NewLocalStorage(cfg)
	}
}
