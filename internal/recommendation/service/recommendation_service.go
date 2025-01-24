package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	log "recommendation-system/pkg/logger"

	"recommendation-system/internal/recommendation/models"
	"recommendation-system/internal/recommendation/repository"
	"recommendation-system/pkg/kafka"
	"recommendation-system/pkg/redis"
	"time"

	kafka_go "github.com/segmentio/kafka-go"
)

type RecommendationService interface {
	GenerateRecommendations(ctx context.Context, userID int64, productID int64) error
	GetLatestRecommendation(ctx context.Context, userID int64) ([]int64, error)
	ProcessKafkaMessage(ctx context.Context, message kafka_go.Message) error
}

type recommendationService struct {
	repo        repository.RecommendationRepository
	kafka       *kafka.KafkaClient
	topic       string
	redisClient *redis.RedisClient
	logger      *log.Logger
}

func NewRecommendationService(repo repository.RecommendationRepository, kafkaClient *kafka.KafkaClient, redisClient *redis.RedisClient, logger *log.Logger) RecommendationService {
	return &recommendationService{
		repo:        repo,
		kafka:       kafkaClient,
		topic:       "recommendation_updates",
		redisClient: redisClient,
		logger:      logger,
	}
}

func (s *recommendationService) GenerateRecommendations(ctx context.Context, userID int64, productID int64) error {
	s.logger.Printf("Generating recommendations for user ID: %d and product ID: %d", userID, productID)
	rec := &models.Recommendation{
		UserID:     userID,
		ProductIDs: []int{int(productID)},
	}

	if err := s.repo.CreateRecommendation(ctx, rec); err != nil {
		s.logger.Printf("Failed to create recommendation: %v", err)
		return err
	}

	cacheKey := fmt.Sprintf("recommendations:user:%d", userID)
	s.redisClient.Delete(ctx, cacheKey)
	s.logger.Printf("Cache cleared for user ID: %d", userID)

	message := map[string]interface{}{
		"event":          "recommendation_created",
		"recommendation": rec,
	}
	s.logger.Println("Publishing recommendation creation event")
	return s.publishMessage(message)
}

func (s *recommendationService) GetLatestRecommendation(ctx context.Context, userID int64) ([]int64, error) {
	s.logger.Printf("Fetching latest recommendations for user ID: %d", userID)
	cacheKey := fmt.Sprintf("recommendations:user:%d", userID)

	cachedData, err := s.redisClient.Get(ctx, cacheKey)
	if err == nil && cachedData != "" {
		s.logger.Printf("Cache hit for user ID: %d", userID)
		var recommendations []int64
		if json.Unmarshal([]byte(cachedData), &recommendations) == nil {
			return recommendations, nil
		}
	}
	s.logger.Printf("Cache miss for user ID: %d, querying repository", userID)

	recommendedPIDs, err := s.repo.GetTopProductsByUserPreference(ctx, userID, 5)
	if err != nil {
		s.logger.Printf("Failed to fetch recommendations from repository: %v", err)
		return nil, err
	}

	dataToCache, _ := json.Marshal(recommendedPIDs)
	s.redisClient.Set(ctx, cacheKey, string(dataToCache), time.Hour)
	s.logger.Printf("Recommendations cached for user ID: %d", userID)

	return recommendedPIDs, nil
}

func (s *recommendationService) ProcessKafkaMessage(ctx context.Context, m kafka_go.Message) error {
	s.logger.Println("Processing Kafka message")

	var msg map[string]interface{}
	if err := json.Unmarshal(m.Value, &msg); err != nil {
		s.logger.Printf("Failed to unmarshal message: %v", err)
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	event, ok := msg["event"].(string)
	if !ok {
		s.logger.Println("Event type missing or invalid in message")
		return errors.New("event type missing or invalid")
	}

	s.logger.Printf("Event type: %s", event)

	switch event {
	case "user_liked":
		userID, productID, err := extractUserAndProductID(msg)
		if err != nil {
			s.logger.Printf("Parse error: %v", err)
			return nil
		}

		category, err := s.repo.GetProductCategory(ctx, productID)
		if err != nil {
			s.logger.Printf("Failed to get product category for product ID %d: %v", productID, err)
			return nil
		}

		if err := s.repo.UpdateUserCategoryScore(ctx, userID, category, 2.0); err != nil {
			s.logger.Printf("Failed to update user category score: %v", err)
		}

	case "user_disliked":
		userID, productID, err := extractUserAndProductID(msg)
		if err != nil {
			s.logger.Printf("Parse error: %v", err)
			return nil
		}

		category, err := s.repo.GetProductCategory(ctx, productID)
		if err != nil {
			s.logger.Printf("Failed to get product category for product ID %d: %v", productID, err)
			return nil
		}

		if err := s.repo.UpdateUserCategoryScore(ctx, userID, category, -1.0); err != nil {
			s.logger.Printf("Failed to update user category score: %v", err)
		}

	case "user_purchased":
		userID, productID, err := extractUserAndProductID(msg)
		if err != nil {
			s.logger.Printf("Parse error: %v", err)
			return nil
		}

		category, err := s.repo.GetProductCategory(ctx, productID)
		if err != nil {
			s.logger.Printf("Failed to get product category for product ID %d: %v", productID, err)
			return nil
		}

		if err := s.repo.UpdateUserCategoryScore(ctx, userID, category, 5.0); err != nil {
			s.logger.Printf("Failed to update user category score: %v", err)
		}

	case "product_created", "product_updated":
		s.logger.Printf("[INFO] Product event: %s", event)

	default:
		s.logger.Printf("[WARN] Unhandled event type: %s", event)
	}

	s.logger.Println("Kafka message processing completed")
	return nil
}

func extractUserAndProductID(msg map[string]interface{}) (int64, int64, error) {
	userIDFloat, ok1 := msg["user_id"].(float64)
	productIDFloat, ok2 := msg["product_id"].(float64)
	if !ok1 || !ok2 {
		return 0, 0, fmt.Errorf("failed to parse user_id or product_id from message")
	}
	return int64(userIDFloat), int64(productIDFloat), nil
}

func (s *recommendationService) publishMessage(message interface{}) error {
	valueBytes, err := json.Marshal(message)
	if err != nil {
		s.logger.Printf("Failed to marshal message: %v", err)
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	s.logger.Printf("Publishing message to topic %s", s.topic)
	return s.kafka.PublishMessage(s.topic, nil, valueBytes)
}
