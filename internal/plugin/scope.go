package plugin

import (
	"context"
	"errors"
	"sync"
)

// Cleanup 定义插件卸载时执行的清理函数。
type Cleanup func(ctx context.Context) error

// ErrScopeClosed 表示插件的作用域已经关闭。
var ErrScopeClosed = errors.New("plugin scope is closed")

// ErrNilCleanup 表示传入了空的清理函数。
var ErrNilCleanup = errors.New("cleanup must not be nil")

// Scope 持有一个插件注册的清理函数。
type Scope struct {
	mu        sync.Mutex
	cleanups  []Cleanup
	closing   bool
	closed    bool
	closeDone chan struct{}
	closeErr  error
}

// NewScope 创建一个插件作用域。
func NewScope() *Scope {
	return &Scope{
		closeDone: make(chan struct{}),
	}
}

// AddCleanup 注册一个插件卸载时执行的清理函数。
func (s *Scope) AddCleanup(cleanup Cleanup) error {
	if cleanup == nil {
		return ErrNilCleanup
	}

	s.mu.Lock()

	defer s.mu.Unlock()

	if s.closing || s.closed {
		return ErrScopeClosed
	}

	s.cleanups = append(s.cleanups, cleanup)

	return nil
}

// Close 按注册顺序的相反顺序执行全部清理函数。
func (s *Scope) Close(ctx context.Context) error {
	s.mu.Lock()

	if s.closed {
		err := s.closeErr
		s.mu.Unlock()
		return err
	}

	done := s.ensureCloseDone()

	if s.closing {
		s.mu.Unlock()

		select {
		case <-done:
			s.mu.Lock()
			err := s.closeErr
			s.mu.Unlock()
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	s.closing = true
	cleanups := append([]Cleanup(nil), s.cleanups...)
	s.cleanups = nil

	s.mu.Unlock()

	var cleanupErrors []error
	var closeErr error

	defer func() {
		s.mu.Lock()
		s.closeErr = closeErr
		s.closed = true
		close(done)
		s.mu.Unlock()
	}()

	for index := len(cleanups) - 1; index >= 0; index-- {
		if err := cleanups[index](ctx); err != nil {
			cleanupErrors = append(cleanupErrors, err)
		}
	}

	closeErr = errors.Join(cleanupErrors...)

	return closeErr

}

func (s *Scope) ensureCloseDone() chan struct{} {
	if s.closeDone == nil {
		s.closeDone = make(chan struct{})
	}

	return s.closeDone
}
