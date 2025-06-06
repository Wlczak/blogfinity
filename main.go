package main

import (
	"net"
	"net/http"
	"text/template"
	"time"

	"github.com/Wlczak/blogfinity/logger"
	"github.com/Wlczak/blogfinity/search"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	zap := logger.GetLogger()
	type PageData struct {
		Year int
	}
	var err error
	tmplf, err := template.ParseFiles("templates/index.tmpl")
	if err != nil {
		zap.Error(err.Error())
	}
	tmpl := template.Must(tmplf, err)

	data := PageData{
		Year: time.Now().Year(),
	}
	err = tmpl.Execute(w, data)

	if err != nil {
		zap.Error(err.Error())
	}
}

func main() {
	zap := logger.GetLogger()

	listener, err := net.Listen("tcp", "localhost:8080")

	if err != nil {
		zap.Error(err.Error())
	}
	http.Handle("/", http.HandlerFunc(indexHandler))

	http.Handle("/search", http.HandlerFunc(search.HandleSearch))
	println("Listening on http://localhost:8080")

	err = http.Serve(listener, nil)

	if err != nil {
		zap.Error(err.Error())
	}
}
