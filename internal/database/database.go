package database

import (
	"log"
	"tripshare/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func New(dbPath string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get underlying sql.DB: %v", err)
	}
	sqlDB.Exec("PRAGMA journal_mode=WAL;")
	sqlDB.Exec("PRAGMA foreign_keys=ON;")

	if err := db.AutoMigrate(&models.Member{}, &models.Trip{}, &models.Payment{}); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	return db
}
