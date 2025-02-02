package models

import "time"

type User struct {
	UserID       uint      `gorm:"primaryKey;autoIncrement"`
	Username     string    `gorm:"unique;not null"`
	Email        string    `gorm:"unique;not null"`
	PasswordHash string    `gorm:"not null"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}
type Post struct {
	PostID    uint      `gorm:"primaryKey;autoIncrement"`
	UserID    uint      `gorm:"not null"`
	Title     string    `gorm:"not null"`
	Category  string    `gorm:"not null"`
	Content   string    `gorm:"type:text;not null"`
	ImageURL  string    `gorm:"type:varchar(255)"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
type Feedback struct {
	FeedbackID uint      `gorm:"primaryKey;autoIncrement"`
	UserID     uint      `gorm:"not null"`
	PostID     uint      `gorm:"not null"`
	Like       bool      `gorm:"default:false"`
	Comment    string    `gorm:"type:text"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime"`
}
