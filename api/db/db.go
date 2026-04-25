package db

import (
	"log"
	"os"
	"path/filepath"

	"godelion/models"

	"golang.org/x/crypto/bcrypt"
	"github.com/glebarez/sqlite" // pure Go SQLite driver
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	// Use current working directory + /godelion/ instead of root /godelion/ to avoid permission issues
	cwd, _ := os.Getwd()
	dbDir := filepath.Join(cwd, "godelion")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		log.Fatalf("Failed to create db directory %s: %v", dbDir, err)
	}
	dbPath := filepath.Join(dbDir, "godelion.db")

	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	err = DB.AutoMigrate(&models.User{}, &models.Container{}, &models.GatewayRule{}, &models.AuditLog{})
	if err != nil {
		log.Fatalf("Failed to auto migrate database: %v", err)
	}

	seedAdmin()
}

func seedAdmin() {
	var count int64
	DB.Model(&models.User{}).Count(&count)
	if count == 0 {
		hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		admin := models.User{
			ID:           "u_admin",
			Username:     "admin",
			PasswordHash: string(hash),
			Role:         "admin",
		}
		DB.Create(&admin)
		log.Println("Created default admin user (admin / admin123)")
	}
}
