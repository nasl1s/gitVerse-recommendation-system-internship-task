package models

import "time"

type Recommendation struct {
    ID         int64     `db:"id" json:"id"`
    UserID     int64     `db:"user_id" json:"user_id"`
    ProductIDs []int     `db:"product_ids" json:"product_ids"`
    CreatedAt  time.Time `db:"created_at" json:"created_at"`
}