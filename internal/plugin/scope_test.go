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

func TestScopeCloseReturnsStoredResultAfterCompletion(t *testing.T) {
	scope := NewScope()
	cleanupErr := errors.New("cleanup failed")

	err := scope.AddCleanup(func(context.Context) error {
		return cleanupErr
	})
	if err != nil {
		t.Fatalf("AddCleanup() error = %v", err)
	}

	firstErr := scope.Close(context.Background())
	if !errors.Is(firstErr, cleanupErr) {
		t.Fatalf("first Close() error = %v, want cleanup error", firstErr)
	}

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	secondErr := scope.Close(canceledCtx)
	if secondErr != firstErr {
		t.Fatalf("second Close() error = %v, want stored error %v", secondErr, firstErr)
	}
}

func TestScopeCloseHonorsCancellationWhileClosing(t *testing.T) {
	scope := NewScope()
	cleanupStarted := make(chan struct{})
	releaseCleanup := make(chan struct{})

	err := scope.AddCleanup(func(context.Context) error {
		close(cleanupStarted)
		<-releaseCleanup
		return nil
	})
	if err != nil {
		t.Fatalf("AddCleanup() error = %v", err)
	}

	firstCloseDone := make(chan error, 1)
	go func() {
		firstCloseDone <- scope.Close(context.Background())
	}()

	<-cleanupStarted

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := scope.Close(canceledCtx); !errors.Is(err, context.Canceled) {
		t.Fatalf("Close() error = %v, want context.Canceled", err)
	}

	close(releaseCleanup)
	if err := <-firstCloseDone; err != nil {
		t.Fatalf("first Close() error = %v", err)
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

func TestScopeCloseFinalizesAfterCleanupPanic(t *testing.T) {
	scope := NewScope()
	panicValue := "cleanup panic"

	err := scope.AddCleanup(func(context.Context) error {
		panic(panicValue)
	})
	if err != nil {
		t.Fatalf("AddCleanup() error = %v", err)
	}

	var recovered any
	func() {
		defer func() {
			recovered = recover()
		}()

		_ = scope.Close(context.Background())
	}()

	if recovered != panicValue {
		t.Fatalf("Close() panic = %v, want %v", recovered, panicValue)
	}

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := scope.Close(canceledCtx); err != nil {
		t.Fatalf("second Close() error = %v, want stored close result", err)
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
