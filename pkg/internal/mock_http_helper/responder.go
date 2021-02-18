package mock_http_helper

import (
	"bytes"
	"encoding/json"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"testing"
)

func NewHttpMockJsonResponder(t *testing.T, method, path string, handler func(t *testing.T, request *http.Request) (interface{}, int)) {
	httpmock.RegisterResponder(method, path, func(request *http.Request) (*http.Response, error) {
		result, status := handler(t, request)
		body := bytes.NewBuffer(nil)
		require.NoError(t, json.NewEncoder(body).Encode(result), "must encode response body")
		response := &http.Response{
			StatusCode:    status,
			Proto:         "HTTP/1.0",
			ProtoMajor:    1,
			ProtoMinor:    0,
			Body:          ioutil.NopCloser(body),
			ContentLength: int64(body.Len()),
			Request:       request,
		}
		return response, nil
	})
}
