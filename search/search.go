package search

import (
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/Wlczak/blogfinity/database"
	"github.com/Wlczak/blogfinity/database/models"
	"github.com/Wlczak/blogfinity/logger"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

func HandleSearch(w http.ResponseWriter, r *http.Request) {
	zap := logger.GetLogger()

	type PageData struct {
		Query string
		Year  int
	}

	query := r.URL.Query().Get("q")

	tmplf, err := template.ParseFiles("templates/search.tmpl")
	if err != nil {
		zap.Error(err.Error())
	}
	tmpl := template.Must(tmplf, err)

	err = tmpl.Execute(w, PageData{
		Query: query,
		Year:  time.Now().Year(),
	})

	if err != nil {
		zap.Error(err.Error())
	}
	searchResults := search(query)

	println("Found " + strconv.Itoa(len(searchResults)) + " results for \"" + query + "\"")
	for _, result := range searchResults {
		println(result.Title)
	}
}

func search(query string) []models.Article {
	zap := logger.GetLogger()
	db, err := database.GetDB()

	if err != nil {
		zap.Error(err.Error())
	}

	articles := models.GetArticles(db, 500)

	return rankedFuzzySearch(articles, query)
}

func rankedFuzzySearch(articles []models.Article, query string) []models.Article {
	var result []models.Article
	var titles []string
	var titleMap map[string]models.Article = make(map[string]models.Article)

	for _, aritcle := range articles {
		titles = append(titles, aritcle.Title)
		titleMap[aritcle.Title] = aritcle
	}

	ranks := fuzzy.RankFindNormalizedFold(query, titles)

	for _, rank := range ranks {
		result = append(result, titleMap[rank.Target])
	}

	return result
}
