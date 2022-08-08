package service

import (
	"context"
	"fmt"
	"net/url"

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

// PaymentTerms will call to the data layer to return a full payment request.
func (p *paymentTermsProxy) PaymentTerms(ctx context.Context, args dpp.PaymentTermsArgs) (*dpp.PaymentTerms, error) {
	if err := validator.New().
		Validate("paymentID", validator.NotEmpty(args.PaymentID)); err.Err() != nil {
		return nil, err
	}
	resp, err := p.preqRdr.PaymentTerms(ctx, args)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read payment request for paymentID %s", args.PaymentID)
	}

	if len(resp.Modes.Hybrid["choiceID0"]["transactions"][0].Outputs.NativeOutputs) == 0 {
		return nil, fmt.Errorf("no outputs received for paymentID %s", args.PaymentID)
	}

	if resp.Modes.Hybrid["choiceID0"]["transactions"][0].Policies.FeeRate == nil {
		return nil, fmt.Errorf("no fees received for paymentID %s", args.PaymentID)
	}

	if p.transCfg.Mode == config.TransportModeHybrid {
		u := url.URL{
			Scheme: "http",
			Host:   p.walletCfg.FQDN,
			Path:   "/api/v1/payment/" + args.PaymentID,
		}
		resp.PaymentURL = u.String()
	}

	return resp, nil
}

// PaymentTerms will call to the data layer to return a full payment request.
func (p *paymentTermsProxy) PaymentTermsSecure(ctx context.Context, args dpp.PaymentTermsArgs) (*envelope.JSONEnvelope, error) {
	if err := validator.New().
		Validate("paymentID", validator.NotEmpty(args.PaymentID)); err.Err() != nil {
		return nil, err
	}
	resp, err := p.preqRdr.PaymentTermsSecure(ctx, args)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read payment request for paymentID %s", args.PaymentID)
	}
	return resp, nil
}
