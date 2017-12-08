package fetchers

import "net/http"

// HTTPClient is a facade for http.Client
type HTTPClient interface {
	Get(url string) (resp *http.Response, err error)
}
