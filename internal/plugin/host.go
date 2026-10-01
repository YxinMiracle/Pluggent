package plugin

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type hostState uint8

// ErrNilPlugin 表示插件列表中包含 nil。
var ErrNilPlugin = errors.New("plugin must not be nil")

// ErrEmptyPluginID 表示插件没有有效 ID。
var ErrEmptyPluginID = errors.New("plugin ID must not be empty")

// ErrDuplicatePluginID 表示插件 ID 重复。
var ErrDuplicatePluginID = errors.New("plugin ID is duplicated")

// ErrHostAlreadyStarted 表示 Host 已经开始过启动流程。
var ErrHostAlreadyStarted = errors.New("plugin host has already started")

// ErrHostStarting 表示 Host 仍在执行插件启动流程。
var ErrHostStarting = errors.New("plugin host is still starting")

const (
	hostStateNew hostState = iota
	hostStateStarting
	hostStateRunning
	hostStateClosing
	hostStateClosed
)

type loadedPlugin struct {
	instance Plugin
	scope    *Scope
}

// Host 管理插件组合、共享服务和插件生命周期。
type Host struct {
	mu sync.Mutex

	state    hostState
	registry *serviceRegistry
	plugins  map[ID]*loadedPlugin
	order    []ID

	closeDone chan struct{}
	closeErr  error
}

// NewHost 创建一个尚未启动的插件宿主。
func NewHost() *Host {
	return &Host{
		registry:  newServiceRegistry(),
		plugins:   make(map[ID]*loadedPlugin),
		closeDone: make(chan struct{}),
	}
}

func validatePlugins(instances []Plugin) error {
	ids := make(map[ID]struct{}, len(instances))

	for index, instance := range instances {
		if instance == nil {
			return fmt.Errorf(
				"plugin at index %d: %w",
				index,
				ErrNilPlugin,
			)
		}

		id := instance.ID()
		if id == "" {
			return fmt.Errorf(
				"plugin at index %d: %w",
				index,
				ErrEmptyPluginID,
			)
		}

		if _, exists := ids[id]; exists {
			return fmt.Errorf(
				"plugin %q: %w",
				id,
				ErrDuplicatePluginID,
			)
		}

		ids[id] = struct{}{}
	}

	return nil
}

func closeLoadedPlugins(
	ctx context.Context,
	loaded []*loadedPlugin,
) error {
	var closeErrors []error

	for index := len(loaded) - 1; index >= 0; index-- {
		current := loaded[index]

		if err := current.scope.Close(ctx); err != nil {
			closeErrors = append(
				closeErrors,
				fmt.Errorf(
					"close plugin %q: %w",
					current.instance.ID(),
					err,
				),
			)
		}
	}

	return errors.Join(closeErrors...)
}

func (h *Host) beginStart() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.state != hostStateNew {
		return ErrHostAlreadyStarted
	}

	h.state = hostStateStarting

	return nil
}

// Start 按给定顺序加载全部插件。
func (h *Host) Start(
	ctx context.Context,
	instances ...Plugin,
) error {
	if err := validatePlugins(instances); err != nil {
		return err
	}

	if err := h.beginStart(); err != nil {
		return err
	}

	loaded := make(
		[]*loadedPlugin,
		0,
		len(instances),
	)

	for _, instance := range instances {
		current := &loadedPlugin{
			instance: instance,
			scope:    newScope(h.registry),
		}

		loaded = append(loaded, current)

		if err := instance.Apply(ctx, current.scope); err != nil {
			applyErr := fmt.Errorf(
				"apply plugin %q: %w",
				instance.ID(),
				err,
			)

			rollbackErr := closeLoadedPlugins(
				context.WithoutCancel(ctx),
				loaded,
			)

			h.mu.Lock()
			h.state = hostStateClosed
			h.closeErr = rollbackErr
			close(h.closeDone)
			h.mu.Unlock()

			return errors.Join(applyErr, rollbackErr)
		}
	}

	h.mu.Lock()

	for _, current := range loaded {
		id := current.instance.ID()

		h.plugins[id] = current
		h.order = append(h.order, id)
	}

	h.state = hostStateRunning
	h.mu.Unlock()

	return nil
}

// Close 按插件加载顺序的相反顺序关闭 Host。
func (h *Host) Close(ctx context.Context) error {
	h.mu.Lock()

	switch h.state {
	case hostStateNew:
		h.state = hostStateClosed
		close(h.closeDone)
		h.mu.Unlock()

		return nil

	case hostStateStarting:
		h.mu.Unlock()

		return ErrHostStarting

	case hostStateRunning:
		h.state = hostStateClosing

		loaded := make(
			[]*loadedPlugin,
			0,
			len(h.order),
		)

		for _, id := range h.order {
			loaded = append(
				loaded,
				h.plugins[id],
			)
		}

		h.mu.Unlock()

		closeErr := closeLoadedPlugins(ctx, loaded)

		h.mu.Lock()
		h.closeErr = closeErr
		h.plugins = make(map[ID]*loadedPlugin)
		h.order = nil
		h.state = hostStateClosed
		close(h.closeDone)
		h.mu.Unlock()

		return closeErr

	case hostStateClosing:
		done := h.closeDone
		h.mu.Unlock()

		select {
		case <-done:
			h.mu.Lock()
			err := h.closeErr
			h.mu.Unlock()

			return err

		case <-ctx.Done():
			return ctx.Err()
		}

	case hostStateClosed:
		err := h.closeErr
		h.mu.Unlock()

		return err

	default:
		h.mu.Unlock()
		panic("unknown plugin host state")
	}
}
