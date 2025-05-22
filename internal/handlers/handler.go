// internal/handlers/handlers.go
package handlers

import (
	"html/template"
	"net/http"
	"trade/internal/models"

	"github.com/gorilla/mux"
)

// SetupRoutes configures all the routes for the application
func SetupRoutes(r *mux.Router) {
	r.HandleFunc("/", HomeHandler).Methods("GET")
	r.HandleFunc("/api/hello", APIHelloHandler).Methods("GET")
}

// HomeHandler serves the main page
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("web/templates/index.html")
	if err != nil {
		http.Error(w, "Template parsing error", http.StatusInternalServerError)
		return
	}

	data := models.PageData{
		Title:   "Trading Dashboard",
		Message: "Hello World from Go Server!",
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
		return
	}
}

// APIHelloHandler returns JSON response
func APIHelloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Hello World from Trading Dashboard API!", "status": "success"}`))
}
