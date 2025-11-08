package database

import (
	"fmt"
	"log"
	"os"
	"registration-system/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect() {
	var err error

	// Get SSL mode from environment, default to disable for local development
	sslMode := os.Getenv("DB_SSL_MODE")
	if sslMode == "" {
		sslMode = "disable"
	}

	// Build DSN with SSL configuration
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		sslMode,
	)

	// Add channel binding if specified and SSL mode requires it
	// Note: channel_binding should be added with space separator in DSN format
	channelBinding := os.Getenv("DB_CHANNEL_BINDING")
	if channelBinding != "" && (sslMode == "require" || sslMode == "verify-ca" || sslMode == "verify-full") {
		dsn += fmt.Sprintf(" channel_binding=%s", channelBinding)
	}

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Database connected successfully")
}

func Migrate() {
	err := DB.AutoMigrate(
		&models.User{},
		&models.Province{},
		&models.District{},
		&models.SubDistrict{},
		&models.Registration{},
		&models.TeacherRegistration{},
		&models.Transaction{},
		&models.ActivityLog{},
		&models.DeviceLog{},
	)

	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	log.Println("Database migrated successfully")
}
