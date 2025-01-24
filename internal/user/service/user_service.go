package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	log "recommendation-system/pkg/logger"

	"recommendation-system/internal/user/models"
	"recommendation-system/internal/user/repository"
	"recommendation-system/pkg/kafka"
	"recommendation-system/pkg/redis"
)

type UserService interface {
	GetUser(ctx context.Context, id int64) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	GetAllUsers(ctx context.Context, limit, offset int) ([]*models.User, error)
	PurchaseProduct(ctx context.Context, userID, productID int64) error
	LikeProduct(ctx context.Context, userID, productID int64) error
	DislikeProduct(ctx context.Context, userID, productID int64) error
	GetUserActions(ctx context.Context, userID int64, productID *int64) (map[string][]interface{}, error)
	GetUserPurchases(ctx context.Context, userID int64, limit, offset int) ([]*models.Purchase, error)
}

type userService struct {
	repo        repository.UserRepository
	kafka       *kafka.KafkaClient
	topic       string
	redisClient *redis.RedisClient
	logger      *log.Logger
}

func NewUserService(repo repository.UserRepository, kafkaClient *kafka.KafkaClient, redisClient *redis.RedisClient, logger *log.Logger) UserService {
	return &userService{
		repo:        repo,
		kafka:       kafkaClient,
		topic:       "user_updates",
		redisClient: redisClient,
		logger:      logger,
	}
}

func (s *userService) GetUser(ctx context.Context, id int64) (*models.User, error) {
	s.logger.Printf("Fetching user with ID: %d", id)
	cacheKey := fmt.Sprintf("user:%d", id)

	cachedData, err := s.redisClient.Get(ctx, cacheKey)
	if err == nil && cachedData != "" {
		var user models.User
		if json.Unmarshal([]byte(cachedData), &user) == nil {
			s.logger.Printf("Cache hit for user ID: %d", id)
			return &user, nil
		}
	}

	s.logger.Printf("Cache miss for user ID: %d, querying repository", id)
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		s.logger.Printf("Failed to fetch user with ID %d: %v", id, err)
		return nil, err
	}

	dataToCache, _ := json.Marshal(user)
	s.redisClient.Set(ctx, cacheKey, string(dataToCache), time.Hour)
	s.logger.Printf("User data cached for ID: %d", id)

	return user, nil
}

func (s *userService) UpdateUser(ctx context.Context, user *models.User) error {
	s.logger.Printf("Updating user with ID: %d", user.ID)
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		s.logger.Printf("Failed to update user: %v", err)
		return err
	}

	cacheKey := fmt.Sprintf("user:%d", user.ID)
	s.redisClient.Delete(ctx, cacheKey)
	s.logger.Printf("Cache invalidated for user ID: %d", user.ID)

	message := map[string]interface{}{
		"event": "user_updated",
		"user":  user,
	}
	s.logger.Println("Publishing user update event")
	return s.publishMessage(message)
}

func (s *userService) GetAllUsers(ctx context.Context, limit, offset int) ([]*models.User, error) {
	s.logger.Printf("Fetching all users with limit: %d, offset: %d", limit, offset)
	users, err := s.repo.GetAllUsers(ctx, limit, offset)
	if err != nil {
		s.logger.Printf("Failed to fetch users: %v", err)
	}
	return users, err
}

func (s *userService) PurchaseProduct(ctx context.Context, userID, productID int64) error {
	s.logger.Printf("User %d purchasing product %d", userID, productID)
	purchase := &models.Purchase{
		UserID:    userID,
		ProductID: productID,
	}

	if err := s.repo.CreatePurchase(ctx, purchase); err != nil {
		s.logger.Printf("Failed to record purchase: %v", err)
		return err
	}

	message := map[string]interface{}{
		"event":      "user_purchased",
		"user_id":    userID,
		"product_id": productID,
		"purchase":   purchase,
	}
	s.logger.Println("Publishing purchase event")
	return s.publishMessage(message)
}

func (s *userService) LikeProduct(ctx context.Context, userID, productID int64) error {
	s.logger.Printf("User %d liking product %d", userID, productID)
	if err := s.repo.RemoveDislikeByUserAndProduct(ctx, userID, productID); err != nil {
		s.logger.Printf("Failed to remove existing dislike: %v", err)
		return fmt.Errorf("failed to remove existing dislike: %w", err)
	}

	exists, err := s.repo.LikeExists(ctx, userID, productID)
	if err != nil {
		s.logger.Printf("Failed to check if like exists: %v", err)
		return err
	}
	if exists {
		s.logger.Printf("Like already exists for user %d and product %d", userID, productID)
		return fmt.Errorf("like already exists for user %d and product %d", userID, productID)
	}

	like := &models.Like{
		UserID:    userID,
		ProductID: productID,
	}
	if err := s.repo.CreateLike(ctx, like); err != nil {
		s.logger.Printf("Failed to create like: %v", err)
		return err
	}

	message := map[string]interface{}{
		"event":      "user_liked",
		"user_id":    userID,
		"product_id": productID,
		"like":       like,
	}
	s.logger.Println("Publishing like event")
	return s.publishMessage(message)
}

func (s *userService) DislikeProduct(ctx context.Context, userID, productID int64) error {
	s.logger.Printf("User %d disliking product %d", userID, productID)
	if err := s.repo.RemoveLikeByUserAndProduct(ctx, userID, productID); err != nil {
		s.logger.Printf("Failed to remove existing like: %v", err)
		return fmt.Errorf("failed to remove existing like: %w", err)
	}

	exists, err := s.repo.DislikeExists(ctx, userID, productID)
	if err != nil {
		s.logger.Printf("Failed to check if dislike exists: %v", err)
		return err
	}
	if exists {
		s.logger.Printf("Dislike already exists for user %d and product %d", userID, productID)
		return fmt.Errorf("dislike already exists for user %d and product %d", userID, productID)
	}

	dislike := &models.Dislike{
		UserID:    userID,
		ProductID: productID,
	}
	if err := s.repo.CreateDislike(ctx, dislike); err != nil {
		s.logger.Printf("Failed to create dislike: %v", err)
		return err
	}

	message := map[string]interface{}{
		"event":      "user_disliked",
		"user_id":    userID,
		"product_id": productID,
		"dislike":    dislike,
	}
	s.logger.Println("Publishing dislike event")
	return s.publishMessage(message)
}

func (s *userService) GetUserActions(ctx context.Context, userID int64, productID *int64) (map[string][]interface{}, error) {
	s.logger.Printf("Fetching user actions for user ID: %d", userID)
	actions, err := s.repo.GetUserActions(ctx, userID, productID)
	if err != nil {
		s.logger.Printf("Failed to fetch user actions: %v", err)
	}
	return actions, err
}

func (s *userService) GetUserPurchases(ctx context.Context, userID int64, limit, offset int) ([]*models.Purchase, error) {
	s.logger.Printf("Fetching purchases for user ID: %d with limit: %d, offset: %d", userID, limit, offset)
	purchases, err := s.repo.GetUserPurchases(ctx, userID, limit, offset)
	if err != nil {
		s.logger.Printf("Failed to fetch purchases: %v", err)
	}
	return purchases, err
}

func (s *userService) publishMessage(message interface{}) error {
	valueBytes, err := json.Marshal(message)
	if err != nil {
		s.logger.Printf("Failed to marshal message: %v", err)
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	s.logger.Printf("Publishing message to topic %s", s.topic)
	return s.kafka.PublishMessage(s.topic, nil, valueBytes)
}
