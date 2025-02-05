package main

import (
	"errors"
	"fmt"
	"forum/internal/models"
	"forum/internal/validator"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type postCreateForm struct {
	Title     string
	Content   string
	ImagePath string
	Category  string
	Author    string
	AuthorID  int
	validator.Validator
	Status string
}
type editPost struct {
	ID        int
	Title     string
	Content   string
	ImagePath string
	Category  string
	Author    string
	AuthorID  int
	validator.Validator
}
type userSignupForm struct {
	Name                string `form:"name"`
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}
type userPost struct {
	ID        int
	Title     string
	Content   string
	ImagePath string
	Category  string
	Author    string
	Created   time.Time
}
type passwordForm struct {
	CurrentPassword     string `form:"currentPassword"`
	NewPassword         string `form:"newPassword"`
	ConfirmPassword     string `form:"confirmPassword"`
	validator.Validator `form:"-"`
}
type userLoginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

func (app *application) manageUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		app.methodNotAllowed(w)
		return
	}
	if r.URL.Path != "/admin/users" {
		http.NotFound(w, r)
		return
	}
	users, err := app.users.GetPendingModerators()
	if err != nil {
		app.serverError(w, err)
		return
	}
	data := app.newTemplateData(w, r)

	data.Users = users
	app.render(w, http.StatusOK, "users.html", data)
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		app.methodNotAllowed(w)
		return
	}
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	selectedCategory := r.URL.Query().Get("Category")
	data := app.newTemplateData(w, r)
	var posts []*models.Post
	var err error
	if selectedCategory == "" {
		posts, err = app.posts.Latest()
	} else {
		posts, err = app.posts.SortByCategory(selectedCategory)
		data.SelectedCategory = selectedCategory
	}
	if err != nil {
		app.serverError(w, err)
		return
	}
	data.Posts = posts
	categories, err := app.categories.GetAll()
	if err != nil {
		app.serverError(w, err)
		return
	}
	userID, err := app.getCurrentUser(r)
	if err != nil && userID == 0 {
		data.Categories = categories
		data.IsAuthenticated = app.isAuthenticated(r)
		app.render(w, http.StatusOK, "home.html", data)
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}
	user, err := app.users.Get(userID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data.Categories = categories
	data.User = user
	data.IsAuthenticated = app.isAuthenticated(r)
	app.render(w, http.StatusOK, "home.html", data)
	return
}

func (app *application) postView(w http.ResponseWriter, r *http.Request) {
	idParam := r.URL.Query().Get("id")
	var id int
	var err error

	if idParam != "" {
		// Если параметр есть, преобразуем его в число
		id, err = strconv.Atoi(idParam)
	} else {
		// Иначе пытаемся извлечь ID из пути
		path := strings.TrimPrefix(r.URL.Path, "/post/view/")
		id, err = strconv.Atoi(path)
	}

	post, err := app.posts.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	comments, err := app.comments.GetByPostID(id)
	if err != nil {
		app.serverError(w, err)
		return
	}
	data := app.newTemplateData(w, r)
	data.Post = post
	data.Comments = comments
	if app.isAuthenticated(r) {
		userId, err := app.getCurrentUser(r)
		if err != nil {
			app.clientError(w, http.StatusUnauthorized)
		}
		user, err := app.users.Get(userId)
		if err != nil {
			app.serverError(w, err)
		}
		data.User = user
	}
	data.IsAuthenticated = app.isAuthenticated(r)
	app.render(w, http.StatusOK, "view.html", data)
}

func (app *application) postCreateForm(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		categories, err := app.categories.GetAll()
		if err != nil {
			app.serverError(w, err)
			return
		}

		data := app.newTemplateData(w, r)
		data.Form = &postCreateForm{
			Validator: validator.Validator{
				FieldErrors: map[string]string{},
			},
		}
		data.Categories = categories
		app.render(w, http.StatusOK, "create.html", data)
		return
	}

	// Если метод POST, обрабатываем данные формы
	if r.Method == http.MethodPost {
		err := r.ParseMultipartForm(20 << 20)
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}

		// Пытаемся получить файл, но проверяем, прикреплен ли он
		file, handler, err := r.FormFile("image")
		if err != nil && err.Error() != "http: no such file" { // проверяем, что файл не был прикреплен
			app.clientError(w, http.StatusBadRequest)
			return
		}

		var filePath string
		if err == nil {
			// Файл был прикреплен, обрабатываем его
			defer file.Close()
			app.infoLog.Printf("Uploaded File: %+v\n", handler.Filename)
			app.infoLog.Printf("File Size: %+v\n", handler.Size)
			app.infoLog.Printf("MIME Header: %+v\n", handler.Header)
			fileName := fmt.Sprintf("%d-%s", time.Now().UnixNano(), handler.Filename)
			filePath = fmt.Sprintf("ui/static/upload/%s", fileName)
			if err := os.MkdirAll("ui/static/upload", os.ModePerm); err != nil {
				app.serverError(w, err)
				return
			}
			dst, err := os.Create(filePath)
			if err != nil {
				app.serverError(w, err)
				return
			}
			defer dst.Close()

			if _, err := io.Copy(dst, file); err != nil {
				app.serverError(w, err)
				return
			}
		} else {
			// Если файл не был прикреплен, оставляем filePath пустым
			filePath = ""
		}

		id, err := app.getCurrentUser(r)
		if err != nil {
			app.serverError(w, err)
			return
		}
		author, err := app.users.Get(id)
		if err != nil {
			app.serverError(w, err)
			return
		}

		// Извлекаем данные из формы
		var statusString string
		if author.Role == "moderator" || author.Role == "admin" {
			statusString = "approved"
		} else {
			statusString = "pending"
		}
		form := postCreateForm{
			Title:     r.PostForm.Get("title"),
			Content:   r.PostForm.Get("content"),
			ImagePath: filePath,
			Category:  r.PostForm.Get("Category"),
			Author:    author.Name,
			AuthorID:  id,
			Status:    statusString,
		}
		form.ImagePath = strings.TrimPrefix(form.ImagePath, "ui/static/upload/")

		// Валидация полей
		form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be blank")
		form.CheckField(validator.MaxChars(form.Title, 100), "title", "This field cannot be longer than 100 characters")
		form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank")
		if !form.Valid() {
			data := app.newTemplateData(w, r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "create.html", data)
			return
		}

		app.infoLog.Printf("User Role: %s, Setting post status to: %s", author.Role, form.Status)
		// Вставляем данные в базу
		id, err = app.posts.Insert(
			form.Title,
			form.Content,
			form.ImagePath,
			form.Category,
			form.Author,
			form.Status,
			form.AuthorID,
		)

		if err != nil {
			app.serverError(w, err)
			return
		}

		app.flash(w, r, "Post created successfully!")
		// Перенаправляем пользователя на страницу с созданным постом
		http.Redirect(w, r, fmt.Sprintf("/post/view/%d", id), http.StatusSeeOther)
		return
	}

	// Если метод не GET и не POST, возвращаем ошибку
	app.clientError(w, http.StatusMethodNotAllowed)
}
