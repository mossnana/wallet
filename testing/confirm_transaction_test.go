package testing

import (
	"context"
	"fmt"
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

func TestConfirmTransaction(t *testing.T) {
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

	t.Run("transaction that verified, should be exist in redis", func(t *testing.T) {
		db.Save(&domains.UserBalance{
			ID:      int64(1),
			Balance: 200,
		})
		result, e := transactionService.VerifyTransaction(ports.VerifyTransactionPayload{
			UserID:        1,
			Amount:        50,
			PaymentMethod: "credit_card",
		})
		require.NoError(t, e)

		r, _ := rdb.Exists(ctx, fmt.Sprintf("transaction_%s", result.TransactionID)).Result()
		if r == 0 {
			t.Errorf("key transaction_%s should be exist", result.TransactionID)
		}

		transactionService.ConfirmTransaction(ports.ConfirmTransactionPayload{
			TransactionID: result.TransactionID,
		})
		confirmedTransaction := &domains.Transaction{
			ID: result.TransactionID,
		}

		db.First(confirmedTransaction)
		if confirmedTransaction.Amount != 50 {
			t.Error("wrong amount")
		}

		userBalance := &domains.UserBalance{
			ID: 1,
		}
		db.First(userBalance)
		if userBalance.Balance != 150 {
			t.Errorf("wrong user balance %v", userBalance.Balance)
		}
		if userBalance.LockBalance != 0 {
			t.Errorf("lock balance not deducted %v", userBalance.LockBalance)
		}
	})

	testcontainers.CleanupContainer(t, postgresContainer)
	testcontainers.CleanupContainer(t, redisContainer)
}
