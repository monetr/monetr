package cache_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/cache"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestLruCache_Set(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		clock := clock.NewMock()
		memoryCache := cache.NewLRUCache(testutils.GetLog(t), clock, config.MemoryCache{
			MaxKeys: 128,
		})

		err := memoryCache.Set(t.Context(), "test:data", TestValue)
		assert.NoError(t, err, "should successfully set value")

		value, err := memoryCache.Get(t.Context(), "test:data")
		assert.NoError(t, err, "should be able to read the value back")
		assert.Equal(t, TestValue, value, "should get back what was set")
	})

	t.Run("overwrite", func(t *testing.T) {
		clock := clock.NewMock()
		memoryCache := cache.NewLRUCache(testutils.GetLog(t), clock, config.MemoryCache{
			MaxKeys: 128,
		})

		assert.NoError(t, memoryCache.Set(t.Context(), "test:data", []byte("first")), "should set first value")
		assert.NoError(t, memoryCache.Set(t.Context(), "test:data", []byte("second")), "should set second value")

		value, err := memoryCache.Get(t.Context(), "test:data")
		assert.NoError(t, err, "should be able to read the value back")
		assert.Equal(t, []byte("second"), value, "second write should win")
	})

	t.Run("never expires", func(t *testing.T) {
		clock := clock.NewMock()
		memoryCache := cache.NewLRUCache(testutils.GetLog(t), clock, config.MemoryCache{
			MaxKeys: 128,
		})

		assert.NoError(t, memoryCache.Set(t.Context(), "test:data", TestValue), "should set value")

		clock.Add(24 * time.Hour)

		value, err := memoryCache.Get(t.Context(), "test:data")
		assert.NoError(t, err, "should be able to read the value back")
		assert.Equal(t, TestValue, value, "value without a ttl should still be there")
	})

	t.Run("evicts oldest", func(t *testing.T) {
		clock := clock.NewMock()
		memoryCache := cache.NewLRUCache(testutils.GetLog(t), clock, config.MemoryCache{
			MaxKeys: 2,
		})

		assert.NoError(t, memoryCache.Set(t.Context(), "test:a", []byte("a")), "must set a")
		assert.NoError(t, memoryCache.Set(t.Context(), "test:b", []byte("b")), "must set b")
		assert.NoError(t, memoryCache.Set(t.Context(), "test:c", []byte("c")), "must set c")

		value, err := memoryCache.Get(t.Context(), "test:a")
		assert.NoError(t, err, "evicted key should not error")
		assert.Nil(t, value, "a should have been evicted")

		value, err = memoryCache.Get(t.Context(), "test:b")
		assert.NoError(t, err, "should be able to read b")
		assert.Equal(t, []byte("b"), value, "b should still be there")

		value, err = memoryCache.Get(t.Context(), "test:c")
		assert.NoError(t, err, "should be able to read c")
		assert.Equal(t, []byte("c"), value, "c should still be there")
	})

	t.Run("recently read survives", func(t *testing.T) {
		clock := clock.NewMock()
		memoryCache := cache.NewLRUCache(testutils.GetLog(t), clock, config.MemoryCache{
			MaxKeys: 2,
		})

		assert.NoError(t, memoryCache.Set(t.Context(), "test:a", []byte("a")), "must set a")
		assert.NoError(t, memoryCache.Set(t.Context(), "test:b", []byte("b")), "must set b")

		_, err := memoryCache.Get(t.Context(), "test:a")
		assert.NoError(t, err, "should be able to read a")

		assert.NoError(t, memoryCache.Set(t.Context(), "test:c", []byte("c")), "must set c")

		value, err := memoryCache.Get(t.Context(), "test:a")
		assert.NoError(t, err, "should be able to read a")
		assert.Equal(t, []byte("a"), value, "a should still be there")

		value, err = memoryCache.Get(t.Context(), "test:b")
		assert.NoError(t, err, "evicted key should not error")
		assert.Nil(t, value, "b should have been evicted")
	})

	t.Run("blank key", func(t *testing.T) {
		clock := clock.NewMock()
		memoryCache := cache.NewLRUCache(testutils.GetLog(t), clock, config.MemoryCache{
			MaxKeys: 128,
		})

		err := memoryCache.Set(t.Context(), "", TestValue)
		assert.EqualError(t, err, "key is blank", "should be blank key error")
	})
}

