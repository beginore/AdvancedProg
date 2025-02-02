package postgresql

import (
	"forum/internal/models"
	"gorm.io/gorm"
)

type FeedbackModel struct {
	Db *gorm.DB
}

// CreateFeedback adds a new feedback entry to the database
func (m *FeedbackModel) CreateFeedback(feedback *models.Feedback) error {
	return m.Db.Create(feedback).Error
}

// GetFeedbackByID retrieves a feedback entry by its ID
func (m *FeedbackModel) GetFeedbackByID(feedbackID uint) (*models.Feedback, error) {
	var feedback models.Feedback
	err := m.Db.First(&feedback, feedbackID).Error
	return &feedback, err
}

// GetFeedbacksByPostID retrieves all feedback entries for a specific post
func (m *FeedbackModel) GetFeedbacksByPostID(postID uint) ([]models.Feedback, error) {
	var feedbacks []models.Feedback
	err := m.Db.Where("post_id = ?", postID).Find(&feedbacks).Error
	return feedbacks, err
}

// UpdateFeedback updates an existing feedback entry
func (m *FeedbackModel) UpdateFeedback(feedback *models.Feedback) error {
	return m.Db.Save(feedback).Error
}

// DeleteFeedback removes a feedback entry by its ID
func (m *FeedbackModel) DeleteFeedback(feedbackID uint) error {
	return m.Db.Delete(&models.Feedback{}, feedbackID).Error
}
