package output

import (
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProgressTimeoutDoesNotInc(t *testing.T) {
	p := NewProgress("loading", "loaded", false)
	p.Start()
	p.Inc()
	p.Stop() // should not hang
	p.Inc()  // should restart progress and inc
	assert.Equal(t, uint64(2), p.Val())
	assert.Equal(t, true, p.(*progress).started.Load())

	p.Stop()
	assert.Equal(t, false, p.(*progress).started.Load())
}

func TestProgressTimeoutDoesNotHang(t *testing.T) {
	p := NewProgress("loading", "loaded", false)
	p.Start()
	time.Sleep(progressTimeout)
	// wait for the internal goroutine to stop after timeout
	for p.(*progress).started.Load() == true {
		runtime.Gosched()
	}
	p.Inc()  // should not hang but inc
	p.Stop() // should not hang
	assert.Equal(t, uint64(1), p.Val())
}

func TestProgress(t *testing.T) {
	progress := NewProgress("loading", "loaded", false)
	progress.Start()
	progress.Inc()
	progress.Inc()
	progress.Inc()
	progress.Stop()
	assert.Equal(t, uint64(3), progress.Val())
}
