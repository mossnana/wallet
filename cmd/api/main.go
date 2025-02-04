package main

import (
	"fmt"
	"os"
	"wallet/internal/adapters/http/handlers"
	"wallet/internal/adapters/repositories"
	"wallet/internal/applications"
	"wallet/internal/domains"
	gormpkg "wallet/pkg/gorm"
	"wallet/pkg/redis"

	"github.com/gofiber/fiber/v2"
)

const DEFAULT_PORT = "3000"

var (
	Version string
)

func main() {
	app := fiber.New()

	// database client
	db := gormpkg.EnvClient()
	rdb := redis.EnvClient()

	// migrate database
	gormpkg.AutoMigrate(
		db,
		&domains.User{},
		&domains.UserBalance{},
		&domains.Transaction{},
	)

	// repository
	userRepository := repositories.NewUserDBRepository(db)
	userBalanceRepository := repositories.NewUserBalanceDBRepository(db)
	transactionDBRepository := repositories.NewTransactionDBRepository(db)
	transactionCacheRepository := repositories.NewTransactionCacheRepository(rdb)

	// services
	transactionService := applications.NewTransactionService(
		userRepository,
		userBalanceRepository,
		transactionDBRepository,
		transactionCacheRepository,
	)

	// handlers
	transactionHandler := handlers.NewTransactionHandler(transactionService)

	walletGroup := app.Group("/wallet")
	walletGroup.Post("/verify", transactionHandler.VerifyTransaction)
	walletGroup.Post("/confirm", transactionHandler.ConfirmTransaction)

	port := os.Getenv("PORT")
	if port == "" {
		port = DEFAULT_PORT
	}

	app.Listen(fmt.Sprintf(":%s", port))
}
