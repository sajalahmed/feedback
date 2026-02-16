package repository

import (
	"feedback-app/models"
	"time"

	"gorm.io/gorm"
)

type FeedbackRepository struct {
	db *gorm.DB
}

func NewFeedbackRepository(db *gorm.DB) *FeedbackRepository {
	return &FeedbackRepository{db: db}
}

func (r *FeedbackRepository) Create(feedback *models.Feedback) error {
	return r.db.Create(feedback).Error
}

func (r *FeedbackRepository) CheckDuplicate(userID uint, content string) (bool, error) {
	var count int64
	window := time.Now().Add(-5 * time.Minute)
	err := r.db.Model(&models.Feedback{}).
		Where("user_id = ? AND content = ? AND created_at > ?", userID, content, window).
		Count(&count).Error
	return count > 0, err
}
