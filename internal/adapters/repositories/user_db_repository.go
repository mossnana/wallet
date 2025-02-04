package repositories

import (
	"context"
	"wallet/internal/domains"

	"gorm.io/gorm"
)

type UserDBRepository struct {
	db *gorm.DB
}

func NewUserDBRepository(db *gorm.DB) *UserDBRepository {
	return &UserDBRepository{
		db: db,
	}
}

func (ub *UserDBRepository) BeginTx(ctx context.Context) (domains.Tx, error) {
	tx := ub.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &GormTx{Tx: tx}, nil
}

func (ub *UserDBRepository) GetByID(userID int64, tx domains.Tx) (user domains.User, e error) {
	user = domains.User{}
	gormTx, ok := tx.(*GormTx)
	if !ok {
		gormTx = &GormTx{Tx: ub.db}
	}

	result := gormTx.Tx.First(&user, "id = ?", userID)
	e = result.Error
	return
}
