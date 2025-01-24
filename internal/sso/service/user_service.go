package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"recommendation-system/internal/sso/models"
	"recommendation-system/internal/sso/repository"
	"recommendation-system/pkg/kafka"
	log "recommendation-system/pkg/logger"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
)

type UserService interface {
	RegisterUser(ctx context.Context, req *models.RegisterRequest) (*models.UserSSO, error)
	LoginUser(ctx context.Context, req *models.LoginRequest) (string, error)
}

type userService struct {
	repo   repository.UserRepository
	kafka  *kafka.KafkaClient
	topic  string
	jwtKey []byte
	logger *log.Logger
}

func NewUserService(repo repository.UserRepository, kafkaClient *kafka.KafkaClient, jwtKey []byte, logger *log.Logger) UserService {
	return &userService{
		repo:   repo,
		kafka:  kafkaClient,
		topic:  "user_updates",
		jwtKey: jwtKey,
		logger: logger,
	}
}

func (s *userService) RegisterUser(ctx context.Context, req *models.RegisterRequest) (*models.UserSSO, error) {
	s.logger.Println("Registering new user")
	user := &models.UserSSO{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: req.Password,
	}

	s.logger.Printf("Creating user: %s", user.Email)
	if err := s.repo.CreateUser(ctx, user); err != nil {
		s.logger.Printf("Failed to create user: %v", err)
		return nil, err
	}

	message := map[string]interface{}{
		"event": "user_created",
		"user": map[string]interface{}{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		},
	}

	s.logger.Printf("Publishing user creation event for user ID: %d", user.ID)
	if err := s.publishMessage(message); err != nil {
		s.logger.Printf("Failed to publish message: %v", err)
		return nil, fmt.Errorf("failed to publish message: %w", err)
	}

	s.logger.Printf("User registered successfully: %s", user.Email)
	return user, nil
}

func (s *userService) LoginUser(ctx context.Context, req *models.LoginRequest) (string, error) {
	s.logger.Printf("Attempting login for email: %s", req.Email)
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		s.logger.Printf("Failed to retrieve user: %v", err)
		return "", ErrInvalidCredentials
	}

	s.logger.Println("Verifying password")
	if err := verifyPassword(user.PasswordHash, req.Password); err != nil {
		s.logger.Printf("Password verification failed: %v", err)
		return "", ErrInvalidCredentials
	}

	s.logger.Println("Generating JWT token")
	token, err := s.generateJWT(user)
	if err != nil {
		s.logger.Printf("Failed to generate token: %v", err)
		return "", err
	}

	s.logger.Printf("User logged in successfully: %s", req.Email)
	return token, nil
}

func (s *userService) publishMessage(message interface{}) error {
	s.logger.Println("Marshalling message for Kafka")
	valueBytes, err := json.Marshal(message)
	if err != nil {
		s.logger.Printf("Failed to marshal message: %v", err)
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	s.logger.Println("Publishing message to Kafka")
	if err := s.kafka.PublishMessage(s.topic, nil, valueBytes); err != nil {
		s.logger.Printf("Failed to publish message to Kafka: %v", err)
		return err
	}

	s.logger.Println("Message published to Kafka successfully")
	return nil
}

func verifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func (s *userService) generateJWT(user *models.UserSSO) (string, error) {
	s.logger.Printf("Generating JWT for user ID: %d", user.ID)
	claims := &jwt.RegisteredClaims{
		Subject:   fmt.Sprintf("%d", user.ID),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.jwtKey)
	if err != nil {
		s.logger.Printf("Failed to sign JWT: %v", err)
		return "", err
	}

	s.logger.Printf("JWT generated successfully for user ID: %d", user.ID)
	return signedToken, nil
}
