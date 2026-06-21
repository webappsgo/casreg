package model

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/casapps/casreg/src/config"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDatabase initializes the database connection
func InitDatabase(cfg *config.Config) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch cfg.Database.Type {
	case "sqlite":
		dsn := cfg.Database.URL
		if dsn == "" {
			dsn = "./casreg.db"
		}
		dialector = sqlite.Open(dsn)

	case "postgres":
		dsn := cfg.Database.URL
		if dsn == "" {
			dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
				cfg.Database.Host,
				cfg.Database.Port,
				cfg.Database.User,
				cfg.Database.Password,
				cfg.Database.Name,
				cfg.Database.SSLMode,
			)
		}
		dialector = postgres.Open(dsn)

	case "mysql":
		dsn := cfg.Database.URL
		if dsn == "" {
			dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
				cfg.Database.User,
				cfg.Database.Password,
				cfg.Database.Host,
				cfg.Database.Port,
				cfg.Database.Name,
			)
		}
		dialector = mysql.Open(dsn)

	case "sqlserver":
		dsn := cfg.Database.URL
		if dsn == "" {
			dsn = fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
				cfg.Database.User,
				cfg.Database.Password,
				cfg.Database.Host,
				cfg.Database.Port,
				cfg.Database.Name,
			)
		}
		dialector = sqlserver.Open(dsn)

	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Database.Type)
	}

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	if cfg.Server.Debug {
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	}

	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.Database.PoolSize)
	sqlDB.SetMaxIdleConns(cfg.Database.PoolSize)

	return db, nil
}

// RunMigrations runs database migrations
func RunMigrations(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&Token{},
		&Organization{},
		&OrganizationMember{},
		&Registry{},
		&Repository{},
		&Tag{},
		&Manifest{},
		&Blob{},
		&Layer{},
		&Ticket{},
		&TicketComment{},
		&Notification{},
		&Issue{},
		&IssueComment{},
		&IssueLabel{},
		&Quota{},
		&AuditLog{},
		&ScanResult{},
		&Vulnerability{},
		&SignatureVerification{},
		&Webhook{},
		&WebhookDelivery{},
	)
}

// CreateFirstAdmin creates the first admin user if no users exist
func CreateFirstAdmin(db *gorm.DB) error {
	var count int64
	if err := db.Model(&User{}).Count(&count).Error; err != nil {
		return err
	}

	if count == 0 {
		rawBytes := make([]byte, 16)
		if _, err := rand.Read(rawBytes); err != nil {
			return fmt.Errorf("generate admin password: %w", err)
		}
		generatedPassword := hex.EncodeToString(rawBytes)

		admin := &User{
			Username:  "admin",
			Email:     "admin@casreg.local",
			FirstName: "System",
			LastName:  "Administrator",
			Role:      RoleAdmin,
			IsActive:  true,
			Theme:     "dark",
		}

		if err := admin.SetPassword(generatedPassword); err != nil {
			return err
		}

		if err := db.Create(admin).Error; err != nil {
			return err
		}

		fmt.Println("===========================================")
		fmt.Println("First admin user created:")
		fmt.Println("  Username: admin")
		fmt.Printf("  Password: %s\n", generatedPassword)
		fmt.Println("  Save this password — it will not be shown again.")
		fmt.Println("===========================================")
	}

	return nil
}
