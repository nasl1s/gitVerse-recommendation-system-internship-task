package repository

import (
	"context"
	"fmt"
	log "recommendation-system/pkg/logger"

	"recommendation-system/internal/recommendation/models"
	"recommendation-system/pkg/db"
)

type RecommendationRepository interface {
	CreateRecommendation(ctx context.Context, rec *models.Recommendation) error
	GetRecommendationsByUserID(ctx context.Context, userID int64) ([]*models.Recommendation, error)
	GetAllUserIDs(ctx context.Context) ([]int64, error)
	UpdateUserCategoryScore(ctx context.Context, userID int64, category string, delta float64) error
	GetTopProductsByUserPreference(ctx context.Context, userID int64, limit int) ([]int64, error)
	GetProductCategory(ctx context.Context, productID int64) (string, error)
}

type recommendationRepository struct {
	db     *db.DB
	logger *log.Logger
}

func NewRecommendationRepository(database *db.DB, logger *log.Logger) RecommendationRepository {
	return &recommendationRepository{db: database, logger: logger}
}

func (r *recommendationRepository) CreateRecommendation(ctx context.Context, rec *models.Recommendation) error {
	r.logger.Printf("Creating recommendation for user ID: %d", rec.UserID)
	query := `
        INSERT INTO recommendations (user_id, product_ids, created_at)
        VALUES ($1, $2, NOW())
        RETURNING id, created_at
    `
	err := r.db.Pool.QueryRow(ctx, query, rec.UserID, rec.ProductIDs).Scan(&rec.ID, &rec.CreatedAt)
	if err != nil {
		r.logger.Printf("Failed to create recommendation: %v", err)
		return fmt.Errorf("failed to create recommendation: %w", err)
	}
	r.logger.Printf("Successfully created recommendation with ID: %d", rec.ID)
	return nil
}

func (r *recommendationRepository) GetRecommendationsByUserID(ctx context.Context, userID int64) ([]*models.Recommendation, error) {
	r.logger.Printf("Fetching recommendations for user ID: %d", userID)
	var recommendations []*models.Recommendation
	query := `
        SELECT id, user_id, product_ids, created_at
        FROM recommendations
        WHERE user_id = $1
    `
	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		r.logger.Printf("Failed to get recommendations: %v", err)
		return nil, fmt.Errorf("failed to get recommendations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var rec models.Recommendation
		if err := rows.Scan(&rec.ID, &rec.UserID, &rec.ProductIDs, &rec.CreatedAt); err != nil {
			r.logger.Printf("Failed to scan recommendation: %v", err)
			return nil, fmt.Errorf("failed to scan recommendation: %w", err)
		}
		recommendations = append(recommendations, &rec)
	}

	if err := rows.Err(); err != nil {
		r.logger.Printf("Rows error: %v", err)
		return nil, fmt.Errorf("rows error: %w", err)
	}

	r.logger.Printf("Successfully fetched %d recommendations for user ID: %d", len(recommendations), userID)
	return recommendations, nil
}

func (r *recommendationRepository) GetAllUserIDs(ctx context.Context) ([]int64, error) {
	r.logger.Println("Fetching all user IDs")
	var userIDs []int64
	query := `
        SELECT id FROM users
    `
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		r.logger.Printf("Failed to get user IDs: %v", err)
		return nil, fmt.Errorf("failed to get user IDs: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			r.logger.Printf("Failed to scan user ID: %v", err)
			return nil, fmt.Errorf("failed to scan user ID: %w", err)
		}
		userIDs = append(userIDs, userID)
	}

	if err := rows.Err(); err != nil {
		r.logger.Printf("Rows error: %v", err)
		return nil, fmt.Errorf("rows error: %w", err)
	}

	r.logger.Printf("Successfully fetched %d user IDs", len(userIDs))
	return userIDs, nil
}

func (r *recommendationRepository) UpdateUserCategoryScore(ctx context.Context, userID int64, category string, delta float64) error {
	r.logger.Printf("Updating category score for user ID: %d, category: %s, delta: %.2f", userID, category, delta)
	query := `
        INSERT INTO user_category_preferences (user_id, category, score)
        VALUES ($1, $2, $3)
        ON CONFLICT (user_id, category)
        DO UPDATE SET score = user_category_preferences.score + EXCLUDED.score
    `
	_, err := r.db.Pool.Exec(ctx, query, userID, category, delta)
	if err != nil {
		r.logger.Printf("Failed to update category score: %v", err)
		return fmt.Errorf("failed to update category score: %w", err)
	}
	r.logger.Printf("Successfully updated category score for user ID: %d, category: %s", userID, category)
	return nil
}

func (r *recommendationRepository) GetTopProductsByUserPreference(ctx context.Context, userID int64, limit int) ([]int64, error) {
	r.logger.Printf("Fetching top products by user preference for user ID: %d", userID)
	catQuery := `
        SELECT category, score 
        FROM user_category_preferences
        WHERE user_id = $1
        ORDER BY score DESC
        LIMIT 10
    `
	rows, err := r.db.Pool.Query(ctx, catQuery, userID)
	if err != nil {
		r.logger.Printf("Failed to get user categories: %v", err)
		return nil, fmt.Errorf("failed to get user categories: %w", err)
	}
	defer rows.Close()

	var topCategories []string
	for rows.Next() {
		var cat string
		var sc float64
		if err := rows.Scan(&cat, &sc); err != nil {
			r.logger.Printf("Failed to scan categories: %v", err)
			return nil, fmt.Errorf("failed to scan categories: %w", err)
		}
		topCategories = append(topCategories, cat)
	}
	if err := rows.Err(); err != nil {
		r.logger.Printf("Rows error: %v", err)
		return nil, fmt.Errorf("rows error: %w", err)
	}

	if len(topCategories) == 0 {
		r.logger.Println("No categories found for user preferences")
		return []int64{}, nil
	}

	prodQuery := `
        SELECT id 
        FROM products
        WHERE category = ANY($1)
        ORDER BY updated_at DESC
        LIMIT $2
    `
	rows2, err := r.db.Pool.Query(ctx, prodQuery, topCategories, limit)
	if err != nil {
		r.logger.Printf("Failed to get top products: %v", err)
		return nil, fmt.Errorf("failed to get top products: %w", err)
	}
	defer rows2.Close()

	var productIDs []int64
	for rows2.Next() {
		var pid int64
		if err := rows2.Scan(&pid); err != nil {
			r.logger.Printf("Failed to scan product ID: %v", err)
			return nil, fmt.Errorf("failed to scan product ID: %w", err)
		}
		productIDs = append(productIDs, pid)
	}
	if err := rows2.Err(); err != nil {
		r.logger.Printf("Rows2 error: %v", err)
		return nil, fmt.Errorf("rows2 error: %w", err)
	}

	r.logger.Printf("Successfully fetched %d top products for user ID: %d", len(productIDs), userID)
	return productIDs, nil
}

func (r *recommendationRepository) GetProductCategory(ctx context.Context, productID int64) (string, error) {
	r.logger.Printf("Fetching category for product ID: %d", productID)
	query := `SELECT category FROM products WHERE id = $1`
	var cat string
	err := r.db.Pool.QueryRow(ctx, query, productID).Scan(&cat)
	if err != nil {
		r.logger.Printf("Failed to get product category: %v", err)
		return "", fmt.Errorf("failed to get product category: %w", err)
	}
	r.logger.Printf("Successfully fetched category for product ID: %d", productID)
	return cat, nil
}
