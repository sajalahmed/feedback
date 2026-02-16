package controllers

import (
	"feedback-app/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type FeedbackController struct {
	service *services.FeedbackService
}

func NewFeedbackController(service *services.FeedbackService) *FeedbackController {
	return &FeedbackController{service: service}
}

type FeedbackRequest struct {
	Content string `json:"content" binding:"required"`
}

func (c *FeedbackController) SubmitFeedback(ctx *gin.Context) {
	var req FeedbackRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Content is required"})
		return
	}

	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}
	userIDValue, ok := userID.(uint)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user context"})
		return
	}

	if err := c.service.SubmitFeedback(userIDValue, req.Content); err != nil {
		if err.Error() == "duplicate feedback submission prevented" {
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit feedback"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Feedback received"})
}
