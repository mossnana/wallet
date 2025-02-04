package domains

import "context"

type UserRepository interface {
	BeginTx(c context.Context) (Tx, error)
	GetByID(userID int64, tx Tx) (User, error)
}
