package repository

import (
	"context"
	"fmt"

	"recommendation-system/internal/user/models"
	"recommendation-system/pkg/db"
	log "recommendation-system/pkg/logger"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByID(ctx context.Context, id int64) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	GetAllUsers(ctx context.Context, limit, offset int) ([]*models.User, error)
	CreatePurchase(ctx context.Context, purchase *models.Purchase) error
	CreateLike(ctx context.Context, like *models.Like) error
	LikeExists(ctx context.Context, userID, productID int64) (bool, error)
	CreateDislike(ctx context.Context, dislike *models.Dislike) error
	DislikeExists(ctx context.Context, userID, productID int64) (bool, error)
	RemoveLikeByUserAndProduct(ctx context.Context, userID, productID int64) error
	RemoveDislikeByUserAndProduct(ctx context.Context, userID, productID int64) error
	GetUserActions(ctx context.Context, userID int64, productID *int64) (map[string][]interface{}, error)
	GetUserPurchases(ctx context.Context, userID int64, limit, offset int) ([]*models.Purchase, error)
}

type userRepository struct {
	db     *db.DB
	logger *log.Logger
}

func NewUserRepository(database *db.DB, logger *log.Logger) UserRepository {
	return &userRepository{db: database, logger: logger}
}

func (r *userRepository) CreateUser(ctx context.Context, user *models.User) error {
	r.logger.Println("Creating a new user")
	query := `
        INSERT INTO users (name, email, created_at, updated_at)
        VALUES ($1, $2, NOW(), NOW())
        RETURNING id, created_at, updated_at
    `
	err := r.db.Pool.QueryRow(ctx, query, user.Name, user.Email).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		r.logger.Printf("Failed to create user: %v", err)
		return fmt.Errorf("failed to create user: %w", err)
	}
	r.logger.Printf("User created successfully with ID: %d", user.ID)
	return nil
}

func (r *userRepository) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	r.logger.Printf("Fetching user by ID: %d", id)
	var user models.User
	query := `
        SELECT id, name, email, created_at, updated_at
        FROM users
        WHERE id = $1
    `
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		r.logger.Printf("Failed to fetch user by ID %d: %v", id, err)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	r.logger.Printf("Successfully fetched user by ID: %d", id)
	return &user, nil
}

func (r *userRepository) UpdateUser(ctx context.Context, user *models.User) error {
	r.logger.Printf("Updating user with ID: %d", user.ID)
	query := `
        UPDATE users
        SET name = $1, email = $2, updated_at = NOW()
        WHERE id = $3
        RETURNING updated_at
    `
	err := r.db.Pool.QueryRow(ctx, query, user.Name, user.Email, user.ID).Scan(&user.UpdatedAt)
	if err != nil {
		r.logger.Printf("Failed to update user with ID %d: %v", user.ID, err)
		return fmt.Errorf("failed to update user: %w", err)
	}
	r.logger.Printf("Successfully updated user with ID: %d", user.ID)
	return nil
}

func (r *userRepository) GetAllUsers(ctx context.Context, limit, offset int) ([]*models.User, error) {
	r.logger.Printf("Fetching all users with limit %d and offset %d", limit, offset)
	var users []*models.User
	query := `
        SELECT id, name, email, created_at, updated_at
        FROM users
        ORDER BY id
        LIMIT $1 OFFSET $2
    `
	rows, err := r.db.Pool.Query(ctx, query, limit, offset)
	if err != nil {
		r.logger.Printf("Failed to fetch users: %v", err)
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt); err != nil {
			r.logger.Printf("Failed to scan user: %v", err)
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		r.logger.Printf("Rows error: %v", err)
		return nil, fmt.Errorf("rows error: %w", err)
	}

	r.logger.Printf("Successfully fetched %d users", len(users))
	return users, nil
}

func (r *userRepository) CreatePurchase(ctx context.Context, purchase *models.Purchase) error {
	r.logger.Printf("Creating purchase for user ID: %d and product ID: %d", purchase.UserID, purchase.ProductID)
	query := `
        INSERT INTO purchases (user_id, product_id, purchased_at)
        VALUES ($1, $2, NOW())
        RETURNING id, purchased_at
    `
	err := r.db.Pool.QueryRow(ctx, query, purchase.UserID, purchase.ProductID).Scan(&purchase.ID, &purchase.PurchasedAt)
	if err != nil {
		r.logger.Printf("Failed to create purchase: %v", err)
		return fmt.Errorf("failed to create purchase: %w", err)
	}
	r.logger.Printf("Purchase created successfully with ID: %d", purchase.ID)
	return nil
}

