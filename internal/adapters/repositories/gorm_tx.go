package repositories

import "gorm.io/gorm"

type GormTx struct {
	Tx *gorm.DB
}

func (g *GormTx) Commit() error {
	return g.Tx.Commit().Error
}

func (g *GormTx) Rollback() error {
	return g.Tx.Rollback().Error
}
