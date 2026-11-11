package schemas_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/monetr/monetr/server/schemas"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type iconSearchRequest struct {
	Name string `json:"name"`
}

func parseIconSearch(t *testing.T, body map[string]any) (*iconSearchRequest, error) {
	encoded, err := json.Marshal(body)
	require.NoError(t, err, "must be able to encode the request body")
	return schemas.Parse[iconSearchRequest](
		t.Context(),
		strings.NewReader(string(encoded)),
		nil,
		schemas.IconSearch,
	)
}

func TestIconSearchSchema(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		result, err := parseIconSearch(t, map[string]any{
			"name": "Amazon",
		})
		assert.NoError(t, err, "a simple name should be valid")
		require.NotNil(t, result, "must have parsed the request")
		assert.Equal(t, "Amazon", result.Name)
	})

	t.Run("name is required", func(t *testing.T) {
		result, err := parseIconSearch(t, map[string]any{})
		assert.Error(t, err, "a missing name must be rejected")
		assert.Nil(t, result)
	})

	t.Run("name cannot be blank", func(t *testing.T) {
		result, err := parseIconSearch(t, map[string]any{
			"name": "   ",
		})
		assert.Error(t, err, "a blank name must be rejected")
		assert.Nil(t, result)
	})

	t.Run("name is capped at 300 characters", func(t *testing.T) {
		result, err := parseIconSearch(t, map[string]any{
			"name": strings.Repeat("a", 300),
		})
		assert.NoError(t, err, "a name right at the cap should be valid")
		assert.NotNil(t, result)

		result, err = parseIconSearch(t, map[string]any{
			"name": strings.Repeat("a", 301),
		})
		assert.Error(t, err, "a name over the cap must be rejected")
		assert.Nil(t, result)
	})
}
