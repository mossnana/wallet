package domains

import "context"

type TransactionRepository interface {
	BeginTx(c context.Context) (Tx, error)
	GetByID(id string, tx Tx) (Transaction, error)
	CreateTransaction(transaction *Transaction, tx Tx) error
	UpdateStatus(userID string, status string, tx Tx) error
}
