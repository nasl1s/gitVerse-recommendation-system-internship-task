package models

import "time"

type Product struct {
    ID            int64     `db:"id" json:"id"`
    Name          string    `db:"name" json:"name"`
    Description   string    `db:"description" json:"description"`
    Price         float64   `db:"price" json:"price"`
    Category      string    `db:"category" json:"category"`
    CreatedAt     time.Time `db:"created_at" json:"created_at"`
    UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
    Likes         []*Like   `json:"likes,omitempty"`
    Dislikes      []*Dislike `json:"dislikes,omitempty"`
    PurchaseCount int       `json:"purchase_count,omitempty"`
}
