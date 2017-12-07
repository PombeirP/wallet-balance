package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadConfigFromJSON(t *testing.T) {
	config := loadConfigFromJSON([]byte("[{\"symbol\": \"BTC\", \"addresses\": [\"a\"]},{\"symbol\": \"DASH\",\"addresses\": [\"b\",\"c\"],\"api_key\": \"apikey1\"},{\"symbol\": \"ETH\",\"addresses\": [\"d\"],\"api_key\": \"apikey2\"}]"))

	require.NotNil(t, config)
	require.Len(t, config, 3, "expected 3 crypto-currencies")

	require.Equal(t, btc, config[0].Symbol)
	require.Equal(t, dash, config[1].Symbol)
	require.Equal(t, eth, config[2].Symbol)

	require.Empty(t, config[0].APIKey)
	require.Equal(t, "apikey1", config[1].APIKey)
	require.Equal(t, "apikey2", config[2].APIKey)

	require.EqualValues(t, []string{"a"}, config[0].Addresses)
	require.EqualValues(t, []string{"b", "c"}, config[1].Addresses)
	require.EqualValues(t, []string{"d"}, config[2].Addresses)

	for idx, c := range config {
		require.Emptyf(t, c.Balance, "Item #%d should have zero balance", idx)
		require.Emptyf(t, c.UsdExchangeRate, "Item #%d should have zero UsdExchangeRate", idx)
		require.Emptyf(t, c.Error, "Item #%d should have nil error", idx)
	}
}
