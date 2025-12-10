package config

import (
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DBConn *gorm.DB

// connectDb
func ConnectDb() {
	// refer https://github.com/go-sql-driver/mysql#dsn-data-source-name for details
	/*
		NOTE:
		To handle time.Time correctly, you need to include parseTime as a parameter. (more parameters)
		To fully support UTF-8 encoding, you need to change charset=utf8 to charset=utf8mb4. See this article for a detailed explanation
	*/

	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
		os.Exit(2)
	}

	// dsn := "root:adesua360@tcp(127.0.0.1:3306)/gnaps?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database. \n", err)
		os.Exit(2)
	}

	log.Println("Database connected")

	// Auto-migrate the ActivityLog model
	// if err := db.AutoMigrate(&models.ActivityLog{}); err != nil {
	// 	log.Printf("Warning: Failed to migrate ActivityLog table: %v", err)
	// } else {
	// 	log.Println("ActivityLog table migration completed")
	// }

	DBConn = db
}
