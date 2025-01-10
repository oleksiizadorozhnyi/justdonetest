package server

import (
	"JustDone/internal/service"
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"strings"
)

type Server struct {
	service *service.OrderService
}

func NewServer(orderService *service.OrderService) (*Server, error) {

	return &Server{service: orderService}, nil
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

	var webhookData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&webhookData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := s.service.CreateOrUpdateOrder(webhookData); err != nil {
		if strings.Contains(err.Error(), "already processed") {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) GetOrders(w http.ResponseWriter, r *http.Request) {
	// Отримуємо параметри запиту
	queryParams := r.URL.Query()

	// Викликаємо метод сервісу для отримання ордерів
	orders, err := s.service.GetOrders(queryParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Відповідаємо списком ордерів
	w.Header().Set("Content-Type", "application/json")
	if len(orders) == 0 {
		w.Write([]byte("[]"))
		return
	}

	json.NewEncoder(w).Encode(orders)
}

func (s *Server) GetOrderDetails(w http.ResponseWriter, r *http.Request) {
	// Виділяємо `order_id` з URL
	orderID := strings.TrimPrefix(r.URL.Path, "/orders/")
	if orderID == "" {
		http.Error(w, "Missing order_id", http.StatusBadRequest)
		return
	}

	// Викликаємо метод сервісу для отримання деталей ордеру
	order, err := s.service.GetOrderDetails(orderID)
	if err != nil {
		if err.Error() == "order not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Відповідаємо JSON з деталями ордеру
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func (s *Server) GetOrderEvents(w http.ResponseWriter, r *http.Request) {
	// Виділяємо `order_id` з URL
	orderID := strings.TrimPrefix(strings.TrimSuffix(r.URL.Path, "/events"), "/orders/")
	if orderID == "" {
		http.Error(w, "Missing order_id", http.StatusBadRequest)
		return
	}

	// Викликаємо метод сервісу для отримання івентів ордеру
	events, err := s.service.GetOrderEvents(orderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Відповідаємо JSON з івентами
	w.Header().Set("Content-Type", "application/json")
	if len(events) == 0 {
		w.Write([]byte("[]"))
		return
	}

	json.NewEncoder(w).Encode(events)
}
