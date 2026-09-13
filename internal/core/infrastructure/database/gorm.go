package database

import (
	"log"
	"os"
	"time"

	"github.com/fiqrikm18/quran-app/internal/core/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDatabaseConnection(config *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(config.Database.URL), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold: 200 * time.Millisecond,
				LogLevel:      logger.Warn,
				Colorful:      true,
			},
		),
	})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(config.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(30 * time.Second)

	return db, nil
}
