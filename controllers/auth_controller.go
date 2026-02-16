package controllers

import (
	"feedback-app/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	service *services.AuthService
}

func NewAuthController(service *services.AuthService) *AuthController {
	return &AuthController{service: service}
}

type LoginRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type VerifyTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

func (c *AuthController) RequestLogin(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	if err := c.service.RequestLogin(req.Email); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process login request"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Please check your email for the login link",
	})
}

func (c *AuthController) VerifyLogin(ctx *gin.Context) {
	token := ctx.Query("token")
	if token == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Token is required"})
		return
	}

	redirectURL, err := c.service.BuildRedirectURL(token)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to build redirect URL"})
		return
	}

	ctx.Redirect(http.StatusFound, redirectURL)
}

func (c *AuthController) CreateSession(ctx *gin.Context) {
	var req VerifyTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Token is required"})
		return
	}

	jwtToken, err := c.service.ExchangeLoginToken(req.Token)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"token": jwtToken,
	})
}
