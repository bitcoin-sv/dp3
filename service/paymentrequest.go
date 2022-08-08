package service

import (
	"context"

	"github.com/libsv/go-bk/envelope"
	"github.com/libsv/go-dpp"
	"github.com/pkg/errors"
	validator "github.com/theflyingcodr/govalidator"
)

type paymentTerms struct {
	prRdr dpp.PaymentTermsReader
}

// NewPaymentTerms will setup and return a new PaymentTerms service that will generate outputs
// using the provided outputter which is defined in server config.
func NewPaymentTerms(prRdr dpp.PaymentTermsReader) *paymentTerms {
	return &paymentTerms{
		prRdr: prRdr,
	}
}

// PaymentTerms handles setting up a new PaymentTerms response and will validate that we have a paymentID.
func (p *paymentTerms) PaymentTerms(ctx context.Context, args dpp.PaymentTermsArgs) (*dpp.PaymentTerms, error) {
	if err := validator.New().
		Validate("paymentID", validator.NotEmpty(args.PaymentID)); err.Err() != nil {
		return nil, err
	}

	pReq, err := p.prRdr.PaymentTerms(ctx, args)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get payment request for paymentID %s", args.PaymentID)
	}
	if pReq.Beneficiary != nil && pReq.Beneficiary.ExtendedData == nil {
		pReq.Beneficiary.ExtendedData = map[string]interface{}{
			"paymentReference": args.PaymentID,
		}
	}

	return pReq, nil
}

// PaymentTerms handles setting up a new PaymentTerms response and will validate that we have a paymentID.
func (p *paymentTerms) PaymentTermsSecure(ctx context.Context, args dpp.PaymentTermsArgs) (*envelope.JSONEnvelope, error) {
	if err := validator.New().
		Validate("paymentID", validator.NotEmpty(args.PaymentID)); err.Err() != nil {
		return nil, err
	}

	pReq, err := p.prRdr.PaymentTermsSecure(ctx, args)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get payment request for paymentID %s", args.PaymentID)
	}

	return pReq, nil
}
