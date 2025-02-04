package ports

type VerifyTransactionPayload struct {
	UserID        int64
	Amount        float64
	PaymentMethod string
}

type VerifyTransactionResult struct {
	UserID        int64
	TransactionID string
	Amount        float64
	PaymentMethod string
	Status        string
	ExpiresAt     string
}

type ConfirmTransactionPayload struct {
	TransactionID string
}

type ConfirmTransactionResult struct {
	UserID        int64
	TransactionID string
	Amount        float64
	Status        string
	Balance       float64
}

type TransactionService interface {
	VerifyTransaction(payload VerifyTransactionPayload) (VerifyTransactionResult, error)
	ConfirmTransaction(payload ConfirmTransactionPayload) (ConfirmTransactionResult, error)
}
