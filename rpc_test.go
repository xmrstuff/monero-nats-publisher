package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func makeServer(t *testing.T, expectedURL string, expectedMethod string, expectedTXID string, respStatus int, respBody string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Make sure we're using the expected URL and HTTP method
		assert.Equal(t, expectedURL, req.URL.String())
		assert.Equal(t, expectedMethod, req.Method)

		// Make sure we're passing the expected request body
		// reqData := RpcRequest{}
		// err := json.NewDecoder(req.Body).Decode(&reqData)
		// assert.Nil(t, err)
		// assert.Equal(t, expectedTXID, reqData.Params.TXID)
		assert.Equal(t, "application/json", req.Header["Content-Type"][0])

		rw.WriteHeader(respStatus)
		rw.Header().Set("Content-Type", "application/json")
		rw.Write([]byte(respBody))
	}))
}
