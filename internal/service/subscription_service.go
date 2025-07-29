package service

import (
	"database/sql"
	"errors"
	"go-subscriptions-service/internal/model"
	"go-subscriptions-service/internal/repo"
	"go-subscriptions-service/pgk/validator"
	"log"
	"time"

	"github.com/google/uuid"
)

type SubscriptionService interface {
	Create(subscription *model.Subscription) error
	GetByID(id uuid.UUID) (*model.Subscription, error)
	GetAll() ([]model.Subscription, error)
	Update(subscription *model.Subscription) error
	Delete(id uuid.UUID) error
	GetTotalAmount(userID uuid.UUID, serviceName *string, from, to time.Time) (int, error)
}

type subscriptionService struct {
	repo repo.SubscriptionRepository
}

func NewSubscriptionService(r repo.SubscriptionRepository) SubscriptionService {
	return &subscriptionService{repo: r}
}

func (s *subscriptionService) Create(subscription *model.Subscription) error {
	log.Printf("Create (service) called: service_name=%v, price=%v, user_id=%v, start_date=%v, end_date=%v", subscription.ServiceName, subscription.Price, subscription.UserID, subscription.StartDate, subscription.EndDate)
	validator.ValidateSubcription(subscription)

	return s.repo.Create(subscription)
}

func (s *subscriptionService) GetByID(id uuid.UUID) (*model.Subscription, error) {
	log.Printf("GetByID (service) called: id=%v", id)
	sub, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Println("GetByID (service) error: subscription not found ", err)
			return nil, sql.ErrNoRows
		}
		log.Println("GetByID (service) error: failed to get subscription by id ", err)
		return nil, err
	}

	log.Println("GetByID (service) success: subscription found ", sub)
	return sub, nil
}

func (s *subscriptionService) GetAll() ([]model.Subscription, error) {
	log.Println("GetAll (service) called")
	subs, err := s.repo.GetAll()
	if err != nil {
		log.Println("GetAll(service) error: failed to get all subscriptions ", err)
		return nil, err
	}
	log.Printf("GetAll (service) success: found %d subscriptions ", len(subs))
	return subs, nil
}

func (s *subscriptionService) Update(subscription *model.Subscription) error {
	log.Printf("Update (service) called: id=%v", subscription.ID)
	validator.ValidateSubcription(subscription)

	err := s.repo.Update(subscription)
	if err != nil {
		log.Println("Update (service) error: failed to update subscription ", err)
		return err
	}

	log.Println("Update (service) success: subscription updated")
	return nil
}

func (s *subscriptionService) Delete(id uuid.UUID) error {
	log.Printf("Delete (service) called: id=%v", id)
	err := s.repo.Delete(id)
	if err != nil {
		log.Println("Delete (service) error: failed to delete subscription ", err)
		return err
	}

	log.Println("Delete (service) success: subscription deleted")
	return nil
}

func (s *subscriptionService) GetTotalAmount(userID uuid.UUID, serviceName *string, from, to time.Time) (int, error) {
	log.Printf("GetTotalAmount (service) called: user_id=%v, service_name=%v, from=%v, to=%v", userID, serviceName, from, to)
	if from.IsZero() || to.IsZero() {
		log.Println("GetTotalAmount (service) error: from or to is zero")
		return 0, errors.New("date range is required")
	}

	if from.After(to) {
		log.Println("GetTotalAmount (service) error: from is after to")
		return 0, errors.New("invalid date range: 'from' is after 'to'")
	}

	total, err := s.repo.GetTotalAmount(userID, serviceName, from, to)
	if err != nil {
		log.Println("GetTotalAmount (service) error: failed to get total amount ", err)
		return 0, err
	}

	log.Printf("GetTotalAmount (service) success: total = %d", total)
	return total, nil
}
