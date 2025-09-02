package main

import (
	"context"
	"fmt"
	"log"
	"time"
	"trade/internal/binance"
	"trade/internal/database"
	db "trade/internal/db/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/joho/godotenv"
)

type BalanceService struct {
	db      *database.Database
	UserID  int32
	Running bool
}

func (s BalanceService) SetRunning(b bool) {
	s.Running = b
}

func (s BalanceService) waitForNextMinute() {
	now := time.Now()
	nextMinute := now.Truncate(time.Minute).Add(time.Minute)

	time.Sleep(time.Until(nextMinute))
}

func (s BalanceService) BalanceSnapshot(ch chan<- string) {
	s.SetRunning(true)
	defer s.SetRunning(false)

	ctx := context.Background()

	for {
		s.waitForNextMinute()
		accounts, err := s.db.Queries.GetUserBinanceAccountsWithStatus(ctx, s.UserID)
		if err != nil {
			fmt.Printf("failed to get accounts: %v", err)
			return
		}
		var clients []binance.Client
		for _, acc := range accounts {
			client, err := binance.New(acc.ApiKey, acc.ApiSecret, acc.BaseUrl.String)
			if err != nil {
				fmt.Printf("failed to create client: %v", err)
				return
			}
			clients = append(clients, *client)
		}

		for i, client := range clients {
			info, err := client.GetMarginAccountInfo()
			if err != nil {
				fmt.Printf("error getting margin account info: %v", err)
				return
			}

			// Convert info.TotalNetAssetOfUSDT to pgtype.Numeric
			var totalUSDT pgtype.Numeric
			if err = totalUSDT.Scan(info.TotalNetAssetOfUSDT); err != nil {
				fmt.Printf("failed to convert TotalNetAssetOfUSDT to numeric: %v", err)
				return
			}

			var now pgtype.Timestamptz
			if err = now.Scan(time.Now()); err != nil {
				fmt.Printf("failed to convert time.Now() to TimneStamptz: %v", err)
				return
			}

			record, err := s.db.Queries.CreateBalanceRecord(ctx, db.CreateBalanceRecordParams{
				BinanceAccountID: accounts[i].ID,
				TotalBalanceUsd:  totalUSDT,
				RecordedAt:       now,
			})
			if err != nil {
				fmt.Printf("failed to create record in db: %v", err)
				return
			}

			str := fmt.Sprintf("%v %v %v %v", record.ID, record.BinanceAccountID, record.TotalBalanceUsd, record.RecordedAt)

			ch <- str
		}
	}
}

func MakeBalanceService(userID int32, db *database.Database) *BalanceService {
	balanceService := BalanceService{
		db:      db,
		UserID:  userID,
		Running: false,
	}

	return &balanceService
}

func main() {
	// The database for queries
	db, err := database.New()
	if err != nil {
		log.Fatal("failed to connect to database,", err)
	}

	users, err := db.Queries.GetUsers(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	var snapshotServices []BalanceService

	for _, user := range users {
		snapshotService := MakeBalanceService(user.ID, db)
		snapshotServices = append(snapshotServices, *snapshotService)
	}

	ch := make(chan string)

	for _, ss := range snapshotServices {
		go ss.BalanceSnapshot(ch)
	}

	reciver(ch)

}

func reciver(ch <-chan string) {
	for message := range ch {
		fmt.Println("Got:", message)
	}
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .evn file")
	}
}
