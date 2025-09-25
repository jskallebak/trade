package binance

import (
	"context"
	"log"
	"trade/internal/database"

	"github.com/adshao/go-binance/v2"
)

func test(key, secret string) {
	_, err := database.New()
	if err != nil {
		log.Fatal("failed to connect database", err)
	}

	type clientAcc struct {
		ApiKey    string
		ApiSecret string
	}

	client := binance.NewClient(key, secret)

	client.NewServerTimeService().Do(context.Background())

}
