package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"runtime"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		render(w, "goajaxtest.html")
	})

	fmt.Println("Starting front end service on port 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Panic(err)
	}
}

func render(w http.ResponseWriter, t string) {
	// Get the file path of the currently executing code file
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		http.Error(w, "Unable to determine the current file path", http.StatusInternalServerError)
		return
	}

	// Build the absolute path to the template file
	templatePath := filepath.Join(filepath.Dir(filename), "templates", t)

	// Parse and execute the template
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
