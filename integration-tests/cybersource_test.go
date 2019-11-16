package test

import (
	"github.com/AlekSi/pointer"
	"github.com/BoltApp/sleet"
	"github.com/BoltApp/sleet/gateways/cybersource"
	"testing"
)

func TestCybersource(t *testing.T) {
	client := cybersource.NewClient(getEnv("CYBERSOURCE_ACCOUNT"), getEnv("CYBERSOURCE_API_KEY"), getEnv("CYBERSOURCE_SHARED_SECRET"))
	options := make(map[string]interface{})
	options["email"] = "test@bolt.com"
	authRequest := baseAuthorizationRequest()
	authRequest.CreditCard = &sleet.CreditCard{
		FirstName:       "Bolt",
		LastName:        "Checkout",
		Number:          "4111111111111111",
		ExpirationMonth: 8,
		ExpirationYear:  2024,
		CVV:             "000",
	}
	authRequest.BillingAddress = &sleet.BillingAddress{
		StreetAddress1: pointer.ToString("77 Geary St"),
		StreetAddress2: pointer.ToString("Floor 4"),
		Locality:       pointer.ToString("San Francisco"),
		RegionCode:     pointer.ToString("CA"),
		PostalCode:     pointer.ToString("94108"),
		CountryCode:    pointer.ToString("US"),
	}
	authRequest.Options = options
	resp, err := client.Authorize(authRequest)
	if err != nil {
		t.Errorf("Expected no error: received: %s", err)
	}
	if resp.Success != true {
		t.Errorf("Expected Success: received: %s", resp.ErrorCode)
	}

	capResp, err := client.Capture(&sleet.CaptureRequest{
		Amount:               authRequest.Amount,
		TransactionReference: resp.TransactionReference,
	})
	if err != nil {
		t.Errorf("Expected no error: received: %s", err)
	}
	if capResp.ErrorCode != nil {
		t.Errorf("Expected No Error Code: received: %s", *capResp.ErrorCode)
	}

	refundResp, err := client.Refund(&sleet.RefundRequest{
		Amount:               authRequest.Amount,
		TransactionReference: resp.TransactionReference,
	})
	if err != nil {
		t.Errorf("Expected no error: received: %s", err)
	}
	if refundResp.ErrorCode != nil {
		t.Errorf("Expected No Error Code: received: %s", *refundResp.ErrorCode)
	}
}
