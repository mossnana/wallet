package responses

type ConfirmTransactionResponse struct {
	TransactionID string  `json:"transaction_id"`
	UserID        int64   `json:"user_id"`
	Amount        float64 `json:"amount"`
	Status        string  `json:"status"`
	Balance       float64 `json:"balance"`
}
