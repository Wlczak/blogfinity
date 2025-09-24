package database

import (
	"github.com/Wlczak/blogfinity/database/models"
	"github.com/Wlczak/blogfinity/logger"
	"github.com/glebarez/sqlite" // Pure go SQLite driver, checkout https://github.com/glebarez/sqlite for details
	"gorm.io/gorm"
)

// github.com/mattn/go-sqlite3
func GetDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("db/gorm.db"), &gorm.Config{})

	return db, err
}

func Migrate(db *gorm.DB) {
	zap := logger.GetLogger()
	err := db.AutoMigrate(&models.Article{})
	if err != nil {
		zap.Error(err.Error())
	}

	err = db.AutoMigrate(&models.Server{})
	if err != nil {
		zap.Error(err.Error())
	}
}
