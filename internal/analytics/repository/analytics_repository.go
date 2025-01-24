package repository

import (
	"context"
	"fmt"
	log "recommendation-system/pkg/logger"

	"recommendation-system/internal/analytics/models"
	"recommendation-system/pkg/db"
)

type AnalyticsRepository interface {
	IncrementProductLikes(ctx context.Context, productID int64) error
	IncrementProductDislikes(ctx context.Context, productID int64) error
	IncrementProductPurchases(ctx context.Context, productID int64) error
	GetProductAnalytics(ctx context.Context, productID int64) (*models.ProductAnalytics, error)

	IncrementUserLikes(ctx context.Context, userID int64) error
	IncrementUserDislikes(ctx context.Context, userID int64) error
	IncrementUserPurchases(ctx context.Context, userID int64) error
	GetUserAnalytics(ctx context.Context, userID int64) (*models.UserAnalytics, error)
}

type analyticsRepository struct {
	db     *db.DB
	logger *log.Logger
}

func NewAnalyticsRepository(database *db.DB, logger *log.Logger) AnalyticsRepository {
	return &analyticsRepository{db: database, logger: logger}
}

func (r *analyticsRepository) IncrementProductLikes(ctx context.Context, productID int64) error {
	r.logger.Printf("Incrementing likes for product ID: %d", productID)
	query := `
        INSERT INTO product_analytics (product_id, likes)
        VALUES ($1, 1)
        ON CONFLICT (product_id)
        DO UPDATE SET likes = product_analytics.likes + 1, updated_at = NOW()
    `
	_, err := r.db.Pool.Exec(ctx, query, productID)
	if err != nil {
		r.logger.Printf("Failed to increment product likes: %v", err)
		return fmt.Errorf("failed to increment product likes: %w", err)
	}
	r.logger.Printf("Successfully incremented likes for product ID: %d", productID)
	return nil
}

func (r *analyticsRepository) IncrementProductDislikes(ctx context.Context, productID int64) error {
	r.logger.Printf("Incrementing dislikes for product ID: %d", productID)
	query := `
        INSERT INTO product_analytics (product_id, dislikes)
        VALUES ($1, 1)
        ON CONFLICT (product_id)
        DO UPDATE SET dislikes = product_analytics.dislikes + 1, updated_at = NOW()
    `
	_, err := r.db.Pool.Exec(ctx, query, productID)
	if err != nil {
		r.logger.Printf("Failed to increment product dislikes: %v", err)
		return fmt.Errorf("failed to increment product dislikes: %w", err)
	}
	r.logger.Printf("Successfully incremented dislikes for product ID: %d", productID)
	return nil
}

func (r *analyticsRepository) IncrementProductPurchases(ctx context.Context, productID int64) error {
	r.logger.Printf("Incrementing purchases for product ID: %d", productID)
	query := `
        INSERT INTO product_analytics (product_id, purchases)
        VALUES ($1, 1)
        ON CONFLICT (product_id)
        DO UPDATE SET purchases = product_analytics.purchases + 1, updated_at = NOW()
    `
	_, err := r.db.Pool.Exec(ctx, query, productID)
	if err != nil {
		r.logger.Printf("Failed to increment product purchases: %v", err)
		return fmt.Errorf("failed to increment product purchases: %w", err)
	}
	r.logger.Printf("Successfully incremented purchases for product ID: %d", productID)
	return nil
}

func (r *analyticsRepository) GetProductAnalytics(ctx context.Context, productID int64) (*models.ProductAnalytics, error) {
	r.logger.Printf("Fetching analytics for product ID: %d", productID)
	query := `
        SELECT id, product_id, likes, dislikes, purchases, updated_at
        FROM product_analytics
        WHERE product_id = $1
    `
	var pa models.ProductAnalytics
	err := r.db.Pool.QueryRow(ctx, query, productID).Scan(
		&pa.ID, &pa.ProductID, &pa.Likes, &pa.Dislikes, &pa.Purchases, &pa.UpdatedAt,
	)
	if err != nil {
		r.logger.Printf("Failed to fetch analytics for product ID: %d, error: %v", productID, err)
		return nil, fmt.Errorf("failed to get product analytics: %w", err)
	}
	r.logger.Printf("Successfully fetched analytics for product ID: %d", productID)
	return &pa, nil
}

func (r *analyticsRepository) IncrementUserLikes(ctx context.Context, userID int64) error {
	r.logger.Printf("Incrementing likes for user ID: %d", userID)
	query := `
        INSERT INTO user_analytics (user_id, total_likes)
        VALUES ($1, 1)
        ON CONFLICT (user_id)
        DO UPDATE SET total_likes = user_analytics.total_likes + 1, updated_at = NOW()
    `
	_, err := r.db.Pool.Exec(ctx, query, userID)
	if err != nil {
		r.logger.Printf("Failed to increment user likes: %v", err)
		return fmt.Errorf("failed to increment user likes: %w", err)
	}
	r.logger.Printf("Successfully incremented likes for user ID: %d", userID)
	return nil
}

func (r *analyticsRepository) IncrementUserDislikes(ctx context.Context, userID int64) error {
	r.logger.Printf("Incrementing dislikes for user ID: %d", userID)
	query := `
        INSERT INTO user_analytics (user_id, total_dislikes)
        VALUES ($1, 1)
        ON CONFLICT (user_id)
        DO UPDATE SET total_dislikes = user_analytics.total_dislikes + 1, updated_at = NOW()
    `
	_, err := r.db.Pool.Exec(ctx, query, userID)
	if err != nil {
		r.logger.Printf("Failed to increment user dislikes: %v", err)
		return fmt.Errorf("failed to increment user dislikes: %w", err)
	}
	r.logger.Printf("Successfully incremented dislikes for user ID: %d", userID)
	return nil
}

func (r *analyticsRepository) IncrementUserPurchases(ctx context.Context, userID int64) error {
	r.logger.Printf("Incrementing purchases for user ID: %d", userID)
	query := `
        INSERT INTO user_analytics (user_id, total_purchases)
        VALUES ($1, 1)
        ON CONFLICT (user_id)
        DO UPDATE SET total_purchases = user_analytics.total_purchases + 1, updated_at = NOW()
    `
	_, err := r.db.Pool.Exec(ctx, query, userID)
	if err != nil {
		r.logger.Printf("Failed to increment user purchases: %v", err)
		return fmt.Errorf("failed to increment user purchases: %w", err)
	}
	r.logger.Printf("Successfully incremented purchases for user ID: %d", userID)
	return nil
}

func (r *analyticsRepository) GetUserAnalytics(ctx context.Context, userID int64) (*models.UserAnalytics, error) {
	r.logger.Printf("Fetching analytics for user ID: %d", userID)
	query := `
        SELECT id, user_id, total_likes, total_dislikes, total_purchases, updated_at
        FROM user_analytics
        WHERE user_id = $1
    `
	var ua models.UserAnalytics
	err := r.db.Pool.QueryRow(ctx, query, userID).Scan(
		&ua.ID, &ua.UserID, &ua.TotalLikes, &ua.TotalDislikes, &ua.TotalPurchases, &ua.UpdatedAt,
	)
	if err != nil {
		r.logger.Printf("Failed to fetch analytics for user ID: %d, error: %v", userID, err)
		return nil, fmt.Errorf("failed to get user analytics: %w", err)
	}
	r.logger.Printf("Successfully fetched analytics for user ID: %d", userID)
	return &ua, nil
}
