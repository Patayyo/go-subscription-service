package validator

import (
	"errors"
	"go-subscriptions-service/internal/dto"
	"go-subscriptions-service/internal/model"
	"go-subscriptions-service/pgk/utils"
	"log"

	"github.com/google/uuid"
)

func ValidateCreateSubscriptionRequest(req *dto.SubscriptionRequest) error {
	log.Println("validateCreateSubscriptionRequest (handler): called with req=", req)
	if req.ServiceName == "" {
		log.Println("validateCreateSubscriptionRequest (handler) error: service_name is required")
		return errors.New("service_name is required")
	}

	if req.Price <= 0 {
		log.Println("validateCreateSubscriptionRequest (handler) error: price must be greater than 0")
		return errors.New("price must be greater than 0")
	}

	if req.UserID == "" {
		log.Println("validateCreateSubscriptionRequest (handler) error: user_id is required")
		return errors.New("user_id is required")
	}

	if _, err := uuid.Parse(req.UserID); err != nil {
		log.Println("validateCreateSubscriptionRequest (handler) error: invalid user_id format")
		return errors.New("invalid user_id format")
	}

	if _, err := utils.ParseMonthYear(req.StartDate); err != nil {
		log.Println("validateCreateSubscriptionRequest (handler) error: invalid start_date format (expected MM-YYYY)")
		return errors.New("invalid start_date format (expected MM-YYYY)")
	}

	if _, err := utils.ParseMonthYear(req.EndDate); err != nil {
		log.Println("validateCreateSubscriptionRequest (handler) error: invalid end_date format (expected MM-YYYY)")
		return errors.New("invalid end_date format (expected MM-YYYY)")
	}

	log.Println("validateCreateSubscriptionRequest (handler) success: request is valid")
	return nil
}

func ValidateSubcription(s *model.Subscription) error {
	log.Printf("validating (service) called: service_name=%v, price=%v, user_id=%v, start_date=%v, end_date=%v", s.ServiceName, s.Price, s.UserID, s.StartDate, s.EndDate)
	if s.ServiceName == "" {
		log.Println("validateSubcription (service) error: service name must not be empty")
		return errors.New("service name must not be empty")
	}
	if s.Price <= 0 {
		log.Println("validateSubcription (service) error: price must be greater than 0")
		return errors.New("price must be greater than 0")
	}

	if s.UserID == uuid.Nil {
		log.Println("validateSubcription (service) error: user ID must not be empty")
		return errors.New("user ID must not be empty")
	}

	if s.StartDate.IsZero() {
		log.Println("validateSubcription (service) error: start date must not be empty")
		return errors.New("start date must not be empty")
	}

	if s.EndDate != nil && s.EndDate.Before(s.StartDate) {
		log.Println("validateSubcription (service) error: end date must be after start date")
		return errors.New("end date must be after start date")
	}

	log.Println("validateSubcription (service) success: subscription is valid")
	return nil
}
