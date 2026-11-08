package micup

import (
	"context"
	"fmt"
	"sync"
)

// WithShard configures a specific shard for the session.
func WithShard(shardID, shardCount int) Option {
	return func(o *SessionOptions) {
		o.ShardID = shardID
		o.ShardCount = shardCount
	}
}

// ShardingManager orchestrates multiple shard sessions for large-scale bot deployments.
type ShardingManager struct {
	Token       string
	ShardCount  int
	Sessions    []*Session
	BaseOptions []Option
	handlers    []interface{}
	mu          sync.RWMutex
}

// NewShardingManager creates a manager for multi-shard bots.
func NewShardingManager(token string, shardCount int, baseOpts ...Option) (*ShardingManager, error) {
	if shardCount < 1 {
		shardCount = 1
	}

	m := &ShardingManager{
		Token:       token,
		ShardCount:  shardCount,
		Sessions:    make([]*Session, shardCount),
		BaseOptions: baseOpts,
		handlers:    make([]interface{}, 0),
	}

	for i := 0; i < shardCount; i++ {
		opts := append([]Option{}, baseOpts...)
		opts = append(opts, WithShard(i, shardCount))

		s, err := New(token, opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize shard %d: %w", i, err)
		}
		m.Sessions[i] = s
	}

	return m, nil
}

// AddHandler registers an event handler across all shards.
func (m *ShardingManager) AddHandler(handler interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.handlers = append(m.handlers, handler)
	for _, s := range m.Sessions {
		s.AddHandler(handler)
	}
}

// Open launches all shard sessions concurrently.
func (m *ShardingManager) Open(ctx context.Context) error {
	var wg sync.WaitGroup
	errChan := make(chan error, m.ShardCount)

	for i, s := range m.Sessions {
		wg.Add(1)
		go func(shardID int, sess *Session) {
			defer wg.Done()
			if err := sess.OpenContext(ctx); err != nil {
				errChan <- fmt.Errorf("shard %d connection failed: %w", shardID, err)
			}
		}(i, s)
	}

	wg.Wait()
	close(errChan)

	if len(errChan) > 0 {
		return <-errChan
	}
	return nil
}

// Close gracefully disconnects all shard sessions.
func (m *ShardingManager) Close() error {
	var firstErr error
	for i, s := range m.Sessions {
		if err := s.Close(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("error closing shard %d: %w", i, err)
		}
	}
	return firstErr
}
