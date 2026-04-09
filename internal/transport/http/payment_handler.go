package http

import (
	"context"
	"errors"
	"net/http"

	"payment-service/internal/domain"

	"github.com/gin-gonic/gin"
)

type paymentUsecase interface {
	CreatePayment(ctx context.Context, orderID string, amount int64) (*domain.Payment, error)
	GetPaymentByOrderID(ctx context.Context, orderID string) (*domain.Payment, error)
}

type Handler struct {
	uc paymentUsecase
}

type createPaymentRequest struct {
	OrderID string `json:"order_id"`
	Amount  int64  `json:"amount"`
}

func NewHandler(uc paymentUsecase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) CreatePayment(c *gin.Context) {
	var req createPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payment, err := h.uc.CreatePayment(c.Request.Context(), req.OrderID, req.Amount)
	if err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toPaymentResponse(payment))
}

func (h *Handler) GetPaymentByOrderID(c *gin.Context) {
	payment, err := h.uc.GetPaymentByOrderID(c.Request.Context(), c.Param("order_id"))
	if err != nil {
		h.respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toPaymentResponse(payment))
}

func (h *Handler) respondError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidAmount), errors.Is(err, domain.ErrOrderIDRequired):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrPaymentNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
