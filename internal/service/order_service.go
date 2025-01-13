package service

import (
	"JustDone/internal/models"
	repository "JustDone/internal/repository/postgres"
	"context"
	"errors"
)

type OrderService struct {
	repo *repository.OrderRepo
}

func NewOrderService(repo *repository.OrderRepo) (*OrderService, error) {
	return &OrderService{repo: repo}, nil
}

func (s *OrderService) CreateOrUpdateOrder(ctx context.Context, event models.WebhookPayload) error {
	if s.repo.IsEventProcessed(ctx, event.EventID) {
		return errors.New("event already processed")
	}

	if s.repo.IsOrderFinalized(ctx, event.OrderID) {
		return errors.New("order is in final state")
	}

	return s.repo.SaveOrderAndEvent(ctx, event)
}

func (s *OrderService) GetOrders(ctx context.Context, filters models.OrderFilter) ([]models.OrderResponse, error) {
	return s.repo.FetchOrders(ctx, filters)
}

func (s *OrderService) GetOrderDetails(ctx context.Context, orderID string) (*models.OrderResponse, error) {
	return s.repo.GetOrderDetails(ctx, orderID)
}

func (s *OrderService) GetOrderEvents(ctx context.Context, orderID string) ([]models.OrderEvent, error) {
	return s.repo.GetOrderEvents(ctx, orderID)
}
