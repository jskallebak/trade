package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"strings"

	db "trade/internal/db/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
)

func (h *UserHandlers) CreateWebhookLog(r *http.Request, responseStatus int, isSuccessful bool) error {
	ctx := r.Context()

	// Prepare headers JSON
	headersMap := make(map[string][]string)
	for name, values := range r.Header {
		if !h.isSensitiveHeader(name) {
			headersMap[name] = values
		}
	}
	headersJSON, err := json.Marshal(headersMap)
	if err != nil {
		return fmt.Errorf("failed to marshal headers: %w", err)
	}

	// Prepare query parameters JSON
	queryParamsJSON, err := json.Marshal(r.URL.Query())
	if err != nil {
		return fmt.Errorf("failed to marshal query params: %w", err)
	}

	// Read request body if needed (be careful with large bodies)
	var requestBody []byte
	if r.Body != nil && r.ContentLength > 0 && r.ContentLength < 10*1024*1024 { // Limit to 10MB
		requestBody, _ = io.ReadAll(r.Body)
		r.Body = io.NopCloser(strings.NewReader(string(requestBody)))
	}

	// Parse IP address
	clientIP := h.getClientIP(r)
	ipAddr, err := netip.ParseAddr(clientIP)
	if err != nil {
		return fmt.Errorf("failed to parse IP address %s: %w", clientIP, err)
	}

	params := db.CreateWebhookLogParams{
		Method:         r.Method,
		UrlPath:        r.URL.Path,
		Headers:        headersJSON,
		QueryParams:    queryParamsJSON,
		RequestBody:    requestBody,
		ResponseStatus: pgtype.Int4{Int32: int32(responseStatus), Valid: true},
		IsSuccessful:   pgtype.Bool{Bool: isSuccessful, Valid: true},
		UserAgent:      pgtype.Text{String: r.UserAgent(), Valid: true},
		IpAddress:      &ipAddr,
	}

	_, err = h.db.Queries.CreateWebhookLog(ctx, params)
	return err
}

func (h *UserHandlers) getClientIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func (h *UserHandlers) Webhook(w http.ResponseWriter, r *http.Request) {
	// Your webhook logic here

	// Process the webhook
	err := h.processWebhook(r)

	// Log the webhook
	var status int
	var success bool

	if err != nil {
		status = http.StatusInternalServerError
		success = false
		http.Error(w, err.Error(), status)
	} else {
		status = http.StatusOK
		success = true
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	}

	// Log the webhook (don't fail the request if logging fails)
	if logErr := h.CreateWebhookLog(r, status, success); logErr != nil {
		// Log error but don't return it to client
		fmt.Printf("Failed to log webhook: %v\n", logErr)
	}
}

func (h *UserHandlers) isSensitiveHeader(headerName string) bool {
	sensitive := []string{
		"Authorization", "X-API-KEY", "X-API-SECRET", "Cookie", "X-Auth-Token", "Bearer", "X-Access-Token",
	}

	headerLower := strings.ToLower(headerName)
	for _, s := range sensitive {
		if strings.ToLower(s) == headerLower {
			return true
		}
	}
	return false
}

func (h *UserHandlers) processWebhook(r *http.Request) error {
	// Your webhook processing logic
	return nil
}
