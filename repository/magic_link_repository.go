package repository

import (
	"errors"
	"feedback-app/models"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrTokenUsed = errors.New("token already used")
var ErrTokenExpired = errors.New("token expired")

type MagicLinkRepository struct {
	db *gorm.DB
}

func NewMagicLinkRepository(db *gorm.DB) *MagicLinkRepository {
	return &MagicLinkRepository{db: db}
}

func (r *MagicLinkRepository) Create(link *models.MagicLink) error {
	return r.db.Create(link).Error
}

func (r *MagicLinkRepository) FindByToken(token string) (*models.MagicLink, error) {
	var link models.MagicLink
	if err := r.db.Where("token = ?", token).First(&link).Error; err != nil {
		return nil, err
	}
	return &link, nil
}

func (r *MagicLinkRepository) MarkUsed(id uint) error {
	return r.db.Model(&models.MagicLink{}).Where("id = ?", id).Update("used", true).Error
}

// ConsumeByToken marks a token as used in a single transaction to prevent reuse.
func (r *MagicLinkRepository) ConsumeByToken(token string, now time.Time) (*models.MagicLink, error) {
	var link models.MagicLink

	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("token = ?", token).First(&link).Error; err != nil {
			return err
		}

		if link.Used {
			return ErrTokenUsed
		}

		if now.After(link.ExpiresAt) {
			return ErrTokenExpired
		}

		return tx.Model(&models.MagicLink{}).Where("id = ?", link.ID).Update("used", true).Error
	})
	if err != nil {
		return nil, err
	}

	return &link, nil
}

// CleanupExpiredTokens is a maintenance helper
func (r *MagicLinkRepository) CleanupExpiredTokens() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&models.MagicLink{}).Error
}
