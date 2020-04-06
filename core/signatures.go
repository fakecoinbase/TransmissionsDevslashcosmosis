package core

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

// A struct to represent a response form a signature validation server.
type ValidationResponse struct {
	ValidSignature bool `json:"valid_signature"`
}

// A function that validates the signature on a transaction by requesting its validity from a validationServerURL.
func ValidateSignature(transaction Transaction, validationServerURL string) bool {
	client := resty.New()

	// Put the transaction into transactionRep format: SENDER_KEY -AMOUNT-> RECIPIENT_KEY (TIMESTAMP_SECONDS)
	transactionRepresentation := fmt.Sprintf("%v -%v-> %v (%v)", transaction.Sender, transaction.Amount, transaction.Recipient, transaction.Timestamp)

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{"signature": transaction.Signature, "transactionRepresentation": transactionRepresentation, "publicKey": transaction.Sender}).
		SetResult(ValidationResponse{}).
		Post(validationServerURL)

	if err != nil {
		log.Panic("Signature validation server is returning an error!")
	}

	return resp.Result().(*ValidationResponse).ValidSignature
}
