package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	log "recommendation-system/pkg/logger"

	"recommendation-system/internal/analytics/models"
	"recommendation-system/internal/analytics/repository"
	"recommendation-system/pkg/kafka"

	kafka_go "github.com/segmentio/kafka-go"
)

type AnalyticsService interface {
	ProcessKafkaMessage(ctx context.Context, message kafka_go.Message) error
	GetProductAnalytics(ctx context.Context, productID int64) (*models.ProductAnalytics, error)
	GetUserAnalytics(ctx context.Context, userID int64) (*models.UserAnalytics, error)
}

type analyticsService struct {
	repo   repository.AnalyticsRepository
	kafka  *kafka.KafkaClient
	logger *log.Logger
}

func NewAnalyticsService(repo repository.AnalyticsRepository, kafkaClient *kafka.KafkaClient, logger *log.Logger) AnalyticsService {
	return &analyticsService{
		repo:   repo,
		kafka:  kafkaClient,
		logger: logger,
	}
}

func (s *analyticsService) ProcessKafkaMessage(ctx context.Context, m kafka_go.Message) error {
	s.logger.Println("Processing Kafka message...")

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
		s.logger.Println("Handling 'user_liked' event")
		userID, productID, err := parseUserAndProductID(msg)
		if err != nil {
			s.logger.Printf("Parse error: %v", err)
			return nil
		}
		if err := s.repo.IncrementProductLikes(ctx, productID); err != nil {
			s.logger.Printf("Failed to increment product likes: %v", err)
		}
		if err := s.repo.IncrementUserLikes(ctx, userID); err != nil {
			s.logger.Printf("Failed to increment user likes: %v", err)
		}

	case "user_disliked":
		s.logger.Println("Handling 'user_disliked' event")
		userID, productID, err := parseUserAndProductID(msg)
		if err != nil {
			s.logger.Printf("Parse error: %v", err)
			return nil
		}
		if err := s.repo.IncrementProductDislikes(ctx, productID); err != nil {
			s.logger.Printf("Failed to increment product dislikes: %v", err)
		}
		if err := s.repo.IncrementUserDislikes(ctx, userID); err != nil {
			s.logger.Printf("Failed to increment user dislikes: %v", err)
		}

	case "user_purchased":
		s.logger.Println("Handling 'user_purchased' event")
		userID, productID, err := parseUserAndProductID(msg)
		if err != nil {
			s.logger.Printf("Parse error: %v", err)
			return nil
		}
		if err := s.repo.IncrementProductPurchases(ctx, productID); err != nil {
			s.logger.Printf("Failed to increment product purchases: %v", err)
		}
		if err := s.repo.IncrementUserPurchases(ctx, userID); err != nil {
			s.logger.Printf("Failed to increment user purchases: %v", err)
		}

	case "product_created", "product_updated":
		s.logger.Printf("[INFO] Product event: %s", event)

	case "user_created", "user_updated":
		s.logger.Printf("[INFO] User event: %s", event)

	default:
		s.logger.Printf("[WARN] Unhandled event type: %s", event)
	}

	s.logger.Println("Kafka message processing completed")
	return nil
}

func parseUserAndProductID(msg map[string]interface{}) (int64, int64, error) {
	u, ok1 := msg["user_id"].(float64)
	p, ok2 := msg["product_id"].(float64)
	if !ok1 || !ok2 {
		return 0, 0, fmt.Errorf("failed to parse user_id or product_id")
	}
	return int64(u), int64(p), nil
}

func (s *analyticsService) GetProductAnalytics(ctx context.Context, productID int64) (*models.ProductAnalytics, error) {
	s.logger.Printf("Fetching analytics for product ID: %d", productID)
	return s.repo.GetProductAnalytics(ctx, productID)
}

func (s *analyticsService) GetUserAnalytics(ctx context.Context, userID int64) (*models.UserAnalytics, error) {
	s.logger.Printf("Fetching analytics for user ID: %d", userID)
	return s.repo.GetUserAnalytics(ctx, userID)
}
