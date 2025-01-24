package models

import "time"

type Purchase struct {
    ID         int64     `db:"id" json:"id"`
    UserID     int64     `db:"user_id" json:"user_id"`
    ProductID  int64     `db:"product_id" json:"product_id"`
    PurchasedAt time.Time `db:"purchased_at" json:"purchased_at"`
}
