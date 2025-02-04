package requests

type ConfirmTransactionRequest struct {
	TransactionID string `json:"transaction_id" validate:"required,min=36,max=36"`
}
