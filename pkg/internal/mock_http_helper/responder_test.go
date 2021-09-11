package mock_http_helper

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestNewHttpMockJsonResponder(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := "https://monetr.test/thing"
	NewHttpMockJsonResponder(t, "GET", url, func(t *testing.T, request *http.Request) (interface{}, int) {
		return map[string]interface{}{
			"value": 123,
		}, http.StatusOK
	}, func(t *testing.T, request *http.Request, response interface{}, status int) map[string][]string {
		return nil
	})

	response, err := http.Get(url)
	assert.NoError(t, err, "http get request must succeed")
	assert.Equal(t, http.StatusOK, response.StatusCode, "status code must be 200")
	assert.Equal(t, "application/json", response.Header.Get("Content-Type"), "content type should always be json")

	body, err := ioutil.ReadAll(response.Body)
	assert.NoError(t, err, "must be able to read the response body")
	assert.NotEmpty(t, body, "body must not be empty")

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	assert.NoError(t, err, "must be able to unmarshal response")

	assert.EqualValues(t, 123, result["value"], "value must match")
	assert.Len(t, result, 1, "must only have one key")

}
