package plugin

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type stubPlugin struct {
	id    ID
	apply func(ctx context.Context, scope *Scope) error
}

func (p stubPlugin) ID() ID {
	return p.id
}

func (p stubPlugin) Apply(ctx context.Context, scope *Scope) error {
	if p.apply == nil {
		return nil
	}

	return p.apply(ctx, scope)
}

func TestValidatePlugins(t *testing.T) {
	tests := []struct {
		name      string
		instances []Plugin
		wantErr   error
	}{
		{
			name: "valid plugins",
			instances: []Plugin{
				stubPlugin{id: "model"},
				stubPlugin{id: "agent"},
			},
		},
		{
			name:      "nil plugin",
			instances: []Plugin{nil},
			wantErr:   ErrNilPlugin,
		},
		{
			name:      "empty plugin ID",
			instances: []Plugin{stubPlugin{}},
			wantErr:   ErrEmptyPluginID,
		},
		{
			name: "duplicate plugin ID",
			instances: []Plugin{
				stubPlugin{id: "model"},
				stubPlugin{id: "model"},
			},
			wantErr: ErrDuplicatePluginID,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validatePlugins(test.instances)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("validatePlugins() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

func TestHostBeginStartOnlyOnce(t *testing.T) {
	host := NewHost()

	if err := host.beginStart(); err != nil {
		t.Fatalf("first beginStart() error = %v", err)
	}

	err := host.beginStart()
	if !errors.Is(err, ErrHostAlreadyStarted) {
		t.Fatalf("second beginStart() error = %v, want ErrHostAlreadyStarted", err)
	}
}

func TestHostStartAppliesInOrderAndSharesServices(t *testing.T) {
	host := NewHost()
	service := DefineService[int]("answer")
	var applyOrder []ID

	provider := stubPlugin{
		id: "provider",
		apply: func(_ context.Context, scope *Scope) error {
			applyOrder = append(applyOrder, "provider")
			return service.Provide(scope, 42)
		},
	}
	consumer := stubPlugin{
		id: "consumer",
		apply: func(_ context.Context, scope *Scope) error {
			applyOrder = append(applyOrder, "consumer")

			value, err := service.Resolve(scope)
			if err != nil {
				return err
			}
			if value != 42 {
				t.Fatalf("Resolve() = %d, want 42", value)
			}

			return nil
		},
	}

	if err := host.Start(context.Background(), provider, consumer); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	wantOrder := []ID{"provider", "consumer"}
	if !reflect.DeepEqual(applyOrder, wantOrder) {
		t.Fatalf("apply order = %v, want %v", applyOrder, wantOrder)
	}
	if !reflect.DeepEqual(host.order, wantOrder) {
		t.Fatalf("host order = %v, want %v", host.order, wantOrder)
	}
	if host.state != hostStateRunning {
		t.Fatalf("host state = %v, want hostStateRunning", host.state)
	}
}

func TestHostStartRollsBackPartialEffectsInReverseOrder(t *testing.T) {
	host := NewHost()
	applyErr := errors.New("apply failed")
	cleanupErr := errors.New("cleanup failed")
	var cleanupOrder []ID

	first := stubPlugin{
		id: "first",
		apply: func(_ context.Context, scope *Scope) error {
			return scope.AddCleanup(func(context.Context) error {
				cleanupOrder = append(cleanupOrder, "first")
				return nil
			})
		},
	}
	second := stubPlugin{
		id: "second",
		apply: func(_ context.Context, scope *Scope) error {
			if err := scope.AddCleanup(func(ctx context.Context) error {
				if ctx.Err() != nil {
					t.Fatalf("rollback context error = %v, want nil", ctx.Err())
				}

				cleanupOrder = append(cleanupOrder, "second")
				return cleanupErr
			}); err != nil {
				return err
			}

			return applyErr
		},
	}

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	err := host.Start(canceledCtx, first, second)
	if !errors.Is(err, applyErr) {
		t.Fatalf("Start() error = %v, want apply error", err)
	}
	if !errors.Is(err, cleanupErr) {
		t.Fatalf("Start() error = %v, want cleanup error", err)
	}

	wantOrder := []ID{"second", "first"}
	if !reflect.DeepEqual(cleanupOrder, wantOrder) {
		t.Fatalf("cleanup order = %v, want %v", cleanupOrder, wantOrder)
	}
	if host.state != hostStateClosed {
		t.Fatalf("host state = %v, want hostStateClosed", host.state)
	}
	if len(host.plugins) != 0 || len(host.order) != 0 {
		t.Fatalf("failed host retained plugins: plugins = %d, order = %v", len(host.plugins), host.order)
	}
}

func TestHostStartRejectsSecondStart(t *testing.T) {
	host := NewHost()

	if err := host.Start(context.Background()); err != nil {
		t.Fatalf("first Start() error = %v", err)
	}

	err := host.Start(context.Background())
	if !errors.Is(err, ErrHostAlreadyStarted) {
		t.Fatalf("second Start() error = %v, want ErrHostAlreadyStarted", err)
	}
}

func TestHostCloseRunsCleanupsInReverseOrderAndStoresResult(t *testing.T) {
	host := NewHost()
	cleanupErr := errors.New("cleanup failed")
	var cleanupOrder []ID

	instances := []Plugin{
		stubPlugin{
			id: "first",
			apply: func(_ context.Context, scope *Scope) error {
				return scope.AddCleanup(func(context.Context) error {
					cleanupOrder = append(cleanupOrder, "first")
					return nil
				})
			},
		},
		stubPlugin{
			id: "second",
			apply: func(_ context.Context, scope *Scope) error {
				return scope.AddCleanup(func(context.Context) error {
					cleanupOrder = append(cleanupOrder, "second")
					return cleanupErr
				})
			},
		},
	}

	if err := host.Start(context.Background(), instances...); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	firstErr := host.Close(context.Background())
	if !errors.Is(firstErr, cleanupErr) {
		t.Fatalf("first Close() error = %v, want cleanup error", firstErr)
	}

	wantOrder := []ID{"second", "first"}
	if !reflect.DeepEqual(cleanupOrder, wantOrder) {
		t.Fatalf("cleanup order = %v, want %v", cleanupOrder, wantOrder)
	}
	if host.state != hostStateClosed {
		t.Fatalf("host state = %v, want hostStateClosed", host.state)
	}
	if len(host.plugins) != 0 || len(host.order) != 0 {
		t.Fatalf("closed host retained plugins: plugins = %d, order = %v", len(host.plugins), host.order)
	}

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	secondErr := host.Close(canceledCtx)
	if !errors.Is(secondErr, cleanupErr) {
		t.Fatalf("second Close() error = %v, want stored cleanup error", secondErr)
	}
}

func TestHostCloseHonorsCancellationWhileClosing(t *testing.T) {
	host := NewHost()
	cleanupStarted := make(chan struct{})
	releaseCleanup := make(chan struct{})

	instance := stubPlugin{
		id: "blocking",
		apply: func(_ context.Context, scope *Scope) error {
			return scope.AddCleanup(func(context.Context) error {
				close(cleanupStarted)
				<-releaseCleanup
				return nil
			})
		},
	}

	if err := host.Start(context.Background(), instance); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	firstCloseDone := make(chan error, 1)
	go func() {
		firstCloseDone <- host.Close(context.Background())
	}()

	<-cleanupStarted

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := host.Close(canceledCtx); !errors.Is(err, context.Canceled) {
		t.Fatalf("Close() error = %v, want context.Canceled", err)
	}

	close(releaseCleanup)
	if err := <-firstCloseDone; err != nil {
		t.Fatalf("first Close() error = %v", err)
	}

	if err := host.Close(canceledCtx); err != nil {
		t.Fatalf("Close() after completion error = %v, want nil", err)
	}
}

func TestHostCloseBeforeStartIsIdempotent(t *testing.T) {
	host := NewHost()

	if err := host.Close(context.Background()); err != nil {
		t.Fatalf("first Close() error = %v", err)
	}
	if err := host.Close(context.Background()); err != nil {
		t.Fatalf("second Close() error = %v", err)
	}
}

func TestHostCloseRejectsStartingState(t *testing.T) {
	host := NewHost()

	if err := host.beginStart(); err != nil {
		t.Fatalf("beginStart() error = %v", err)
	}

	if err := host.Close(context.Background()); !errors.Is(err, ErrHostStarting) {
		t.Fatalf("Close() error = %v, want ErrHostStarting", err)
	}
}
