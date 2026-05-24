package store

import (
	"sync"
	"time"
)

type AsyncWriteBuffer struct {
	backend       HitStore
	queue         []HitEvent
	mu            sync.Mutex
	flushInterval time.Duration
	maxBuffer     int
	batchSize     int
	done          chan struct{}
	wg            sync.WaitGroup
}

func NewAsyncWriteBuffer(backend HitStore, flushInterval time.Duration, maxBuffer, batchSize int) *AsyncWriteBuffer {
	if flushInterval == 0 {
		flushInterval = 2 * time.Second
	}
	if maxBuffer == 0 {
		maxBuffer = 10_000
	}
	if batchSize == 0 {
		batchSize = 100
	}
	return &AsyncWriteBuffer{
		backend:       backend,
		flushInterval: flushInterval,
		maxBuffer:     maxBuffer,
		batchSize:     batchSize,
		done:          make(chan struct{}),
	}
}

func (b *AsyncWriteBuffer) Start() {
	b.wg.Add(1)
	go func() {
		defer b.wg.Done()
		b.flushWorker()
	}()
}

func (b *AsyncWriteBuffer) Record(hit HitEvent) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.queue) >= b.maxBuffer {
		b.queue = b.queue[1:] // drop oldest on overflow
	}
	b.queue = append(b.queue, hit)
	return nil
}

func (b *AsyncWriteBuffer) flushWorker() {
	ticker := time.NewTicker(b.flushInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			b.flush()
		case <-b.done:
			b.flush()
			return
		}
	}
}

func (b *AsyncWriteBuffer) flush() {
	b.mu.Lock()
	if len(b.queue) == 0 {
		b.mu.Unlock()
		return
	}
	n := len(b.queue)
	if n > b.batchSize {
		n = b.batchSize
	}
	batch := make([]HitEvent, n)
	copy(batch, b.queue[:n])
	b.queue = b.queue[n:]
	b.mu.Unlock()

	for _, hit := range batch {
		b.backend.Record(hit) //nolint:errcheck
	}
}

// FlushAll drains the queue synchronously. Used in tests.
func (b *AsyncWriteBuffer) FlushAll() {
	for {
		b.mu.Lock()
		n := len(b.queue)
		b.mu.Unlock()
		if n == 0 {
			break
		}
		b.flush()
	}
}

func (b *AsyncWriteBuffer) RecentHits(query HitQuery) ([]HitEvent, error) {
	return b.backend.RecentHits(query)
}

func (b *AsyncWriteBuffer) EndpointSummaries(sinceDays int) ([]EndpointSummary, error) {
	return b.backend.EndpointSummaries(sinceDays)
}

func (b *AsyncWriteBuffer) TotalSummary(sinceDays int) (TotalSummary, error) {
	return b.backend.TotalSummary(sinceDays)
}

func (b *AsyncWriteBuffer) Close() error {
	close(b.done)
	b.wg.Wait()
	return b.backend.Close()
}
