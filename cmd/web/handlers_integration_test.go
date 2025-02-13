// handlers_integration_test.go
package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestHomeHandler(t *testing.T) {
	app := newTestApplication(t)

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	app.home(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func newTestApplication(t *testing.T) *application {
	t.Helper()

	templateCache, err := newTemplateCache1()
	if err != nil {
		t.Fatal(err)
	}

	return &application{
		templateCache: templateCache,
	}
}

func TestLoginRedirectWhenUnauthenticated(t *testing.T) {
	app := newTestApplication(t)

	req := httptest.NewRequest("GET", "/user/profile", nil)
	rr := httptest.NewRecorder()

	app.profile(rr, req)

	if rr.Code != http.StatusFound {
		t.Errorf("Expected redirect status %d, got %d", http.StatusFound, rr.Code)
	}

	location := rr.Header().Get("Location")
	if location != "/user/login" {
		t.Errorf("Expected redirect to /user/login, got %s", location)
	}
}

func newTemplateCache1() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	// Получаем абсолютные пути
	baseDir, err := filepath.Abs("./ui/html")
	if err != nil {
		return nil, err
	}

	pages, err := filepath.Glob(filepath.Join(baseDir, "pages", "*.html"))
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		// Специальная обработка для страницы ошибок
		if name == "errors.html" {
			ts, err := template.New(name).Funcs(functions).ParseFiles(page)
			if err != nil {
				return nil, err
			}
			cache[name] = ts
			continue
		}

		// Обычные страницы
		basePath := filepath.Join(baseDir, "base.html")
		partialsPath := filepath.Join(baseDir, "partials", "*.html")

		ts, err := template.New(name).Funcs(functions).ParseFiles(basePath)
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob(partialsPath)
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseFiles(page)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	// Добавляем шаблон ошибок, если он не был найден
	if _, ok := cache["errors.html"]; !ok {
		errorsPath := filepath.Join(baseDir, "pages", "errors.html")
		ts, err := template.New("errors.html").ParseFiles(errorsPath)
		if err != nil {
			return nil, fmt.Errorf("error template not found: %v", err)
		}
		cache["errors.html"] = ts
	}

	return cache, nil
}
