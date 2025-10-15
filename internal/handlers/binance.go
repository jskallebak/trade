// Package handlers something
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/adshao/go-binance/v2"
)

func (h *UserHandlers) BnTestBiance(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
	}

	client := h.clients[req.Name]
	client.NewSetServerTimeService().Do(ctx)

	price, err := client.NewListPricesService().Symbol("BTCUSDT").Do(ctx)
	if err != nil {
		http.Error(w, "error getting prices", http.StatusInternalServerError)
		fmt.Println("MISTAKE")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(price)
}

func (h *UserHandlers) AddClient(name string, client *binance.Client) {
	h.clients[name] = client
}

func (h *UserHandlers) BnGetMarginAccInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	client := h.clients[req.Name]

	marginAccount, err := client.NewGetMarginAccountService().Do(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting margin account: %v", err), http.StatusInternalServerError)
	}

	asset, err := btc2Usdt(ctx, marginAccount.TotalNetAssetOfBTC, client)
	if err != nil {
		http.Error(w, fmt.Sprintf("error converting btc2usdt %v", err), http.StatusInternalServerError)
	}

	fmt.Println("func", marginAccount.TotalCollateralValueInUSDT, asset)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(marginAccount); err != nil {
		http.Error(w, fmt.Sprint("Failed to encode response: %w", err), http.StatusInternalServerError)
	}
}

func (h *UserHandlers) GetOrCreateClient(apiKey, apiSecret, name string) (*binance.Client, error) {
	h.mu.RLock()
	if client, exists := h.clients[name]; exists {
		h.mu.RUnlock()
		return client, nil
	}
	h.mu.RUnlock()

	h.mu.Lock()
	defer h.mu.Unlock()

	// Double-check in case another goroutine created it
	if client, exists := h.clients[name]; exists {
		return client, nil
	}

	if apiKey != "" || apiSecret != "" {
		client := binance.NewClient(apiKey, apiSecret)

		h.clients[name] = client
		return client, nil
	}

	return nil, fmt.Errorf("error finding or creating client")
}

func (h *UserHandlers) CacheUserClients(ctx context.Context, userID int32) error {
	accounts, err := h.db.Queries.GetUserBinanceAccounts(ctx, userID)
	if err != nil {
		return err
	}

	for _, account := range accounts {
		h.GetOrCreateClient(account.ApiKey, account.ApiSecret, account.Name)
	}

	return nil
}

func btc2Usdt(ctx context.Context, asset string, c *binance.Client) (float64, error) {
	totalAsset, err := strconv.ParseFloat(asset, 64)
	if err != nil {
		return 0.0, fmt.Errorf("error converting TotalAssetOfBtc to float: %v", err)
	}

	priceService := c.NewAveragePriceService()
	priceService.Symbol("BTCUSDT")
	priceObj, err := priceService.Do(ctx)
	if err != nil {
		return 0.0, fmt.Errorf("error getting avg price %v", err)
	}

	price, err := strconv.ParseFloat(priceObj.Price, 64)
	if err != nil {
		return 0.0, fmt.Errorf("error converting price data to float: %v", err)
	}

	balance := price * totalAsset
	return balance, nil
}
