package service

import (
	"JustDone/internal/models"
	repository "JustDone/internal/repository/postgres"
	"errors"
)

type OrderService struct {
	repo *repository.OrderRepo
}

func NewOrderService(repo *repository.OrderRepo) (*OrderService, error) {
	return &OrderService{repo: repo}, nil
}

func (s *OrderService) CreateOrUpdateOrder(event models.WebhookPayload) error {
	if s.repo.IsEventProcessed(event.EventID) {
		return errors.New("event already processed")
	}

	if s.repo.IsOrderFinalized(event.OrderID) {
		return errors.New("order is in final state")
	}

	return s.repo.SaveOrderAndEvent(event)
}

func (s *OrderService) GetOrders(filters models.OrderFilter) ([]models.OrderResponse, error) {
	return s.repo.FetchOrders(filters)
}

func (s *OrderService) GetOrderDetails(orderID string) (*models.OrderResponse, error) {
	return s.repo.GetOrderDetails(orderID)
}

func (s *OrderService) GetOrderEvents(orderID string) ([]models.OrderEvent, error) {
	return s.repo.GetOrderEvents(orderID)
}
