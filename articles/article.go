package articles

import (
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/Wlczak/blogfinity/ai"
	"github.com/Wlczak/blogfinity/database"
	"github.com/Wlczak/blogfinity/database/models"
	"github.com/Wlczak/blogfinity/logger"
)

func HandleArticle(w http.ResponseWriter, r *http.Request, queue chan ai.AiQuery) {
	zap := logger.GetLogger()
	type PageData struct {
		Article models.Article
		Year    int
	}

	urlParts := strings.Split(r.URL.Path, "/")
	articleId := urlParts[len(urlParts)-1]
	//fmt.Println(articleId)

	tmplf, err := template.ParseFiles("templates/article.tmpl")

	if err != nil {
		zap.Error(err.Error())
	}
	tmpl := template.Must(tmplf, err)

	db, _ := database.GetDB()

	articleIdInt, err := strconv.Atoi(articleId)
	if err != nil {
		zap.Error(err.Error())
		return
	}

	article := models.GetArticleById(db, articleIdInt)

	aiQuery := ai.AiQuery{
		Query:   article.Title,
		Type:    "body",
		Article: article,
	}
	queue <- aiQuery

	err = tmpl.Execute(w, PageData{
		Article: article,
		Year:    time.Now().Year(),
	})

	if err != nil {
		zap.Error(err.Error())
	}
}
