package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/adshao/go-binance/v2"
)

func (h *UserHandlers) BnTestBiance(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	key := os.Getenv("TEST_API_KEY")
	secret := os.Getenv("TEST_API_SECRET")

	client := binance.NewClient(key, secret)
	client.NewSetServerTimeService().Do(ctx)

	price, err := client.NewListPricesService().Symbol("BTCUSDT").Do(ctx)
	if err != nil {
		http.Error(w, "error getting prices", http.StatusInternalServerError)
		fmt.Println("MISTAKE")
		return
	}

	fmt.Println(price)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(price)
}

func (h *UserHandlers) AddClient(name string, client *binance.Client) {
	h.clients[name] = client
}

func (h *UserHandlers) BnGetMarginAccountInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	fmt.Println(req.Name)

	client, exits := h.clients[req.Name]
	if !exits {
		http.Error(w, "Client not found", http.StatusInternalServerError)
		return
	}

	marginAccount, err := client.NewGetMarginAccountService().Do(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting margin account: %v", err), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(marginAccount); err != nil {
		http.Error(w, fmt.Sprint("Failed to encode response: %w", err), http.StatusInternalServerError)
	}
}
