package main

import (
	"net"
	"net/http"
	"text/template"
	"time"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {

	type PageData struct {
		Year int
	}

	tmpl := template.Must(template.ParseFiles("index.tmpl"))

	data := PageData{
		Year: time.Now().Year(),
	}
	tmpl.Execute(w, data)
}

func main() {
	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}

	if err != nil {
		panic(err)
	}
	http.Handle("/", http.HandlerFunc(indexHandler))

	println("Listening on http://localhost:8080")

	http.Serve(listener, nil)
}
