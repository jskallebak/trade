package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"

	"trade/internal/auth"
	"trade/internal/database"
	db "trade/internal/db/sqlc"
	"trade/internal/middleware"
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
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Email == "" || req.Password == "" {
		http.Error(w, "Name, email and password are required", http.StatusBadRequest)
		return
	}

	password_hash, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	params := db.CreateUserParams{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: password_hash,
	}

	user, err := h.db.Queries.CreateUser(ctx, params)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
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
	ctx := r.Context()

	vars := mux.Vars(r)
	idStr := vars["id"]

	Id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Email == "" {
		http.Error(w, "Name and Email are required", http.StatusBadRequest)
		return
	}

	params := db.UpdateUserParams{
		ID:    int32(Id),
		Name:  req.Name,
		Email: req.Email,
	}

	if req.Password != "" {
		passwordHash, err := auth.HashPassword(req.Password)
		if err != nil {
			http.Error(w, "Error hashing password", http.StatusInternalServerError)
			return
		}
		params.PasswordHash = passwordHash
	}

	user, err := h.db.Queries.UpdateUser(ctx, params)
	if err != nil {
		http.Error(w, "Error updating the user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
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

func loginPageHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("web/templates/login.html")
	if err != nil {
		http.Error(w, "Failed to parse template", http.StatusInternalServerError)
		return
	}

	data := models.PageData{
		Title: "Trading Dashboard",
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
	}
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("web/templates/dashboard.html") // You'll create this
	if err != nil {
		http.Error(w, "Failed to parse template", http.StatusInternalServerError)
		return
	}

	data := models.PageData{
		Title:   "Dashboard",
		Message: "Welcome to your dashboard!",
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
	}
}

func helloWorld(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("web/templates/index.html")
	if err != nil {
		http.Error(w, "Failed to parse template", http.StatusInternalServerError)
		return
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

func (h UserHandlers) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and Password are required", http.StatusBadRequest)
		return
	}

	user, err := h.db.Queries.GetUserByEmail(ctx, req.Email)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := auth.CheckPassword(req.Password, user.PasswordHash); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateJWT(user.ID, user.Email)
	if err != nil {
		http.Error(w, "Error generating JWT", http.StatusInternalServerError)
		return
	}

	// Set JWT token as HTTP-only cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   24 * 60 * 60, // 24 hours
	})

	type LoginResponse struct {
		ID      int32  `json:"id"`
		Name    string `json:"name"`
		Email   string `json:"email"`
		Token   string `json:"token"`
		Message string `json:"message"`
	}

	loginResponse := LoginResponse{
		ID:      user.ID,
		Name:    user.Name,
		Email:   user.Email,
		Token:   token,
		Message: "login successful",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(loginResponse)

}

func (h UserHandlers) GetProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := ctx.Value("userID").(int32)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	user, err := h.db.Queries.GetUser(ctx, userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h UserHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	// Clear the authentication cookie
	http.SetCookie(w, &http.Cookie{
		Name:   "auth_token",
		Value:  "",
		MaxAge: -1,
		Path:   "/",
	})

	// For API calls, return JSON response
	if r.Header.Get("Content-Type") == "application/json" || r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Logged out successfully",
		})
		return
	}

	// For web page calls, redirect to login
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// Dashboard API endpoints
func (h UserHandlers) GetDashboardMetrics(w http.ResponseWriter, r *http.Request) {
	// Mock dashboard metrics - in a real app, this would come from your trading system
	metrics := map[string]interface{}{
		"total_pnl":      5000.00,
		"annualized_roi": 18.5,
		"max_drawdown":   12.3,
		"uptime":         99.8,
		"active_bots":    2,
		"total_trades":   540,
		"win_rate":       52.3,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func (h UserHandlers) GetBotStats(w http.ResponseWriter, r *http.Request) {
	// Mock bot statistics
	botStats := []map[string]interface{}{
		{
			"bot_name":      "Alpha1",
			"status":        "RUNNING",
			"win_rate":      52,
			"profit_factor": 2.6,
			"trades":        340,
			"pnl":           5000,
		},
		{
			"bot_name":      "Beta2",
			"status":        "STOPPED",
			"win_rate":      49,
			"profit_factor": 2.1,
			"trades":        200,
			"pnl":           -300,
		},
		{
			"bot_name":      "Gamma3",
			"status":        "RUNNING",
			"win_rate":      58,
			"profit_factor": 3.1,
			"trades":        150,
			"pnl":           2500,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(botStats)
}

func (h UserHandlers) GetPositions(w http.ResponseWriter, r *http.Request) {
	// Mock open positions
	positions := []map[string]interface{}{
		{
			"trade_id": "#25678",
			"bot":      "Alpha1",
			"position": "LONG",
			"entry":    "20000",
			"current":  "22000",
			"pnl":      2000,
			"time":     "2h 45m",
			"symbol":   "BTC/USD",
		},
		{
			"trade_id": "#25679",
			"bot":      "Gamma3",
			"position": "SHORT",
			"entry":    "1850",
			"current":  "1820",
			"pnl":      300,
			"time":     "1h 15m",
			"symbol":   "ETH/USD",
		},
		{
			"trade_id": "#25680",
			"bot":      "Alpha1",
			"position": "LONG",
			"entry":    "0.52",
			"current":  "0.48",
			"pnl":      -400,
			"time":     "3h 20m",
			"symbol":   "ADA/USD",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(positions)
}

func SetupRoutes(db *database.Database) *mux.Router {
	userHandler := NewUserHandler(db)
	r := mux.NewRouter()

	// Apply middleware to ALL routes (will skip auth for login routes)
	r.Use(middleware.AuthMiddleware)

	// Static files (public)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static/"))))

	// Public routes (middleware will skip auth for these)
	r.HandleFunc("/login", loginPageHandler).Methods("GET")               // Login page
	r.HandleFunc("/api/login", userHandler.Login).Methods("POST")         // Login API
	r.HandleFunc("/api/register", userHandler.CreateUser).Methods("POST") // Registration API (optional)

	// Protected web pages (require authentication)
	r.HandleFunc("/", helloWorld).Methods("GET")                // Dashboard/home
	r.HandleFunc("/dashboard", dashboardHandler).Methods("GET") // Dashboard page
	// r.HandleFunc("/profile", profilePageHandler).Methods("GET") // Profile page

	// Logout route (can be accessed both authenticated and unauthenticated)
	r.HandleFunc("/logout", userHandler.Logout).Methods("GET", "POST")
	r.HandleFunc("/api/logout", userHandler.Logout).Methods("POST")

	// Protected API routes (require authentication)
	r.HandleFunc("/api/hello", apiHelloWorld).Methods("GET") // Test API endpoint
	r.HandleFunc("/api/profile", userHandler.GetProfile).Methods("GET")
	r.HandleFunc("/api/users", userHandler.ListUsers).Methods("GET")
	r.HandleFunc("/api/users/{id}", userHandler.GetUser).Methods("GET")
	r.HandleFunc("/api/users/{id}", userHandler.UpdateUser).Methods("PUT")
	r.HandleFunc("/api/users/{id}", userHandler.DeleteUser).Methods("DELETE")

	// Dashboard API endpoints
	r.HandleFunc("/api/dashboard/metrics", userHandler.GetDashboardMetrics).Methods("GET")
	r.HandleFunc("/api/dashboard/bot-stats", userHandler.GetBotStats).Methods("GET")
	r.HandleFunc("/api/dashboard/positions", userHandler.GetPositions).Methods("GET")

	return r
}
