package search

import (
	"html/template"
	"net/http"
	"time"

	"github.com/Wlczak/blogfinity/ai"
	"github.com/Wlczak/blogfinity/database"
	"github.com/Wlczak/blogfinity/database/models"
	"github.com/Wlczak/blogfinity/logger"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

func HandleSearch(w http.ResponseWriter, r *http.Request, queue chan *ai.AiQuery) {
	zap := logger.GetLogger()

	type PageData struct {
		Query        string
		Year         int
		Results      []models.Article
		Models       []string
		Model        string
		ServerOnline bool
	}

	query := r.URL.Query().Get("q")
	model := r.URL.Query().Get("model")

	tmplf, err := template.ParseFiles("templates/search.tmpl")
	if err != nil {
		zap.Error(err.Error())
	}
	tmpl := template.Must(tmplf, err)

	searchResults := search(query)

	// println("Found " + strconv.Itoa(len(searchResults)) + " results for \"" + query + "\"")
	// for _, result := range searchResults {
	// 	println(result.Title)
	// }
	var resultCount = ai.MaxSearchResults
	if len(searchResults) < resultCount {
		for i := len(searchResults); i < resultCount; i++ {
			searchResults = append(searchResults, models.Article{})
			aiQuery := ai.AiQuery{
				Query:   query,
				Type:    "title",
				Article: models.Article{},
				Model:   model,
			}
			if ai.IsServerOnline() {
				queue <- &aiQuery
			}
		}
	}
	// fmt.Println(model)
	err = tmpl.Execute(w, PageData{
		Query:        query,
		Year:         time.Now().Year(),
		Results:      searchResults,
		Models:       ai.GetModels(),
		Model:        model,
		ServerOnline: ai.IsServerOnline(),
	})

	if err != nil {
		zap.Error(err.Error())
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
	titleMap := make(map[string]models.Article)

	for _, aritcle := range articles {
		titles = append(titles, aritcle.Title)
		titleMap[aritcle.Title] = aritcle
	}

	ranks := fuzzy.RankFindNormalizedFold(query, titles)

	for _, rank := range ranks {
		if rank.Distance < ai.SearchDistance {
			result = append(result, titleMap[rank.Target])
		}
		//fmt.Println(rank.Distance)
	}

	return result
}
