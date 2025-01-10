package service

import (
	repository "JustDone/internal/repository/postgres"
	"database/sql"
	"errors"
)

type OrderService struct {
	repo *repository.OrderRepo
}

func NewOrderService(repo *repository.OrderRepo) (*OrderService, error) {
	return &OrderService{repo: repo}, nil
}

func (s *OrderService) CreateOrUpdateOrder(webhookData map[string]interface{}) error {
	// Validate and process webhook data
	orderID := webhookData["order_id"].(string)
	eventID := webhookData["event_id"].(string)
	_ = webhookData["status"].(string)

	if s.repo.IsEventProcessed(eventID) {
		return errors.New("event already processed")
	}

	if s.repo.IsOrderFinalized(orderID) {
		return errors.New("order is in final state")
	}

	return s.repo.SaveOrderAndEvent(webhookData)
}

func (s *OrderService) GetOrders(filters map[string][]string) ([]map[string]interface{}, error) {
	return s.repo.FetchOrders(filters)
}

func (s *OrderService) GetOrderDetails(orderID string) (map[string]interface{}, error) {
	// Fetch order details from the repository
	order, err := s.repo.GetOrderDetails(orderID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("order not found")
		}
		return nil, err
	}

	return order, nil
}

func (s *OrderService) GetOrderEvents(orderID string) ([]map[string]interface{}, error) {
	// Fetch events from the repository
	events, err := s.repo.GetOrderEvents(orderID)
	if err != nil {
		return nil, err
	}

	return events, nil
}
