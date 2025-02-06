package applications

import (
	"context"
	"errors"
	"time"
	"wallet/internal/domains"
	"wallet/internal/ports"
)

type TransactionService struct {
	userDBRepository           domains.UserRepository
	userBalanceRepository      domains.UserBalanceRepository
	transactionDBRepository    domains.TransactionRepository
	transactionCacheRepository domains.TransactionRepository
	logger                     domains.Logger
}

func NewTransactionService(userDBRepository domains.UserRepository, userBalanceRepository domains.UserBalanceRepository, transactionDBRepository domains.TransactionRepository, transactionCacheRepository domains.TransactionRepository, logger domains.Logger) *TransactionService {
	return &TransactionService{
		userDBRepository:           userDBRepository,
		transactionDBRepository:    transactionDBRepository,
		transactionCacheRepository: transactionCacheRepository,
		userBalanceRepository:      userBalanceRepository,
		logger:                     logger,
	}
}

func (t *TransactionService) VerifyTransaction(payload ports.VerifyTransactionPayload) (result ports.VerifyTransactionResult, e error) {
	result = ports.VerifyTransactionResult{}

	// begin transaction
	tx, e := t.userDBRepository.BeginTx(context.Background())
	if e != nil {
		return
	}
	t.logger.Debug("begin transaction", "user_id", payload.UserID, "amount", payload.Amount)

	// check user exist
	user, e := t.userDBRepository.GetByID(payload.UserID, tx)
	if e != nil {
		tx.Rollback()
		return
	}
	t.logger.Info("found user", "user_id", user.ID)

	// ensure user balance can create transaction
	e = t.userBalanceRepository.VerifyUserBalanceWithAmount(user.ID, payload.Amount, tx)
	if e != nil {
		tx.Rollback()
		return
	}
	t.logger.Info("verify user balance")

	// allocate balance for this transaction
	e = t.userBalanceRepository.AllocateUserBalance(user.ID, payload.Amount, tx)
	if e != nil {
		tx.Rollback()
		return
	}
	t.logger.Info("allcated user balance")

	// store transaction
	newTransaction := domains.Transaction{
		UserID:        user.ID,
		Amount:        payload.Amount,
		PaymentMethod: payload.PaymentMethod,
		Status:        "verified",
		ExpiresAt:     time.Now().Truncate(time.Minute).AddDate(0, 0, 1).Add(-time.Second),
	}
	e = t.transactionDBRepository.CreateTransaction(&newTransaction, tx)
	if e != nil {
		tx.Rollback()
		return
	}
	e = t.transactionCacheRepository.CreateTransaction(&newTransaction, tx)
	if e != nil {
		tx.Rollback()
		return
	}
	t.logger.Info("create transaction", "transaction_id", newTransaction.ID)

	e = tx.Commit()

	result = ports.VerifyTransactionResult{
		UserID:        newTransaction.UserID,
		TransactionID: newTransaction.ID,
		Amount:        newTransaction.Amount,
		Status:        newTransaction.Status,
		PaymentMethod: newTransaction.PaymentMethod,
		ExpiresAt:     newTransaction.ExpiresAt.Format("2006-01-02 15:04:05Z"),
	}

	return
}

func (t *TransactionService) ConfirmTransaction(payload ports.ConfirmTransactionPayload) (result ports.ConfirmTransactionResult, e error) {
	result = ports.ConfirmTransactionResult{}

	// retrieve transaction by transaction id from cache
	transaction, e := t.transactionCacheRepository.GetByID(payload.TransactionID, nil)
	if e != nil {
		return
	}

	// begin
	tx, e := t.transactionDBRepository.BeginTx(context.Background())
	if e != nil {
		return
	}

	// if not found in cache, get transaction from database
	if transaction.ID == "" {
		transaction, e = t.transactionDBRepository.GetByID(payload.TransactionID, tx)
		if e != nil {
			tx.Rollback()
			return
		}
	}

	// check transaction not expired
	now := time.Now()
	if transaction.Status != "verified" && transaction.ExpiresAt.After(now) {
		tx.Rollback()
		e = errors.New("transaction expired")
		return
	}

	// update transaction status
	e = t.transactionDBRepository.UpdateStatus(transaction.ID, "confirmed", tx)
	if e != nil {
		tx.Rollback()
		return
	}
	e = t.transactionCacheRepository.UpdateStatus(transaction.ID, "confirmed", tx)
	if e != nil {
		tx.Rollback()
		return
	}

	// update user balance
	e = t.userBalanceRepository.UpdateUserBalance(transaction.UserID, transaction.Amount, tx)
	if e != nil {
		tx.Rollback()
		return
	}

	// get current balance
	userBalance, _ := t.userBalanceRepository.GetByUserID(result.UserID, tx)

	// commit
	e = tx.Commit()

	result = ports.ConfirmTransactionResult{
		UserID:        transaction.UserID,
		TransactionID: transaction.ID,
		Amount:        transaction.Amount,
		Status:        "confirmed",
		Balance:       userBalance.Balance,
	}

	return
}
