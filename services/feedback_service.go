package services

import (
	"errors"
	"feedback-app/models"
	"feedback-app/platform/slack"
	"feedback-app/repository"
	"fmt"
	"time"
)

type FeedbackService struct {
	repo        *repository.FeedbackRepository
	slackClient slack.Client
}

func NewFeedbackService(repo *repository.FeedbackRepository, slackClient slack.Client) *FeedbackService {
	return &FeedbackService{
		repo:        repo,
		slackClient: slackClient,
	}
}

func (s *FeedbackService) SubmitFeedback(userID uint, content string) error {
	isDuplicate, err := s.repo.CheckDuplicate(userID, content)
	if err != nil {
		return err
	}
	if isDuplicate {
		return errors.New("duplicate feedback submission prevented")
	}

	feedback := &models.Feedback{
		UserID:    userID,
		Content:   content,
		CreatedAt: time.Now(),
	}

	if err := s.repo.Create(feedback); err != nil {
		return err
	}

	go func() {
		msg := fmt.Sprintf("New user feedback (User ID: %d): %s", userID, content)
		_ = s.slackClient.PostMessage("feedbacks", msg)
	}()

	return nil
}
