package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"payment-service/internal/domain"
)

type PaymentRepository interface {
	Create(ctx context.Context, payment *domain.Payment) error
	GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error)
}

type PaymentUsecase struct {
	repo PaymentRepository
}

func NewPaymentUsecase(repo PaymentRepository) *PaymentUsecase {
	return &PaymentUsecase{repo: repo}
}

func (uc *PaymentUsecase) CreatePayment(ctx context.Context, orderID string, amount int64) (*domain.Payment, error) {
	if orderID == "" {
		return nil, domain.ErrOrderIDRequired
	}
	if amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}

	payment := &domain.Payment{
		ID:      newID("pay"),
		OrderID: orderID,
		Amount:  amount,
		Status:  domain.PaymentStatusAuthorized,
	}

	if amount > 100000 {
		payment.Status = domain.PaymentStatusDeclined
	} else {
		payment.TransactionID = newID("tx")
	}

	if err := uc.repo.Create(ctx, payment); err != nil {
		return nil, err
	}

	return payment, nil
}

func (uc *PaymentUsecase) GetPaymentByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	if orderID == "" {
		return nil, domain.ErrOrderIDRequired
	}

	return uc.repo.GetByOrderID(ctx, orderID)
}

func newID(prefix string) string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		panic(fmt.Sprintf("generate id: %v", err))
	}

	return prefix + "-" + hex.EncodeToString(buf)
}
