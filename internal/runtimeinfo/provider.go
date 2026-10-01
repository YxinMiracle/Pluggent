package runtimeinfo

import (
	"context"

	"github.com/yxinmiracle/pluggent/internal/plugin"
)

// ProviderPluginID 是运行信息提供方插件的唯一标识。
const ProviderPluginID plugin.ID = "runtime-info-provider"

// RuntimeInfoProviderPlugin 向其他插件提供当前进程的运行信息。
type RuntimeInfoProviderPlugin struct {
	Version string
}

// 确认 RuntimeInfoProviderPlugin 实现了通用插件接口。
var _ plugin.Plugin = RuntimeInfoProviderPlugin{}

// ID 返回运行信息插件的唯一标识。
func (RuntimeInfoProviderPlugin) ID() plugin.ID {
	return ProviderPluginID
}

// Apply 把当前运行信息注册到插件作用域。
func (p RuntimeInfoProviderPlugin) Apply(
	ctx context.Context,
	scope *plugin.Scope,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	info := Info{
		Version: p.Version,
	}

	return RuntimeInfoService.Provide(scope, info)
}
