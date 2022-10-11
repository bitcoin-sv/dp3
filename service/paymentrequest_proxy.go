package service

import (
	"context"

	"github.com/bitcoin-sv/dpp-proxy/config"
	"github.com/libsv/go-bk/envelope"
	"github.com/libsv/go-dpp"
	"github.com/pkg/errors"
	validator "github.com/theflyingcodr/govalidator"
)

// paymentTermsProxy simply acts as a pass-through to the data layer
// where another service will create the PaymentTerms.
// TODO - remove the other payment request service.
type paymentTermsProxy struct {
	preqRdr   dpp.PaymentTermsReader
	transCfg  *config.Transports
	walletCfg *config.Server
}

// NewPaymentTermsProxy will setup and return a new PaymentTerms service that will generate outputs
// using the provided outputter which is defined in server config.
func NewPaymentTermsProxy(preqRdr dpp.PaymentTermsReader, transCfg *config.Transports, walletCfg *config.Server) *paymentTermsProxy {
	return &paymentTermsProxy{
		preqRdr:   preqRdr,
		transCfg:  transCfg,
		walletCfg: walletCfg,
	}
}

// PaymentTerms will call to the data layer to return a signed JSON envelope containing payment terms.
func (p *paymentTermsProxy) PaymentTerms(ctx context.Context, args dpp.PaymentTermsArgs) (*envelope.JSONEnvelope, error) {
	if err := validator.New().
		Validate("paymentID", validator.NotEmpty(args.PaymentID)); err.Err() != nil {
		return nil, err
	}
	resp, err := p.preqRdr.PaymentTerms(ctx, args)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read payment request for paymentID %s", args.PaymentID)
	}
	return resp, nil
}
