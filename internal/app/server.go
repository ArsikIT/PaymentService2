package app

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"time"

	paymentv1 "github.com/ArsikIT/generated-proto-go/proto/payment/v1"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"payment-service/internal/repository"
	transportgrpc "payment-service/internal/transport/grpc"
	transporthttp "payment-service/internal/transport/http"
	"payment-service/internal/usecase"
)

type Server struct {
	httpServer   *http.Server
	grpcServer   *grpc.Server
	grpcListener net.Listener
	dbPool       *pgxpool.Pool
}

func NewServer(cfg Config) (*Server, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.PostgresDSN)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	repo := repository.NewPostgresPaymentRepository(pool)
	paymentUsecase := usecase.NewPaymentUsecase(repo)

	httpHandler := transporthttp.NewHandler(paymentUsecase)
	router := gin.Default()
	router.POST("/payments", httpHandler.CreatePayment)
	router.GET("/payments/:order_id", httpHandler.GetPaymentByOrderID)

	grpcHandler := transportgrpc.NewHandler(paymentUsecase)
	grpcServer := grpc.NewServer()
	paymentv1.RegisterPaymentServiceServer(grpcServer, grpcHandler)

	grpcListener, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		pool.Close()
		return nil, err
	}

	return &Server{
		httpServer: &http.Server{
			Addr:    cfg.HTTPAddr,
			Handler: router,
		},
		grpcServer:   grpcServer,
		grpcListener: grpcListener,
		dbPool:       pool,
	}, nil
}

func (s *Server) Run() error {
	group := new(errgroup.Group)

	group.Go(func() error {
		log.Printf("payment-service http listening on %s", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})

	group.Go(func() error {
		log.Printf("payment-service grpc listening on %s", s.grpcListener.Addr().String())
		if err := s.grpcServer.Serve(s.grpcListener); err != nil && !errors.Is(err, net.ErrClosed) {
			return err
		}
		return nil
	})

	return group.Wait()
}

func (s *Server) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = s.httpServer.Shutdown(ctx)
	s.grpcServer.GracefulStop()
	_ = s.grpcListener.Close()
	s.dbPool.Close()
}
