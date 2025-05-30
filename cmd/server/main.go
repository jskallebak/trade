package main

import (
	"fmt"
	"log"
	"net/http"

	"trade/internal/database"
	"trade/internal/handlers"
	"trade/internal/middleware"

	"github.com/joho/godotenv"
)

func main() {
	db, err := database.New()
	if err != nil {
		log.Fatal("failed to connect to database,", err)
	}

	mux := handlers.SetupRoutes(db)
	mux.Use(middleware.LoggingMiddleware)
	mux.Use(middleware.CORSMiddleware)

	mux.PathPrefix("/static").Handler(http.StripPrefix("/static", http.FileServer(http.Dir("./web/static"))))

	fmt.Println("Starting server on :8080")
	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}
