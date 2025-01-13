package server

import (
	"JustDone/internal/models"
	"JustDone/internal/service"
	"encoding/json"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Server struct {
	service *service.OrderService
}

func NewServer(orderService *service.OrderService) (*Server, error) {
	return &Server{
		service: orderService,
	}, nil
}

func (s *Server) Run(addr string) error {
	r := mux.NewRouter()

	r.HandleFunc("/webhooks/payments/orders", s.HandleWebhook).Methods("POST")
	r.HandleFunc("/orders", s.GetOrders).Methods("GET")
	r.HandleFunc("/orders/{order_id}", s.GetOrderDetails).Methods("GET")
	r.HandleFunc("/orders/{order_id}/events", s.GetOrderEvents).Methods("GET")

	return http.ListenAndServe(addr, r)
}

func (s *Server) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	payload := models.WebhookPayload{}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		zap.L().Error("Invalid JSON payload", zap.Error(err))
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if !isValidStatus(payload.Status) {
		http.Error(w, "Invalid status value", http.StatusBadRequest)
		return
	}

	if err := s.service.CreateOrUpdateOrder(r.Context(), payload); err != nil {
		if strings.Contains(err.Error(), "event already processed") {
			http.Error(w, "Event already processed", http.StatusConflict)
		} else if strings.Contains(err.Error(), "order is in final state") {
			http.Error(w, "order is in final state", http.StatusUnprocessableEntity)
		} else {
			zap.L().Error("Failed to create or update order", zap.Error(err))
			http.Error(w, "Failed to process request", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) GetOrders(w http.ResponseWriter, r *http.Request) {
	filters := ParseOrderFilter(r.URL.Query())

	orders, err := s.service.GetOrders(r.Context(), filters)
	if err != nil {
		zap.L().Error("Failed to fetch orders", zap.Error(err))
		http.Error(w, "Failed to fetch orders", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if len(orders) == 0 {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("[]"))
		return
	}

	json.NewEncoder(w).Encode(orders)
}

func (s *Server) GetOrderDetails(w http.ResponseWriter, r *http.Request) {
	orderID := strings.TrimPrefix(r.URL.Path, "/orders/")
	if orderID == "" {
		http.Error(w, "Missing order_id", http.StatusBadRequest)
		return
	}

	order, err := s.service.GetOrderDetails(r.Context(), orderID)
	if err != nil {
		if err.Error() == "order not found" {
			http.Error(w, "Order not found", http.StatusNotFound)
		} else {
			zap.L().Error("Failed to fetch order details", zap.Error(err))
			http.Error(w, "Failed to fetch order details", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func (s *Server) GetOrderEvents(w http.ResponseWriter, r *http.Request) {
	orderID := strings.TrimPrefix(strings.TrimSuffix(r.URL.Path, "/events"), "/orders/")
	if orderID == "" {
		http.Error(w, "Missing order_id", http.StatusBadRequest)
		return
	}

	events, err := s.service.GetOrderEvents(r.Context(), orderID)
	if err != nil {
		zap.L().Error("Failed to fetch order events", zap.Error(err))
		http.Error(w, "Failed to fetch order events", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if len(events) == 0 {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("[]"))
		return
	}

	json.NewEncoder(w).Encode(events)
}

func ParseOrderFilter(queryParams url.Values) models.OrderFilter {
	filter := models.OrderFilter{
		Limit:     10,
		Offset:    0,
		SortBy:    "created_at",
		SortOrder: "DESC",
	}

	if statuses, ok := queryParams["status"]; ok {
		filter.Status = strings.Split(statuses[0], ",")
	}

	if userID, ok := queryParams["user_id"]; ok {
		filter.UserID = userID[0]
	}

	if limit, ok := queryParams["limit"]; ok {
		if parsedLimit, err := strconv.Atoi(limit[0]); err == nil {
			filter.Limit = parsedLimit
		}
	}

	if offset, ok := queryParams["offset"]; ok {
		if parsedOffset, err := strconv.Atoi(offset[0]); err == nil {
			filter.Offset = parsedOffset
		}
	}

	if sortBy, ok := queryParams["sort_by"]; ok {
		if sortBy[0] == "created_at" || sortBy[0] == "updated_at" {
			filter.SortBy = sortBy[0]
		}
	}

	if sortOrder, ok := queryParams["sort_order"]; ok {
		if sortOrder[0] == "asc" || sortOrder[0] == "desc" {
			filter.SortOrder = strings.ToUpper(sortOrder[0])
		}
	}

	return filter
}

func isValidStatus(status string) bool {
	validStatuses := map[string]bool{
		"created": true,
		"updated": true,
		"failed":  true,
		"success": true,
	}
	return validStatuses[status]
}
