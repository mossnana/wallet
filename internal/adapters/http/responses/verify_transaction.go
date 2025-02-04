package responses

type VerifyTransactionResponse struct {
	TransactionID string  `json:"transaction_id"`
	UserID        int64   `json:"user_id"`
	Amount        float64 `json:"amount"`
	PaymentMethod string  `json:"payment_method"`
	Status        string  `json:"status"`
	ExpiresAt     string  `json:"expires_at"`
}
