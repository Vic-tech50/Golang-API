package config

import (
	"fmt"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// DB is the shared database connection used by the application.
var DB *gorm.DB

// Connect establishes a connection to MySQL.
func Connect() error {
	// Read database settings from environment variables.
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	// Check that required settings are present.
	if host == "" || port == "" || user == "" || dbName == "" {
		return fmt.Errorf("database configuration is incomplete")
	}

	// Build the MySQL connection string.
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, port, dbName,
	)

	// Open the GORM database connection.
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	// Retrieve the underlying SQL connection for connection-pool settings.
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get SQL connection: %w", err)
	}

	// Configure the database connection pool.
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// Confirm that MySQL is reachable.
	if err := sqlDB.Ping(); err != nil {
		_ = sqlDB.Close()
		return fmt.Errorf("ping database: %w", err)
	}

	// Save the connection for other packages to use.
	DB = db

	return nil
}