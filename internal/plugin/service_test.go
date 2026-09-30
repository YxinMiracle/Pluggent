package plugin

import (
	"context"
	"errors"
	"testing"
)

type greetingService interface {
	Greet(name string) string
}

type greetingServiceFunc func(name string) string

func (f greetingServiceFunc) Greet(name string) string {
	return f(name)
}

func TestServiceProvideAndResolve(t *testing.T) {
	scope := NewScope()
	service := DefineService[greetingService]("greeting")
	implementation := greetingServiceFunc(func(name string) string {
		return "你好，" + name
	})

	if err := service.Provide(scope, implementation); err != nil {
		t.Fatalf("Provide() error = %v", err)
	}

	resolved, err := service.Resolve(scope)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if got := resolved.Greet("小明"); got != "你好，小明" {
		t.Fatalf("Greet() = %q, want %q", got, "你好，小明")
	}
}

func TestServiceRejectsDuplicateProvider(t *testing.T) {
	scope := NewScope()
	service := DefineService[int]("answer")

	if err := service.Provide(scope, 42); err != nil {
		t.Fatalf("first Provide() error = %v", err)
	}

	err := service.Provide(scope, 43)
	if !errors.Is(err, ErrServiceAlreadyProvided) {
		t.Fatalf("second Provide() error = %v, want ErrServiceAlreadyProvided", err)
	}

	resolved, err := service.Resolve(scope)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if resolved != 42 {
		t.Fatalf("Resolve() = %d, want 42", resolved)
	}
}

func TestServiceResolveReportsMissingProvider(t *testing.T) {
	scope := NewScope()
	service := DefineService[int]("answer")

	_, err := service.Resolve(scope)
	if !errors.Is(err, ErrServiceNotFound) {
		t.Fatalf("Resolve() error = %v, want ErrServiceNotFound", err)
	}
}

func TestServiceRegistrationFollowsScopeLifecycle(t *testing.T) {
	scope := NewScope()
	service := DefineService[int]("answer")
	resolvedDuringCleanup := 0

	if err := service.Provide(scope, 42); err != nil {
		t.Fatalf("Provide() error = %v", err)
	}
	if err := scope.AddCleanup(func(context.Context) error {
		value, err := service.Resolve(scope)
		if err != nil {
			return err
		}

		resolvedDuringCleanup = value
		return nil
	}); err != nil {
		t.Fatalf("AddCleanup() error = %v", err)
	}

	if err := scope.Close(context.Background()); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if resolvedDuringCleanup != 42 {
		t.Fatalf("value resolved during cleanup = %d, want 42", resolvedDuringCleanup)
	}

	_, err := service.Resolve(scope)
	if !errors.Is(err, ErrServiceNotFound) {
		t.Fatalf("Resolve() after Close() error = %v, want ErrServiceNotFound", err)
	}
}

func TestServiceRegistryIsSharedByScopes(t *testing.T) {
	registry := newServiceRegistry()
	providerScope := newScope(registry)
	consumerScope := newScope(registry)
	service := DefineService[int]("answer")

	if err := service.Provide(providerScope, 42); err != nil {
		t.Fatalf("Provide() error = %v", err)
	}

	resolved, err := service.Resolve(consumerScope)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if resolved != 42 {
		t.Fatalf("Resolve() = %d, want 42", resolved)
	}

	if err := providerScope.Close(context.Background()); err != nil {
		t.Fatalf("provider Close() error = %v", err)
	}

	_, err = service.Resolve(consumerScope)
	if !errors.Is(err, ErrServiceNotFound) {
		t.Fatalf("Resolve() after provider Close() error = %v, want ErrServiceNotFound", err)
	}

	if err := consumerScope.Close(context.Background()); err != nil {
		t.Fatalf("consumer Close() error = %v", err)
	}
}

func TestServiceSupportsZeroValueScope(t *testing.T) {
	var scope Scope
	service := DefineService[int]("answer")

	if err := service.Provide(&scope, 42); err != nil {
		t.Fatalf("Provide() error = %v", err)
	}

	resolved, err := service.Resolve(&scope)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if resolved != 42 {
		t.Fatalf("Resolve() = %d, want 42", resolved)
	}
}

func TestServiceRejectsUndefinedDefinition(t *testing.T) {
	scope := NewScope()
	var service Service[int]

	if err := service.Provide(scope, 42); !errors.Is(err, ErrInvalidService) {
		t.Fatalf("Provide() error = %v, want ErrInvalidService", err)
	}

	_, err := service.Resolve(scope)
	if !errors.Is(err, ErrInvalidService) {
		t.Fatalf("Resolve() error = %v, want ErrInvalidService", err)
	}
}
