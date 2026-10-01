package plugin

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// ErrServiceAlreadyProvided 表示同一个服务已经有提供者。
var ErrServiceAlreadyProvided = errors.New("service is already provided")

// ErrInvalidService 表示使用了未定义的服务。
var ErrInvalidService = errors.New("service definition is invalid")

// ErrServiceNotFound 表示作用域中没有对应的服务。
var ErrServiceNotFound = errors.New("service is not found")

// ErrServiceTypeMismatch 表示服务值与定义的类型不一致。
var ErrServiceTypeMismatch = errors.New("service type does not match its definition")

// serviceKey 保存服务的唯一身份和诊断名称。
type serviceKey struct {
	name string
}

type serviceValue[T any] struct {
	value T
}

// serviceRegistry 保存一次插件组合共享的服务实现。
type serviceRegistry struct {
	mu     sync.Mutex
	values map[*serviceKey]any
}

func newServiceRegistry() *serviceRegistry {
	return &serviceRegistry{
		values: make(map[*serviceKey]any),
	}
}

// Service 是提供者和消费者共享的类型化服务定义。
type Service[T any] struct {
	key *serviceKey
}

// DefineService 创建一个类型化服务定义。
func DefineService[T any](name string) Service[T] {
	if name == "" {
		panic("service name must not be empty")
	}

	return Service[T]{
		key: &serviceKey{name: name},
	}
}

// Name 返回用于诊断的服务名称。
func (s Service[T]) Name() string {
	if s.key == nil {
		return ""
	}

	return s.key.name
}

func (s *Scope) provideService(key *serviceKey, value any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closing || s.closed {
		return ErrScopeClosed
	}

	registry := s.ensureServiceRegistry()

	registry.mu.Lock()
	defer registry.mu.Unlock()

	if _, exists := registry.values[key]; exists {
		return ErrServiceAlreadyProvided
	}

	registry.values[key] = value

	s.cleanups = append(s.cleanups, func(context.Context) error {
		registry.mu.Lock()
		delete(registry.values, key)
		registry.mu.Unlock()

		return nil
	})

	return nil
}

// Provide 在作用域中注册服务实现。
func (s Service[T]) Provide(scope *Scope, value T) error {
	if s.key == nil {
		return ErrInvalidService
	}

	err := scope.provideService(
		s.key,
		serviceValue[T]{
			value: value,
		},
	)
	if err != nil {
		return fmt.Errorf(
			"provide service %q: %w", s.Name(), err)
	}

	return nil
}

func (s *Scope) resolveService(key *serviceKey) (any, bool) {
	s.mu.Lock()
	registry := s.ensureServiceRegistry()
	s.mu.Unlock()

	registry.mu.Lock()
	defer registry.mu.Unlock()

	value, exists := registry.values[key]

	return value, exists
}

// Resolve 从作用域中获取服务实现。
func (s Service[T]) Resolve(scope *Scope) (T, error) {
	var zero T

	if s.key == nil {
		return zero, ErrInvalidService
	}

	stored, exists := scope.resolveService(s.key)
	if !exists {
		return zero, fmt.Errorf(
			"resolve service %q: %w",
			s.Name(),
			ErrServiceNotFound,
		)
	}

	entry, ok := stored.(serviceValue[T])
	if !ok {
		return zero, fmt.Errorf(
			"resolve service %q: %w",
			s.Name(),
			ErrServiceTypeMismatch,
		)
	}

	return entry.value, nil
}
