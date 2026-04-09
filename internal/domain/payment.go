package domain

import "errors"

const (
	PaymentStatusAuthorized = "Authorized"
	PaymentStatusDeclined   = "Declined"
)

var (
	ErrInvalidAmount   = errors.New("amount must be greater than 0")
	ErrOrderIDRequired = errors.New("order_id is required")
	ErrPaymentNotFound = errors.New("payment not found")
)

type Payment struct {
	ID            string
	OrderID       string
	TransactionID string
	Amount        int64
	Status        string
}
