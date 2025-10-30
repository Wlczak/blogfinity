package statistics

import (
	"encoding/json"
	"net/http"
	"text/template"
	"time"

	"github.com/Wlczak/blogfinity/ai"
	"github.com/Wlczak/blogfinity/logger"
)

type PageData struct {
	ServerOnline bool
	Models       []string
	Model        string
	Year         int
	Ongoing      Ongoing
	TotalSlots   int
}

type Ongoing struct {
	ArticleRequests int
	TitleRequests   int
	TotalRequests   int
}

func HandleStats(w http.ResponseWriter, r *http.Request, queue *ai.Queue) {
	zap := logger.GetLogger()

	tmplf, err := template.ParseFiles("templates/stats.tmpl")
	if err != nil {
		zap.Error(err.Error())
	}
	tmpl := template.Must(tmplf, err)

	articleCount, titleCount := getStats(queue)

	model := r.URL.Query().Get("model")

	err = tmpl.Execute(w, PageData{
		Year:         time.Now().Year(),
		Models:       ai.GetModels(),
		Model:        model,
		ServerOnline: ai.IsServerOnline(),
		TotalSlots:   ai.MaxAiQueueSize,
		Ongoing: Ongoing{
			ArticleRequests: articleCount,
			TitleRequests:   titleCount,
			TotalRequests:   articleCount + titleCount,
		},
	})
	if err != nil {
		zap.Error(err.Error())
	}

}

func HandleStatsApi(w http.ResponseWriter, r *http.Request, queue *ai.Queue) {
	zap := logger.GetLogger()
	articleCount, titleCount := getStats(queue)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(Ongoing{
		ArticleRequests: articleCount,
		TitleRequests:   titleCount,
		TotalRequests:   articleCount + titleCount,
	})
	if err != nil {
		zap.Error(err.Error())
	}
}

func getStats(queue *ai.Queue) (articleCount int, titleCount int) {
	articleCount = 0
	titleCount = 0
	for _, val := range queue.Copy() {
		if val.Type == "body" {
			articleCount++
		} else {
			titleCount++
		}
	}
	return articleCount, titleCount
}
