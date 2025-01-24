package repository

import (
	"context"
	"fmt"
	log "recommendation-system/pkg/logger"

	"recommendation-system/internal/product/models"
	"recommendation-system/pkg/db"
)

type ProductRepository interface {
	CreateProduct(ctx context.Context, product *models.Product) error
	GetProductByID(ctx context.Context, id int64) (*models.Product, error)
	UpdateProduct(ctx context.Context, product *models.Product) error
	GetAllProducts(ctx context.Context, limit, offset int) ([]*models.Product, error)
	DeleteProduct(ctx context.Context, id int64) error
}

type productRepository struct {
	db     *db.DB
	logger *log.Logger
}

func NewProductRepository(database *db.DB, logger *log.Logger) ProductRepository {
	return &productRepository{db: database, logger: logger}
}

func (r *productRepository) CreateProduct(ctx context.Context, product *models.Product) error {
	r.logger.Printf("Creating product: %s", product.Name)
	query := `
        INSERT INTO products (name, description, price, category, created_at, updated_at)
        VALUES ($1, $2, $3, $4, NOW(), NOW())
        RETURNING id, created_at, updated_at
    `
	err := r.db.Pool.QueryRow(ctx, query,
		product.Name,
		product.Description,
		product.Price,
		product.Category,
	).Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)
	if err != nil {
		r.logger.Printf("Failed to create product: %v", err)
		return fmt.Errorf("failed to create product: %w", err)
	}
	r.logger.Printf("Successfully created product with ID: %d", product.ID)
	return nil
}

func (r *productRepository) GetProductByID(ctx context.Context, id int64) (*models.Product, error) {
	r.logger.Printf("Fetching product with ID: %d", id)
	var product models.Product
	query := `
        SELECT id, name, description, price, category, created_at, updated_at
        FROM products
        WHERE id = $1
    `
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.Category,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err != nil {
		r.logger.Printf("Failed to fetch product: %v", err)
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	r.logger.Printf("Fetching likes for product ID: %d", id)
	likesQuery := `
        SELECT id, user_id, product_id, liked_at
        FROM likes
        WHERE product_id = $1
    `
	rows, err := r.db.Pool.Query(ctx, likesQuery, id)
	if err != nil {
		r.logger.Printf("Failed to fetch likes: %v", err)
		return nil, fmt.Errorf("failed to get likes for product: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var like models.Like
		if err := rows.Scan(&like.ID, &like.UserID, &like.ProductID, &like.LikedAt); err != nil {
			r.logger.Printf("Failed to scan like: %v", err)
			return nil, fmt.Errorf("failed to scan like: %w", err)
		}
		product.Likes = append(product.Likes, &like)
	}
	if err := rows.Err(); err != nil {
		r.logger.Printf("Rows error while fetching likes: %v", err)
		return nil, fmt.Errorf("rows error: %w", err)
	}

	r.logger.Printf("Fetching dislikes for product ID: %d", id)
	dislikesQuery := `
        SELECT id, user_id, product_id, disliked_at
        FROM dislikes
        WHERE product_id = $1
    `
	rows, err = r.db.Pool.Query(ctx, dislikesQuery, id)
	if err != nil {
		r.logger.Printf("Failed to fetch dislikes: %v", err)
		return nil, fmt.Errorf("failed to get dislikes for product: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var dislike models.Dislike
		if err := rows.Scan(&dislike.ID, &dislike.UserID, &dislike.ProductID, &dislike.DislikedAt); err != nil {
			r.logger.Printf("Failed to scan dislike: %v", err)
			return nil, fmt.Errorf("failed to scan dislike: %w", err)
		}
		product.Dislikes = append(product.Dislikes, &dislike)
	}
	if err := rows.Err(); err != nil {
		r.logger.Printf("Rows error while fetching dislikes: %v", err)
		return nil, fmt.Errorf("rows error: %w", err)
	}

	r.logger.Printf("Fetching purchase count for product ID: %d", id)
	purchaseCountQuery := `
        SELECT COUNT(*) FROM purchases WHERE product_id = $1
    `
	err = r.db.Pool.QueryRow(ctx, purchaseCountQuery, id).Scan(&product.PurchaseCount)
	if err != nil {
		r.logger.Printf("Failed to fetch purchase count: %v", err)
		return nil, fmt.Errorf("failed to get purchase count for product: %w", err)
	}

	r.logger.Printf("Successfully fetched product with ID: %d", id)
	return &product, nil
}

func (r *productRepository) UpdateProduct(ctx context.Context, product *models.Product) error {
	r.logger.Printf("Updating product with ID: %d", product.ID)
	query := `
        UPDATE products
        SET name = $1, description = $2, price = $3, category = $4, updated_at = NOW()
        WHERE id = $5
        RETURNING updated_at
    `
	err := r.db.Pool.QueryRow(ctx, query,
		product.Name,
		product.Description,
		product.Price,
		product.Category,
		product.ID,
	).Scan(&product.UpdatedAt)
	if err != nil {
		r.logger.Printf("Failed to update product: %v", err)
		return fmt.Errorf("failed to update product: %w", err)
	}
	r.logger.Printf("Successfully updated product with ID: %d", product.ID)
	return nil
}

func (r *productRepository) GetAllProducts(ctx context.Context, limit, offset int) ([]*models.Product, error) {
	r.logger.Printf("Fetching all products with limit: %d, offset: %d", limit, offset)
	var products []*models.Product
	query := `
        SELECT id, name, description, price, category, created_at, updated_at
        FROM products
        ORDER BY id
        LIMIT $1 OFFSET $2
    `
	rows, err := r.db.Pool.Query(ctx, query, limit, offset)
	if err != nil {
		r.logger.Printf("Failed to fetch products: %v", err)
		return nil, fmt.Errorf("failed to get products: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var product models.Product
		if err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.Category,
			&product.CreatedAt,
			&product.UpdatedAt,
		); err != nil {
			r.logger.Printf("Failed to scan product: %v", err)
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, &product)
	}

	if err := rows.Err(); err != nil {
		r.logger.Printf("Rows error while fetching products: %v", err)
		return nil, fmt.Errorf("rows error: %w", err)
	}

	r.logger.Printf("Successfully fetched %d products", len(products))
	return products, nil
}

func (r *productRepository) DeleteProduct(ctx context.Context, id int64) error {
	r.logger.Printf("Deleting product with ID: %d", id)
	query := `
        DELETE FROM products
        WHERE id = $1
    `
	cmdTag, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		r.logger.Printf("Failed to delete product: %v", err)
		return fmt.Errorf("failed to delete product: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		r.logger.Printf("No product found with ID: %d", id)
		return fmt.Errorf("no product found with id %d", id)
	}
	r.logger.Printf("Successfully deleted product with ID: %d", id)
	return nil
}
