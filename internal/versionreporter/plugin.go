package versionreporter

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/yxinmiracle/pluggent/internal/plugin"
	"github.com/yxinmiracle/pluggent/internal/runtimeinfo"
)

// ErrNilWriter 表示版本信息没有可用的输出目标。
var ErrNilWriter = errors.New("version reporter writer must not be nil")

// PluginID 是版本输出插件的唯一标识。
const PluginID plugin.ID = "version-reporter"

// VersionReporterPlugin 消费运行信息服务，并把当前 Pluggent 版本写入指定输出。
type VersionReporterPlugin struct {
	Writer io.Writer
}

// 确认 VersionReporterPlugin 实现了通用插件接口。
var _ plugin.Plugin = VersionReporterPlugin{}

// ID 返回版本输出插件的唯一标识。
func (VersionReporterPlugin) ID() plugin.ID {
	return PluginID
}

// Apply 获取运行信息并输出当前 Pluggent 版本。
func (p VersionReporterPlugin) Apply(
	ctx context.Context,
	scope *plugin.Scope,
) error {
	if p.Writer == nil {
		return ErrNilWriter
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	info, err := runtimeinfo.RuntimeInfoService.Resolve(scope)
	if err != nil {
		return fmt.Errorf(
			"resolve runtime info: %w",
			err,
		)
	}

	_, err = fmt.Fprintf(
		p.Writer,
		"pluggent %s\n",
		info.Version,
	)

	if err != nil {
		return fmt.Errorf(
			"write version: %w",
			err,
		)
	}

	return nil
}
