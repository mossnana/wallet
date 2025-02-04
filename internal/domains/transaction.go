package domains

import "time"

type Transaction struct {
	ID            string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID        int64
	Amount        float64
	PaymentMethod string
	Status        string
	ExpiresAt     time.Time
}
