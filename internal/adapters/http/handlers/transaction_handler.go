package handlers

import (
	"net/http"
	"wallet/internal/adapters/http/requests"
	"wallet/internal/adapters/http/responses"
	"wallet/internal/ports"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type TransactionHandler struct {
	transactionService ports.TransactionService
}

func NewTransactionHandler(transactionService ports.TransactionService) *TransactionHandler {
	return &TransactionHandler{
		transactionService: transactionService,
	}
}

func (h *TransactionHandler) VerifyTransaction(c *fiber.Ctx) error {
	var req requests.VerifyTransactionRequest

	if e := c.BodyParser(&req); e != nil {
		return c.Status(http.StatusBadRequest).JSON(responses.NewErrorResponse(http.StatusBadRequest, "Invalid request body"))
	}

	if e := validator.New().Struct(&req); e != nil {
		errorMsgs := make([]string, 0)
		for _, fieldError := range e.(validator.ValidationErrors) {
			errorMsgs = append(errorMsgs, fieldError.Error())
		}
		return c.Status(http.StatusBadRequest).JSON(responses.NewErrorResponse(http.StatusBadRequest, errorMsgs...))
	}

	result, e := h.transactionService.VerifyTransaction(ports.VerifyTransactionPayload{
		UserID:        req.UserID,
		Amount:        req.Amount,
		PaymentMethod: req.PaymentMethod,
	})
	if e != nil {
		return c.Status(http.StatusInternalServerError).JSON(responses.NewErrorResponse(http.StatusInternalServerError, "Invalid request body"))
	}

	return c.JSON(responses.VerifyTransactionResponse{
		TransactionID: result.TransactionID,
		UserID:        result.UserID,
		Amount:        result.Amount,
		PaymentMethod: result.PaymentMethod,
		Status:        result.Status,
		ExpiresAt:     result.ExpiresAt,
	})
}

func (h *TransactionHandler) ConfirmTransaction(c *fiber.Ctx) error {
	var req requests.ConfirmTransactionRequest

	if e := c.BodyParser(&req); e != nil {
		return e
	}

	if e := validator.New().Struct(&req); e != nil {
		return e
	}

	result, e := h.transactionService.ConfirmTransaction(ports.ConfirmTransactionPayload{
		TransactionID: req.TransactionID,
	})
	if e != nil {
		return c.Status(http.StatusInternalServerError).JSON(responses.NewErrorResponse(http.StatusInternalServerError, "Invalid request body"))
	}

	return c.JSON(responses.ConfirmTransactionResponse{
		TransactionID: result.TransactionID,
		UserID:        result.UserID,
		Amount:        result.Amount,
		Status:        result.Status,
		Balance:       result.Balance,
	})
}
