package requests

type VerifyTransactionRequest struct {
	UserID        int64   `json:"user_id" validate:"required,number,min=1"`
	Amount        float64 `json:"amount" validate:"required,number,min=1"`
	PaymentMethod string  `json:"payment_method" validate:"required,oneof='credit_card'"`
}
