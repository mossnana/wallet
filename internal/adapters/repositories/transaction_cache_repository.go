package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"wallet/internal/domains"

	"github.com/redis/go-redis/v9"
)

type TransactionCacheRepository struct {
	client *redis.Client
	logger domains.Logger
}

func NewTransactionCacheRepository(client *redis.Client, logger domains.Logger) *TransactionCacheRepository {
	return &TransactionCacheRepository{client: client, logger: logger}
}

func (t *TransactionCacheRepository) BeginTx(ctx context.Context) (domains.Tx, error) {
	return &RedisTx{}, nil
}

func (t *TransactionCacheRepository) GetByID(id string, tx domains.Tx) (domains.Transaction, error) {
	key := fmt.Sprintf("transaction_%v", id)
	val, e := t.client.Get(context.Background(), key).Result()
	if e != nil {
		if e == redis.Nil {
			return domains.Transaction{}, nil
		}
		return domains.Transaction{}, e
	}

	var transaction domains.Transaction
	e = json.Unmarshal([]byte(val), &transaction)
	t.logger.Info("hit cache !!!", "transaction_id", transaction.ID)

	return transaction, e
}

func (t *TransactionCacheRepository) CreateTransaction(transaction *domains.Transaction, tx domains.Tx) error {
	key := fmt.Sprintf("transaction_%v", transaction.ID)

	b, e := json.Marshal(transaction)
	if e != nil {
		return e
	}

	e = t.client.Set(context.Background(), key, string(b), time.Hour*23).Err()
	t.logger.Info("create cache !!!")

	return e
}

func (t *TransactionCacheRepository) UpdateStatus(id string, status string, tx domains.Tx) error {
	transaction, e := t.GetByID(id, nil)
	if e != nil {
		return e
	}

	transaction.Status = status

	e = t.CreateTransaction(&transaction, nil)
	if e != nil {
		return e
	}

	t.logger.Info("update cache !!!")

	return nil
}
