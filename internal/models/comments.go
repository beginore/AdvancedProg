package models

import (
	"database/sql"
	"time"
)

// Comment represents a comment in the database
type Comment struct {
	ID       int
	PostID   int
	Content  string
	Likes    int
	Dislikes int
	UserID   int
	Author   string
	Created  time.Time
}

// CommentModel provides methods for interacting with the comments table
type CommentModel struct {
	DB *sql.DB
}

// GetByPostID retrieves comments for a specific post
func (m *CommentModel) GetByPostID(postID int) ([]*Comment, error) {
	stmt := `
        SELECT id, post_id, content, likes, dislikes, user_id, author, created 
        FROM comments 
        WHERE post_id = $1 
        ORDER BY created ASC
    `
	rows, err := m.DB.Query(stmt, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*Comment
	for rows.Next() {
		comment := &Comment{}
		err := rows.Scan(&comment.ID, &comment.PostID, &comment.Content, &comment.Likes, &comment.Dislikes, &comment.UserID, &comment.Author, &comment.Created)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

// Insert adds a new comment to the database
func (m *CommentModel) Insert(comment *Comment) error {
	stmt := `
        INSERT INTO comments (post_id, content, likes, dislikes, user_id, author, created) 
        VALUES ($1, $2, $3, $4, $5, $6, NOW()) 
        RETURNING id
    `
	var id int
	err := m.DB.QueryRow(stmt, comment.PostID, comment.Content, comment.Likes, comment.Dislikes, comment.UserID, comment.Author).Scan(&id)
	if err != nil {
		return err
	}
	comment.ID = id
	return nil
}

// Delete removes a comment by ID
func (m *CommentModel) Delete(commentID int) error {
	stmt := `DELETE FROM comments WHERE id = $1`
	_, err := m.DB.Exec(stmt, commentID)
	if err != nil {
		return err
	}
	return nil
}

// Update modifies an existing comment
func (m *CommentModel) Update(commentID int, content string) error {
	stmt := `UPDATE comments SET content = $1, created = NOW() WHERE id = $2`
	_, err := m.DB.Exec(stmt, content, commentID)
	if err != nil {
		return err
	}
	return nil
}

// GetByID retrieves a comment by ID
func (m *CommentModel) GetByID(commentID int) (*Comment, error) {
	stmt := `
        SELECT id, post_id, content, likes, dislikes, user_id, author, created 
        FROM comments 
        WHERE id = $1
    `
	row := m.DB.QueryRow(stmt, commentID)
	comment := &Comment{}
	err := row.Scan(&comment.ID, &comment.PostID, &comment.Content, &comment.Likes, &comment.Dislikes, &comment.UserID, &comment.Author, &comment.Created)
	if err != nil {
		return nil, err
	}
	return comment, nil
}

// UserComments retrieves comments made by a specific user
func (m *CommentModel) UserComments(userID int) ([]*Comment, error) {
	stmt := `
        SELECT id, post_id, content, likes, dislikes, user_id, author, created 
        FROM comments 
        WHERE user_id = $1 
        ORDER BY created ASC
    `
	rows, err := m.DB.Query(stmt, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*Comment
	for rows.Next() {
		comment := &Comment{}
		err := rows.Scan(&comment.ID, &comment.PostID, &comment.Content, &comment.Likes, &comment.Dislikes, &comment.UserID, &comment.Author, &comment.Created)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}
