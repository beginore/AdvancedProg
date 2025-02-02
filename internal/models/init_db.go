package models

import (
	"database/sql"
	"fmt"
)

// Функция для инициализации базы данных (создание таблиц и т.д.)
func InitDB(db *sql.DB) error {
	createTablePosts := `	
	CREATE TABLE IF NOT EXISTS posts (
		id SERIAL PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		content TEXT NOT NULL,
		created_at TIMESTAMPTZ DEFAULT NOW()
	);`

	createTableUsers := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(100) UNIQUE NOT NULL,
		password TEXT NOT NULL,
		created_at TIMESTAMPTZ DEFAULT NOW()
	);`

	createTableComments := `
	CREATE TABLE IF NOT EXISTS comments (
		id SERIAL PRIMARY KEY,
		post_id INT REFERENCES posts(id),
		content TEXT NOT NULL,
		created_at TIMESTAMPTZ DEFAULT NOW()
	);`

	// Выполнение SQL запросов
	_, err := db.Exec(createTablePosts)
	if err != nil {
		return fmt.Errorf("error creating posts table: %v", err)
	}

	_, err = db.Exec(createTableUsers)
	if err != nil {
		return fmt.Errorf("error creating users table: %v", err)
	}

	_, err = db.Exec(createTableComments)
	if err != nil {
		return fmt.Errorf("error creating comments table: %v", err)
	}

	return nil
}
