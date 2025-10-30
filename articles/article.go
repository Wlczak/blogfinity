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
	"github.com/google/uuid"
)

func HandleArticle(w http.ResponseWriter, r *http.Request, queue chan ai.AiQuery) {
	zap := logger.GetLogger()
	type PageData struct {
		Article models.Article
		Year    int
		Query   string
		Models  []string
		Model   string
	}

	query := r.URL.Query().Get("q")
	model := r.URL.Query().Get("model")
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

	rid := uuid.New()

	aiQuery := ai.AiQuery{
		Query:     article.Title,
		Type:      "body",
		Article:   article,
		Model:     model,
		RequestId: rid.String(),
	}
	queue <- aiQuery

	err = tmpl.Execute(w, PageData{
		Article: article,
		Year:    time.Now().Year(),
		Query:   query,
		Models:  ai.GetModels(),
		Model:   model,
	})

	if err != nil {
		zap.Error(err.Error())
	}
}