func TestLruCache_SetTTL(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		clock := clock.NewMock()
		memoryCache := cache.NewLRUCache(testutils.GetLog(t), clock, config.MemoryCache{
			MaxKeys: 128,
		})

		err := memoryCache.SetTTL(t.Context(), "test:data", TestValue, time.Minute)
		assert.NoError(t, err, "should successfully set value")

		clock.Add(30 * time.Second)

		value, err := memoryCache.Get(t.Context(), "test:data")
		assert.NoError(t, err, "should be able to read the value back")
		assert.Equal(t, TestValue, value, "value should still be there before it expires")
	})

	t.Run("expired", func(t *testing.T) {
		clock := clock.NewMock()
		memoryCache := cache.NewLRUCache(testutils.GetLog(t), clock, config.MemoryCache{
			MaxKeys: 128,
		})

		err := memoryCache.SetTTL(t.Context(), "test:data", TestValue, time.Minute)
		assert.NoError(t, err, "should successfully set value")

		clock.Add(2 * time.Minute)

		value, err := memoryCache.Get(t.Context(), "test:data")
		assert.NoError(t, err, "expired key should not error")
		assert.Nil(t, value, "value should be gone after it expires")
	})

	t.Run("blank key", func(t *testing.T) {
		clock := clock.NewMock()
		memoryCache := cache.NewLRUCache(testutils.GetLog(t), clock, config.MemoryCache{
			MaxKeys: 128,
		})

		err := memoryCache.SetTTL(t.Context(), "", TestValue, time.Minute)
		assert.EqualError(t, err, "key is blank", "should be blank key error")
	})
}

func TestLruCache_Get(t *testing.T) {
	t.Run("miss", func(t *testing.T) {
		// This test is here to prove a fix for a nil pointer panic on a cache miss.
		clock := clock.NewMock()
		memoryCache := cache.NewLRUCache(testutils.GetLog(t), clock, config.MemoryCache{
			MaxKeys: 128,
		})

		value, err := memoryCache.Get(t.Context(), "test:missing")
		assert.NoError(t, err, "miss should not error")
		assert.Nil(t, value, "miss should not return anything")
	})

	t.Run("blank key", func(t *testing.T) {
		clock := clock.NewMock()
		memoryCache := cache.NewLRUCache(testutils.GetLog(t), clock, config.MemoryCache{
			MaxKeys: 128,
		})

		_, err := memoryCache.Get(t.Context(), "")
		assert.EqualError(t, err, "key is blank", "should be blank key error")
	})
}

func TestLruCache_SetEz(t *testing.T) {
	type thing struct {
		Name  string
		Count int
	}

	t.Run("simple", func(t *testing.T) {
		clock := clock.NewMock()
		memoryCache := cache.NewLRUCache(testutils.GetLog(t), clock, config.MemoryCache{
			MaxKeys: 128,
		})

		expected := thing{
			Name:  "monetr",
			Count: 5,
		}
		err := memoryCache.SetEz(t.Context(), "test:thing", expected)
		assert.NoError(t, err, "should successfully set object")

		var result thing
		err = memoryCache.GetEz(t.Context(), "test:thing", &result)
		assert.NoError(t, err, "should be able to read the object back")
		assert.Equal(t, expected, result, "should get back the same object")
	})

	t.Run("with ttl", func(t *testing.T) {
		clock := clock.NewMock()
		memoryCache := cache.NewLRUCache(testutils.GetLog(t), clock, config.MemoryCache{
			MaxKeys: 128,
		})

		err := memoryCache.SetEzTTL(t.Context(), "test:thing", thing{
			Name:  "monetr",
			Count: 5,
		}, time.Minute)
		assert.NoError(t, err, "should successfully set object")

		clock.Add(2 * time.Minute)

		var result thing
		err = memoryCache.GetEz(t.Context(), "test:thing", &result)
		assert.NoError(t, err, "expired key should not error")
		assert.Zero(t, result, "output should be untouched after the key expires")
	})

	t.Run("blank key", func(t *testing.T) {
		clock := clock.NewMock()
		memoryCache := cache.NewLRUCache(testutils.GetLog(t), clock, config.MemoryCache{
			MaxKeys: 128,
		})

		err := memoryCache.SetEz(t.Context(), "", thing{
			Name: "monetr",
		})
		assert.EqualError(t, err, "key is blank", "should be blank key error")
	})
}

