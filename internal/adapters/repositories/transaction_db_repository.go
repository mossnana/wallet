package repositories

import (
	"context"
	"errors"
	"wallet/internal/domains"

	"gorm.io/gorm"
)

type TransactionDBRepository struct {
	db     *gorm.DB
	logger domains.Logger
}

func NewTransactionDBRepository(db *gorm.DB, logger domains.Logger) *TransactionDBRepository {
	return &TransactionDBRepository{db: db, logger: logger}
}

func (t *TransactionDBRepository) BeginTx(ctx context.Context) (domains.Tx, error) {
	tx := t.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &GormTx{Tx: tx}, nil
}

func (t *TransactionDBRepository) GetByID(id string, tx domains.Tx) (transaction domains.Transaction, e error) {
	transaction = domains.Transaction{}
	txGorm, ok := tx.(*GormTx)
	if !ok {
		return transaction, errors.New("invalid tx type")
	}

	result := txGorm.Tx.First(&transaction, "id = ?", id)
	e = result.Error
	return
}

func (t *TransactionDBRepository) CreateTransaction(transaction *domains.Transaction, tx domains.Tx) error {
	txGorm, ok := tx.(*GormTx)
	if !ok {
		return errors.New("invalid tx type")
	}

	result := txGorm.Tx.Create(transaction)
	return result.Error
}

func (t *TransactionDBRepository) UpdateStatus(id string, status string, tx domains.Tx) error {
	txGorm, ok := tx.(*GormTx)
	if !ok {
		return errors.New("invalid tx type")
	}

	if err := txGorm.Tx.Model(&domains.Transaction{}).Where("id = ?", id).Update("status", status).Error; err != nil {
		return err
	}

	return nil
}
