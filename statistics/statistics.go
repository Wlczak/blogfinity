package statistics

import (
	"net/http"
	"text/template"
	"time"

	"github.com/Wlczak/blogfinity/ai"
	"github.com/Wlczak/blogfinity/logger"
)

func HandleStats(w http.ResponseWriter, r *http.Request) {
	zap := logger.GetLogger()

	tmplf, err := template.ParseFiles("templates/stats.tmpl")
	if err != nil {
		zap.Error(err.Error())
	}
	tmpl := template.Must(tmplf, err)

	type PageData struct {
		ServerOnline bool
		Models       []string
		Model        string
		Year         int
		Ongoing      struct {
			ArticleRequests int
			TitleRequests   int
		}
	}
	model := r.URL.Query().Get("model")
	err = tmpl.Execute(w, PageData{
		Year:         time.Now().Year(),
		Models:       ai.GetModels(),
		Model:        model,
		ServerOnline: ai.IsServerOnline(),
		Ongoing: struct {
			ArticleRequests int
			TitleRequests   int
		}{
			ArticleRequests: 2,
			TitleRequests:   56,
		},
	})
	if err != nil {
		zap.Error(err.Error())
	}
}
