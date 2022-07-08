//go:build !race

// https://github.com/dshulyak/art

package art

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOlock(t *testing.T) {

	t.Run("concurrent readers", func(t *testing.T) {
		var lock olock
		version, restart := lock.RLock()
		require.Empty(t, version)
		require.False(t, restart)
		version, restart = lock.RLock()
		require.Empty(t, version)
		require.False(t, restart)
	})

	t.Run("reader invalidated", func(t *testing.T) {
		var lock olock
		version, restart := lock.RLock()
		require.Empty(t, version)
		require.False(t, restart)
		lock.Lock()
		restart = lock.RUnlock(version, nil)
		require.True(t, restart)
	})

	t.Run("writer blocks reader", func(t *testing.T) {
		var lock olock
		lock.Lock()
		versionc := make(chan uint64, 1)
		go func() {
			version, restart := lock.RLock()
			require.False(t, restart)
			versionc <- version
		}()
		select {
		case <-time.After(100 * time.Millisecond):
		case <-versionc:
			require.FailNow(t, "reader must be blocked")
		}
		lock.Unlock()
		select {
		case <-time.After(100 * time.Millisecond):
			require.FailNow(t, "reader must succeed when writer unlocks")
		case version := <-versionc:
			// 4 - +2 locked, +2 unlocked
			require.Equal(t, uint64(4), version)
		}
	})

	t.Run("read obsolete", func(t *testing.T) {
		var lock olock
		lock.Lock()
		lock.UnlockObsolete()

		version, restart := lock.RLock()
		require.True(t, restart)
		// +2 - lock , +3 - unlock obsolete
		require.Equal(t, uint64(5), version)
	})

	t.Run("read check", func(t *testing.T) {
		var lock olock
		version, restart := lock.RLock()
		require.False(t, restart)
		require.False(t, lock.Check(version))
		lock.Lock()
		require.True(t, lock.Check(version))
	})

	t.Run("upgrade", func(t *testing.T) {
		var lock olock
		version, restart := lock.RLock()
		require.False(t, restart)
		require.False(t, lock.Upgrade(version, nil))
	})
}

type integer struct {
	lock  olock
	value int
}

func (i *integer) inc() {
	for {
		version, _ := i.lock.RLock()
		if i.lock.Upgrade(version, nil) {
			continue
		}
		i.value++
		i.lock.Unlock()
		return
	}
}

func (i *integer) cas(from, to int) {
	for {
		version, _ := i.lock.RLock()
		if i.value != from {
			if i.lock.RUnlock(version, nil) {
				continue
			}
			return
		}

		if i.lock.Upgrade(version, nil) {
			continue
		}
		i.value += to
		i.lock.Unlock()
		return
	}
}

func TestOlockIncrement(t *testing.T) {
	var (
		it integer
		wg sync.WaitGroup
	)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			it.inc()
			wg.Done()
		}()
	}
	wg.Wait()
	require.Equal(t, 100, it.value)
}

func TestOlockCas(t *testing.T) {
	var (
		it integer
		wg sync.WaitGroup
	)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			it.cas(0, 100)
			wg.Done()
		}()
	}
	wg.Wait()
	require.Equal(t, 100, it.value)
}
