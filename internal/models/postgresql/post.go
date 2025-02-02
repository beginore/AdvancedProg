package postgresql

import (
	"forum/internal/models"
	"gorm.io/gorm"
)

type PostModel struct {
	Db *gorm.DB
}

func (m *PostModel) CreatePost(post *models.Post) error {
	return m.Db.Create(post).Error
}

// GetPostByID retrieves a post by ID
func (m *PostModel) GetPostByID(postID uint) (*models.Post, error) {
	var post models.Post
	err := m.Db.First(&post, postID).Error
	return &post, err
}

// GetPostsByUserID retrieves all posts by a user
func (m *PostModel) GetPostsByUserID(userID uint) ([]models.Post, error) {
	var posts []models.Post
	err := m.Db.Where("user_id = ?", userID).Find(&posts).Error
	return posts, err
}

// UpdatePost updates an existing post
func (m *PostModel) UpdatePost(post *models.Post) error {
	return m.Db.Save(post).Error
}

// DeletePost removes a post by ID
func (m *PostModel) DeletePost(postID uint) error {
	return m.Db.Delete(&models.Post{}, postID).Error
}
