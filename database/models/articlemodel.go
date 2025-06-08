package models

import (
	"gorm.io/gorm"
)

type Article struct {
	gorm.Model
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	Author string `json:"author"`
}

func GetArticles(db *gorm.DB, limit int) []Article {

	var articles []Article

	db.Select("Title", "Body", "Author", "CreatedAt", "ID").Limit(500).Find(&articles)

	return articles
}

func GetArticleById(db *gorm.DB, id int) Article {
	var article Article
	db.First(&article, id)
	return article
}

func (a *Article) Create(db *gorm.DB) {
	db.Create(&a)
}

func (a *Article) Update(db *gorm.DB) {
	db.Save(&a)
}

func (a *Article) HasBody(db *gorm.DB) bool {
	db.First(&a, a.ID)

	return a.Body != ""
}
