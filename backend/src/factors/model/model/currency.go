package model

import(
	"time"
)

type Currency struct {
	Currency   string          `json:"currency"`
	InrValue   float64         `json:"int_value"`
	Date       int64      	   `json:"date"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}