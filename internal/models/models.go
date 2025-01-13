package models

import (
	"time"
)

type WebhookPayload struct {
	EventID   string                 `json:"event_id"`
	OrderID   string                 `json:"order_id"`
	UserID    string                 `json:"user_id"`
	Status    string                 `json:"status"`
	UpdatedAt time.Time              `json:"updated_at"`
	CreatedAt time.Time              `json:"created_at"`
	Meta      map[string]interface{} `json:"meta"`
}

type Order struct {
	OrderID   string                 `json:"order_id"`
	UserID    string                 `json:"user_id"`
	Status    string                 `json:"status"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	Meta      map[string]interface{} `json:"meta"`
}

type OrderResponse struct {
	OrderID   string    `json:"order_id"`
	UserID    string    `json:"user_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type OrderEvent struct {
	EventID   string    `json:"event_id"`
	OrderID   string    `json:"order_id"`
	UserID    string    `json:"user_id"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

type OrderFilter struct {
	Status    []string `json:"status"`
	UserID    string   `json:"user_id"`
	Limit     int      `json:"limit"`
	Offset    int      `json:"offset"`
	SortBy    string   `json:"sort_by"`
	SortOrder string   `json:"sort_order"`
}
