package repositories

type RedisTx struct {
}

func (r *RedisTx) Commit() error {
	return nil
}

func (r *RedisTx) Rollback() error {
	return nil
}
