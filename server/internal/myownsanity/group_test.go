package myownsanity_test

import (
	"testing"

	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/stretchr/testify/assert"
)

type groupItem struct {
	Key   string
	Value int
}

var groupItems = []groupItem{
	{Key: "b", Value: 1},
	{Key: "a", Value: 2},
	{Key: "b", Value: 3},
	{Key: "c", Value: 4},
	{Key: "a", Value: 5},
}

func TestGroupByV(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		result := myownsanity.GroupByV(
			groupItems,
			func(item groupItem) string { return item.Key },
			func(item groupItem) int { return item.Value },
		)
		assert.Equal(t, []myownsanity.Group[string, int]{
			{Key: "b", Items: []int{1, 3}},
			{Key: "a", Items: []int{2, 5}},
			{Key: "c", Items: []int{4}},
		}, result, "groups should be in the order their keys were first seen")
	})

	t.Run("empty", func(t *testing.T) {
		result := myownsanity.GroupByV(
			[]groupItem{},
			func(item groupItem) string { return item.Key },
			func(item groupItem) int { return item.Value },
		)
		assert.Empty(t, result)
	})
}

func TestGroupByMap(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		result := myownsanity.GroupByMap(
			groupItems,
			func(item groupItem) string { return item.Key },
		)
		assert.Equal(t, map[string][]groupItem{
			"a": {{Key: "a", Value: 2}, {Key: "a", Value: 5}},
			"b": {{Key: "b", Value: 1}, {Key: "b", Value: 3}},
			"c": {{Key: "c", Value: 4}},
		}, result)
	})

	t.Run("empty", func(t *testing.T) {
		result := myownsanity.GroupByMap(
			[]groupItem{},
			func(item groupItem) string { return item.Key },
		)
		assert.Empty(t, result)
	})
}

func TestGroupByMapV(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		result := myownsanity.GroupByMapV(
			groupItems,
			func(item groupItem) string { return item.Key },
			func(item groupItem) int { return item.Value },
		)
		assert.Equal(t, map[string][]int{
			"a": {2, 5},
			"b": {1, 3},
			"c": {4},
		}, result)
	})

	t.Run("empty", func(t *testing.T) {
		result := myownsanity.GroupByMapV(
			[]groupItem{},
			func(item groupItem) string { return item.Key },
			func(item groupItem) int { return item.Value },
		)
		assert.Empty(t, result)
	})
}
