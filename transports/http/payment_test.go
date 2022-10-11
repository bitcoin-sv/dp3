package http

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/bitcoin-sv/dpp-proxy/log"
	"github.com/bitcoin-sv/dpp-proxy/transports/client_errors"
	"github.com/bitcoin-sv/dpp-proxy/transports/http/middleware"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/libsv/go-dpp"
	dppMocks "github.com/libsv/go-dpp/mocks"
	"github.com/stretchr/testify/assert"
)

func TestPaymentHandler_CreatedPayment(t *testing.T) {
	tests := map[string]struct {
		paymentCreateFunc func(context.Context, dpp.PaymentCreateArgs, dpp.Payment) (*dpp.PaymentACK, error)
		reqBody           dpp.Payment
		paymentID         string
		expResponse       dpp.PaymentACK
		expTextResponse   string
		expStatusCode     int
		expErr            error
	}{
		"successful post": {
			paymentCreateFunc: func(ctx context.Context, args dpp.PaymentCreateArgs, req dpp.Payment) (*dpp.PaymentACK, error) {
				return &dpp.PaymentACK{}, nil
			},
			paymentID: "abc123",
			reqBody:   dpp.Payment{},
			expResponse: dpp.PaymentACK{},
			expStatusCode: http.StatusCreated,
		},
		"error response returns 422": {
			paymentCreateFunc: func(ctx context.Context, args dpp.PaymentCreateArgs, req dpp.Payment) (*dpp.PaymentACK, error) {
				return nil, client_errors.NewErrUnprocessable("422", "failed")
			},
			paymentID:       "abc123",
			reqBody:         dpp.Payment{},
			expStatusCode:   http.StatusUnprocessableEntity,
			expTextResponse: "\"failed\"\n",
		},
		"payment create service error is handled": {
			paymentCreateFunc: func(ctx context.Context, args dpp.PaymentCreateArgs, req dpp.Payment) (*dpp.PaymentACK, error) {
				return nil, client_errors.NewErrBadRequest("400", "ohnonono")
			},
			paymentID:     "abc123",
			reqBody:       dpp.Payment{},
			expStatusCode: http.StatusBadRequest,
			expTextResponse: "\"ohnonono\"\n",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			e := echo.New()
			h := NewPaymentHandler(&dppMocks.PaymentServiceMock{
				PaymentCreateFunc: test.paymentCreateFunc,
			})
			g := e.Group("/")
			e.HideBanner = true
			h.RegisterRoutes(g)

			body, err := json.Marshal(test.reqBody)
			assert.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body))
			req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)
			ctx.SetPath("/api/v1/payment/:paymentID")
			ctx.SetParamNames("paymentID")
			ctx.SetParamValues(test.paymentID)

			err = h.createPayment(ctx)
			middleware.ErrorHandler(log.Noop{})(err, ctx)

			if test.expErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, test.expErr, err.Error())
				return
			}

			response := rec.Result()
			defer response.Body.Close()
			assert.Equal(t, test.expStatusCode, response.StatusCode)

			if test.expTextResponse != "" {
				var bodyData []byte
				bodyData, err = io.ReadAll(response.Body)
				assert.Nil(t, err)
				assert.Equal(t, test.expTextResponse, string(bodyData))
			} else {
				var ack dpp.PaymentACK
				assert.NoError(t, json.NewDecoder(response.Body).Decode(&ack))
				assert.Equal(t, test.expResponse, ack)
			}
		})
	}
}
