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

	// GORM automatically manages these timestamps.
	CreatedAt time.Time `json:"created_at"`
}
