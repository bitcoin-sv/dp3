package sockets

import (
	"context"
	"fmt"
	"time"

	server "github.com/bitcoin-sv/dpp-proxy"
	"github.com/google/uuid"
	"github.com/libsv/go-bk/envelope"
	"github.com/pkg/errors"
	"github.com/theflyingcodr/lathos/errs"
	"github.com/theflyingcodr/sockets"

	"github.com/libsv/go-dpp"
)

// Routes contain the unique keys for socket messages used in the payment protocol.
const (
	RoutePayment              = "payment"
	RoutePaymentACK           = "payment.ack"
	RoutePaymentError         = "payment.error"
	RouteProofCreate          = "proof.create"
	RoutePaymentTermsCreate   = "paymentterms.create"
	RoutePaymentTermsResponse = "paymentterms.response"
	RoutePaymentTermsError    = "paymentterms.error"

	appID = "dpp"
)

type PaymentStore struct {
	s sockets.ServerChannelBroadcaster
}

// NewPaymentStore will setup and return a new payd socket data store.
func NewPaymentStore(b sockets.ServerChannelBroadcaster) *PaymentStore {
	return &PaymentStore{s: b}
}

// ProofCreate will broadcast the proof to all currently listening clients on the socket channel.
func (p *PaymentStore) ProofCreate(ctx context.Context, args dpp.ProofCreateArgs, req envelope.JSONEnvelope) error {
	msg := sockets.NewMessage(RouteProofCreate, "", args.PaymentReference)
	msg.AppID = appID
	msg.CorrelationID = args.TxID
	if err := msg.WithBody(req); err != nil {
		return err
	}
	msg.Headers.Add("x-tx-id", args.TxID)
	p.s.Broadcast(args.PaymentReference, msg)
	return nil
}

// PaymentTerms will send a socket request to a payd client for a payment request.
// It will wait on a response before returnign the payment request.
func (p *PaymentStore) PaymentTerms(ctx context.Context, args dpp.PaymentTermsArgs) (*envelope.JSONEnvelope, error) {
	msg := sockets.NewMessage(RoutePaymentTermsCreate, "", args.PaymentID)
	msg.AppID = appID
	msg.CorrelationID = uuid.NewString()

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	resp, err := p.s.BroadcastAwait(ctx, args.PaymentID, msg)
	if err != nil {
		if errors.Is(err, sockets.ErrChannelNotFound) {
			return nil, errs.NewErrNotFound("N00001", "invoice not found")
		}
		return nil, errors.Wrap(err, "failed to broadcast message for payment terms (secure)")
	}
	switch resp.Key() {
	case RoutePaymentTermsResponse:
		var pr *envelope.JSONEnvelope
		if err := resp.Bind(&pr); err != nil {
			return nil, errors.Wrap(err, "failed to bind payment terms (secure) response")
		}
		return pr, nil
	case RoutePaymentTermsError:
		var clientErr server.ClientError
		if err := resp.Bind(&clientErr); err != nil {
			return nil, errors.Wrap(err, "failed to bind error response")
		}
		return nil, toLathosErr(clientErr)
	}

	return nil, fmt.Errorf("unexpected response key '%s'", resp.Key())
}

// PaymentCreate will send a request to payd to create and process the payment.
func (p *PaymentStore) PaymentCreate(ctx context.Context, args dpp.PaymentCreateArgs, req dpp.Payment) (*dpp.PaymentACK, error) {
	msg := sockets.NewMessage(RoutePayment, "", args.PaymentID)
	msg.AppID = appID
	msg.CorrelationID = uuid.NewString()
	if err := msg.WithBody(req); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	resp, err := p.s.BroadcastAwait(ctx, args.PaymentID, msg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send payment message for payment")
	}
	switch resp.Key() {
	case RoutePaymentACK:
		var pr *dpp.PaymentACK
		if err := resp.Bind(&pr); err != nil {
			return nil, errors.Wrap(err, "failed to bind payment ack response")
		}
		return pr, nil
	case RoutePaymentError:
		var clientErr server.ClientError
		if err := resp.Bind(&clientErr); err != nil {
			return nil, errors.Wrap(err, "failed to bind error response")
		}
		return nil, toLathosErr(clientErr)
	}

	return nil, fmt.Errorf("unexpected response key '%s'", resp.Key())
}

func toLathosErr(c server.ClientError) error {
	switch c.Code {
	case "404", "N0001":
		return errs.NewErrNotFound(c.Code, c.Message)
	}

	return c
}