func TestLruCache_GetEz(t *testing.T) {
	t.Run("miss", func(t *testing.T) {
		clock := clock.NewMock()
		memoryCache := cache.NewLRUCache(testutils.GetLog(t), clock, config.MemoryCache{
			MaxKeys: 128,
		})

		result := map[string]any{
			"untouched": true,
		}
		err := memoryCache.GetEz(t.Context(), "test:missing", &result)
		assert.NoError(t, err, "miss should not error")
		assert.Equal(t, map[string]any{
			"untouched": true,
		}, result, "output should be untouched on a miss")
	})

	t.Run("not msgpack", func(t *testing.T) {
		clock := clock.NewMock()
		memoryCache := cache.NewLRUCache(testutils.GetLog(t), clock, config.MemoryCache{
			MaxKeys: 128,
		})

		// 0xc1 is the one byte msgpack never uses for anything.
		assert.NoError(t, memoryCache.Set(t.Context(), "test:data", []byte{0xc1}), "must set raw value")

		var result map[string]any
		err := memoryCache.GetEz(t.Context(), "test:data", &result)
		assert.EqualError(t, err, "failed to unmarshal from cache: msgpack: unexpected code=c1 decoding map length", "should fail to unmarshal")
	})

	t.Run("blank key", func(t *testing.T) {
		clock := clock.NewMock()
		memoryCache := cache.NewLRUCache(testutils.GetLog(t), clock, config.MemoryCache{
			MaxKeys: 128,
		})

		var result map[string]any
		err := memoryCache.GetEz(t.Context(), "", &result)
		assert.EqualError(t, err, "key is blank", "should be blank key error")
	})
}

func TestLruCache_Delete(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		clock := clock.NewMock()
		memoryCache := cache.NewLRUCache(testutils.GetLog(t), clock, config.MemoryCache{
			MaxKeys: 128,
		})

		assert.NoError(t, memoryCache.Set(t.Context(), "test:data", TestValue), "must set value")

		err := memoryCache.Delete(t.Context(), "test:data")
		assert.NoError(t, err, "should successfully delete value")

		value, err := memoryCache.Get(t.Context(), "test:data")
		assert.NoError(t, err, "deleted key should not error")
		assert.Nil(t, value, "value should be gone after delete")
	})

	t.Run("missing key", func(t *testing.T) {
		clock := clock.NewMock()
		memoryCache := cache.NewLRUCache(testutils.GetLog(t), clock, config.MemoryCache{
			MaxKeys: 128,
		})

		err := memoryCache.Delete(t.Context(), "test:missing")
		assert.NoError(t, err, "deleting a key that is not there should not error")
	})

	t.Run("blank key", func(t *testing.T) {
		clock := clock.NewMock()
		memoryCache := cache.NewLRUCache(testutils.GetLog(t), clock, config.MemoryCache{
			MaxKeys: 128,
		})

		err := memoryCache.Delete(t.Context(), "")
		assert.EqualError(t, err, "key is blank", "should be blank key error")
	})
}

