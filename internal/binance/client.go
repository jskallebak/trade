package binance

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Client struct {
	ApiKey     string
	ApiSecret  string
	BaseURL    string
	HttpClient *http.Client
}

type AccountInfo struct {
	Balances []Balance
}

type Balance struct {
	Asset  string `json:"asset"` // Make sure this matches Binance API
	Free   string `json:"free"`
	Locked string `json:"locked"`
}

type CrossMarginAccount struct {
	BorrowEnabled       bool               `json:"borrowEnabled"`
	MarginLevel         string             `json:"marginLevel"`
	TotalAssetOfBtc     string             `json:"totalAssetOfBtc"`
	TotalLiabilityOfBtc string             `json:"totalLiabilityOfBtc"`
	TotalNetAssetOfBtc  string             `json:"totalNetAssetOfBtc"`
	TotalNetAssetOfUSDT string             `json:"totalNetAssetOfUsdt"`
	TradeEnabled        bool               `json:"tradeEnabled"`
	TransferEnabled     bool               `json:"transferEnabled"`
	UserAssets          []CrossMarginAsset `json:"userAssets"`
}

type CrossMarginAsset struct {
	Asset    string `json:"asset"`
	Borrowed string `json:"borrowed"`
	Free     string `json:"free"`
	Interest string `json:"interest"`
	Locked   string `json:"locked"`
	NetAsset string `json:"netAsset"`
}

type ValidationResult struct {
	IsValid       bool   `json:"is_valid"`
	SpotEnabled   bool   `json:"spot_enabled"`
	MarginEnabled bool   `json:"margin_enabled"`
	ErrorMessage  string `json:"error_message"`
}

func (c Client) ValidateAccount() (ValidationResult, error) {
	var res ValidationResult

	_, err := c.GetAccountInfo()
	if err != nil {
		res.IsValid = false
		res.SpotEnabled = false
		res.MarginEnabled = false
		res.ErrorMessage = fmt.Sprintf("Spot API failed: %s", err.Error())
		return res, nil
	}

	res.IsValid = true
	res.SpotEnabled = true

	_, err = c.GetMarginAccountInfo()
	if err != nil {
		res.MarginEnabled = false
	} else {
		res.MarginEnabled = true
	}

	return res, nil
}

func New(key, secret, baseURL string) (*Client, error) {
	if baseURL == "" {
		baseURL = "https://api.binance.com"
	}

	if key == "" || secret == "" {
		return nil, errors.New("key and secret must be set")
	}

	httpClient := http.Client{
		Timeout: 10 * time.Second,
	}

	client := Client{
		ApiKey:     key,
		ApiSecret:  secret,
		BaseURL:    baseURL,
		HttpClient: &httpClient,
	}

	return &client, nil
}

func (c Client) signRequest(queryString string) string {
	mac := hmac.New(sha256.New, []byte(c.ApiSecret))
	mac.Write([]byte(queryString))
	return fmt.Sprintf("%x", (mac.Sum(nil)))
}

func (c Client) GetMarginAccountInfo() (CrossMarginAccount, error) {
	timeStamp := time.Now().UnixMilli()
	queryString := fmt.Sprintf("timestamp=%d", timeStamp)
	signature := c.signRequest(queryString)
	finalQuery := fmt.Sprintf("%s&signature=%s", queryString, signature)
	url := fmt.Sprintf("%s/sapi/v1/margin/account?%s", c.BaseURL, finalQuery)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return CrossMarginAccount{}, fmt.Errorf("error making new request %v", err)
	}

	req.Header.Set("X-MBX-APIKEY", c.ApiKey)
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return CrossMarginAccount{}, fmt.Errorf("error sending the request %v", err)
	}
	defer resp.Body.Close()

	err = c.CheckStatus(resp)
	if err != nil {
		return CrossMarginAccount{}, err
	}

	var marginAccount CrossMarginAccount
	err = json.NewDecoder(resp.Body).Decode(&marginAccount)
	if err != nil {
		return CrossMarginAccount{}, fmt.Errorf("error decoding the response %v", err)
	}

	asset, err := c.BtcAsset2Usdt(marginAccount.TotalAssetOfBtc)
	if err != nil {
		return CrossMarginAccount{}, fmt.Errorf("error getting the margin balance %v", err)
	}

	liability, err := c.BtcAsset2Usdt(marginAccount.TotalLiabilityOfBtc)
	if err != nil {
		return CrossMarginAccount{}, fmt.Errorf("error getting the liabillities %v", err)
	}

	marginAccount.TotalNetAssetOfUSDT = fmt.Sprintf("%.2f", (asset - liability))

	return marginAccount, nil
}

