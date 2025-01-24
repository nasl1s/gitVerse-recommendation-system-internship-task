package models

import "time"

type Like struct {
    ID          int64     `db:"id" json:"id"`
    UserID      int64     `db:"user_id" json:"user_id"`
    ProductID   int64     `db:"product_id" json:"product_id"`
    LikedAt     time.Time `db:"liked_at" json:"liked_at"`
}
