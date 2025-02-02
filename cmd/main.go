package main

import (
	"fmt"
	"forum/config"
	"forum/internal/models"
	"forum/internal/models/postgresql"
	_ "github.com/lib/pq"
	"log"
)

func main() {
	// Loads env variables
	config.LoadConfig()

	// Connect to PostgreSQL
	db, err := postgresql.OpenDB()
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	// AutoMigrate (Ensures tables exist)
	err = db.AutoMigrate(&models.User{}, &models.Post{}, &models.Feedback{})
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	// Creating a new user
	user := models.User{
		Username:     "John",
		Email:        "john@gmail.com",
		PasswordHash: "qwer",
	}

	userCrud := &postgresql.UserModel{Db: db}
	err = userCrud.CreateUser(&user)
	if err != nil {
		log.Fatalf("Failed to create post: %v", err)
	}

	fmt.Printf("✅ User created successfully: %+v\n", user)
}
