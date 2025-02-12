package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"forum/internal/models"
	"html/template"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv" // Import the godotenv package
	_ "github.com/lib/pq"      // PostgreSQL driver
)

type application struct {
	errorLog           *log.Logger
	infoLog            *log.Logger
	posts              *models.PostModel
	users              *models.UserModel
	comments           *models.CommentModel
	categories         *models.CategoryModel
	reactions          *models.ReactionModel
	notificationsModel *models.NotificationModel
	templateCache      map[string]*template.Template
	sessions           map[string]int
	mu                 sync.Mutex
	reports            *models.ReportModel
}

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get environment variables
	addr := flag.String("addr", ":4000", "http service address")

	// Create the PostgreSQL DSN using environment variables
	dsn := getDBConnectionString()

	flag.Parse()

	// Логгеры для ошибок и информации
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// Открытие базы данных
	db, err := openDB(dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer db.Close()

	// Create tables if they don't exist
	err = createTables(db)
	if err != nil {
		errorLog.Fatal(err)
	}

	// Инициализация кэша шаблонов
	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}

	// Инициализация структуры приложения
	app := application{
		errorLog:           errorLog,
		infoLog:            infoLog,
		posts:              &models.PostModel{DB: db},
		users:              &models.UserModel{DB: db},
		comments:           &models.CommentModel{DB: db},
		categories:         &models.CategoryModel{DB: db},
		notificationsModel: &models.NotificationModel{DB: db},
		reactions:          &models.ReactionModel{DB: db},
		templateCache:      templateCache,
		sessions:           make(map[string]int),
		reports:            &models.ReportModel{DB: db},
	}

	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	// Инициализация структуры сервера для использования errorLog и роутера
	srv := &http.Server{
		Addr:         *addr,
		ErrorLog:     errorLog,
		Handler:      app.routes(),
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Запуск сервера с поддержкой HTTPS
	infoLog.Printf("Starting server on http://localhost%s", *addr)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

// Get the connection string for PostgreSQL using environment variables
func getDBConnectionString() string {
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dbSslMode := os.Getenv("DB_SSLMODE")

	// Construct the connection string
	return "postgres://" + dbUser + ":" + dbPassword + "@" + dbHost + ":" + dbPort + "/" + dbName + "?sslmode=" + dbSslMode
}

func openDB(dsn string) (*sql.DB, error) {
	// Open the PostgreSQL database using the pq driver
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

// createTables will run the table creation queries
func createTables(db *sql.DB) error {
	queries := []string{
		// Users table
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			hashed_password TEXT NOT NULL,
			provider TEXT,
			provider_id TEXT,
			created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			role TEXT NOT NULL
		);`,
		// Posts table
		`CREATE TABLE IF NOT EXISTS posts (
			id SERIAL PRIMARY KEY,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			image_path TEXT,
			category TEXT NOT NULL,
			likes INTEGER DEFAULT 0,
			dislikes INTEGER DEFAULT 0,
			author TEXT NOT NULL,
			author_id INTEGER NOT NULL,
			created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			status TEXT NOT NULL
		);`,
		// Comments table
		`CREATE TABLE IF NOT EXISTS comments (
			id SERIAL PRIMARY KEY,
			post_id INTEGER NOT NULL,
			content TEXT NOT NULL,
			likes INTEGER DEFAULT 0,
			dislikes INTEGER DEFAULT 0,
			user_id INTEGER NOT NULL,
			author TEXT NOT NULL,
			created TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,
		// Categories table
		`CREATE TABLE IF NOT EXISTS categories (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL UNIQUE
		);`,
		// Notifications table
		`CREATE TABLE IF NOT EXISTS notifications (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL,
			type TEXT NOT NULL,
			post_id INTEGER,
			comment_id INTEGER,
			created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			is_read BOOLEAN DEFAULT FALSE,
			actor_id INTEGER,
			actor_name TEXT
		);`,
		// Reports table
		`CREATE TABLE IF NOT EXISTS reports (
			id SERIAL PRIMARY KEY,
			post_id INTEGER NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
			reporter_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			reason TEXT DEFAULT 'NO REASON' NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			answer TEXT DEFAULT 'NO ANSWER',
			solved INTEGER DEFAULT 0,
			admin_id INTEGER DEFAULT 0 REFERENCES users(id) ON DELETE SET NULL
		);`,

		`CREATE TABLE IF NOT EXISTS comments (
    id SERIAL PRIMARY KEY,
    post_id INTEGER NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    likes INTEGER DEFAULT 0,
    dislikes INTEGER DEFAULT 0,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    author TEXT NOT NULL
);`,
		`CREATE TABLE IF NOT EXISTS comment_dislikes (
    comment_id INTEGER NOT NULL REFERENCES comments(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (comment_id, user_id)
);`,
		`CREATE TABLE IF NOT EXISTS comment_likes (
    comment_id INTEGER NOT NULL REFERENCES comments(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (comment_id, user_id)
);`,
		`CREATE TABLE IF NOT EXISTS post_dislikes (
    post_id INTEGER NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (post_id, user_id)
);`,
		`CREATE TABLE IF NOT EXISTS post_likes (
    post_id INTEGER NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (post_id, user_id)
);`,
	}

	// Execute each query
	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			return err
		}
	}
	return nil
}
