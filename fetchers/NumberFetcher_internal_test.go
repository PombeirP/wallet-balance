package fetchers

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockHTTPClient struct {
	mock.Mock
	HTTPClient
}

func (m *mockHTTPClient) Get(url string) (resp *http.Response, err error) {
	args := m.Called(url)
	resp, _ = args.Get(0).(*http.Response)
	err = args.Error(1)
	return
}

func TestWebNumberFetcherFetch(t *testing.T) {
	cases := []struct {
		specifiedURL            string
		returnedStatus          string
		returnedStatusCode      int
		returnedGetErrorMessage string
		returnedBody            string
		expectedErrorMessage    string
		expectedValue           float64
	}{
		{"http://test1", "200", 200, "", "190.123", "", 190.123},
		{"http://test2", "200", 200, "", "100", "", 100.},
		{"http://test3", "", 0, "No connection", "", "No connection", 0.},
		{"http://test4", "301", 301, "", "Redirected", "Redirected", 0.},
		{"http://test5", "404", 404, "", "Not found", "Not found", 0.},
		{"http://test6", "500", 500, "", "", "500", 0.},
		{"http://test7", "200", 200, "", "1a0", `strconv.ParseFloat: parsing "1a0": invalid syntax`, 0.},
	}

	for _, testCase := range cases {
		clientMock := new(mockHTTPClient)

		err := error(nil)
		if testCase.returnedGetErrorMessage != "" {
			err = errors.New(testCase.returnedGetErrorMessage)
		}
		clientMock.On("Get", testCase.specifiedURL).Return(&http.Response{Status: testCase.returnedStatus, StatusCode: testCase.returnedStatusCode, Body: ioutil.NopCloser(bytes.NewBuffer([]byte(testCase.returnedBody)))}, err).Once()

		fetcher := NewWebNumberFetcher(clientMock)
		resultsChan := make(chan float64)
		errorsChan := make(chan error)
		go fetcher.Fetch(testCase.specifiedURL, resultsChan, errorsChan)
		select {
		case result := <-resultsChan:
			require.Empty(t, testCase.returnedGetErrorMessage)
			require.Equal(t, testCase.expectedValue, result)
		case err := <-errorsChan:
			require.Error(t, err, testCase.specifiedURL)
			require.Equalf(t, testCase.expectedErrorMessage, err.Error(), `%s: Expected error message to be "%s"`, testCase.specifiedURL, testCase.expectedErrorMessage)
		}

		clientMock.AssertExpectations(t)
	}
}
