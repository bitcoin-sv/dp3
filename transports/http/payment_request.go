package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/libsv/go-dpp"
	"github.com/pkg/errors"
)

type (
	// PaymentTermsHandler is an http handler that supports BIP-270 requests.
	PaymentTermsHandler struct {
		svc dpp.PaymentTermsService
	}
)

// NewPaymentTermsHandler will create and return a new PaymentTermsHandler.
func NewPaymentTermsHandler(svc dpp.PaymentTermsService) *PaymentTermsHandler {
	return &PaymentTermsHandler{
		svc: svc,
	}
}

// RegisterRoutes will setup all routes with an echo group.
func (h *PaymentTermsHandler) RegisterRoutes(g *echo.Group) {
	g.GET(RouteV1PaymentTerms, h.buildPaymentTerms)
	g.GET(RouteV1PaymentTermsSecure, h.buildPaymentTermsSecure)
}

// buildPaymentTerms will setup and return a new payment request.
// @Summary Request to pay an invoice and receive back outputs to use when constructing the payment transaction
// @Description Creates a payment request based on a payment id (the identifier for an invoice).
// @Tags Payment
// @Accept json
// @Produce json
// @Param paymentID path string true "Payment ID"
// @Success 201 {object} dpp.PaymentTerms "contains outputs, merchant data and expiry information, used by the payee to construct a transaction"
// @Failure 404 {object} server.ClientError "returned if the paymentID has not been found"
// @Failure 400 {object} server.ClientError "returned if the user input is invalid, usually an issue with the paymentID"
// @Failure 500 {string} string "returned if there is an unexpected internal error"
// @Router /api/v1/payment/{paymentID} [GET].
func (h *PaymentTermsHandler) buildPaymentTerms(e echo.Context) error {
	var args dpp.PaymentTermsArgs
	if err := e.Bind(&args); err != nil {
		return errors.Wrap(err, "failed to bind request")
	}
	resp, err := h.svc.PaymentTerms(e.Request().Context(), args)
	if err != nil {
		return errors.WithStack(err)
	}
	return e.JSON(http.StatusOK, resp)
}

// buildPaymentTermsSecure will setup and return a new payment request.
// @Summary Request to pay an invoice and receive back outputs to use when constructing the payment transaction
// @Description Creates a payment request based on a payment id (the identifier for an invoice).
// @Tags Payment
// @Accept json
// @Produce json
// @Param paymentID path string true "Payment ID"
// @Success 201 {object} envelope.JSONEnvelope "contains the signed PaymentTerms"
// @Failure 404 {object} server.ClientError "returned if the paymentID has not been found"
// @Failure 400 {object} server.ClientError "returned if the user input is invalid, usually an issue with the paymentID"
// @Failure 500 {string} string "returned if there is an unexpected internal error"
// @Router /api/v1/payment/sec/{paymentID} [GET].
func (h *PaymentTermsHandler) buildPaymentTermsSecure(e echo.Context) error {
	var args dpp.PaymentTermsArgs
	if err := e.Bind(&args); err != nil {
		return errors.Wrap(err, "failed to bind request")
	}
	resp, err := h.svc.PaymentTermsSecure(e.Request().Context(), args)
	if err != nil {
		return errors.WithStack(err)
	}
	return e.JSON(http.StatusOK, resp)
}
