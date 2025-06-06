package search

import (
	"html/template"
	"net/http"
	"time"

	"github.com/Wlczak/blogfinity/logger"
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
}
