package models

import (
	"database/sql"
	"errors"
	"time"
)

// Post structure for storing post data
type Post struct {
	ID        int
	Title     string
	Content   string
	ImagePath string
	Category  string
	Likes     int
	Dislikes  int
	Author    string
	AuthorID  int
	Created   time.Time
	Status    string
}

// PostModel wrapper for database connection
type PostModel struct {
	DB *sql.DB
}

// Insert adds a new post to the database
func (m *PostModel) Insert(title, content, imagePath, category, author, status string, authorID int) (int, error) {
	stmt := `INSERT INTO posts (title, content, image_path, category, author, author_id, created, status) 
             VALUES ($1, $2, $3, $4, $5, $6, NOW(), $7)
             RETURNING id`
	var id int
	err := m.DB.QueryRow(stmt, title, content, imagePath, category, author, authorID, status).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// Get returns a post by ID
func (m *PostModel) Get(id int) (*Post, error) {
	stmt := `SELECT id, title, content, image_path, category, likes, dislikes, author, author_id, created, status 
             FROM posts 
             WHERE id = $1`
	row := m.DB.QueryRow(stmt, id)
	p := &Post{}
	err := row.Scan(&p.ID, &p.Title, &p.Content, &p.ImagePath, &p.Category, &p.Likes, &p.Dislikes, &p.Author, &p.AuthorID, &p.Created, &p.Status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		}
		return nil, err
	}
	return p, nil
}

// Latest returns the 10 latest posts
func (m *PostModel) Latest() ([]*Post, error) {
	stmt := `SELECT id, title, content, image_path, category, likes, dislikes, author, author_id, created, status 
             FROM posts 
             WHERE status = 'approved' 
             ORDER BY created DESC 
             LIMIT 10`
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Post
	for rows.Next() {
		p := &Post{}
		err = rows.Scan(&p.ID, &p.Title, &p.Content, &p.ImagePath, &p.Category, &p.Likes, &p.Dislikes, &p.Author, &p.AuthorID, &p.Created, &p.Status)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}

// UserPosts returns posts by a specific user
func (m *PostModel) UserPosts(userID int) ([]*Post, error) {
	stmt := `SELECT id, title, content, image_path, category, likes, dislikes, author, author_id, created, status 
             FROM posts 
             WHERE author_id = $1`
	rows, err := m.DB.Query(stmt, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Post
	for rows.Next() {
		p := &Post{}
		err = rows.Scan(&p.ID, &p.Title, &p.Content, &p.ImagePath, &p.Category, &p.Likes, &p.Dislikes, &p.Author, &p.AuthorID, &p.Created, &p.Status)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}

// UpdatePost updates an existing post
func (m *PostModel) UpdatePost(title, content, imagePath, category, author string, authorID, id int) error {
	stmt := `UPDATE posts 
             SET title = $1, content = $2, image_path = $3, category = $4, author = $5, author_id = $6 
             WHERE id = $7`
	_, err := m.DB.Exec(stmt, title, content, imagePath, category, author, authorID, id)
	if err != nil {
		return err
	}
	return nil
}

// DeletePost deletes a post by ID
func (m *PostModel) DeletePost(id int) (string, error) {
	// Fetch the image path first
	stmt1 := `SELECT image_path FROM posts WHERE id = $1 FOR UPDATE`
	var imagePath string
	err := m.DB.QueryRow(stmt1, id).Scan(&imagePath)
	if err != nil {
		return "", err
	}

	// Delete the post
	stmt2 := `DELETE FROM posts WHERE id = $1`
	_, err = m.DB.Exec(stmt2, id)
	if err != nil {
		return "", err
	}
	return imagePath, nil
}

// SortByCategory retrieves posts by category
func (m *PostModel) SortByCategory(category string) ([]*Post, error) {
	stmt := `SELECT id, title, content, image_path, category, likes, dislikes, author, author_id, created, status 
             FROM posts 
             WHERE category = $1 AND status = 'approved'`
	rows, err := m.DB.Query(stmt, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Post
	for rows.Next() {
		post := &Post{}
		err = rows.Scan(&post.ID, &post.Title, &post.Content, &post.ImagePath, &post.Category, &post.Likes, &post.Dislikes, &post.Author, &post.AuthorID, &post.Created, &post.Status)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}

// GetPendingPosts retrieves posts that are pending approval
func (m *PostModel) GetPendingPosts() ([]*Post, error) {
	stmt := `SELECT id, title, content, author, created 
             FROM posts 
             WHERE status = 'pending' 
             ORDER BY created DESC`
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Post
	for rows.Next() {
		p := &Post{}
		err = rows.Scan(&p.ID, &p.Title, &p.Content, &p.Author, &p.Created)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}

// ApprovePost changes the status of a post to approved
func (m *PostModel) ApprovePost(postID int) error {
	stmt := `UPDATE posts 
             SET status = 'approved' 
             WHERE id = $1`
	_, err := m.DB.Exec(stmt, postID)
	return err
}
