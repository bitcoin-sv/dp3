package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/libsv/go-bk/envelope"
	"github.com/libsv/go-dpp"
	dppMocks "github.com/libsv/go-dpp/mocks"
	"github.com/stretchr/testify/assert"
)

func TestPaymentTermsHandler_BuildPaymentTerms(t *testing.T) {
	tests := map[string]struct {
		paymentTermsFunc func(context.Context, dpp.PaymentTermsArgs) (*envelope.JSONEnvelope, error)
		paymentID        string
		expResponse      dpp.PaymentTerms
		expStatusCode    int
		expErr           error
	}{
		"successful post": {
			paymentTermsFunc: func(ctx context.Context, args dpp.PaymentTermsArgs) (*envelope.JSONEnvelope, error) {
				return envelope.NewJSONEnvelope(&dpp.PaymentTerms{
					Memo: fmt.Sprintf("payment %s", args.PaymentID),
				})
			},
			paymentID: "abc123",
			expResponse: dpp.PaymentTerms{
				Memo: "payment abc123",
			},
			expStatusCode: http.StatusOK,
		},
		"error is reported back": {
			paymentTermsFunc: func(ctx context.Context, args dpp.PaymentTermsArgs) (*envelope.JSONEnvelope, error) {
				return nil, errors.New("nah darn")
			},
			paymentID: "abc123",
			expErr:    errors.New("nah darn"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			e := echo.New()
			h := NewPaymentTermsHandler(&dppMocks.PaymentTermsServiceMock{
				PaymentTermsFunc: test.paymentTermsFunc,
			})
			g := e.Group("/")
			e.HideBanner = true
			h.RegisterRoutes(g)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			ctx := e.NewContext(req, rec)
			ctx.SetPath("/api/v1/payment/:paymentID")
			ctx.SetParamNames("paymentID")
			ctx.SetParamValues(test.paymentID)

			err := h.buildPaymentTerms(ctx)
			if test.expErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, test.expErr, err.Error())
				return
			}

			response := rec.Result()
			defer response.Body.Close()
			assert.Equal(t, test.expStatusCode, response.StatusCode)

			// These are created with an ephemeral public key that we do not keep track of and it
			// is not our responsibility to test the correctness of the envelope in this limited
			// scenario.
			var env envelope.JSONEnvelope
			assert.NoError(t, json.NewDecoder(response.Body).Decode(&env))

			var actualTerms dpp.PaymentTerms
			err = json.Unmarshal([]byte(env.Payload), &actualTerms)
			assert.Nil(t, err)
			assert.Equal(t, test.expResponse, actualTerms)
		})
	}
}
