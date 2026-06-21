package main

import (
	"embed"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/casapps/casreg/src/cli"
	"github.com/casapps/casreg/src/config"
	"github.com/casapps/casreg/src/handler"
	"github.com/casapps/casreg/src/middleware"
	"github.com/casapps/casreg/src/model"
	"github.com/casapps/casreg/src/scheduler"
	"github.com/casapps/casreg/src/storage"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
)

var (
	Version      = "devel"
	CommitID     = "N/A"
	BuildDate    = "N/A"
	OfficialSite = ""

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

	// Server mode — load config and configure logging before any output
	cfg, err := config.Load()
	if err != nil {
		logrus.Fatalf("Failed to load configuration: %v", err)
	}

	setupLogging(cfg)
	logrus.WithFields(logrus.Fields{
		"version": Version,
		"commit":  CommitID,
		"built":   BuildDate,
	}).Info("casreg starting")

	// Initialize database
	db, err := model.InitDatabase(cfg)
	if err != nil {
		logrus.Fatalf("Failed to initialize database: %v", err)
	}

	// Run migrations
	if err := model.RunMigrations(db); err != nil {
		logrus.Fatalf("Failed to run migrations: %v", err)
	}

	// Create first admin user if none exists
	if err := model.CreateFirstAdmin(db); err != nil {
		logrus.Fatalf("Failed to create first admin: %v", err)
	}

	// Initialize storage backend
	store, err := initStorage(cfg)
	if err != nil {
		logrus.Fatalf("Failed to initialize storage: %v", err)
	}

	// Initialize scheduler
	sched := scheduler.New(cfg, db, store)
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
		w.Write([]byte(`{"status":"healthy","version":"` + Version + `"}`))
	})

	// Metrics endpoint (Prometheus compatible)
	r.Get("/metrics", handler.MetricsHandler(cfg, db, store))

	// API v1 routes
	r.Route("/v1", func(r chi.Router) {
		// Authentication routes (no auth required)
		r.Post("/auth/login", handler.Login(cfg, db))
		r.Post("/auth/register", handler.Register(cfg, db))
		r.Post("/auth/refresh", handler.RefreshToken(cfg, db))

		// Public routes (no auth required)
		r.Get("/registries", handler.ListPublicRegistries(db))
		r.Get("/support/docs", handler.ListDocs(docsAssets))
		r.Get("/support/docs/{id}", handler.GetDoc(docsAssets))

		// Authenticated routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate(cfg, db))

			// User management
			r.Get("/users/me", handler.GetCurrentUser(db))
			r.Put("/users/me", handler.UpdateCurrentUser(db))
			r.Post("/users/me/password", handler.ChangePassword(cfg, db))
			r.Get("/users/{id}", handler.GetUser(db))

			// Token management
			r.Get("/tokens", handler.ListTokens(db))
			r.Post("/tokens", handler.CreateToken(cfg, db))
			r.Delete("/tokens/{id}", handler.DeleteToken(db))
			r.Post("/tokens/{id}/rotate", handler.RotateToken(cfg, db))

			// Organization management
			r.Get("/organizations", handler.ListOrganizations(db))
			r.Post("/organizations", handler.CreateOrganization(db))
			r.Get("/organizations/{name}", handler.GetOrganization(db))
			r.Put("/organizations/{name}", handler.UpdateOrganization(db))
			r.Delete("/organizations/{name}", handler.DeleteOrganization(db))
			r.Post("/organizations/{name}/members", handler.AddOrgMember(db))
			r.Get("/organizations/{name}/members", handler.ListOrgMembers(db))
			r.Delete("/organizations/{name}/members/{username}", handler.RemoveOrgMember(db))

			// Registry management
			r.Get("/registries", handler.ListRegistries(db))
			r.Post("/registries", handler.CreateRegistry(db, store))
			r.Get("/registries/{name}", handler.GetRegistry(db))
			r.Put("/registries/{name}", handler.UpdateRegistry(db))
			r.Delete("/registries/{name}", handler.DeleteRegistry(db, store))

			// Repository management
			r.Get("/registries/{registry}/repositories", handler.ListRepositories(db))
			r.Post("/registries/{registry}/repositories", handler.CreateRepository(db))
			r.Get("/registries/{registry}/repositories/{repo}", handler.GetRepository(db))
			r.Put("/registries/{registry}/repositories/{repo}", handler.UpdateRepository(db))
			r.Delete("/registries/{registry}/repositories/{repo}", handler.DeleteRepository(db, store))

			// Tag management
			r.Get("/registries/{registry}/repositories/{repo}/tags", handler.ListTags(db))
			r.Get("/registries/{registry}/repositories/{repo}/tags/{tag}", handler.GetTag(db))
			r.Delete("/registries/{registry}/repositories/{repo}/tags/{tag}", handler.DeleteTag(db, store))
			r.Post("/registries/{registry}/repositories/{repo}/tags/{tag}/scan", handler.ScanTag(cfg, db, store))
			r.Get("/registries/{registry}/repositories/{repo}/tags/{tag}/scan-results", handler.GetScanResults(db))

			// Support system
			r.Get("/support/tickets", handler.ListTickets(db))
			r.Post("/support/tickets", handler.CreateTicket(db))
			r.Get("/support/tickets/{id}", handler.GetTicket(db))
			r.Post("/support/tickets/{id}/comments", handler.AddTicketComment(db))

			// Admin routes
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireAdmin)

				r.Get("/admin/users", handler.AdminListUsers(db))
				r.Post("/admin/users", handler.AdminCreateUser(cfg, db))
				r.Put("/admin/users/{id}", handler.AdminUpdateUser(db))
				r.Delete("/admin/users/{id}", handler.AdminDeleteUser(db))

				r.Get("/admin/organizations", handler.AdminListOrganizations(db))
				r.Get("/admin/registries", handler.AdminListRegistries(db))
				r.Get("/admin/system/stats", handler.AdminSystemStats(cfg, db, store))
				r.Post("/admin/system/cleanup", handler.AdminCleanup(db, store))
			})
		})
	})

	// Docker Registry V2 API
	r.Route("/v2", func(r chi.Router) {
		r.Get("/", handler.RegistryVersion(Version))

		// Optional auth for public registries
		r.Use(middleware.OptionalAuth(cfg, db))

		r.Get("/{registry}/{repository}/manifests/{reference}", handler.GetManifest(db, store))
		r.Put("/{registry}/{repository}/manifests/{reference}", handler.PutManifest(db, store))
		r.Delete("/{registry}/{repository}/manifests/{reference}", handler.DeleteManifest(db, store))

		r.Get("/{registry}/{repository}/blobs/{digest}", handler.GetBlob(db, store))
		r.Delete("/{registry}/{repository}/blobs/{digest}", handler.DeleteBlob(db, store))

		r.Post("/{registry}/{repository}/blobs/uploads/", handler.StartBlobUpload(db, store))
		r.Get("/{registry}/{repository}/blobs/uploads/{uuid}", handler.GetBlobUploadStatus(db, store))
		r.Patch("/{registry}/{repository}/blobs/uploads/{uuid}", handler.UploadBlobChunk(db, store))
		r.Put("/{registry}/{repository}/blobs/uploads/{uuid}", handler.CompleteBlobUpload(db, store))
		r.Delete("/{registry}/{repository}/blobs/uploads/{uuid}", handler.CancelBlobUpload(db, store))

		r.Get("/{registry}/{repository}/tags/list", handler.ListTagsV2(db))
	})

	// Swagger UI
	r.Get("/swagger/*", handler.SwaggerUI())

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
