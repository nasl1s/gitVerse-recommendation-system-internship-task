package models

import "time"

type Dislike struct {
    ID            int64     `db:"id" json:"id"`
    UserID        int64     `db:"user_id" json:"user_id"`
    ProductID     int64     `db:"product_id" json:"product_id"`
    DislikedAt    time.Time `db:"disliked_at" json:"disliked_at"`
}
