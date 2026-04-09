package grpc

import (
	"context"
	"errors"

	paymentv1 "github.com/ArsikIT/generated-proto-go/proto/payment/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"payment-service/internal/domain"
)

type paymentUsecase interface {
	CreatePayment(ctx context.Context, orderID string, amount int64) (*domain.Payment, error)
}

type Handler struct {
	paymentv1.UnimplementedPaymentServiceServer
	uc paymentUsecase
}

func NewHandler(uc paymentUsecase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) ProcessPayment(ctx context.Context, req *paymentv1.ProcessPaymentRequest) (*paymentv1.ProcessPaymentResponse, error) {
	payment, err := h.uc.CreatePayment(ctx, req.GetOrderId(), req.GetAmount())
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidAmount), errors.Is(err, domain.ErrOrderIDRequired):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}

	message := "payment authorized"
	if payment.Status == domain.PaymentStatusDeclined {
		message = "payment declined"
	}

	return &paymentv1.ProcessPaymentResponse{
		PaymentId:     payment.ID,
		OrderId:       payment.OrderID,
		Status:        payment.Status,
		TransactionId: payment.TransactionID,
		Message:       message,
		ProcessedAt:   timestamppb.Now(),
	}, nil
}
