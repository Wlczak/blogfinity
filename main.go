package main

import (
	"net"
	"net/http"
	"text/template"
	"time"

	"github.com/Wlczak/blogfinity/logger"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	zap := logger.GetLogger()
	type PageData struct {
		Year int
	}

	tmpl := template.Must(template.ParseFiles("index.tmpl"))

	data := PageData{
		Year: time.Now().Year(),
	}
	err := tmpl.Execute(w, data)

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

	println("Listening on http://localhost:8080")

	err = http.Serve(listener, nil)

	if err != nil {
		zap.Error(err.Error())
	}
}
