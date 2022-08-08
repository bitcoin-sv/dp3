package payd

import (
	"context"
	"fmt"
	"net/http"

	"github.com/libsv/go-bk/envelope"
	"github.com/pkg/errors"

	"github.com/bitcoin-sv/dpp-proxy/config"
	"github.com/bitcoin-sv/dpp-proxy/data"
	"github.com/bitcoin-sv/dpp-proxy/data/payd/models"
	"github.com/libsv/go-dpp"
)

// Known endpoints for the payd wallet implementing the payment protocol interface.
const (
	urlPayments      = "%s/api/v1/payments/%s"
	urlProofs        = "%s/api/v1/proofs/%s"
	protocolInsecure = "http"
	protocolSecure   = "https"
)

type payd struct {
	client data.HTTPClient
	cfg    *config.PayD
}

// NewPayD will setup a new store that can interface with a payd wallet implementing
// the Payment Protocol Interface.
func NewPayD(cfg *config.PayD, client data.HTTPClient) *payd {
	return &payd{
		cfg:    cfg,
		client: client,
	}
}

// PaymentTerms will fetch a payment request message from payd for a given messages.
func (p *payd) PaymentTerms(ctx context.Context, args dpp.PaymentTermsArgs) (*dpp.PaymentTerms, error) {
	var resp dpp.PaymentTerms
	if err := p.client.Do(ctx, http.MethodGet, fmt.Sprintf(urlPayments, p.baseURL(), args.PaymentID), http.StatusOK, nil, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// PaymentTerms will fetch a payment request message from payd for a given messages.
func (p *payd) PaymentTermsSecure(ctx context.Context, args dpp.PaymentTermsArgs) (*envelope.JSONEnvelope, error) {
	return nil, errors.New("PayD unsupported for payment terms (secure)")
}

// PaymentCreate will post a request to payd to validate and add the txos to the wallet.
//
// If invalid a non 204 status code is returned.
func (p *payd) PaymentCreate(ctx context.Context, args dpp.PaymentCreateArgs, req dpp.Payment) (*dpp.PaymentACK, error) {
	paymentReq := models.PayDPayment{
		ModeID:      req.ModeID,
		Mode:        req.Mode,
		Originator:  req.Originator,
		Transaction: req.Transaction,
		Memo:        req.Memo,
	}
	var ack dpp.PaymentACK
	if err := p.client.Do(ctx, http.MethodPost, fmt.Sprintf(urlPayments, p.baseURL(), args.PaymentID), http.StatusNoContent, paymentReq, &ack); err != nil {
		return nil, err
	}
	return &ack, nil
}

// ProofCreate will pass on the proof to a payd instance for storage.
func (p *payd) ProofCreate(ctx context.Context, args dpp.ProofCreateArgs, req envelope.JSONEnvelope) error {
	return errors.WithStack(p.client.Do(ctx, http.MethodPost, fmt.Sprintf(urlProofs, p.baseURL(), args.TxID), http.StatusCreated, req, nil))
}

// baseURL will return http or https depending on if we're using TLS.
func (p *payd) baseURL() string {
	if p.cfg.Secure {
		return fmt.Sprintf("%s://%s%s", protocolSecure, p.cfg.Host, p.cfg.Port)
	}
	return fmt.Sprintf("%s://%s%s", protocolInsecure, p.cfg.Host, p.cfg.Port)
}
