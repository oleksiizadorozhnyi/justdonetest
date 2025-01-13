package repository

import (
	"JustDone/internal/config"
	"JustDone/internal/models"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"strconv"
	"strings"
)

type OrderRepo struct {
	db *sql.DB
}

func NewOrderRepo(cfg config.Postgres) (*OrderRepo, error) {
	conStr := fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=%v", cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.SSLMode)
	db, err := sql.Open("pgx", conStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to apply migrations to database: %v", err)
	}

	return &OrderRepo{db: db}, nil
}

func runMigrations(db *sql.DB) error {
	if err := goose.Up(db, "./migrations"); err != nil {
		return fmt.Errorf("goose migration failed: %w", err)
	}
	return nil
}

func (r *OrderRepo) IsEventProcessed(ctx context.Context, eventID string) bool {
	var count int
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM order_events WHERE event_id = $1", eventID).Scan(&count)
	return count > 0
}

func (r *OrderRepo) IsOrderFinalized(ctx context.Context, orderID string) bool {
	var status string
	r.db.QueryRowContext(ctx, "SELECT status FROM orders WHERE order_id = $1", orderID).Scan(&status)
	return status == "failed" || status == "success"
}

func (r *OrderRepo) SaveOrderAndEvent(ctx context.Context, event models.WebhookPayload) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	metaJSON, err := json.Marshal(event.Meta)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.ExecContext(ctx, `
        INSERT INTO orders (order_id, user_id, status, created_at, updated_at, meta)
        VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT (order_id) DO UPDATE SET
            status = EXCLUDED.status,
            updated_at = EXCLUDED.updated_at,
            meta = EXCLUDED.meta
    `, event.OrderID, event.UserID, event.Status, event.CreatedAt, event.UpdatedAt, metaJSON)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.ExecContext(ctx, `
        INSERT INTO order_events (event_id, order_id, user_id, status, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `, event.EventID, event.OrderID, event.UserID, event.Status, event.CreatedAt, event.UpdatedAt)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *OrderRepo) FetchOrders(ctx context.Context, filter models.OrderFilter) ([]models.OrderResponse, error) {
	query := `
        SELECT order_id, user_id, status, created_at, updated_at
        FROM orders
    `
	args := []interface{}{}
	whereClauses := []string{}

	if len(filter.Status) > 0 {
		whereClauses = append(whereClauses, "status = ANY($"+strconv.Itoa(len(args)+1)+")")
		args = append(args, pq.Array(filter.Status))
	}

	if filter.UserID != "" {
		whereClauses = append(whereClauses, "user_id = $"+strconv.Itoa(len(args)+1))
		args = append(args, filter.UserID)
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	query += " ORDER BY " + filter.SortBy + " " + filter.SortOrder

	query += " LIMIT $" + strconv.Itoa(len(args)+1) + " OFFSET $" + strconv.Itoa(len(args)+2)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.OrderResponse
	for rows.Next() {
		var order models.OrderResponse

		err := rows.Scan(
			&order.OrderID,
			&order.UserID,
			&order.Status,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		orders = append(orders, order)
	}

	return orders, nil
}

func (r *OrderRepo) GetOrderDetails(ctx context.Context, orderID string) (*models.OrderResponse, error) {
	var order models.OrderResponse

	err := r.db.QueryRowContext(ctx, `
        SELECT order_id, user_id, status, created_at, updated_at
        FROM orders
        WHERE order_id = $1
    `, orderID).Scan(
		&order.OrderID,
		&order.UserID,
		&order.Status,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (r *OrderRepo) GetOrderEvents(ctx context.Context, orderID string) ([]models.OrderEvent, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT event_id, order_id, user_id, status, created_at, updated_at
        FROM order_events
        WHERE order_id = $1
    `, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.OrderEvent
	for rows.Next() {
		var event models.OrderEvent
		err := rows.Scan(
			&event.EventID,
			&event.OrderID,
			&event.UserID,
			&event.Status,
			&event.CreatedAt,
			&event.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}
