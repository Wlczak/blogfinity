package main

import (
	"fmt"
	"net"
	"net/http"
	"text/template"
	"time"

	"github.com/Wlczak/blogfinity/ai"
	"github.com/Wlczak/blogfinity/articles"
	"github.com/Wlczak/blogfinity/database"
	"github.com/Wlczak/blogfinity/logger"
	"github.com/Wlczak/blogfinity/search"
	"github.com/Wlczak/blogfinity/statistics"
	"github.com/Wlczak/blogfinity/ws"
	"github.com/joho/godotenv"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	zap := logger.GetLogger()

	model := r.URL.Query().Get("model")

	type PageData struct {
		Year         int
		Models       []string
		Model        string
		ServerOnline bool
	}
	var err error
	tmplf, err := template.ParseFiles("templates/index.tmpl")
	if err != nil {
		zap.Error(err.Error())
	}
	tmpl := template.Must(tmplf, err)

	data := PageData{
		Year:         time.Now().Year(),
		Models:       ai.GetModels(),
		Model:        model,
		ServerOnline: ai.IsServerOnline(),
	}
	err = tmpl.Execute(w, data)

	if err != nil {
		zap.Error(err.Error())
	}
}

func main() {
	zap := logger.GetLogger()

	err := godotenv.Load("conf/.env")
	if err != nil {
		zap.Error(err.Error())
	}

	db, err := database.GetDB()
	if err != nil {
		zap.Error(err.Error())
	}
	database.Migrate(db)

	queueTransport := make(chan ai.AiQuery)
	var queue = ai.NewQueue()
	go ai.HandleQueue(queueTransport, queue)

	address := "0.0.0.0:8080"
	listener, err := net.Listen("tcp", address)

	if err != nil {
		zap.Error(err.Error())
	}
	http.Handle("/", http.HandlerFunc(indexHandler))

	http.Handle("/static/icon.png", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "static/icon.png") }))

	http.Handle("/search", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { search.HandleSearch(w, r, queueTransport) }))

	http.Handle("/article/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { articles.HandleArticle(w, r, queueTransport) }))

	http.Handle("/ws/article/{articleId}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { ws.HandleWsArticle(w, r, queue) }))

	http.Handle("/stats", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { statistics.HandleStats(w, r, queue) }))

	http.Handle("/stats/api", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { statistics.HandleStatsApi(w, r, queue) }))

	http.Handle("/sitemap.xml", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { articles.HandleSitemap(w, r) }))

	fmt.Println("Listening on ", address)

	err = http.Serve(listener, nil)

	if err != nil {
		zap.Error(err.Error())
	}
}
