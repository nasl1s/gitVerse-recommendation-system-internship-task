package models

import "time"

type ProductAnalytics struct {
    ID        int64     `db:"id" json:"-"`
    ProductID int64     `db:"product_id" json:"product_id"`
    Likes     int       `db:"likes" json:"likes"`
    Dislikes  int       `db:"dislikes" json:"dislikes"`
    Purchases int       `db:"purchases" json:"purchases"`
    UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type UserAnalytics struct {
    ID             int64     `db:"id" json:"-"`
    UserID         int64     `db:"user_id" json:"user_id"`
    TotalLikes     int       `db:"total_likes" json:"total_likes"`
    TotalDislikes  int       `db:"total_dislikes" json:"total_dislikes"`
    TotalPurchases int       `db:"total_purchases" json:"total_purchases"`
    UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}
