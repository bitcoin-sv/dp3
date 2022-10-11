package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/libsv/go-bk/envelope"
	"github.com/libsv/go-dpp/modes/hybridmode"
	"github.com/libsv/go-dpp/nativetypes"

	"github.com/bitcoin-sv/dpp-proxy/service"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-dpp"
	dppMocks "github.com/libsv/go-dpp/mocks"
	"github.com/stretchr/testify/assert"
)

func TestPaymentTerms_PaymentTerms(t *testing.T) {
	created := time.Now()
	expired := created.Add(time.Hour * 24)
	tests := map[string]struct {
		paymentTermsFunc func(context.Context, dpp.PaymentTermsArgs) (*envelope.JSONEnvelope, error)
		args             dpp.PaymentTermsArgs
		expResp          *dpp.PaymentTerms
		expErr           error
	}{
		"successful request": {
			args: dpp.PaymentTermsArgs{
				PaymentID: "abc123",
			},
			paymentTermsFunc: func(context.Context, dpp.PaymentTermsArgs) (*envelope.JSONEnvelope, error) {
				return envelope.NewJSONEnvelope(&dpp.PaymentTerms{
					Network:             "regtest",
					Version:             "1.0",
					CreationTimestamp:   created.Unix(),
					ExpirationTimestamp: expired.Unix(),
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
					PaymentURL: "http://iamsotest/api/v1/payment/abc123",
					Memo:       "invoice abc123",
					Beneficiary: &dpp.Beneficiary{
						ExtendedData: map[string]interface{}{"paymentReference": "abc123"},
					},
				})
			},
			expResp: &dpp.PaymentTerms{
				Network:             "regtest",
				Version:             "1.0",
				CreationTimestamp:   created.Unix(),
				ExpirationTimestamp: expired.Unix(),
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
				PaymentURL: "http://iamsotest/api/v1/payment/abc123",
				Memo:       "invoice abc123",
				Beneficiary: &dpp.Beneficiary{
					ExtendedData: map[string]interface{}{"paymentReference": "abc123"},
				},
			},
		},
		"successful request with nil extended data": {
			args: dpp.PaymentTermsArgs{
				PaymentID: "abc123",
			},
			paymentTermsFunc: func(context.Context, dpp.PaymentTermsArgs) (*envelope.JSONEnvelope, error) {
				return envelope.NewJSONEnvelope(&dpp.PaymentTerms{
					Network:             "regtest",
					Version:             "1.0",
					CreationTimestamp:   created.Unix(),
					ExpirationTimestamp: expired.Unix(),
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
					Beneficiary: &dpp.Beneficiary{
						ExtendedData: map[string]interface{}{"paymentReference": "abc123"},
					},
					PaymentURL: "http://iamsotest/api/v1/payment/abc123",
					Memo:       "invoice abc123",
				})
			},
			expResp: &dpp.PaymentTerms{
				Network:             "regtest",
				Version:             "1.0",
				CreationTimestamp:   created.Unix(),
				ExpirationTimestamp: expired.Unix(),
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
				PaymentURL: "http://iamsotest/api/v1/payment/abc123",
				Memo:       "invoice abc123",
				Beneficiary: &dpp.Beneficiary{
					ExtendedData: map[string]interface{}{"paymentReference": "abc123"},
				},
			},
		},
		"invalid args rejected": {
			expErr: errors.New("[paymentID: value cannot be empty]"),
		},
		"payment request reader error handled and reported": {
			args: dpp.PaymentTermsArgs{
				PaymentID: "abc123",
			},
			paymentTermsFunc: func(context.Context, dpp.PaymentTermsArgs) (*envelope.JSONEnvelope, error) {
				return nil, errors.New("oh boi")
			},
			expErr: errors.New("failed to get payment request for paymentID abc123: oh boi"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			svc := service.NewPaymentTerms(&dppMocks.PaymentTermsServiceMock{
				PaymentTermsFunc: test.paymentTermsFunc,
			})

			resp, err := svc.PaymentTerms(context.TODO(), test.args)
			if test.expErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, test.expErr.Error())
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, resp)

			var pr dpp.PaymentTerms
			// Testing the validity of the envelope is for the package that code comes from.
			err = json.Unmarshal([]byte(resp.Payload), &pr)
			assert.Nil(t, err)
			assert.Equal(t, *test.expResp, pr)
		})
	}
}
