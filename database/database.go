package database

import (
	"github.com/Wlczak/blogfinity/database/models"
	"github.com/glebarez/sqlite" // Pure go SQLite driver, checkout https://github.com/glebarez/sqlite for details
	"gorm.io/gorm"
)

// github.com/mattn/go-sqlite3
func GetDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})

	return db, err
}

func Migrate(db *gorm.DB) {
	db.AutoMigrate(&models.Article{})
}
