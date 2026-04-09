package app

import (
	"context"
	"time"

	"payment-service/internal/repository"
	transporthttp "payment-service/internal/transport/http"
	"payment-service/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewServer(cfg Config) (*gin.Engine, func(), error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.PostgresDSN)
	if err != nil {
		return nil, nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, nil, err
	}

	repo := repository.NewPostgresPaymentRepository(pool)
	paymentUsecase := usecase.NewPaymentUsecase(repo)
	handler := transporthttp.NewHandler(paymentUsecase)

	router := gin.Default()
	router.POST("/payments", handler.CreatePayment)
	router.GET("/payments/:order_id", handler.GetPaymentByOrderID)

	cleanup := func() {
		pool.Close()
	}

	return router, cleanup, nil
}
