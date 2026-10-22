package models

import "time"

// Product represents a product stored in MySQL.
type Product struct {
	// GORM automatically creates an ID primary key.
	ID uint `gorm:"primaryKey" json:"id"`

	// Product name is required.
	Name string `gorm:"size:150;not null" json:"name"`

	// Description is optional.
	Description string `gorm:"type:text" json:"description"`

	// Price stores the product price.
	Price float64 `gorm:"not null" json:"price"`

	// Stock stores the available quantity.
	Stock int `gorm:"not null;default:0" json:"stock"`

	// ImagePath stores the file path, not the actual image.
	ImagePath string `gorm:"size:255" json:"image_path"`

	// GORM automatically manages these timestamps.
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type User struct {
	// GORM automatically creates an ID primary key.
	ID uint `gorm:"primaryKey" json:"id"`

	// Product name is required.
	Name string `gorm:"size:150;not null" json:"name"`

	// Description is optional.
	Email string `gorm:"type:text" json:"email"`

	// Price stores the product price.
	Password string `gorm:"not null" json:"password"`

	Role string `gorm:"size:20;not null;default:user" json:"role"`

	// GORM automatically manages these timestamps.
	CreatedAt time.Time `json:"created_at"`
}

type Payment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Email     string    `gorm:"size:150;not null" json:"email"`
	Reference string    `gorm:"size:100;uniqueIndex;not null" json:"reference"`
	Amount    int64     `gorm:"not null" json:"amount"`
	Currency  string    `gorm:"size:3;not null;default:NGN" json:"currency"`
	Status    string    `gorm:"size:20;not null;default:pending" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// FlutterwavePayment represents a payment stored in MySQL.
type FlutterwavePayment struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Email         string    `gorm:"size:150;not null" json:"email"`
	Reference     string    `gorm:"size:100;uniqueIndex;not null" json:"reference"`
	TransactionID string    `gorm:"size:100" json:"transaction_id"`
	Amount        float64   `gorm:"not null" json:"amount"`
	Currency      string    `gorm:"size:3;not null;default:NGN" json:"currency"`
	Status        string    `gorm:"size:20;not null;default:pending" json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// StripePayment stores the payment details in MySQL.
type StripePayment struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	Email             string    `gorm:"size:150;not null" json:"email"`
	Reference         string    `gorm:"size:100;uniqueIndex;not null" json:"reference"`
	CheckoutSessionID string    `gorm:"size:150;uniqueIndex" json:"checkout_session_id"`
	Amount            int64     `gorm:"not null" json:"amount"`
	Currency          string    `gorm:"size:3;not null;default:usd" json:"currency"`
	Status            string    `gorm:"size:20;not null;default:pending" json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}