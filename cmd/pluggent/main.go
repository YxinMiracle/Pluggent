package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/yxinmiracle/pluggent/internal/app"
	appconfig "github.com/yxinmiracle/pluggent/internal/config"
)

var version = "dev"

func main() {
	os.Exit(run())
}

func run() int {
	if err := appconfig.LoadEnvironmentFile(
		appconfig.DefaultEnvironmentPath,
	); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		return 1
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
	)
	defer stop()

	err := app.Run(ctx, app.Options{
		Args:       os.Args[1:],
		ConfigPath: appconfig.ResolveConfigPath(os.LookupEnv),
		Stdout:     os.Stdout,
		Version:    version,
	})
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}
