package sockets

import (
	"github.com/theflyingcodr/sockets"

	"context"

	"github.com/theflyingcodr/sockets/server"
)


type paymentTerms struct {
}

// NewPaymentTerms will setup a new instance of a paymentTerms handler.
func NewPaymentTerms() *paymentTerms {
	return &paymentTerms{}
}

// Register will register new handler/s with the socket server.
func (p *paymentTerms) Register(s *server.SocketServer) {
	s.RegisterChannelHandler("paymentterms.create", p.buildPaymentTerms)
	s.RegisterChannelHandler("paymentterms.response", p.paymentTermsResponse)
}

// buildPaymentTerms will forward a paymentterms.create message to all connected clients.
func (p *paymentTerms) buildPaymentTerms(ctx context.Context, msg *sockets.Message) (*sockets.Message, error) {
	return msg, nil
}

// buildPaymentTerms will forward a paymentterms.response message to all connected clients.
func (p *paymentTerms) paymentTermsResponse(ctx context.Context, msg *sockets.Message) (*sockets.Message, error) {
	return msg, nil
}
