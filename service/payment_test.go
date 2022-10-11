package service_test

import (
	"context"
	"errors"
	"github.com/libsv/go-bc/spv"
	"github.com/libsv/go-dpp/modes/hybridmode"
	"testing"

	"github.com/bitcoin-sv/dpp-proxy/log"
	"github.com/bitcoin-sv/dpp-proxy/service"
	"github.com/libsv/go-dpp"
	dppMocks "github.com/libsv/go-dpp/mocks"
	"github.com/stretchr/testify/assert"
)

func TestPayment_Create(t *testing.T) {
	tests := map[string]struct {
		paymentCreateFn func(context.Context, dpp.PaymentCreateArgs, dpp.Payment) (*dpp.PaymentACK, error)
		args            dpp.PaymentCreateArgs
		req             dpp.Payment
		expErr          error
	}{
		"successful payment create": {
			paymentCreateFn: func(context.Context, dpp.PaymentCreateArgs, dpp.Payment) (*dpp.PaymentACK, error) {
				return &dpp.PaymentACK{}, nil
			},
			req: dpp.Payment{
				ModeID: "ef63d9775da5",
				Mode: hybridmode.Payment{
					OptionID:     "choiceID0",
					Transactions: []string{"tx1 hex", "tx2 hex"},
					Ancestors:    map[string]spv.TSCAncestryJSON{},
				},
				Originator: dpp.Originator{
					Name: 		"Bob the builder",
					Paymail: 	"bob@bestpaymail.com",
					Avatar:  	"https://iamges.com/vtwe4eerf",
					ExtendedData: map[string]interface{}{"paymentReference": "omgwow"},
				},
			},
			args: dpp.PaymentCreateArgs{
				PaymentID: "abc123",
			},
		},
		"invalid args errors": {
			paymentCreateFn: func(context.Context, dpp.PaymentCreateArgs, dpp.Payment) (*dpp.PaymentACK, error) {
				return &dpp.PaymentACK{}, nil
			},
			args: dpp.PaymentCreateArgs{},
			req: dpp.Payment{
				ModeID: "ef63d9775da5",
				Mode: hybridmode.Payment{
					OptionID:     "choiceID0",
					Transactions: []string{"tx1 hex", "tx2 hex"},
					Ancestors:    map[string]spv.TSCAncestryJSON{},
				},
				Originator: dpp.Originator{
					Name: 		"Bob the builder",
					Paymail: 	"bob@bestpaymail.com",
					Avatar:  	"https://iamges.com/vtwe4eerf",
					ExtendedData: map[string]interface{}{"paymentReference": "omgwow"},
				},
			},
			expErr: errors.New("[paymentID: value cannot be empty]"),
		},
		"missing mode errors": {
			paymentCreateFn: func(context.Context, dpp.PaymentCreateArgs, dpp.Payment) (*dpp.PaymentACK, error) {
				return &dpp.PaymentACK{}, nil
			},
			args: dpp.PaymentCreateArgs{
				PaymentID: "abc123",
			},
			req: dpp.Payment{
				ModeID: "ef63d9775da5",
				Originator: dpp.Originator{
					Name: 		"Bob the builder",
					Paymail: 	"bob@bestpaymail.com",
					Avatar:  	"https://iamges.com/vtwe4eerf",
					ExtendedData: map[string]interface{}{"paymentReference": "omgwow"},
				},
			},
			expErr: errors.New("[mode.optionId: value cannot be empty], [mode.transactions: value cannot be empty], [mode: value cannot be empty]"),
		},
		"error on payment create is handled": {
			paymentCreateFn: func(context.Context, dpp.PaymentCreateArgs, dpp.Payment) (*dpp.PaymentACK, error) {
				return nil, errors.New("lol oh boi")
			},
			args: dpp.PaymentCreateArgs{
				PaymentID: "abc123",
			},
			req: dpp.Payment{
				ModeID: "ef63d9775da5",
				Mode: hybridmode.Payment{
					OptionID:     "choiceID0",
					Transactions: []string{"tx1 hex", "tx2 hex"},
					Ancestors:    map[string]spv.TSCAncestryJSON{},
				},
				Originator: dpp.Originator{
					Name: 		"Bob the builder",
					Paymail: 	"bob@bestpaymail.com",
					Avatar:  	"https://iamges.com/vtwe4eerf",
					ExtendedData: map[string]interface{}{"paymentReference": "omgwow"},
				},
			},
			expErr: errors.New("lol oh boi"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			svc := service.NewPayment(
				log.Noop{},
				&dppMocks.PaymentWriterMock{
					PaymentCreateFunc: test.paymentCreateFn,
				})

			_, err := svc.PaymentCreate(context.TODO(), test.args, test.req)
			if test.expErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, test.expErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
