package repositories

import (
	"context"
	"errors"
	"wallet/internal/domains"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserBalanceDBRepository struct {
	db *gorm.DB
}

func NewUserBalanceDBRepository(db *gorm.DB) *UserBalanceDBRepository {
	return &UserBalanceDBRepository{db: db}
}

func (ub *UserBalanceDBRepository) BeginTx(ctx context.Context) (domains.Tx, error) {
	tx := ub.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &GormTx{Tx: tx}, nil
}

func (ub *UserBalanceDBRepository) VerifyUserBalanceWithAmount(userID int64, amount float64, tx domains.Tx) error {
	txGorm, ok := tx.(*GormTx)
	if !ok {
		return errors.New("invalid tx type")
	}

	var userBalance domains.UserBalance
	if err := txGorm.Tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", userID).First(&userBalance).Error; err != nil {
		return errors.New("invalid tx type")
	}

	if userBalance.Balance < amount {
		return errors.New("insufficient balance")
	}

	return nil
}

func (ub *UserBalanceDBRepository) UpdateUserBalance(userID int64, amount float64, tx domains.Tx) error {
	txGorm, ok := tx.(*GormTx)
	if !ok {
		return errors.New("invalid tx type")
	}

	if e := txGorm.Tx.Model(&domains.UserBalance{}).Where("id = ?", userID).Update("balance", gorm.Expr("balance - ?", amount)).Error; e != nil {
		return e
	}

	return nil
}
