package noop

import (
	"context"
	"time"

	"github.com/libsv/go-dpp/modes/hybridmode"

	"github.com/bitcoin-sv/dpp-proxy/log"
	"github.com/libsv/go-bk/envelope"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-dpp"
	"github.com/libsv/go-dpp/nativetypes"
)

type noop struct {
	l log.Logger
}

// NewNoOp will setup and return a new no operational data store for
// testing purposes. Useful if you want to explore endpoints without
// integrating with a wallet.
func NewNoOp(l log.Logger) *noop {
	l.Info("using NOOP data store")
	return &noop{}
}

// PaymentCreate will post a request to payd to validate and add the txos to the wallet.
//
// If invalid a non 204 status code is returned.
func (n *noop) PaymentCreate(ctx context.Context, args dpp.PaymentCreateArgs, req dpp.Payment) (*dpp.PaymentACK, error) {
	n.l.Info("hit noop.PaymentCreate")
	return &dpp.PaymentACK{}, nil
}

func (n noop) PaymentTerms(ctx context.Context, args dpp.PaymentTermsArgs) (*dpp.PaymentTerms, error) {
	return &dpp.PaymentTerms{
		Network:             "noop",
		Version:             "1.0",
		CreationTimestamp:   time.Now().Unix(),
		ExpirationTimestamp: time.Now().Add(time.Hour).Unix(),
		Memo:                "noop",
		PaymentURL:          "noop",
		Beneficiary: &dpp.Beneficiary{
			AvatarURL:        "noop",
			Name:             "noop",
			Email:            "noop",
			Address:          "noop",
			ExtendedData:     nil,
			PaymentReference: "noop",
		},
		Modes: &dpp.PaymentTermsModes{
			Hybrid: hybridmode.PaymentTerms{
				"choiceID0": {
					"transactions": {
						hybridmode.TransactionTerms{
							Outputs: hybridmode.Outputs{NativeOutputs: []nativetypes.NativeOutput{
								{
									Amount: 1000,
									LockingScript: func() *bscript.Script {
										ls, _ := bscript.NewFromHexString(
											"76a91493d0d43918a5df78f08cfe22a4e022846b6736c288ac")
										return ls
									}(),
									Description: "noop description",
								},
							}},
							Inputs:   hybridmode.Inputs{},
							Policies: &hybridmode.Policies{},
						},
					},
				},
			},
		},
	}, nil
}

func (n noop) PaymentTermsSecure(ctx context.Context, args dpp.PaymentTermsArgs) (*envelope.JSONEnvelope, error) {
	var signature string = "3044022004cf2c5711f34f0de11fd316074c44ce0f63a525840aae0cf61d9dee04b317b102201a56049354449ddce3d8b059403b2d866662b6d1f9d0064365d420406d8d992d"
	var publicKey string = "03d546057437f3279f66d6ae91a03ffe1120ef3a79b8f186d9b6a8f1e0582ccf78"
	return &envelope.JSONEnvelope{
		Payload:   "{\"network\":\"mainnet\",\"creationTimestamp\":{},\"expirationTimestamp\":{},\"url\":\"https://localhost:3443/api/v1/payment/123456\",\"memo\":\"string\",\"beneficiary\":{\"name\":\"beneficiary 1\",\"avatar\":\"http://url.com\",\"extensions\":{\"email\":\"beneficiary@m.com\",\"address\":\"1 the street, the town, B1 1AA\",\"additionalProp1\":{}}},\"outputs\":[{\"amount\":100000,\"script\":\"76a91455b61be43392125d127f1780fb038437cd67ef9c88ac\",\"description\":\"paymentReference 123456\"}],\"fees\":{\"data\":{\"satoshis\":0,\"bytes\":0},\"standard\":{\"satoshis\":0,\"bytes\":0}},\"ancestry\":{\"format\":\"binary\",\"minDepth\":0}}",
		Signature: &signature,
		PublicKey: &publicKey,
		Encoding:  "UTF-8",
		MimeType:  "application/json",
	}, nil
}
