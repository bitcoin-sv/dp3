package middleware_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bitcoin-sv/dpp-proxy/log"
	"github.com/bitcoin-sv/dpp-proxy/transports/client_errors"
	"github.com/bitcoin-sv/dpp-proxy/transports/http/middleware"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	validator "github.com/theflyingcodr/govalidator"
)

func TestErrorHandler(t *testing.T) {
	tests := map[string]struct {
		err           error
		expResp       interface{}
		expStatusCode int
	}{
		"client error 400": {
			err: validator.ErrValidation{
				"paymentID": []string{"no style", "no class"},
			},
			expResp: map[string]interface{}{
				"errors": map[string]interface{}{
					"paymentID": []interface{}{"no style", "no class"},
				},
			},
			expStatusCode: http.StatusBadRequest,
		},
		"internal server error 500": {
			err: errors.New("There was an unexpected server error"),
			expResp: "There was an unexpected server error",
			expStatusCode: http.StatusInternalServerError,
		},
		"not found 404": {
			err: client_errors.NewErrNotFound("404", "invoice not found"),
			expResp: "invoice not found",
			expStatusCode: http.StatusNotFound,
		},
		"conflict 409": {
			err: client_errors.NewErrDuplicate("409", "item already exists"),
			expResp: "item already exists",
			expStatusCode: http.StatusConflict,
		},
		"not auth'd 401": {
			err: client_errors.NewErrNotAuthenticated("401", "will ya login"),
			expResp: "will ya login",
			expStatusCode: http.StatusUnauthorized,
		},
		"forbidden 403": {
			err: client_errors.NewErrNotAuthorised("403", "lol nice try buddy"),
			expResp: "lol nice try buddy",
			expStatusCode: http.StatusForbidden,
		},
		"cannot process 422": {
			err: client_errors.NewErrUnprocessable("422", "what did you even send?"),
			expResp: "what did you even send?",
			expStatusCode: http.StatusUnprocessableEntity,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()

			e := echo.New()
			ctx := e.NewContext(req, rec)
			middleware.ErrorHandler(log.Noop{})(test.err, ctx)

			response := rec.Result()
			defer response.Body.Close()

			var mm interface{}
			assert.NoError(t, json.NewDecoder(response.Body).Decode(&mm))
			if m, ok := mm.(map[string]interface{}); ok {
				delete(m, "id")
			}

			assert.Equal(t, test.expResp, mm)
			assert.Equal(t, test.expStatusCode, response.StatusCode)
		})
	}
}
