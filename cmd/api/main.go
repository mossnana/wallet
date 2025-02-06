package main

import (
	"fmt"
	"os"
	"wallet/internal/adapters/http/handlers"
	"wallet/internal/adapters/repositories"
	"wallet/internal/applications"
	"wallet/internal/domains"
	gormpkg "wallet/pkg/gorm"
	loggerpkg "wallet/pkg/logger"
	"wallet/pkg/redis"

	"github.com/gofiber/fiber/v2"
)

const DEFAULT_PORT = "3000"

var (
	Version string
)

func main() {
	app := fiber.New()

	// logger
	logger := loggerpkg.StandardOutputLogger()

	// database client
	db := gormpkg.EnvClient()
	rdb := redis.EnvClient()
	logger.Info("connected databases")

	// migrate database
	gormpkg.AutoMigrate(
		db,
		&domains.User{},
		&domains.UserBalance{},
		&domains.Transaction{},
	)
	logger.Info("migrated database")

	// repository
	userRepository := repositories.NewUserDBRepository(db, logger)
	userBalanceRepository := repositories.NewUserBalanceDBRepository(db, logger)
	transactionDBRepository := repositories.NewTransactionDBRepository(db, logger)
	transactionCacheRepository := repositories.NewTransactionCacheRepository(rdb, logger)
	logger.Info("declared repositories")

	// services
	transactionService := applications.NewTransactionService(
		userRepository,
		userBalanceRepository,
		transactionDBRepository,
		transactionCacheRepository,
		logger,
	)
	logger.Info("declared services")

	// handlers
	transactionHandler := handlers.NewTransactionHandler(transactionService)
	logger.Info("declared handlers")

	walletGroup := app.Group("/wallet")
	walletGroup.Post("/verify", transactionHandler.VerifyTransaction)
	walletGroup.Post("/confirm", transactionHandler.ConfirmTransaction)
	logger.Info("declared endpoints")

	port := os.Getenv("PORT")
	logger.Info("getting port from env", "port", port)
	if port == "" {
		port = DEFAULT_PORT
	}

	app.Listen(fmt.Sprintf(":%s", port))
}
