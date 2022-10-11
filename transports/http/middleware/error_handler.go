package middleware

import (
	"errors"
	"net/http"

	server "github.com/bitcoin-sv/dpp-proxy"
	"github.com/bitcoin-sv/dpp-proxy/log"
	"github.com/bitcoin-sv/dpp-proxy/transports/client_errors"
	"github.com/labstack/echo/v4"
	validator "github.com/theflyingcodr/govalidator"
	"github.com/theflyingcodr/lathos"
	"github.com/theflyingcodr/lathos/errs"
)

// ErrorHandler we can flesh this out.
func ErrorHandler(l log.Logger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if err == nil {
			return
		}
		var valErr validator.ErrValidation
		if errors.As(err, &valErr) {
			resp := map[string]interface{}{
				"errors": valErr,
			}
			_ = c.JSON(http.StatusBadRequest, resp)
			return
		}

		if errors.Is(err, echo.ErrNotFound) {
			err = client_errors.NewErrNotFound("404", "Not Found")
		}

		var cErr server.ClientError
		if errors.As(err, &cErr) {
			_ = c.JSON(http.StatusBadRequest, cErr.Message)
			return
		}

		// Internal server error, log it to a system and return small detail
		if !lathos.IsClientError(err) {
			internalErr := errs.NewErrInternal(err, "500")
			l.Error(internalErr, "Internal Server Error")
			_ = c.JSON(http.StatusInternalServerError, internalErr.Error())
			return
		}
		var clientErr lathos.ClientError
		errors.As(err, &clientErr)
		resp := server.ClientError{
			ID:      clientErr.ID(),
			Code:    clientErr.Code(),
			Title:   clientErr.Title(),
			Message: clientErr.Detail(),
		}
		if lathos.IsNotFound(err) {
			_ = c.JSON(http.StatusNotFound, resp.Message)
			return
		}
		if lathos.IsDuplicate(err) {
			_ = c.JSON(http.StatusConflict, resp.Message)
			return
		}
		if lathos.IsNotAuthenticated(err) {
			_ = c.JSON(http.StatusUnauthorized, resp.Message)
			return
		}
		if lathos.IsNotAuthorised(err) {
			_ = c.JSON(http.StatusForbidden, resp.Message)
			return
		}
		if lathos.IsCannotProcess(err) {
			_ = c.JSON(http.StatusUnprocessableEntity, resp.Message)
			return
		}
		if lathos.IsBadRequest(err) {
			_ = c.JSON(http.StatusBadRequest, resp.Message)
			return
		}
	}
}
