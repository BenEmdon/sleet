package orbital

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"

	"github.com/BoltApp/sleet"
)

type Credentials struct {
	username   string
	password   string
	merchantID int
}

type OrbitalClient struct {
	host            string
	credentials     Credentials
	clientRequestID string
	httpClient      *http.Client
}

func (client *OrbitalClient) Authorize(request *sleet.AuthorizationRequest) (*sleet.AuthorizationResponse, error) {

	orbitalAuthRequest := buildAuthRequest(request)
	orbitalAuthRequest.Body.OrbitalConnectionUsername = client.credentials.username
	orbitalAuthRequest.Body.OrbitalConnectionPassword = client.credentials.password
	orbitalAuthRequest.Body.MerchantID = client.credentials.merchantID
	orbitalAuthRequest.Body.OrderID = *request.ClientTransactionReference

	orbitalAuthRequest.Body.XMLName = xml.Name{Local: RequestTypeAuth}

	orbitalResponse, err := client.sendRequest(orbitalAuthRequest)
	if err != nil {
		return nil, err
	}

	if orbitalResponse.Body.ProcStatus != 0 {
		response := sleet.AuthorizationResponse{Success: false, ErrorCode: string(orbitalResponse.Body.ProcStatus)}
		return &response, nil
	}

	return &sleet.AuthorizationResponse{
		Success:              true,
		TransactionReference: string(orbitalResponse.Body.TxRefNum),
		AvsResult:            translateAvs(orbitalResponse.Body.AVSRespCode),
		CvvResult:            translateCvv(orbitalResponse.Body.CVV2RespCode),
		Response:             string(orbitalResponse.Body.ApprovalStatus),
		AvsResultRaw:         string(orbitalResponse.Body.AVSRespCode),
		CvvResultRaw:         string(orbitalResponse.Body.CVV2RespCode),
	}, nil
}

func (client *OrbitalClient) Capture(request *sleet.CaptureRequest) (*sleet.CaptureResponse, error) {

	orbitalAuthRequest := buildCaptureRequest(request)
	orbitalAuthRequest.Body.OrbitalConnectionUsername = client.credentials.username
	orbitalAuthRequest.Body.OrbitalConnectionPassword = client.credentials.password
	orbitalAuthRequest.Body.MerchantID = client.credentials.merchantID
	orbitalAuthRequest.Body.OrderID = *request.ClientTransactionReference

	orbitalAuthRequest.Body.XMLName = xml.Name{Local: RequestTypeCapture}

	orbitalResponse, err := client.sendRequest(orbitalAuthRequest)
	if err != nil {
		return nil, err
	}

	if orbitalResponse.Body.ProcStatus != 0 {
		errorCode := string(orbitalResponse.Body.ProcStatus)
		return &sleet.CaptureResponse{
			Success:   false,
			ErrorCode: &errorCode,
		}, nil
	}

	return &sleet.CaptureResponse{
		Success:              true,
		TransactionReference: string(orbitalResponse.Body.TxRefNum),
	}, nil
}

func (client *OrbitalClient) Void(request *sleet.VoidRequest) (*sleet.VoidResponse, error) {

	orbitalAuthRequest := buildVoidRequest(request)
	orbitalAuthRequest.Body.OrbitalConnectionUsername = client.credentials.username
	orbitalAuthRequest.Body.OrbitalConnectionPassword = client.credentials.password
	orbitalAuthRequest.Body.MerchantID = client.credentials.merchantID
	orbitalAuthRequest.Body.OrderID = *request.ClientTransactionReference

	orbitalAuthRequest.Body.XMLName = xml.Name{Local: RequestTypeVoid}

	orbitalResponse, err := client.sendRequest(orbitalAuthRequest)
	if err != nil {
		return nil, err
	}

	if orbitalResponse.Body.ProcStatus != 0 {
		errorCode := string(orbitalResponse.Body.ProcStatus)
		return &sleet.VoidResponse{
			Success:   false,
			ErrorCode: &errorCode,
		}, nil
	}

	return &sleet.VoidResponse{
		Success:              true,
		TransactionReference: string(orbitalResponse.Body.TxRefNum),
	}, nil
}

func (client *OrbitalClient) Refund(request *sleet.RefundRequest) (*sleet.RefundResponse, error) {

	orbitalAuthRequest := buildRefundRequest(request)
	orbitalAuthRequest.Body.OrbitalConnectionUsername = client.credentials.username
	orbitalAuthRequest.Body.OrbitalConnectionPassword = client.credentials.password
	orbitalAuthRequest.Body.MerchantID = client.credentials.merchantID
	orbitalAuthRequest.Body.OrderID = *request.ClientTransactionReference

	orbitalAuthRequest.Body.XMLName = xml.Name{Local: RequestTypeRefund}

	orbitalResponse, err := client.sendRequest(orbitalAuthRequest)
	if err != nil {
		return nil, err
	}

	if orbitalResponse.Body.ProcStatus != 0 {
		errorCode := string(orbitalResponse.Body.ProcStatus)
		return &sleet.RefundResponse{
			Success:   false,
			ErrorCode: &errorCode,
		}, nil
	}

	return &sleet.RefundResponse{
		Success:              true,
		TransactionReference: string(orbitalResponse.Body.TxRefNum),
	}, nil
}
func (client *OrbitalClient) sendRequest(data Request) (*Response, error) {

	bodyXML, err := xml.Marshal(data)
	if err != nil {
		return nil, err
	}

	reader := bytes.NewReader(bodyXML)
	request, err := http.NewRequest(http.MethodPost, client.host, reader)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", "application/xml; charset=utf-8")

	resp, err := client.httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var orbitalResponse Response

	err = json.Unmarshal(body, &orbitalResponse)
	if err != nil {
		return nil, err
	}

	return &orbitalResponse, nil
}
