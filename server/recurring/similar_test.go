package recurring

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimilarTransactions_TFIDF_DBSCAN(t *testing.T) {
	t.Run("monetr mercury dataset", func(t *testing.T) {
		data := GetFixtures(t, "monetr_sample_data_1.json")

		detector := NewSimilarTransactions_TFIDF_DBSCAN()

		for i := range data {
			detector.AddTransaction(&data[i])
		}

		groups := detector.DetectSimilarTransactions(context.Background())
		assert.NotEmpty(t, groups, "must return an array of groups of similar transactions")
		for _, group := range groups {
			assert.NotEmpty(t, group.Members, "a groups matches should not be empty!")
			assert.NotEmpty(t, group.Name, "a groups name should not be empty!")
			assert.NotEmpty(t, group.Signature, "a groups signature should not be empty!")
		}
		// TODO, add specific assertions here about what the groups are.
		j, _ := json.MarshalIndent(groups, "", "  ")
		fmt.Println(string(j))
	})

	t.Run("amazon dataset", func(t *testing.T) {
		data := GetFixtures(t, "amazon_sample_data_1.json")

		detector := NewSimilarTransactions_TFIDF_DBSCAN()

		for i := range data {
			detector.AddTransaction(&data[i])
		}

		groups := detector.DetectSimilarTransactions(context.Background())
		assert.NotEmpty(t, groups, "must return an array of groups of similar transactions")
		// Since the amazon dataset only has amazon transactions there should only be a single group.
		assert.Len(t, groups, 1, "should have only a single group")
	})
}
