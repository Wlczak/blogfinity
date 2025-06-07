package models

import (
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

	var articles []Article

	db.Select("Title").Limit(500).Find(&articles)

	return articles
}
