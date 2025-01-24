package service

import (
	"context"
	"encoding/json"
	"fmt"
	log "recommendation-system/pkg/logger"

	"recommendation-system/internal/product/models"
	"recommendation-system/internal/product/repository"
	"recommendation-system/pkg/kafka"
)

type ProductService interface {
	CreateProduct(ctx context.Context, product *models.Product) error
	GetProduct(ctx context.Context, id int64) (*models.Product, error)
	UpdateProduct(ctx context.Context, product *models.Product) error
	GetAllProducts(ctx context.Context, limit, offset int) ([]*models.Product, error)
	DeleteProduct(ctx context.Context, id int64) error
}

type productService struct {
	repo   repository.ProductRepository
	kafka  *kafka.KafkaClient
	topic  string
	logger *log.Logger
}

func NewProductService(repo repository.ProductRepository, kafkaClient *kafka.KafkaClient, logger *log.Logger) ProductService {
	return &productService{
		repo:   repo,
		kafka:  kafkaClient,
		topic:  "product_updates",
		logger: logger,
	}
}

func (s *productService) CreateProduct(ctx context.Context, product *models.Product) error {
	s.logger.Printf("Creating product with ID: %d", product.ID)
	if err := s.repo.CreateProduct(ctx, product); err != nil {
		s.logger.Printf("Failed to create product: %v", err)
		return err
	}

	message := map[string]interface{}{
		"event":   "product_created",
		"product": product,
	}
	s.logger.Println("Publishing product creation event")
	return s.publishMessage(message)
}

func (s *productService) GetProduct(ctx context.Context, id int64) (*models.Product, error) {
	s.logger.Printf("Fetching product with ID: %d", id)
	product, err := s.repo.GetProductByID(ctx, id)
	if err != nil {
		s.logger.Printf("Failed to fetch product: %v", err)
	}
	return product, err
}

func (s *productService) UpdateProduct(ctx context.Context, product *models.Product) error {
	s.logger.Printf("Updating product with ID: %d", product.ID)
	if err := s.repo.UpdateProduct(ctx, product); err != nil {
		s.logger.Printf("Failed to update product: %v", err)
		return err
	}

	message := map[string]interface{}{
		"event":   "product_updated",
		"product": product,
	}
	s.logger.Println("Publishing product update event")
	return s.publishMessage(message)
}

func (s *productService) GetAllProducts(ctx context.Context, limit, offset int) ([]*models.Product, error) {
	s.logger.Printf("Fetching all products with limit: %d, offset: %d", limit, offset)
	products, err := s.repo.GetAllProducts(ctx, limit, offset)
	if err != nil {
		s.logger.Printf("Failed to fetch products: %v", err)
	}
	return products, err
}

func (s *productService) DeleteProduct(ctx context.Context, id int64) error {
	s.logger.Printf("Deleting product with ID: %d", id)
	if err := s.repo.DeleteProduct(ctx, id); err != nil {
		s.logger.Printf("Failed to delete product: %v", err)
		return err
	}

	message := map[string]interface{}{
		"event":      "product_deleted",
		"product_id": id,
	}
	s.logger.Println("Publishing product deletion event")
	return s.publishMessage(message)
}

func (s *productService) publishMessage(message interface{}) error {
	valueBytes, err := json.Marshal(message)
	if err != nil {
		s.logger.Printf("Failed to marshal message: %v", err)
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	s.logger.Printf("Publishing message to topic %s", s.topic)
	return s.kafka.PublishMessage(s.topic, nil, valueBytes)
}