func TestLruCache_CompareAndSwap(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		clock := clock.NewMock()
		memoryCache := cache.NewLRUCache(testutils.GetLog(t), clock, config.MemoryCache{
			MaxKeys: 128,
		})

		assert.NoError(t, memoryCache.SetTTL(t.Context(), "test:cas", []byte("unused"), time.Minute), "must seed key")

		swapped, err := memoryCache.CompareAndSwap(t.Context(), "test:cas", []byte("unused"), []byte("consumed"), time.Minute)
		assert.NoError(t, err, "should be able to compare and swap")
		assert.True(t, swapped, "should swap because the current value matched expected")

		value, err := memoryCache.Get(t.Context(), "test:cas")
		assert.NoError(t, err, "should be able to read back the swapped value")
		assert.Equal(t, []byte("consumed"), value, "value must have been swapped")
	})

	t.Run("mismatch", func(t *testing.T) {
		clock := clock.NewMock()
		memoryCache := cache.NewLRUCache(testutils.GetLog(t), clock, config.MemoryCache{
			MaxKeys: 128,
		})

		assert.NoError(t, memoryCache.SetTTL(t.Context(), "test:cas", []byte("consumed"), time.Minute), "must seed key")

		swapped, err := memoryCache.CompareAndSwap(t.Context(), "test:cas", []byte("unused"), []byte("consumed"), time.Minute)
		assert.NoError(t, err, "comparison itself should not error")
		assert.False(t, swapped, "must not swap when the current value does not match expected")
	})

	t.Run("missing key", func(t *testing.T) {
		clock := clock.NewMock()
		memoryCache := cache.NewLRUCache(testutils.GetLog(t), clock, config.MemoryCache{
			MaxKeys: 128,
		})

		swapped, err := memoryCache.CompareAndSwap(t.Context(), "test:cas", []byte("unused"), []byte("consumed"), time.Minute)
		assert.NoError(t, err, "comparison itself should not error")
		assert.False(t, swapped, "must not swap when the key does not exist")

		value, err := memoryCache.Get(t.Context(), "test:cas")
		assert.NoError(t, err, "should be able to read back the key")
		assert.Nil(t, value, "no value should have been written")
	})

	t.Run("expired key", func(t *testing.T) {
		clock := clock.NewMock()
		memoryCache := cache.NewLRUCache(testutils.GetLog(t), clock, config.MemoryCache{
			MaxKeys: 128,
		})

		assert.NoError(t, memoryCache.SetTTL(t.Context(), "test:cas", []byte("unused"), time.Minute), "must seed key")

		clock.Add(2 * time.Minute)

		swapped, err := memoryCache.CompareAndSwap(t.Context(), "test:cas", []byte("unused"), []byte("consumed"), time.Minute)
		assert.NoError(t, err, "comparison itself should not error")
		assert.False(t, swapped, "must not swap a key that has expired")
	})

	t.Run("zero ttl", func(t *testing.T) {
		// If the ttl is zero then we want to keep whatever expiry the key already
		// had, the same way KEEPTTL works in valkey.
		clock := clock.NewMock()
		memoryCache := cache.NewLRUCache(testutils.GetLog(t), clock, config.MemoryCache{
			MaxKeys: 128,
		})

		assert.NoError(t, memoryCache.SetTTL(t.Context(), "test:cas", []byte("unused"), time.Minute), "must seed key")

		swapped, err := memoryCache.CompareAndSwap(t.Context(), "test:cas", []byte("unused"), []byte("consumed"), 0)
		assert.NoError(t, err, "should be able to compare and swap")
		assert.True(t, swapped, "should swap because the current value matched expected")

		value, err := memoryCache.Get(t.Context(), "test:cas")
		assert.NoError(t, err, "should be able to read back the swapped value")
		assert.Equal(t, []byte("consumed"), value, "swapped value should still be there")

		clock.Add(2 * time.Minute)

		value, err = memoryCache.Get(t.Context(), "test:cas")
		assert.NoError(t, err, "expired key should not error")
		assert.Nil(t, value, "swapped value should expire with the original ttl")
	})

	t.Run("refreshes ttl", func(t *testing.T) {
		clock := clock.NewMock()
		memoryCache := cache.NewLRUCache(testutils.GetLog(t), clock, config.MemoryCache{
			MaxKeys: 128,
		})

		assert.NoError(t, memoryCache.SetTTL(t.Context(), "test:cas", []byte("unused"), 10*time.Second), "must seed key")

		swapped, err := memoryCache.CompareAndSwap(t.Context(), "test:cas", []byte("unused"), []byte("consumed"), time.Minute)
		assert.NoError(t, err, "should be able to compare and swap")
		assert.True(t, swapped, "should swap because the current value matched expected")

		clock.Add(30 * time.Second)

		value, err := memoryCache.Get(t.Context(), "test:cas")
		assert.NoError(t, err, "should be able to read back the swapped value")
		assert.Equal(t, []byte("consumed"), value, "expiry should have been refreshed to the longer ttl")
	})

	t.Run("only one winner", func(t *testing.T) {
		// This is what stops a proof of work challenge from being used twice, if a
		// bunch of requests race to consume the same one only one of them can win.
		clock := clock.NewMock()
		memoryCache := cache.NewLRUCache(testutils.GetLog(t), clock, config.MemoryCache{
			MaxKeys: 128,
		})

		assert.NoError(t, memoryCache.SetTTL(t.Context(), "test:cas", []byte("unused"), time.Minute), "must seed key")

		var wins atomic.Int32
		var wg sync.WaitGroup
		for range 100 {
			wg.Go(func() {
				swapped, err := memoryCache.CompareAndSwap(t.Context(), "test:cas", []byte("unused"), []byte("consumed"), time.Minute)
				assert.NoError(t, err, "comparison itself should not error")
				if swapped {
					wins.Add(1)
				}
			})
		}
		wg.Wait()

		assert.EqualValues(t, 1, wins.Load(), "one swap should win")
	})

	t.Run("blank key", func(t *testing.T) {
		clock := clock.NewMock()
		memoryCache := cache.NewLRUCache(testutils.GetLog(t), clock, config.MemoryCache{
			MaxKeys: 128,
		})

		_, err := memoryCache.CompareAndSwap(t.Context(), "", []byte("unused"), []byte("consumed"), time.Minute)
		assert.EqualError(t, err, "key is blank", "should be blank key error")
	})
}

func TestLruCache_Close(t *testing.T) {
	clock := clock.NewMock()
	memoryCache := cache.NewLRUCache(testutils.GetLog(t), clock, config.MemoryCache{
		MaxKeys: 128,
	})

	assert.NoError(t, memoryCache.Set(t.Context(), "test:data", TestValue), "must set value")

	err := memoryCache.Close()
	assert.NoError(t, err, "should close without error")

	value, err := memoryCache.Get(t.Context(), "test:data")
	assert.NoError(t, err, "should be able to read after close")
	assert.Nil(t, value, "everything should be purged after close")
}
