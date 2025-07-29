package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"go-subscriptions-service/internal/dto"
	"go-subscriptions-service/internal/model"
	"go-subscriptions-service/internal/service"
	"go-subscriptions-service/pgk/utils"
	"go-subscriptions-service/pgk/validator"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type SubscriptionHandler struct {
	service service.SubscriptionService
}

func NewSubscriptionHandler(s service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{service: s}
}

func (h *SubscriptionHandler) RegisterRouters(r *mux.Router) {
	r.HandleFunc("/subscription/total_amount", h.GetTotalAmount).Methods("GET")
	r.HandleFunc("/subscription", h.CreateSubscription).Methods("POST")
	r.HandleFunc("/subscription/{id}", h.GetSubscriptionsByID).Methods("GET")
	r.HandleFunc("/subscription", h.GetAllSubscriptions).Methods("GET")
	r.HandleFunc("/subscription/{id}", h.UpdateSubscription).Methods("PATCH")
	r.HandleFunc("/subscription/{id}", h.DeleteSubscription).Methods("DELETE")
}

// GetTotalAmount godoc
// @Summary Получить сумму подписок за период
// @Description Считает сумму подписок за период пользователя в заданном диапазоне
// @Tags subscription
// @Accept json
// @Produce json
// @Param user_id query string true "ID пользователя"
// @Param from query string true "Дата начала периода (yyyy-mm-dd)"
// @Param to query string true "Дата окончания периода (yyyy-mm-dd)"
// @Param service_name query string false "Название сервиса (опционально)"
// @Success 200 {object} map[string]int
// @Failure 400 {string} string "Неверные параметры запроса"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /subscription/total_amount [get]
func (h *SubscriptionHandler) GetTotalAmount(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		log.Println("GetTotalAmount (handler) error: user_id is required")
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	userIDUUID, err := uuid.Parse(userID)
	if err != nil {
		log.Println("GetTotalAmount (handler) error: uuid.Parse failed: ", err)
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	if from == "" || to == "" {
		log.Println("GetTotalAmount (handler) error: from and to are required")
		http.Error(w, "from and to are required", http.StatusBadRequest)
		return
	}

	fromDate, err := time.Parse("2006-01-02", from)
	if err != nil {
		log.Println("GetTotalAmount (handler) error: from time.Parse failed: ", err)
		http.Error(w, "invalid from date", http.StatusBadRequest)
		return
	}

	toDate, err := time.Parse("2006-01-02", to)
	if err != nil {
		log.Println("GetTotalAmount (handler) error: to time.Parse failed: ", err)
		http.Error(w, "invalid to date", http.StatusBadRequest)
		return
	}

	var servName *string

	serviceName := r.URL.Query().Get("service_name")
	if serviceName != "" {
		servName = &serviceName
	}

	res, err := h.service.GetTotalAmount(userIDUUID, servName, fromDate, toDate)
	if err != nil {
		log.Println("GetTotalAmount (handler) error: failed to get total amount: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type Res struct {
		TotalAmount int `json:"total_amount"`
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Res{TotalAmount: res})
	log.Printf("GetTotalAmount (handler) success: user_id=%v, total=%d", userIDUUID, res)
}

// CreateSubscription godoc
// @Summary Создать подписку
// @Description Создать новую подписку
// @Tags subscription
// @Accept json
// @Produce json
// @Param request body dto.SubscriptionRequest true "Данные для создания подписки"
// @Success 201 {object} model.Subscription
// @Failure 400 {string} string "Неверные данные"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /subscription [post]
func (h *SubscriptionHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req dto.SubscriptionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("CreateSubscription (handler) error: json.NewDecoder failed: ", err)
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if err := validator.ValidateCreateSubscriptionRequest(&req); err != nil {
		log.Println("CreateSubscription (handler) error: validateCreateSubscriptionRequest failed: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID, _ := uuid.Parse(req.UserID)
	startDate, _ := utils.ParseMonthYear(req.StartDate)
	endDate, _ := utils.ParseMonthYear(req.EndDate)

	sub := model.Subscription{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      userID,
		StartDate:   startDate,
		EndDate:     &endDate,
	}

	if err := h.service.Create(&sub); err != nil {
		log.Println("CreateSubscription (handler) error: failed to create subscription: ", err)
		http.Error(w, "failed to create subscription", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(sub)
	log.Println("CreateSubscription (handler) success: subscription created")
}

// GetSubscriptionsByID godoc
// @Summary Получить подписку по ID
// @Description Возвращает подписку по ID
// @Tags subscription
// @Accept json
// @Produce json
// @Param id path string true "ID подписки"
// @Success 200 {object} model.Subscription
// @Failure 400 {string} string "Неверный ID"
// @Failure 404 {string} string "Подписка не найдена"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /subscription/{id} [get]
func (h *SubscriptionHandler) GetSubscriptionsByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Println("GetSubscriptionsByID (handler) error: uuid.Parse failed: ", err)
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	res, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Println("GetSubscriptionsByID (handler) error: subscription not found: ", err)
			http.Error(w, sql.ErrNoRows.Error(), http.StatusNotFound)
			return
		}
		log.Println("GetSubscriptionsByID (handler) error: failed to get subscription by id: ", err)
		http.Error(w, "failed to get subscription by id", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
	log.Println("GetSubscriptionsByID (handler) success: subscription found")
}

// GetAllSubscriptions godoc
// @Summary Получить все подписки
// @Description Возвращает список всех подписок
// @Tags subscription
// @Accept json
// @Produce json
// @Success 200 {array} model.Subscription
// @Failure 500 {string} string "Ошибка сервера"
// @Router /subscription [get]
func (h *SubscriptionHandler) GetAllSubscriptions(w http.ResponseWriter, r *http.Request) {
	subscriptions, err := h.service.GetAll()
	if err != nil {
		log.Println("GetAllSubscriptions (handler) error: failed to get all subscriptions: ", err)
		http.Error(w, "failed to get all subscriptions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(subscriptions)
	log.Println("GetAllSubscriptions (handler) success: all subscriptions found")
}

// UpdateSubscription godoc
// @Summary Обновить подписку
// @Description Обновляет существующую подписку
// @Tags subscription
// @Accept json
// @Produce json
// @Param id path string true "ID подписки"
// @Param request body dto.SubscriptionRequest true "Оновленные данные подписки"
// @Success 200 {object} model.Subscription
// @Failure 400 {string} string "Неверный ID или тело запроса"
// @Failure 404 {string} string "Подписка не найдена"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /subscription/{id} [patch]
func (h *SubscriptionHandler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Println("UpdateSubscription (handler) error: uuid.Parse failed: ", err)
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req dto.SubscriptionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("UpdateSubscription (handler) error: json.NewDecoder failed: ", err)
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if err := validator.ValidateCreateSubscriptionRequest(&req); err != nil {
		log.Println("UpdateSubscription (handler) error: validateCreateSubscriptionRequest failed: ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID, _ := uuid.Parse(req.UserID)
	startDate, _ := utils.ParseMonthYear(req.StartDate)
	endDate, _ := utils.ParseMonthYear(req.EndDate)

	sub := model.Subscription{
		ID:          id,
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      userID,
		StartDate:   startDate,
		EndDate:     &endDate,
	}

	if err := h.service.Update(&sub); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Println("UpdateSubscription (handler) error: subscription not found: ", err)
			http.Error(w, sql.ErrNoRows.Error(), http.StatusNotFound)
			return
		}
		log.Println("UpdateSubscription (handler) error: failed to update subscription: ", err)
		http.Error(w, "failed to update subscription", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(sub)
	log.Println("UpdateSubscription (handler) success: subscription updated")
}

// DeleteSubscription godoc
// @Summary Удалить подписку
// @Description Удаляет подписку по ID
// @Tags subscription
// @Accept json
// @Produce json
// @Param id path string true "ID подписки"
// @Success 204 {string} string ""
// @Failure 400 {string} string "Неверный ID"
// @Failure 404 {string} string "Подписка не найдена"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /subscription/{id} [delete]
func (h *SubscriptionHandler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Println("DeleteSubscription (handler) error: uuid.Parse failed: ", err)
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Println("DeleteSubscription (handler) error: subscription not found: ", err)
			http.Error(w, sql.ErrNoRows.Error(), http.StatusNotFound)
			return
		}
		log.Println("DeleteSubscription (handler) error: failed to delete subscription: ", err)
		http.Error(w, "failed to delete subscription", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	log.Println("DeleteSubscription (handler) success: subscription deleted")
}
