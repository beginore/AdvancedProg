package main

import (
	"database/sql"
	"flag"
	"fmt"
	models2 "forum-app/internal/models"
	"github.com/prometheus/client_golang/prometheus"
	"html/template"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type application struct {
	errorLog           *log.Logger
	infoLog            *log.Logger
	posts              *models2.PostModel
	users              *models2.UserModel
	comments           *models2.CommentModel
	categories         *models2.CategoryModel
	reactions          *models2.ReactionModel
	notificationsModel *models2.NotificationModel
	templateCache      map[string]*template.Template
	sessions           map[string]int
	mu                 sync.Mutex
	reports            *models2.ReportModel
}

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"path", "method", "status"}, // Добавляем метку status
	)
	httpDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_duration_seconds",
			Help:    "Histogram of response times",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path"},
	)
	dbQueryDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Histogram of database query durations",
			Buckets: prometheus.DefBuckets,
		},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal, httpDuration, dbQueryDuration)
}

func main() {
	// Адрес порта
	addr := flag.String("addr", ":4000", "http service address")
	dsn := "./data/forum.db"
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

	// Инициализация кэша шаблонов
	templateCache, err := newTemplateCache()
	fmt.Println(templateCache)
	if err != nil {
		errorLog.Fatal(err)
	}

	// Инициализация структуры приложения
	app := application{
		errorLog:           errorLog,
		infoLog:            infoLog,
		posts:              &models2.PostModel{DB: db},
		users:              &models2.UserModel{DB: db},
		comments:           &models2.CommentModel{DB: db},
		categories:         &models2.CategoryModel{DB: db},
		notificationsModel: &models2.NotificationModel{DB: db},
		reactions:          &models2.ReactionModel{DB: db},
		templateCache:      templateCache,
		sessions:           make(map[string]int),
		reports:            &models2.ReportModel{DB: db}, // Добавляем поле reports корректно
	}

	rateLimiter := NewRateLimiter(&app, 3, 5)
	limitedRouter := rateLimiter.Limit(app.routes())

	// Инициализация структуры сервера для использования errorLog и роутера
	srv := &http.Server{
		Addr:         *addr,
		ErrorLog:     errorLog,
		Handler:      limitedRouter,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Запуск сервера с поддержкой HTTPS
	infoLog.Printf("Starting server on http://localhost%s", *addr)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dsn+"?_foreign_keys=on&_busy_timeout=5000")
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
