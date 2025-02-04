package domains

import (
	"context"
)

type UserBalanceRepository interface {
	BeginTx(c context.Context) (Tx, error)
	VerifyUserBalanceWithAmount(userID int64, amount float64, tx Tx) error
	UpdateUserBalance(userID int64, amount float64, tx Tx) error
}
