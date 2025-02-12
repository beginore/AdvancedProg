package models

import (
	"database/sql"
	"time"
)

// Notification represents a notification in the database
type Notification struct {
	ID        int
	UserID    int
	Type      string
	PostID    int
	CommentID int
	Created   time.Time
	IsRead    bool
	ActorID   int
	ActorName string
}

// NotificationModel provides methods for interacting with the notifications table
type NotificationModel struct {
	DB *sql.DB
}

// Insert adds a new notification to the database
func (m *NotificationModel) Insert(userID, actorID int, ntype string, postID, commentID int) error {
	stmt := `
        INSERT INTO notifications (user_id, type, post_id, comment_id, created, is_read, actor_id) 
        VALUES ($1, $2, $3, $4, NOW(), false, $5)
    `
	_, err := m.DB.Exec(stmt, userID, ntype, postID, commentID, actorID)
	return err
}

// GetUnreadCount returns the count of unread notifications for a user
func (m *NotificationModel) GetUnreadCount(userID int) (int, error) {
	var count int
	stmt := `
        SELECT COUNT(*) 
        FROM notifications 
        WHERE user_id = $1 AND is_read = false
    `
	err := m.DB.QueryRow(stmt, userID).Scan(&count)
	return count, err
}

// GetAll retrieves all notifications for a user
func (m *NotificationModel) GetAll(userID int) ([]*Notification, error) {
	stmt := `
        SELECT n.id, n.type, n.post_id, n.comment_id, n.created, n.is_read,
               u.id, u.name
        FROM notifications n
        JOIN users u ON n.actor_id = u.id
        WHERE n.user_id = $1
        ORDER BY n.created DESC
    `
	rows, err := m.DB.Query(stmt, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*Notification
	for rows.Next() {
		var n Notification
		err := rows.Scan(
			&n.ID, &n.Type, &n.PostID, &n.CommentID, &n.Created, &n.IsRead,
			&n.ActorID, &n.ActorName,
		)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, &n)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return notifications, nil
}

// MarkAllAsRead marks all notifications as read for a user
func (m *NotificationModel) MarkAllAsRead(userID int) error {
	stmt := `
        UPDATE notifications 
        SET is_read = true 
        WHERE user_id = $1
    `
	_, err := m.DB.Exec(stmt, userID)
	return err
}
