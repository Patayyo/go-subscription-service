package main

import (
	"fmt"
	"go-subscriptions-service/db"
	"go-subscriptions-service/internal/handler"
	"go-subscriptions-service/internal/repo"
	"go-subscriptions-service/internal/service"
	"log"
	"net/http"

	_ "go-subscriptions-service/docs"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Subscriptions Service API
// @version 1.0
// @description REST API сервис для управления онлайн-подписками пользователей
// @host localhost:8080
// @BasePath /
func main() {
	db.InitEnv()

	conn := db.Connect()
	defer conn.Close()

	if err := conn.Ping(); err != nil {
		log.Fatal("Cannot connect to db: ", err)
	}

	subscriptionRepo := repo.NewSubscriptionRepo(conn)
	subscriptionService := service.NewSubscriptionService(subscriptionRepo)
	subscriptionHandler := handler.NewSubscriptionHandler(subscriptionService)

	router := mux.NewRouter()
	subscriptionHandler.RegisterRouters(router)

	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	fmt.Println("Server is listening on port 8080")
	http.ListenAndServe(":8080", router)
}
