package plugin

import "context"

// ID 标识一次组合中的插件实例。
type ID string

// Plugin 定义 Pluggent 插件必须实现的行为。
type Plugin interface {
	ID() ID
	Apply(ctx context.Context, scope *Scope) error
}
