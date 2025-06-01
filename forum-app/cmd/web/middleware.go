package main

import (
	"fmt"
	"golang.org/x/time/rate"
	"net/http"
	"regexp"
	"sync"
	"time"
)

func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Note: This is split across multiple lines for readability. You don't
		// need to do this in your own code.
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline' fonts.googleapis.com; font-src fonts.gstatic.com")

		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")

		next.ServeHTTP(w, r)
	})
}

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.infoLog.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())

		next.ServeHTTP(w, r)
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a deferred function (which will always be run in the event
		// of a panic as Go unwinds the stack).
		defer func() {
			// Use the builtin recover function to check if there has been a
			// panic or not. If there has...
			if err := recover(); err != nil {
				// Set a "Connection: close" header on the response.
				w.Header().Set("Connection", "close")
				// Call the app.serverError helper method to return a 500
				// Internal Server response.
				app.serverError(w, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !app.isAuthenticated(r) {
			app.flash(w, r, "You should login before to do that")
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}

		w.Header().Add("Cache-Control", "no-store")

		next.ServeHTTP(w, r)
	})
}

func (app *application) requireRole(role string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := app.getCurrentUser(r)
		if err != nil {
			app.clientError(w, http.StatusUnauthorized)
			return
		}

		user, err := app.users.Get(userID)
		if err != nil || user.Role != role {
			app.clientError(w, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Структура для хранения лимитеров по IP
type RateLimiter struct {
	app      *application
	visitors map[string]*rate.Limiter
	mu       sync.Mutex
	limit    rate.Limit
	burst    int
}

// Создаем новый RateLimiter
func NewRateLimiter(app *application, limit rate.Limit, burst int) *RateLimiter {
	return &RateLimiter{
		app:      app, // Передаем приложение
		visitors: make(map[string]*rate.Limiter),
		limit:    limit,
		burst:    burst,
	}
}

// Получение (или создание) лимитера для IP
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if limiter, exists := rl.visitors[ip]; exists {
		return limiter
	}

	limiter := rate.NewLimiter(rl.limit, rl.burst)
	rl.visitors[ip] = limiter

	// Удаляем IP из списка через 10 минут бездействия
	go func() {
		time.Sleep(10 * time.Minute)
		rl.mu.Lock()
		delete(rl.visitors, ip)
		rl.mu.Unlock()
	}()

	return limiter
}

func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		limiter := rl.getLimiter(ip)

		if !limiter.Allow() {
			rl.app.clientError(w, http.StatusTooManyRequests) // Используем app из структуры
			return
		}

		next.ServeHTTP(w, r)
	})
}

func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		rw := &responseWriter{w, http.StatusOK}
		next.ServeHTTP(rw, r)
		duration := time.Since(start).Seconds()

		normalizedPath := normalizePath(r.URL.Path)

		httpRequestsTotal.WithLabelValues(normalizedPath, r.Method, fmt.Sprintf("%d", rw.statusCode)).Inc()
		httpDuration.WithLabelValues(normalizedPath).Observe(duration)
	})
}

// Функция нормализации пути
func normalizePath(path string) string {
	// Пример: заменяем числа в путях на ":id"
	return regexp.MustCompile(`/\d+`).ReplaceAllString(path, "/:id")
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
