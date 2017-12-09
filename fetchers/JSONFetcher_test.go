package fetchers_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/PombeirP/wallet-balance/fetchers"
	"github.com/stretchr/testify/require"
)

func TestEtherscanJSONFetcherFetch(t *testing.T) {
	cases := []struct {
		url            string
		body           string
		status         string
		statusCode     int
		errorMessage   string
		expectedStatus string
		expectedValues []string
	}{
		{"https://api.etherscan.io/api?module=account&action=balancemulti&address=0,1&tag=latest", `{"status":"1","message":"OK","result":[{"account":"0","balance":"190.123"},{"account":"1","balance":"100"}]}`, "200", 200, "", "1", []string{"190.123", "100"}},
		{"https://api.etherscan.io/api?module=account&action=balancemulti&address=1&tag=latest", `{"status":"1","message":"OK","result":[{"account":"1","balance":"100"}]}`, "200", 200, "", "1", []string{"100"}},
		{"https://api.etherscan.io/api?module=account&action=balancemulti", `{"status":"0","message":"NOTOK","result":"Error!"}`, "200", 200, "NOTOK", "", nil},
		{"https://api.etherscan.io/api2", `Server Error in '/' Application.`, "404", 404, "Server Error in '/' Application.", "", nil},
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

		err := error(nil)
		if testCase.errorMessage != "" {
			err = errors.New(testCase.errorMessage)
		}
		clientMock.On("Get", testCase.url).Return(&http.Response{Status: testCase.status, StatusCode: testCase.statusCode, Body: ioutil.NopCloser(bytes.NewBuffer([]byte(testCase.body)))}, err).Once()

		fetcher := fetchers.NewEtherscanJSONFetcher(clientMock)
		responseReadyChan := make(chan bool)
		errorsChan := make(chan error)
		response := &etherscanAccountBalanceResponse{}
		go fetcher.Fetch(testCase.url, response, responseReadyChan, errorsChan)
		select {
		case _ = <-responseReadyChan:
			require.Empty(t, testCase.errorMessage, testCase.url)
			require.Equal(t, testCase.expectedStatus, response.Status, testCase.url)
			for idx := range response.Result {
				require.Equal(t, testCase.expectedValues[idx], response.Result[idx].Balance, testCase.url)
			}
		case err := <-errorsChan:
			require.Error(t, err, testCase.url)
			require.Equalf(t, testCase.errorMessage, err.Error(), `%s: Expected error message to be "%s"`, testCase.url, testCase.errorMessage)
		}

		clientMock.AssertExpectations(t)
	}
}