func (c Client) BtcAsset2Usdt(asset string) (float64, error) {
	totalAsset, err := strconv.ParseFloat(asset, 64)
	if err != nil {
		return 0.0, fmt.Errorf("error converting TotalAssetOfBtc to float: %v", err)
	}

	priceData, err := c.GetPrice("BTCUSDT")
	if err != nil {
		return 0.0, fmt.Errorf("error getting BTCUSDT price: %v", err)
	}

	price, err := strconv.ParseFloat(priceData.Price, 64)
	if err != nil {
		return 0.0, fmt.Errorf("error converting price data to float: %v", err)
	}

	balance := price * totalAsset
	return balance, nil
}

func (c Client) GetAccountInfo() (AccountInfo, error) {
	timeStamp := time.Now().UnixMilli()
	queryString := fmt.Sprintf("timestamp=%d", timeStamp)
	signature := c.signRequest(queryString)

	// Debug output
	log.Printf("API Key: %s", c.ApiKey)
	log.Printf("API Secret: %s", c.ApiSecret)
	log.Printf("Timestamp: %d", timeStamp)
	log.Printf("Query string: '%s'", queryString)
	log.Printf("Signature: %s", signature)

	finalQuery := fmt.Sprintf("%s&signature=%s", queryString, signature)
	log.Printf("Final URL: %s/api/v3/account?%s", c.BaseURL, finalQuery)

	url := fmt.Sprintf("%s/api/v3/account?%s", c.BaseURL, finalQuery)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return AccountInfo{}, fmt.Errorf("error making new request %v", err)
	}

	req.Header.Set("X-MBX-APIKEY", c.ApiKey)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return AccountInfo{}, fmt.Errorf("error sending the request %v", err)
	}
	defer resp.Body.Close()

	err = c.CheckStatus(resp)
	if err != nil {
		return AccountInfo{}, err
	}

	var accountInfo AccountInfo
	err = json.NewDecoder(resp.Body).Decode(&accountInfo)
	if err != nil {
		return AccountInfo{}, fmt.Errorf("error decoding the response %v", err)
	}

	return accountInfo, nil
}

type PriceData struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

func (c Client) GetPrice(symbol string) (PriceData, error) {
	url := fmt.Sprintf("%s/api/v3/ticker/price?symbol=%s", c.BaseURL, symbol)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return PriceData{}, fmt.Errorf("error making the request %v", err)
	}

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return PriceData{}, fmt.Errorf("error sending the request %v", err)
	}
	defer resp.Body.Close()

	err = c.CheckStatus(resp)
	if err != nil {
		return PriceData{}, err
	}

	var priceData PriceData

	err = json.NewDecoder(resp.Body).Decode(&priceData)
	if err != nil {
		return PriceData{}, fmt.Errorf("error decoding the response %v", err)
	}

	return priceData, nil
}

func (c Client) CheckStatus(resp *http.Response) error {
	switch resp.StatusCode {
	case 200:
		return nil
	case 401:
		return fmt.Errorf("unauthorized - check your API key")
	case 403:
		return fmt.Errorf("forbidden - check your API permissions")
	case 429:
		return fmt.Errorf("rate limit exceeded")
	default:
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
}
