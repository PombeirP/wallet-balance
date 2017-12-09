package fetchers_test

import (
	"net/http"

	"github.com/PombeirP/wallet-balance/fetchers"
	"github.com/stretchr/testify/mock"
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
