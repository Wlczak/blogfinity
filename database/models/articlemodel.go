package models

import (
	"github.com/Wlczak/blogfinity/database"
	"github.com/Wlczak/blogfinity/logger"
	"gorm.io/gorm"
)

type Article struct {
	gorm.Model
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	Author int    `json:"author"`
}

func GetArticles(db *gorm.DB, limit int) []Article {
	zap := logger.GetLogger()
	db, err := database.GetDB()

	var articles []Article

	db.Select("Title").Limit(500).Find(&articles)

	if err != nil {
		zap.Error(err.Error())
	}
	return articles
}
