package repositories

import (
	"context"
	"errors"
	"wallet/internal/domains"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserBalanceDBRepository struct {
	db     *gorm.DB
	logger domains.Logger
}

func NewUserBalanceDBRepository(db *gorm.DB, logger domains.Logger) *UserBalanceDBRepository {
	return &UserBalanceDBRepository{db: db, logger: logger}
}

func (ub *UserBalanceDBRepository) BeginTx(ctx context.Context) (domains.Tx, error) {
	tx := ub.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &GormTx{Tx: tx}, nil
}

func (ub *UserBalanceDBRepository) GetByUserID(userID int64, tx domains.Tx) (domains.UserBalance, error) {
	txGorm, ok := tx.(*GormTx)
	if !ok {
		return domains.UserBalance{}, errors.New("invalid tx type")
	}

	var userBalance domains.UserBalance
	e := txGorm.Tx.First(&userBalance).Error

	return userBalance, e
}

func (ub *UserBalanceDBRepository) VerifyUserBalanceWithAmount(userID int64, amount float64, tx domains.Tx) error {
	txGorm, ok := tx.(*GormTx)
	if !ok {
		return errors.New("invalid tx type")
	}

	var userBalance domains.UserBalance
	if e := txGorm.Tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", userID).First(&userBalance).Error; e != nil {
		return e
	}
	ub.logger.Info("select user balance for update", "id", userID)

	if userBalance.Balance < (userBalance.LockBalance + amount) {
		ub.logger.Error("insufficient balance", "current balance", userBalance.Balance, "lock balance", userBalance.LockBalance, "amount", amount)
		return errors.New("insufficient balance")
	}

	return nil
}

func (ub *UserBalanceDBRepository) AllocateUserBalance(userID int64, amount float64, tx domains.Tx) error {
	txGorm, ok := tx.(*GormTx)
	if !ok {
		return errors.New("invalid tx type")
	}

	if amount < 0 {
		return errors.New("amount must greater than 0")
	}
	ub.logger.Debug("start allocate user balance")

	if e := txGorm.Tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&domains.UserBalance{}).Where("id = ?", userID).Update("lock_balance", gorm.Expr("lock_balance + ?", amount)).Error; e != nil {
		ub.logger.Error(e.Error())
		return e
	}
	ub.logger.Info("updated locked balance")

	return nil
}

func (ub *UserBalanceDBRepository) UpdateUserBalance(userID int64, amount float64, tx domains.Tx) error {
	txGorm, ok := tx.(*GormTx)
	if !ok {
		return errors.New("invalid tx type")
	}

	ub.logger.Debug("start update user balance")
	if e := txGorm.Tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&domains.UserBalance{}).Where("id = ?", userID).Update("balance", gorm.Expr("balance - ?", amount)).Update("lock_balance", gorm.Expr("lock_balance - ?", amount)).Error; e != nil {
		ub.logger.Error(e.Error())
		return e
	}
	ub.logger.Debug("updated user balance")

	return nil
}
