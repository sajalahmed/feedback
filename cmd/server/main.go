package main

import (
	"feedback-app/config"
	"feedback-app/controllers"
	"feedback-app/db"
	"feedback-app/middleware"
	"feedback-app/platform/email"
	"feedback-app/platform/slack"
	"feedback-app/repository"
	"feedback-app/services"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	gormDB, err := db.InitDB(cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("Failed to init db: %v", err)
	}

	userRepo := repository.NewUserRepository(gormDB)
	magicLinkRepo := repository.NewMagicLinkRepository(gormDB)
	feedbackRepo := repository.NewFeedbackRepository(gormDB)

	slackClient := slack.NewMockClient()
	emailClient := email.NewSMTPClient(cfg.SMTP)

	authService := services.NewAuthService(userRepo, magicLinkRepo, emailClient, services.AuthConfig{
		JWTSecret:     cfg.JWTSecret,
		JWTExpiration: time.Duration(cfg.JWTTokenExpireMinutes) * time.Minute,
		AppURL:        cfg.AppURL,
		AppEnv:        cfg.AppEnv,
		DeepLinkURL:   cfg.DeepLinkURL,
		LoginLinkTTL:  time.Duration(cfg.LoginLinkExpireMinutes) * time.Minute,
	})

	feedbackService := services.NewFeedbackService(feedbackRepo, slackClient)

	authController := controllers.NewAuthController(authService)
	feedbackController := controllers.NewFeedbackController(feedbackService)

	r := gin.Default()

	loginRateLimiter := middleware.NewRateLimiter(time.Duration(cfg.RateLimitSeconds) * time.Second)

	auth := r.Group("/auth")
	{
		auth.POST("/login", loginRateLimiter.Limit(), authController.RequestLogin)
		auth.GET("/verify", authController.VerifyLogin)
		auth.POST("/session", authController.CreateSession)
	}

	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		api.POST("/feedback", feedbackController.SubmitFeedback)
	}

	log.Printf("Server starting on %s", cfg.ServerPort)
	if err := r.Run(cfg.ServerPort); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
