package domains

import (
	"context"
)

type UserBalanceRepository interface {
	BeginTx(c context.Context) (Tx, error)
	GetByUserID(userID int64, tx Tx) (UserBalance, error)
	VerifyUserBalanceWithAmount(userID int64, amount float64, tx Tx) error
	AllocateUserBalance(userID int64, amount float64, tx Tx) error
	UpdateUserBalance(userID int64, amount float64, tx Tx) error
}
