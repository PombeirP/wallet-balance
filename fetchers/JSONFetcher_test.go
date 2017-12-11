package fetchers_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/PombeirP/wallet-balance/fetchers"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockHTTPClient struct {
	mock.Mock
	fetchers.HTTPClient
}

func (m *mockHTTPClient) Get(url string) (resp *http.Response, err error) {
	args := m.Called(url)
	resp, _ = args.Get(0).(*http.Response)
	err = args.Error(1)
	return
}

func TestEtherscanJSONFetcherFetch(t *testing.T) {
	cases := []struct {
		specifiedUrl            string
		returnedBody            string
		returnedStatus          string
		returnedStatusCode      int
		returnedGetErrorMessage string
		expectedErrorMessage    string
		expectedStatus          string
		expectedValues          []string
	}{
		{"https://api.etherscan.io/api?module=account&action=balancemulti&address=0,1&tag=latest", `{"status":"1","message":"OK","result":[{"account":"0","balance":"190.123"},{"account":"1","balance":"100"}]}`, "200", 200, "", "", "1", []string{"190.123", "100"}},
		{"https://api.etherscan.io/api?module=account&action=balancemulti&address=1&tag=latest", `{"status":"1","message":"OK","result":[{"account":"1","balance":"100"}]}`, "200", 200, "", "", "1", []string{"100"}},
		{"https://api.etherscan.io/api?module=account&action=balancemulti", `{"status":"0","message":"NOTOK","result":"Error!"}`, "200", 200, "", "NOTOK", "", nil},
		{"https://somesite", `Hello!`, "200", 200, "", "invalid character 'H' looking for beginning of value", "", nil},
		{"https://api.etherscan.io/api2", `Server Error in '/' Application.`, "404", 404, "Server Error in '/' Application.", "Server Error in '/' Application.", "", nil},
	}

	type etherscanResponseHeader struct {
		Status  string `json:"status,omitempty"`
		Message string `json:"message,omitempty"`
	}

	type etherscanAccountBalanceResult struct {
		Account string `json:"account,omitempty"`
		Balance string `json:"balance,omitempty"`
	}

	type etherscanAccountBalanceResponse struct {
		etherscanResponseHeader
		Result []*etherscanAccountBalanceResult `json:"result,omitempty"`
	}

	for _, testCase := range cases {
		clientMock := new(mockHTTPClient)

		returnedErr := error(nil)
		if testCase.returnedGetErrorMessage != "" {
			returnedErr = errors.New(testCase.returnedGetErrorMessage)
		}
		clientMock.On("Get", testCase.specifiedUrl).Return(&http.Response{Status: testCase.returnedStatus, StatusCode: testCase.returnedStatusCode, Body: ioutil.NopCloser(bytes.NewBuffer([]byte(testCase.returnedBody)))}, returnedErr).Once()

		fetcher := fetchers.NewEtherscanJSONFetcher(clientMock)
		response := &etherscanAccountBalanceResponse{}
		err := fetcher.Fetch(testCase.specifiedUrl, response)
		if err == nil {
			require.Empty(t, testCase.returnedGetErrorMessage, testCase.specifiedUrl)
			require.Empty(t, testCase.expectedErrorMessage, testCase.specifiedUrl)
			require.Equal(t, testCase.expectedStatus, response.Status, testCase.specifiedUrl)
			for idx := range response.Result {
				require.Equal(t, testCase.expectedValues[idx], response.Result[idx].Balance, testCase.specifiedUrl)
			}
		} else {
			require.Error(t, err, testCase.specifiedUrl)
			require.Equalf(t, testCase.expectedErrorMessage, err.Error(), `%s: Expected error message to be "%s"`, testCase.specifiedUrl, testCase.expectedErrorMessage)
		}

		clientMock.AssertExpectations(t)
	}
}
