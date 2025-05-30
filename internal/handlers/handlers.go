package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"

	"trade/internal/database"
	db "trade/internal/db/sqlc"
	"trade/internal/models"

	"github.com/gorilla/mux"
)

type UserHandlers struct {
	db *database.Database
}

func NewUserHandler(db *database.Database) *UserHandlers {
	return &UserHandlers{db: db}
}

func (h *UserHandlers) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Email == "" {
		http.Error(w, "Name and email are required", http.StatusBadRequest)
		return
	}

	params := db.CreateUserParams{
		Name:  req.Name,
		Email: req.Email,
	}

	user, err := h.db.Queries.CreateUser(ctx, params)
	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandlers) ListUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	users, err := h.db.Queries.ListUsers(ctx)
	if err != nil {
		http.Error(w, "Error getting list of users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)

}

func (h *UserHandlers) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	idStr := vars["id"]

	ID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.db.Queries.GetUser(ctx, int32(ID))
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandlers) UpdateUser(w http.ResponseWriter, r *http.Request) {

}

func (h *UserHandlers) DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	idStr := vars["id"]

	ID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	err = h.db.Queries.DeleteUser(ctx, int32(ID))
	if err != nil {
		http.Error(w, "Error deleting the user", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusNoContent)
}

func helloWorld(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("web/templates/index.html")
	if err != nil {
		http.Error(w, "Failed to parse template", http.StatusInternalServerError)
	}

	data := models.PageData{
		Title:   "Trading Dashboard",
		Message: "Welcome!",
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
		return
	}
}

func apiHelloWorld(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := models.APIResponse{
		Message: "Hello from API!",
		Status:  "success",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		return
	}
}

func SetupRoutes(db *database.Database) *mux.Router {
	userHandler := NewUserHandler(db)
	r := mux.NewRouter()
	r.HandleFunc("/", helloWorld).Methods("GET")
	r.HandleFunc("/api/hello", apiHelloWorld).Methods("GET")

	r.HandleFunc("/api/users", userHandler.CreateUser).Methods("POST")
	r.HandleFunc("/api/users", userHandler.ListUsers).Methods("GET")
	r.HandleFunc("/api/users/{id}", userHandler.GetUser).Methods("GET")
	// r.HandleFunc("/api/users/{id}", UpdateUserHandler).Methods("PUT")
	r.HandleFunc("/api/users/{id}", userHandler.DeleteUser).Methods("DELETE")

	return r
}
