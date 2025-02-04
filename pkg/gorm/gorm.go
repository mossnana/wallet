package gormpkg

import (
	"fmt"
	"os"
	"wallet/internal/domains"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func EnvClient() *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		"disable",
		os.Getenv("TZ"),
	)

	db, e := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if e != nil {
		panic("failed to connect database")
	}

	return db
}

func AutoMigrate(db *gorm.DB, dst ...interface{}) {
	db.AutoMigrate(
		&domains.User{},
		&domains.UserBalance{},
		&domains.Transaction{},
	)
	db.AutoMigrate(dst...)

	// for testing only (auto insert test user)
	db.FirstOrCreate(&domains.User{
		ID: int64(1),
	})
	db.FirstOrCreate(&domains.UserBalance{
		ID:      int64(1),
		Balance: 1000000000.0,
	})
}
