package app

import (
	"errors"
	"fmt"

	appconfig "github.com/yxinmiracle/pluggent/internal/config"
	"github.com/yxinmiracle/pluggent/internal/plugin"
	"github.com/yxinmiracle/pluggent/internal/runtimeinfo"
	"github.com/yxinmiracle/pluggent/internal/versionreporter"
)

// ErrUnknownPluginID 表示配置引用了当前程序没有注册的插件工厂。
var ErrUnknownPluginID = errors.New("configured plugin ID is not registered")

type pluginFactory func() plugin.Plugin

func buildPlugins(
	entries []appconfig.PluginEntry,
	options Options,
) ([]plugin.Plugin, error) {
	factories := map[plugin.ID]pluginFactory{
		runtimeinfo.ProviderPluginID: func() plugin.Plugin {
			return runtimeinfo.RuntimeInfoProviderPlugin{
				Version: options.Version,
			}
		},
		versionreporter.PluginID: func() plugin.Plugin {
			return versionreporter.VersionReporterPlugin{
				Writer: options.Stdout,
			}
		},
	}

	instances := make([]plugin.Plugin, 0, len(entries))

	for index, entry := range entries {
		if entry.Disabled {
			continue
		}

		factory, exists := factories[entry.ID]
		if !exists {
			return nil, fmt.Errorf(
				"plugin at index %d with ID %q: %w",
				index,
				entry.ID,
				ErrUnknownPluginID,
			)
		}

		instances = append(instances, factory())
	}

	return instances, nil
}