func (r *userRepository) LikeExists(ctx context.Context, userID, productID int64) (bool, error) {
	r.logger.Printf("Checking if like exists for user ID: %d and product ID: %d", userID, productID)
	query := `
        SELECT 1 
        FROM likes
        WHERE user_id = $1 AND product_id = $2
        LIMIT 1
    `
	var dummy int
	err := r.db.Pool.QueryRow(ctx, query, userID, productID).Scan(&dummy)
	if err != nil {
		r.logger.Printf("Like does not exist for user ID: %d and product ID: %d", userID, productID)
		return false, nil
	}
	r.logger.Printf("Like exists for user ID: %d and product ID: %d", userID, productID)
	return true, nil
}

func (r *userRepository) CreateLike(ctx context.Context, like *models.Like) error {
	r.logger.Printf("Creating like for user ID: %d and product ID: %d", like.UserID, like.ProductID)
	query := `
        INSERT INTO likes (user_id, product_id, liked_at)
        VALUES ($1, $2, NOW())
        RETURNING id, liked_at
    `
	err := r.db.Pool.QueryRow(ctx, query, like.UserID, like.ProductID).Scan(&like.ID, &like.LikedAt)
	if err != nil {
		r.logger.Printf("Failed to create like: %v", err)
		return fmt.Errorf("failed to create like: %w", err)
	}
	r.logger.Printf("Like created successfully with ID: %d", like.ID)
	return nil
}

func (r *userRepository) DislikeExists(ctx context.Context, userID, productID int64) (bool, error) {
	r.logger.Printf("Checking if dislike exists for user ID: %d and product ID: %d", userID, productID)
	query := `
        SELECT 1 
        FROM dislikes
        WHERE user_id = $1 AND product_id = $2
        LIMIT 1
    `
	var dummy int
	err := r.db.Pool.QueryRow(ctx, query, userID, productID).Scan(&dummy)
	if err != nil {
		r.logger.Printf("Dislike does not exist for user ID: %d and product ID: %d", userID, productID)
		return false, nil
	}
	r.logger.Printf("Dislike exists for user ID: %d and product ID: %d", userID, productID)
	return true, nil
}

func (r *userRepository) CreateDislike(ctx context.Context, dislike *models.Dislike) error {
	r.logger.Printf("Creating dislike for user ID: %d and product ID: %d", dislike.UserID, dislike.ProductID)
	query := `
        INSERT INTO dislikes (user_id, product_id, disliked_at)
        VALUES ($1, $2, NOW())
        RETURNING id, disliked_at
    `
	err := r.db.Pool.QueryRow(ctx, query, dislike.UserID, dislike.ProductID).Scan(&dislike.ID, &dislike.DislikedAt)
	if err != nil {
		r.logger.Printf("Failed to create dislike: %v", err)
		return fmt.Errorf("failed to create dislike: %w", err)
	}
	r.logger.Printf("Dislike created successfully with ID: %d", dislike.ID)
	return nil
}

func (r *userRepository) RemoveLikeByUserAndProduct(ctx context.Context, userID, productID int64) error {
	r.logger.Printf("Removing like for user ID: %d and product ID: %d", userID, productID)
	query := `DELETE FROM likes WHERE user_id = $1 AND product_id = $2`
	_, err := r.db.Pool.Exec(ctx, query, userID, productID)
	if err != nil {
		r.logger.Printf("Failed to remove like: %v", err)
		return fmt.Errorf("failed to remove like: %w", err)
	}
	r.logger.Printf("Successfully removed like for user ID: %d and product ID: %d", userID, productID)
	return nil
}

func (r *userRepository) RemoveDislikeByUserAndProduct(ctx context.Context, userID, productID int64) error {
	r.logger.Printf("Removing dislike for user ID: %d and product ID: %d", userID, productID)
	query := `DELETE FROM dislikes WHERE user_id = $1 AND product_id = $2`
	_, err := r.db.Pool.Exec(ctx, query, userID, productID)
	if err != nil {
		r.logger.Printf("Failed to remove dislike: %v", err)
		return fmt.Errorf("failed to remove dislike: %w", err)
	}
	r.logger.Printf("Successfully removed dislike for user ID: %d and product ID: %d", userID, productID)
	return nil
}

