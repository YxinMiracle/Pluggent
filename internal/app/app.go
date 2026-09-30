package app

import (
	"context"
	"fmt"
	"io"
)

// Options 保存运行 Pluggent 所需的进程输入。
type Options struct {
	Args    []string
	Stdout  io.Writer
	Version string
}

// Run 启动 Pluggent 应用。
func Run(ctx context.Context, options Options) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	_, err := fmt.Fprintf(
		options.Stdout,
		"pluggent %s\n",
		options.Version,
	)
	return err
}
