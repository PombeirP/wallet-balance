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

func TestWebNumberFetcherFetch(t *testing.T) {
	cases := []struct {
		url           string
		status        string
		statusCode    int
		errorMessage  string
		body          string
		expectedValue float64
	}{
		{"http://test1", "200", 200, "", "190.123", 190.123},
		{"http://test2", "200", 200, "", "100", 100.},
		{"http://test3", "301", 301, "Redirected", "", 0.},
		{"http://test4", "404", 404, "Not found", "", 0.},
		{"http://test5", "500", 500, "Internal server error", "Internal server error", 0.},
	}

	for _, testCase := range cases {
		clientMock := new(mockHTTPClient)

		err := error(nil)
		if testCase.errorMessage != "" {
			err = errors.New(testCase.errorMessage)
		}
		clientMock.On("Get", testCase.url).Return(&http.Response{Status: testCase.status, StatusCode: testCase.statusCode, Body: ioutil.NopCloser(bytes.NewBuffer([]byte(testCase.body)))}, err).Once()

		fetcher := fetchers.NewWebNumberFetcher(clientMock)
		resultsChan := make(chan float64)
		errorsChan := make(chan error)
		go fetcher.Fetch(testCase.url, resultsChan, errorsChan)
		select {
		case result := <-resultsChan:
			require.Empty(t, testCase.errorMessage)
			require.Equal(t, testCase.expectedValue, result)
		case err := <-errorsChan:
			require.Error(t, err, testCase.url)
			require.Equalf(t, testCase.errorMessage, err.Error(), `%s: Expected error message to be "%s"`, testCase.url, testCase.errorMessage)
		}

		clientMock.AssertExpectations(t)
	}
}
