package hw04lrucache

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	t.Run("empty cache", func(t *testing.T) {
		c := NewCache(10)

		_, ok := c.Get("aaa")
		require.False(t, ok)

		_, ok = c.Get("bbb")
		require.False(t, ok)
	})

	t.Run("simple", func(t *testing.T) {
		c := NewCache(5)

		wasInCache := c.Set("aaa", 100)
		require.False(t, wasInCache)

		wasInCache = c.Set("bbb", 200)
		require.False(t, wasInCache)

		val, ok := c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 100, val)

		val, ok = c.Get("bbb")
		require.True(t, ok)
		require.Equal(t, 200, val)

		wasInCache = c.Set("aaa", 300)
		require.True(t, wasInCache)

		val, ok = c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 300, val)

		val, ok = c.Get("ccc")
		require.False(t, ok)
		require.Nil(t, val)
	})

	t.Run("purge logic 1", func(t *testing.T) {
		c := NewCache(3)

		// кладём 3 элемента
		require.False(t, c.Set("k1", 1))
		require.False(t, c.Set("k2", 2))
		require.False(t, c.Set("k3", 3))

		// все на месте
		v, ok := c.Get("k1")
		require.True(t, ok)
		require.Equal(t, 1, v)

		v, ok = c.Get("k2")
		require.True(t, ok)
		require.Equal(t, 2, v)

		v, ok = c.Get("k3")
		require.True(t, ok)
		require.Equal(t, 3, v)

		// добавляем 4-й -> должен вытолкнуться самый старый (k1)
		require.False(t, c.Set("k4", 4))

		_, ok = c.Get("k1")
		require.False(t, ok)

		v, ok = c.Get("k2")
		require.True(t, ok)
		require.Equal(t, 2, v)

		v, ok = c.Get("k3")
		require.True(t, ok)
		require.Equal(t, 3, v)

		v, ok = c.Get("k4")
		require.True(t, ok)
		require.Equal(t, 4, v)
	})

	t.Run("purge logic 2", func(t *testing.T) {
		c := NewCache(3)

		require.False(t, c.Set("k1", 1))
		require.False(t, c.Set("k2", 2))
		require.False(t, c.Set("k3", 3))

		v, ok := c.Get("k1")
		require.True(t, ok)
		require.Equal(t, 1, v)

		require.True(t, c.Set("k2", 200))

		require.False(t, c.Set("k4", 4))

		_, ok = c.Get("k3")
		require.False(t, ok, "k3 should be evicted as least recently used")

		v, ok = c.Get("k1")
		require.True(t, ok)
		require.Equal(t, 1, v)

		v, ok = c.Get("k2")
		require.True(t, ok)
		require.Equal(t, 200, v)

		v, ok = c.Get("k4")
		require.True(t, ok)
		require.Equal(t, 4, v)
	})
}

func TestCacheMultithreading(t *testing.T) {
	t.Skip() // Remove me if task with asterisk completed.

	c := NewCache(10)
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Set(Key(strconv.Itoa(i)), i)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Get(Key(strconv.Itoa(rand.Intn(1_000_000))))
		}
	}()

	wg.Wait()
}
