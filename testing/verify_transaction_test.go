package testing

import (
	"context"
	"testing"
	"time"
	"wallet/internal/adapters/repositories"
	"wallet/internal/applications"
	"wallet/internal/domains"
	"wallet/internal/ports"
	"wallet/pkg/logger"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	pg "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	TEST_DB_USER     = "test"
	TEST_DB_PASSWORD = "user"
	TEST_DB_NAME     = "password"
)

func TestValidateTransaction(t *testing.T) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "redis:7.2-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}
	redisContainer, e := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, e)

	postgresContainer, e := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase(TEST_DB_NAME),
		postgres.WithUsername(TEST_DB_USER),
		postgres.WithPassword(TEST_DB_PASSWORD),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	require.NoError(t, e)

	endpoint, e := redisContainer.Endpoint(ctx, "")
	require.NoError(t, e)

	rdb := redis.NewClient(&redis.Options{
		Addr: endpoint,
	})

	dbURL, _ := postgresContainer.ConnectionString(ctx)
	db, _ := gorm.Open(pg.Open(dbURL))

	db.AutoMigrate(&domains.User{}, &domains.UserBalance{}, &domains.Transaction{})
	db.Save(&domains.User{
		ID: int64(1),
	})

	// logger
	lg := logger.StandardOutputLogger()

	// repository
	userRepository := repositories.NewUserDBRepository(db, lg)
	userBalanceRepository := repositories.NewUserBalanceDBRepository(db, lg)
	transactionDBRepository := repositories.NewTransactionDBRepository(db, lg)
	transactionCacheRepository := repositories.NewTransactionCacheRepository(rdb, lg)

	// services
	transactionService := applications.NewTransactionService(
		userRepository,
		userBalanceRepository,
		transactionDBRepository,
		transactionCacheRepository,
		lg,
	)

	t.Run("transaction should deduct in lock balance not balance", func(t *testing.T) {
		db.FirstOrCreate(&domains.UserBalance{
			ID:      int64(1),
			Balance: 1000000000.0,
		})
		result, e := transactionService.VerifyTransaction(ports.VerifyTransactionPayload{
			UserID:        1,
			Amount:        50,
			PaymentMethod: "credit_card",
		})
		require.NoError(t, e)

		userBalance := &domains.UserBalance{
			ID: 1,
		}
		e = db.First(userBalance).Error
		require.NoError(t, e)

		if result.TransactionID == "" {
			t.Error("transaction id should exist")
		}

		if userBalance.LockBalance != 50 {
			t.Errorf("lock balance should be 50 not %v", userBalance.LockBalance)
		}

		if userBalance.Balance != 1000000000 {
			t.Errorf("lock balance should be 1000000000 not %v", userBalance.Balance)
		}
	})

	t.Run("if allocate balance + new coming amount more than current balance will error insufficient balance", func(t *testing.T) {
		db.Save(&domains.UserBalance{
			ID:      int64(1),
			Balance: 200,
		})
		transactionService.VerifyTransaction(ports.VerifyTransactionPayload{
			UserID:        1,
			Amount:        50,
			PaymentMethod: "credit_card",
		})
		transactionService.VerifyTransaction(ports.VerifyTransactionPayload{
			UserID:        1,
			Amount:        100,
			PaymentMethod: "credit_card",
		})
		_, exceedError := transactionService.VerifyTransaction(ports.VerifyTransactionPayload{
			UserID:        1,
			Amount:        51,
			PaymentMethod: "credit_card",
		})
		if exceedError == nil {
			t.Errorf("should error insufficient balance")
		}
	})

	testcontainers.CleanupContainer(t, postgresContainer)
	testcontainers.CleanupContainer(t, redisContainer)
}
