package dpp

import (
	"context"

	"github.com/libsv/go-bk/envelope"
	"github.com/libsv/go-dpp/modes/hybridmode"
	"github.com/libsv/go-dpp/nativetypes"
)

// These structures are defined in the TSC spec:
// See https://tsc.bitcoinassociation.net/standards/direct_payment_protocol

// PaymentTermsModes message used in DPP TSC spec. for the `PaymentTerms` message.
type PaymentTermsModes struct {
	// Hybrid contains a key value map of possible payment terms modalities - currently there is only one option:
	// `HybridPaymentMode` with BRFCID: "ef63d9775da5".
	Hybrid hybridmode.PaymentTerms `json:"ef63d9775da5"`
}

// PaymentTerms message as defined in the DPP T$C spec.
type PaymentTerms struct {
	// Network  Always set to "bitcoin" (but seems to be set to 'bitcoin-sv'
	// outside bip270 spec, see https://handcash.github.io/handcash-merchant-integration/#/merchant-payments)
	// {enum: bitcoin, bitcoin-sv, test}
	// Required.
	Network string `json:"network" binding:"required" example:"mainnet" enums:"mainnet,testnet,stn,regtest"`
	// Version version of DPP TSC spec.
	// Required.
	Version string `json:"version" binding:"required" example:"1.0"`
	// Outputs an array of outputs. DEPRECATED but included for backward compatibility.
	// Optional.
	Outputs []nativetypes.NativeOutput `json:"outputs,omitempty"`
	// CreationTimestamp Unix timestamp (seconds since 1-Jan-1970 UTC) when the PaymentTerms were created.
	// Required.
	CreationTimestamp int64 `json:"creationTimestamp" binding:"required" swaggertype:"primitive,int" example:"1648163657"`
	// ExpirationTimestamp Unix timestamp (UTC) after which the PaymentTerms should be considered invalid.
	// Optional.
	ExpirationTimestamp int64 `json:"expirationTimestamp" binding:"required" swaggertype:"primitive,int" example:"1648164657"`
	// PaymentURL secure HTTPS location where a Payment message (see below) will be sent to obtain a PaymentACK.
	// Maximum length is 4000 characters
	PaymentURL string `json:"paymentUrl" binding:"required" example:"http://localhost:3443/api/v1/payment/123456"`
	// Memo note that should be displayed to the customer, explaining what these PaymentTerms are for.
	// Maximum length is 50 characters.
	// Optional.
	Memo string `json:"memo,omitempty" example:"invoice number 123456"`
	// Beneficiary Arbitrary data that may be used by the payment host to identify the PaymentTerms
	// May be omitted if the payment host does not need to associate Payments with PaymentTerms
	// or if they associate each PaymentTerms with a separate payment address.
	// Maximum length is 10000 characters.
	// Optional.
	Beneficiary *Beneficiary `json:"beneficiary,omitempty"`
	// PaymentTermsModes TSC payment messages specified by ID (and well defined) messages customer can choose to pay
	// A key-value map. required field but not if legacy BIP270 outputs are provided
	Modes *PaymentTermsModes `json:"modes"`
}

// PaymentTermsArgs are request arguments that can be passed to the service.
type PaymentTermsArgs struct {
	// PaymentID is an identifier for an invoice.
	PaymentID string `param:"paymentID"`
}

// PaymentTermsService can be implemented to enforce business rules
// and process in order to fulfil PaymentTerms.
type PaymentTermsService interface {
	PaymentTermsReader
}

// PaymentTermsReader will return a new payment request.
type PaymentTermsReader interface {
	PaymentTerms(ctx context.Context, args PaymentTermsArgs) (*PaymentTerms, error)
	PaymentTermsSecure(ctx context.Context, args PaymentTermsArgs) (*envelope.JSONEnvelope, error)
}
