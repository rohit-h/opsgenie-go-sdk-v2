package client

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	BaseURL     = "https://api.opsgenie.com"
	Endpoint    = "v2/alerts"
	EndpointURL = BaseURL + "/" + Endpoint
	BadEndpoint = ":"
)

type testRequest struct {
	MandatoryField string
	ExtraField     string
}

func (tr testRequest) Validate() (bool, error) {
	if tr.MandatoryField == "" {
		return false, errors.New("mandatory field cannot be empty")
	}
	return true, nil
}

func (tr testRequest) Endpoint() string {
	return "/an-enpoint"
}

func (tr testRequest) Method() string {
	return "POST"
}

type testResult struct {
	ResponseMeta
	Data string
}

func TestExec(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{
    		"Data": "processed"}`)
	}))
	defer ts.Close()

	ogClient, err := NewOpsGenieClient(Config{
		ApiKey: "apiKey",
	})

	request := testRequest{MandatoryField: "afield", ExtraField: "extra"}
	result := &testResult{}
	ogClient.Config.apiUrl = ts.URL
	err = ogClient.Exec(nil, request, result)
	assert.Equal(t, result.Data, "processed")
	if err != nil {
		t.Fail()
	}
}

func TestExecWhenRequestIsNotValid(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{
    		"Data": "processed"}`)
	}))
	defer ts.Close()

	ogClient, err := NewOpsGenieClient(Config{
		ApiKey: "apiKey",
	})
	ogClient.Config.apiUrl = ts.URL

	request := testRequest{ExtraField: "extra"}
	result := &testResult{}

	err = ogClient.Exec(nil, request, result)
	assert.Equal(t, err.Error(), "mandatory field cannot be empty")
}

func TestExecWhenApiReturns422(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintln(w, `{
    "message": "Request body is not processable. Please check the errors.",
    "errors": {
        "recipients#type": "Invalid recipient type 'bb'"
    },
    "took": 0.083,
    "requestId": "Id"
}`)
	}))
	defer ts.Close()

	ogClient, err := NewOpsGenieClient(Config{
		ApiKey: "apiKey",
	})
	ogClient.Config.apiUrl = ts.URL
	request := testRequest{MandatoryField: "afield", ExtraField: "extra"}
	result := &testResult{}

	err = ogClient.Exec(nil, request, result)
	fmt.Println(err.Error())
	assert.Contains(t, err.Error(), "422")
	assert.Contains(t, err.Error(), "Invalid recipient")

}

func TestExecWhenApiReturns5XX(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, `{
    "message": "Internal Server Error",
    "took": 0.083,
    "requestId": "6c20ec4e-076a-4422-8d65-7b8ca92067ab"
}`)
	}))
	defer ts.Close()

	ogClient, err := NewOpsGenieClient(Config{
		ApiKey:     "apiKey",
		RetryCount: 1,
	})
	ogClient.Config.apiUrl = ts.URL
	request := testRequest{MandatoryField: "afield", ExtraField: "extra"}
	result := &testResult{}

	err = ogClient.Exec(nil, request, result)
	fmt.Println(err.Error())
	assert.Contains(t, err.Error(), "Internal Server Error")
	assert.Contains(t, err.Error(), "500")

}