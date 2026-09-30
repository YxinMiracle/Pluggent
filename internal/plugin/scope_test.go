package plugin

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestScopeCloseRunsCleanupsInReverseOrder(t *testing.T) {
	scope := NewScope()
	var order []string

	for _, name := range []string{"first", "second", "third"} {
		name := name
		err := scope.AddCleanup(func(context.Context) error {
			order = append(order, name)
			return nil
		})
		if err != nil {
			t.Fatalf("AddCleanup() error = %v", err)
		}
	}

	if err := scope.Close(context.Background()); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	want := []string{"third", "second", "first"}
	if !reflect.DeepEqual(order, want) {
		t.Fatalf("cleanup order = %v, want %v", order, want)
	}
}

func TestScopeCloseIsIdempotent(t *testing.T) {
	scope := NewScope()
	cleanupCalls := 0

	err := scope.AddCleanup(func(context.Context) error {
		cleanupCalls++
		return nil
	})
	if err != nil {
		t.Fatalf("AddCleanup() error = %v", err)
	}

	if err := scope.Close(context.Background()); err != nil {
		t.Fatalf("first Close() error = %v", err)
	}
	if err := scope.Close(context.Background()); err != nil {
		t.Fatalf("second Close() error = %v", err)
	}

	if cleanupCalls != 1 {
		t.Fatalf("cleanup calls = %d, want 1", cleanupCalls)
	}
}

func TestScopeCloseContinuesAfterCleanupError(t *testing.T) {
	scope := NewScope()
	firstErr := errors.New("first cleanup failed")
	secondErr := errors.New("second cleanup failed")
	cleanupCalls := 0

	for _, cleanupErr := range []error{firstErr, secondErr} {
		cleanupErr := cleanupErr
		err := scope.AddCleanup(func(context.Context) error {
			cleanupCalls++
			return cleanupErr
		})
		if err != nil {
			t.Fatalf("AddCleanup() error = %v", err)
		}
	}

	err := scope.Close(context.Background())
	if !errors.Is(err, firstErr) {
		t.Fatalf("Close() error = %v, want first cleanup error", err)
	}
	if !errors.Is(err, secondErr) {
		t.Fatalf("Close() error = %v, want second cleanup error", err)
	}
	if cleanupCalls != 2 {
		t.Fatalf("cleanup calls = %d, want 2", cleanupCalls)
	}
}

func TestScopeRejectsCleanupAfterClose(t *testing.T) {
	scope := NewScope()

	if err := scope.Close(context.Background()); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	err := scope.AddCleanup(func(context.Context) error { return nil })
	if !errors.Is(err, ErrScopeClosed) {
		t.Fatalf("AddCleanup() error = %v, want ErrScopeClosed", err)
	}
}

func TestScopeRejectsNilCleanup(t *testing.T) {
	scope := NewScope()

	err := scope.AddCleanup(nil)
	if !errors.Is(err, ErrNilCleanup) {
		t.Fatalf("AddCleanup() error = %v, want ErrNilCleanup", err)
	}
}

func TestScopeZeroValueIsUsable(t *testing.T) {
	var scope Scope
	cleanupCalls := 0

	err := scope.AddCleanup(func(context.Context) error {
		cleanupCalls++
		return nil
	})
	if err != nil {
		t.Fatalf("AddCleanup() error = %v", err)
	}

	if err := scope.Close(context.Background()); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if cleanupCalls != 1 {
		t.Fatalf("cleanup calls = %d, want 1", cleanupCalls)
	}
}
