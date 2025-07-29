package repo

import (
	"database/sql"
	"errors"
	"fmt"
	"go-subscriptions-service/internal/model"
	"log"
	"time"

	"github.com/google/uuid"
)

type SubscriptionRepository interface {
	Create(subscription *model.Subscription) error
	GetByID(id uuid.UUID) (*model.Subscription, error)
	GetAll() ([]model.Subscription, error)
	Update(subscription *model.Subscription) error
	Delete(id uuid.UUID) error
	GetTotalAmount(userID uuid.UUID, serviceName *string, from, to time.Time) (int, error)
}

type subscriptionRepo struct {
	db *sql.DB
}

func NewSubscriptionRepo(db *sql.DB) SubscriptionRepository {
	return &subscriptionRepo{db: db}
}

func (r *subscriptionRepo) Create(subscription *model.Subscription) error {
	log.Printf("Create  (repo): inserting subscription for user_id=%v, service_name=%v", subscription.UserID, subscription.ServiceName)

	err := r.db.QueryRow(
		`
		INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
		`, subscription.ServiceName, subscription.Price, subscription.UserID, subscription.StartDate, subscription.EndDate).Scan(&subscription.ID)
	if err != nil {
		log.Printf("Create (repo) error: %v", err)
		return fmt.Errorf("failed to create subscription: %v", err)
	}

	log.Printf("Create (repo) success: created subscription with id=%v", subscription.ID)
	return nil
}

func (r *subscriptionRepo) GetByID(id uuid.UUID) (*model.Subscription, error) {
	log.Printf("GetByID (repo): retrieving subscription for id=%v", id)
	var s model.Subscription

	err := r.db.QueryRow(
		`
		SELECT id, service_name, price, user_id, start_date, end_date
		FROM subscriptions
		WHERE id = $1
		`, id).Scan(&s.ID, &s.ServiceName, &s.Price, &s.UserID, &s.StartDate, &s.EndDate)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("GetByID (repo) not found: %v", err)
			return nil, sql.ErrNoRows
		}
		log.Printf("GetByID (repo) error: %v", err)
		return nil, fmt.Errorf("failed to get subscription by id: %v", err)
	}

	log.Printf("GetByID (repo) success: subscription found with id=%v", s.ID)
	return &s, nil
}

func (r *subscriptionRepo) GetAll() ([]model.Subscription, error) {
	log.Println("GetAll (repo): fetching all subscriptions")
	rows, err := r.db.Query(`
	SELECT id, service_name, price, user_id, start_date, end_date
	FROM subscriptions
	`)
	if err != nil {
		log.Printf("GetAll (repo) query error: %v", err)
		return nil, fmt.Errorf("failed to get subscriptions: %v", err)
	}
	defer rows.Close()

	var subscriptions []model.Subscription

	for rows.Next() {
		var s model.Subscription

		err = rows.Scan(&s.ID, &s.ServiceName, &s.Price, &s.UserID, &s.StartDate, &s.EndDate)
		if err != nil {
			log.Printf("GetAll (repo) scan error: %v", err)
			return nil, fmt.Errorf("failed to scan subscription: %v", err)
		}
		subscriptions = append(subscriptions, s)
	}

	if err = rows.Err(); err != nil {
		log.Printf("GetAll (repo) rows error: %v", err)
		return nil, fmt.Errorf("failed to get subscriptions: %v", err)
	}

	log.Printf("GetAll (repo) success: found %d subscriptions", len(subscriptions))
	return subscriptions, nil
}

func (r *subscriptionRepo) Update(subscription *model.Subscription) error {
	log.Printf("Update (repo): updating subscription for id=%v", subscription.ID)
	var exists int

	tx, err := r.db.Begin()
	if err != nil {
		log.Printf("Update (repo) transaction error: %v", err)
		return err
	}

	err = tx.QueryRow(
		`
		SELECT 1
		FROM subscriptions
		WHERE id = $1
		FOR UPDATE
		`, subscription.ID).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("Update (repo) not found: %v", err)
			tx.Rollback()
			return fmt.Errorf("subscription not found: %v", err)
		}
		log.Printf("Update (repo) query error: %v", err)
		tx.Rollback()
		return fmt.Errorf("failed to update subscription: %v", err)
	}

	_, err = tx.Exec(
		`
		UPDATE subscriptions
		SET service_name = $2, price = $3, user_id = $4, start_date = $5, end_date = $6
		WHERE id = $1
		`, subscription.ID, subscription.ServiceName, subscription.Price, subscription.UserID, subscription.StartDate, subscription.EndDate)
	if err != nil {
		log.Printf("Update (repo) error: %v", err)
		tx.Rollback()
		return fmt.Errorf("failed to update subscription: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		log.Panicf("Update (repo) commit error: %v", err)
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	log.Printf("Update (repo) success: updated subscription with id=%v", subscription.ID)
	return nil
}

func (r *subscriptionRepo) Delete(id uuid.UUID) error {
	log.Printf("Delete (repo): deleting subscription for id=%v", id)
	_, err := r.db.Exec(
		`
		DELETE FROM subscriptions
		WHERE id = $1
		`, id)
	if err != nil {
		log.Printf("Delete (repo) error: %v", err)
		return fmt.Errorf("failed to delete subscription: %v", err)
	}

	log.Printf("Delete (repo) success: deleted subscription with id=%v", id)
	return nil
}

func (r *subscriptionRepo) GetTotalAmount(userID uuid.UUID, serviceName *string, from, to time.Time) (int, error) {
	log.Printf("GetTotalAmount (repo): getting total amount for user_id=%v, service_name=%v, from=%v, to=%v", userID, serviceName, from, to)
	var totalAmount sql.NullInt64

	query := `SELECT SUM(price)
	FROM subscriptions
	WHERE user_id = $1
	AND start_date BETWEEN $2 AND $3`

	args := []interface{}{userID, from, to}

	if serviceName != nil {
		query += " AND service_name = $4"
		args = append(args, *serviceName)
	}

	err := r.db.QueryRow(query, args...).Scan(&totalAmount)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("GetTotalAmount (repo) not found: %v", err)
			return 0, sql.ErrNoRows
		}

		log.Printf("GetTotalAmount (repo) error: %v", err)
		return 0, fmt.Errorf("failed to get total amount: %v", err)
	}

	if !totalAmount.Valid {
		log.Println("GetTotalAmount (repo): result is NULL, returning 0")
		return 0, nil
	}

	log.Printf("GetTotalAmount (repo) success: total amount = %d", totalAmount.Int64)
	return int(totalAmount.Int64), nil
}
