package app

import (
	"context"
	"fmt"
	"io"

	appconfig "github.com/yxinmiracle/pluggent/internal/config"
	"github.com/yxinmiracle/pluggent/internal/plugin"
)

// Options 保存运行 Pluggent 所需的进程输入。
type Options struct {
	Args       []string
	ConfigPath string
	Stdout     io.Writer
	Version    string
}

// Run 启动 Pluggent 应用。
func Run(ctx context.Context, options Options) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	configuration, err := appconfig.Load(options.ConfigPath)
	if err != nil {
		return fmt.Errorf("load Pluggent config: %w", err)
	}

	instances, err := buildPlugins(configuration.Plugins, options)
	if err != nil {
		return fmt.Errorf("compose Pluggent plugins: %w", err)
	}

	host := plugin.NewHost()

	if err := host.Start(ctx, instances...); err != nil {
		return fmt.Errorf(
			"start Pluggent: %w",
			err,
		)
	}

	if err := host.Close(context.WithoutCancel(ctx)); err != nil {
		return fmt.Errorf(
			"close Pluggent: %w",
			err,
		)
	}

	return nil
}
