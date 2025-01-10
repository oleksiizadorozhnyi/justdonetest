package repository

import (
	"database/sql"
	"fmt"
	"github.com/jackc/pgtype"
	_ "github.com/jackc/pgx/v4/stdlib"
	"strconv"
	"strings"
)

type OrderRepo struct {
	db *sql.DB
}

func NewOrderRepo(conStr string) (*OrderRepo, error) {
	db, err := sql.Open("pgx", conStr)
	if err != nil {
		return nil, err
	}

	// Test the database connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	return &OrderRepo{db: db}, nil
}

func (r *OrderRepo) IsEventProcessed(eventID string) bool {
	var count int
	r.db.QueryRow("SELECT COUNT(*) FROM order_events WHERE event_id = $1", eventID).Scan(&count)
	return count > 0
}

func (r *OrderRepo) IsOrderFinalized(orderID string) bool {
	var status string
	r.db.QueryRow("SELECT status FROM orders WHERE order_id = $1", orderID).Scan(&status)
	return status == "failed" || status == "success"
}

func (r *OrderRepo) SaveOrderAndEvent(data map[string]interface{}) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	// Зберігаємо ордер
	_, err = tx.Exec(`
		INSERT INTO orders (order_id, user_id, status, created_at, updated_at, meta)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (order_id)
		DO UPDATE SET 
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at,
			meta = EXCLUDED.meta
	`, data["order_id"], data["user_id"], data["status"], data["created_at"], data["updated_at"], data["meta"])
	if err != nil {
		tx.Rollback()
		return err
	}

	// Зберігаємо івент
	_, err = tx.Exec(`
		INSERT INTO order_events (event_id, order_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`, data["event_id"], data["order_id"], data["status"], data["created_at"], data["updated_at"])
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *OrderRepo) FetchOrders(filters map[string][]string) ([]map[string]interface{}, error) {
	query := `
		SELECT order_id, user_id, status, created_at, updated_at
		FROM orders
	`
	args := []interface{}{}
	whereClauses := []string{}

	// Filter by statuses (optional)
	if statuses, ok := filters["status"]; ok && len(statuses) > 0 {
		whereClauses = append(whereClauses, "status = ANY($"+strconv.Itoa(len(args)+1)+")")

		// Use pgtype.TextArray for PostgreSQL arrays
		textArray := &pgtype.TextArray{}
		if err := textArray.Set(strings.Split(statuses[0], ",")); err != nil {
			return nil, err
		}
		args = append(args, textArray)
	}

	// Filter by user_id (optional)
	if userID, ok := filters["user_id"]; ok && len(userID) > 0 {
		whereClauses = append(whereClauses, "user_id = $"+strconv.Itoa(len(args)+1))

		// Use pgtype.UUID for PostgreSQL UUID
		uuid := &pgtype.UUID{}
		if err := uuid.Set(userID[0]); err != nil {
			return nil, err
		}
		args = append(args, uuid)
	}

	// Add WHERE clause if there are filters
	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Sorting
	sortBy := "created_at"
	if val, ok := filters["sort_by"]; ok && len(val) > 0 {
		sortBy = val[0]
	}
	sortOrder := "DESC"
	if val, ok := filters["sort_order"]; ok && len(val) > 0 && (val[0] == "asc" || val[0] == "desc") {
		sortOrder = strings.ToUpper(val[0])
	}
	query += " ORDER BY " + sortBy + " " + sortOrder

	// Pagination
	limit := 10
	if val, ok := filters["limit"]; ok && len(val) > 0 {
		if parsed, err := strconv.Atoi(val[0]); err == nil {
			limit = parsed
		}
	}
	offset := 0
	if val, ok := filters["offset"]; ok && len(val) > 0 {
		if parsed, err := strconv.Atoi(val[0]); err == nil {
			offset = parsed
		}
	}
	query += " LIMIT $" + strconv.Itoa(len(args)+1) + " OFFSET $" + strconv.Itoa(len(args)+2)
	args = append(args, limit, offset)

	// Execute the query
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Process results
	var orders []map[string]interface{}
	for rows.Next() {
		var orderID, userID, status string
		var createdAt, updatedAt string

		if err := rows.Scan(&orderID, &userID, &status, &createdAt, &updatedAt); err != nil {
			return nil, err
		}

		order := map[string]interface{}{
			"order_id":   orderID,
			"user_id":    userID,
			"status":     status,
			"created_at": createdAt,
			"updated_at": updatedAt,
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func (r *OrderRepo) GetOrderDetails(orderID string) (map[string]interface{}, error) {
	query := `
        SELECT order_id, user_id, status, created_at, updated_at
        FROM orders
        WHERE order_id = $1
    `

	row := r.db.QueryRow(query, orderID)

	var orderIDResult, userID, status string
	var createdAt, updatedAt string

	err := row.Scan(&orderIDResult, &userID, &status, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	order := map[string]interface{}{
		"order_id":   orderIDResult,
		"user_id":    userID,
		"status":     status,
		"created_at": createdAt,
		"updated_at": updatedAt,
	}

	return order, nil
}

func (r *OrderRepo) GetOrderEvents(orderID string) ([]map[string]interface{}, error) {
	query := `
        SELECT event_id, order_id, status, created_at, updated_at
        FROM order_events
        WHERE order_id = $1
        ORDER BY created_at ASC
    `

	rows, err := r.db.Query(query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []map[string]interface{}

	for rows.Next() {
		var eventID, orderIDResult, status string
		var createdAt, updatedAt string

		if err := rows.Scan(&eventID, &orderIDResult, &status, &createdAt, &updatedAt); err != nil {
			return nil, err
		}

		event := map[string]interface{}{
			"event_id":   eventID,
			"order_id":   orderIDResult,
			"status":     status,
			"created_at": createdAt,
			"updated_at": updatedAt,
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}
