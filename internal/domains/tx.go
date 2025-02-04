package domains

type Tx interface {
	Commit() error
	Rollback() error
}