func (r *userRepository) GetUserActions(ctx context.Context, userID int64, productID *int64) (map[string][]interface{}, error) {
	r.logger.Printf("Fetching actions for user ID: %d", userID)
	actions := make(map[string][]interface{})

	var likesQuery string
	var likesArgs []interface{}
	if productID != nil {
		likesQuery = `
            SELECT id, user_id, product_id, liked_at
            FROM likes
            WHERE user_id = $1 AND product_id = $2
        `
		likesArgs = []interface{}{userID, *productID}
	} else {
		likesQuery = `
            SELECT id, user_id, product_id, liked_at
            FROM likes
            WHERE user_id = $1
        `
		likesArgs = []interface{}{userID}
	}

	rows, err := r.db.Pool.Query(ctx, likesQuery, likesArgs...)
	if err != nil {
		r.logger.Printf("Failed to get likes: %v", err)
		return nil, fmt.Errorf("failed to get likes: %w", err)
	}
	defer rows.Close()

	var likes []*models.Like
	for rows.Next() {
		var like models.Like
		if err := rows.Scan(&like.ID, &like.UserID, &like.ProductID, &like.LikedAt); err != nil {
			r.logger.Printf("Failed to scan like: %v", err)
			return nil, fmt.Errorf("failed to scan like: %w", err)
		}
		likes = append(likes, &like)
	}
	if err := rows.Err(); err != nil {
		r.logger.Printf("Rows error (likes): %v", err)
		return nil, fmt.Errorf("rows error (likes): %w", err)
	}

	actions["likes"] = make([]interface{}, len(likes))
	for i, v := range likes {
		actions["likes"][i] = v
	}

	var dislikesQuery string
	var dislikesArgs []interface{}
	if productID != nil {
		dislikesQuery = `
            SELECT id, user_id, product_id, disliked_at
            FROM dislikes
            WHERE user_id = $1 AND product_id = $2
        `
		dislikesArgs = []interface{}{userID, *productID}
	} else {
		dislikesQuery = `
            SELECT id, user_id, product_id, disliked_at
            FROM dislikes
            WHERE user_id = $1
        `
		dislikesArgs = []interface{}{userID}
	}

	rows, err = r.db.Pool.Query(ctx, dislikesQuery, dislikesArgs...)
	if err != nil {
		r.logger.Printf("Failed to get dislikes: %v", err)
		return nil, fmt.Errorf("failed to get dislikes: %w", err)
	}
	defer rows.Close()

	var dislikes []*models.Dislike
	for rows.Next() {
		var dislike models.Dislike
		if err := rows.Scan(&dislike.ID, &dislike.UserID, &dislike.ProductID, &dislike.DislikedAt); err != nil {
			r.logger.Printf("Failed to scan dislike: %v", err)
			return nil, fmt.Errorf("failed to scan dislike: %w", err)
		}
		dislikes = append(dislikes, &dislike)
	}
	if err := rows.Err(); err != nil {
		r.logger.Printf("Rows error (dislikes): %v", err)
		return nil, fmt.Errorf("rows error (dislikes): %w", err)
	}

	actions["dislikes"] = make([]interface{}, len(dislikes))
	for i, v := range dislikes {
		actions["dislikes"][i] = v
	}

	purchases, err := r.GetUserPurchases(ctx, userID, 999999, 0)
	if err != nil {
		r.logger.Printf("Failed to get purchases: %v", err)
		return nil, fmt.Errorf("failed to get purchases: %w", err)
	}

	var filteredPurchases []interface{}
	for _, p := range purchases {
		if productID == nil || p.ProductID == *productID {
			filteredPurchases = append(filteredPurchases, p)
		}
	}
	actions["purchases"] = filteredPurchases

	r.logger.Printf("Successfully fetched actions for user ID: %d", userID)
	return actions, nil
}

func (r *userRepository) GetUserPurchases(ctx context.Context, userID int64, limit, offset int) ([]*models.Purchase, error) {
	r.logger.Printf("Fetching purchases for user ID: %d with limit %d and offset %d", userID, limit, offset)
	var purchases []*models.Purchase
	query := `
        SELECT id, user_id, product_id, purchased_at
        FROM purchases
        WHERE user_id = $1
        ORDER BY purchased_at DESC
        LIMIT $2 OFFSET $3
    `
	rows, err := r.db.Pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		r.logger.Printf("Failed to get purchases: %v", err)
		return nil, fmt.Errorf("failed to get purchases: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var purchase models.Purchase
		if err := rows.Scan(&purchase.ID, &purchase.UserID, &purchase.ProductID, &purchase.PurchasedAt); err != nil {
			r.logger.Printf("Failed to scan purchase: %v", err)
			return nil, fmt.Errorf("failed to scan purchase: %w", err)
		}
		purchases = append(purchases, &purchase)
	}

	if err := rows.Err(); err != nil {
		r.logger.Printf("Rows error: %v", err)
		return nil, fmt.Errorf("rows error: %w", err)
	}

	r.logger.Printf("Successfully fetched %d purchases for user ID: %d", len(purchases), userID)
	return purchases, nil
}
