package services

import (
	"bytes"
	"errors"
	"feedback-app/models"
	"feedback-app/platform/email"
	"feedback-app/repository"
	"feedback-app/utils"
	"fmt"
	"html/template"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthService struct {
	userRepo      *repository.UserRepository
	magicLinkRepo *repository.MagicLinkRepository
	emailClient   email.Client
	jwtSecret     string
	jwtExpiration time.Duration
	appURL        string
	appEnv        string
	deepLinkURL   string
	LoginLinkTTL  time.Duration
}

type AuthConfig struct {
	JWTSecret     string
	JWTExpiration time.Duration
	AppURL        string
	AppEnv        string
	DeepLinkURL   string
	LoginLinkTTL  time.Duration
}

func NewAuthService(uRepo *repository.UserRepository, mRepo *repository.MagicLinkRepository, emailClient email.Client, cfg AuthConfig) *AuthService {
	return &AuthService{
		userRepo:      uRepo,
		magicLinkRepo: mRepo,
		emailClient:   emailClient,
		jwtSecret:     cfg.JWTSecret,
		jwtExpiration: cfg.JWTExpiration,
		appURL:        cfg.AppURL,
		appEnv:        cfg.AppEnv,
		deepLinkURL:   cfg.DeepLinkURL,
		LoginLinkTTL:  cfg.LoginLinkTTL,
	}
}

func (s *AuthService) RequestLogin(emailAddr string) error {
	user, err := s.userRepo.FindByEmail(emailAddr)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		user = &models.User{Email: emailAddr, CreatedAt: time.Now(), UpdatedAt: time.Now()}
		if err := s.userRepo.Create(user); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	token := uuid.New().String()
	magicLink := &models.MagicLink{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(s.LoginLinkTTL),
		CreatedAt: time.Now(),
	}

	if err := s.magicLinkRepo.Create(magicLink); err != nil {
		return err
	}

	link := fmt.Sprintf("%s/auth/verify?token=%s", s.appURL, token)
	body, err := s.renderLoginEmail(link)
	if err != nil {
		log.Printf("Failed to render login email: %v", err)
		return errors.New("failed to render login email")
	}

	if err := s.emailClient.Send(emailAddr, "Login to Feedback App", body); err != nil {
		log.Printf("Failed to send email to %s: %v", emailAddr, err)
		return errors.New("failed to send login email")
	}

	return nil
}

func (s *AuthService) ExchangeLoginToken(token string) (string, error) {
	link, err := s.magicLinkRepo.ConsumeByToken(token, time.Now())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("invalid token")
		}
		if errors.Is(err, repository.ErrTokenUsed) {
			return "", errors.New("token already used")
		}
		if errors.Is(err, repository.ErrTokenExpired) {
			return "", errors.New("token expired")
		}
		return "", err
	}

	jwtToken, err := utils.GenerateJWT(link.UserID, s.jwtSecret, s.jwtExpiration)
	if err != nil {
		return "", err
	}

	return jwtToken, nil
}

func (s *AuthService) BuildRedirectURL(token string) (string, error) {
	if s.deepLinkURL == "" {
		return "", errors.New("redirect URL not configured")
	}

	parsed, err := url.Parse(s.deepLinkURL)
	if err != nil {
		return "", err
	}

	query := parsed.Query()
	query.Set("token", token)
	parsed.RawQuery = query.Encode()

	return parsed.String(), nil
}

func (s *AuthService) renderLoginEmail(link string) (string, error) {
	content, err := os.ReadFile(filepath.Join("templates", "email", "login.html"))
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("login.html").Parse(string(content))
	if err != nil {
		return "", err
	}

	data := struct {
		Link          string
		ExpiryMinutes int
	}{
		Link:          link,
		ExpiryMinutes: int(s.LoginLinkTTL.Minutes()),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
