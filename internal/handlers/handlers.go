package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"

	"trade/internal/auth"
	"trade/internal/database"
	db "trade/internal/db/sqlc"
	"trade/internal/middleware"
	"trade/internal/models"

	"github.com/gorilla/mux"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
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
	type Stats struct {
		ID           int32   `json:"id"`
		Name         string  `json:"name"`
		Status       string  `json:"status"`
		WinRate      float64 `json:"win_rate"`
		ProfitFactor float64 `json:"profit_factor"`
		Trades       int32   `json:"trades"`
		Pnl          float64 `json:"pnl"`
		Strategy     string  `json:"strategy"`
	}

	ctx := r.Context()
	userID, ok := ctx.Value("userID").(int32)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	bots, err := h.db.Queries.GetUserBots(ctx, userID)
	if err != nil {
		http.Error(w, "Error getting bots from DB", http.StatusInternalServerError)
		return
	}

	botStats := []Stats{}
	for _, bot := range bots {
		winRate, err := bot.WinRate.Float64Value()
		if err != nil {
			http.Error(w, "Error handling bot winrate data", http.StatusInternalServerError)
			return
		}

		profitFactor, err := bot.ProfitFactor.Float64Value()
		if err != nil {
			http.Error(w, "Error handling bot profitFactor data", http.StatusInternalServerError)
			return
		}

		holding, err := bot.Holding.Float64Value()
		if err != nil {
			http.Error(w, "Error handling bot holding data", http.StatusInternalServerError)
			return
		}

		initialHolding, err := bot.InitialHolding.Float64Value()
		if err != nil {
			http.Error(w, "Error handling bot initial_holding data", http.StatusInternalServerError)
			return

		}

		data := Stats{
			ID:           bot.ID,
			Name:         bot.Name,
			Status:       bot.Status.String,
			WinRate:      winRate.Float64,
			ProfitFactor: profitFactor.Float64,
			Trades:       bot.Trades.Int32,
			Pnl:          holding.Float64 - initialHolding.Float64,
			Strategy:     bot.Strategy,
		}
		botStats = append(botStats, data)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(botStats)
}

func (h *UserHandlers) UpdateBot(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := ctx.Value("userID").(int32)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	botIDStr := vars["botID"]

	botID, err := strconv.Atoi(botIDStr)
	if err != nil {
		http.Error(w, "Invalid bot ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Name           string          `json:"name"`
		Strategy       string          `json:"strategy"`
		InitialHolding decimal.Decimal `json:"initial_holding"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Strategy == "" {
		http.Error(w, "Name and strategy are required", http.StatusBadRequest)
		return
	}

	initialHolding, err := decimalToPgNumeric(req.InitialHolding)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		return
	}

	params := db.UpdateBotParams{
		ID:             int32(botID),
		UserID:         userID,
		Name:           req.Name,
		Strategy:       req.Strategy,
		InitialHolding: initialHolding,
	}

	bot, err := h.db.Queries.UpdateBot(ctx, params)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			http.Error(w, "Bot not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to update bot status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bot)
}

func (h UserHandlers) GetPositions(w http.ResponseWriter, r *http.Request) {
	// Mock open positions
	positions := []map[string]any{
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

func (h *UserHandlers) CreateBot(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		Name           string          `json:"name"`
		Strategy       string          `json:"strategy"`
		InitialHolding decimal.Decimal `json:"initial_holding"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	UserID, ok := ctx.Value("userID").(int32)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	if req.Name == "" || req.Strategy == "" {
		http.Error(w, "Name and Strategy are required", http.StatusBadRequest)
		return
	}

	initialHolding, err := decimalToPgNumeric(req.InitialHolding)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		return
	}

	params := db.CreateBotParams{
		UserID:         UserID,
		Name:           req.Name,
		Strategy:       req.Strategy,
		InitialHolding: initialHolding,
	}

	res, err := h.db.Queries.CreateBot(ctx, params)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			http.Error(w, "Bot name already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to create bot", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (h UserHandlers) GetUserBots(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := ctx.Value("userID").(int32)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	bots, err := h.db.Queries.GetUserBots(ctx, userID)
	if err != nil {
		http.Error(w, "Error getting bots from DB", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bots)
}

func (h UserHandlers) UpdateBotStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	idStr := vars["botID"]

	botID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid bot ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	status := models.BotStatus(req.Status)
	if !status.IsValid() {
		http.Error(w, "Invalid status. Must be STOPPED, RUNNING, PAUSED, or ERROR", http.StatusBadRequest)
		return
	}
	pgtextStatus := pgtype.Text{String: string(status), Valid: true}

	userID, ok := ctx.Value("userID").(int32)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	params := db.UpdateBotStatusParams{
		ID:     int32(botID),
		UserID: userID,
		Status: pgtextStatus,
	}

	bot, err := h.db.Queries.UpdateBotStatus(ctx, params)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			http.Error(w, "Bot not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to update bot status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bot)
}

func (h UserHandlers) DeleteBot(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	idStr := vars["botID"]

	botID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid bot ID", http.StatusBadRequest)
		return
	}

	userID, ok := ctx.Value("userID").(int32)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	params := db.DeleteBotParams{
		ID:     int32(botID),
		UserID: userID,
	}

	err = h.db.Queries.DeleteBot(ctx, params)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			http.Error(w, "Bot not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Error deleting bot", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func decimalToPgNumeric(d decimal.Decimal) (pgtype.Numeric, error) {
	var pgNum pgtype.Numeric
	err := pgNum.Scan(d.String())
	if err != nil {
		return pgtype.Numeric{}, fmt.Errorf("error converting decimal to numeric: %v", err)
	}
	return pgNum, nil
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

	// Bot api endpoints
	r.HandleFunc("/api/bots", userHandler.CreateBot).Methods("POST")
	r.HandleFunc("/api/bots", userHandler.GetUserBots).Methods("GET")
	r.HandleFunc("/api/bots/{botID}/status", userHandler.UpdateBotStatus).Methods("PUT")
	r.HandleFunc("/api/bots/{botID}", userHandler.DeleteBot).Methods("DELETE")
	r.HandleFunc("/api/bots/{botID}", userHandler.UpdateBot).Methods("PUT")

	return r
}
