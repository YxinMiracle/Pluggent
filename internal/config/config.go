// Package config 负责读取和校验 Pluggent 的声明式配置。
package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/yxinmiracle/pluggent/internal/plugin"
	"go.yaml.in/yaml/v3"
)

// ErrEmptyPath 表示没有提供配置文件路径。
var ErrEmptyPath = errors.New("config path must not be empty")

// ErrNoPlugins 表示配置没有声明任何插件。
var ErrNoPlugins = errors.New("config must declare at least one plugin")

// ErrEmptyPluginID 表示配置中的插件没有有效 ID。
var ErrEmptyPluginID = errors.New("configured plugin ID must not be empty")

// ErrDuplicatePluginID 表示配置重复声明了同一个插件 ID。
var ErrDuplicatePluginID = errors.New("configured plugin ID is duplicated")

// ErrMultipleDocuments 表示配置文件包含多个 YAML 文档。
var ErrMultipleDocuments = errors.New("config must contain exactly one YAML document")

// Config 是 Pluggent 的完整声明式配置。
type Config struct {
	Plugins []PluginEntry `yaml:"plugins"`
}

// PluginEntry 描述一个插件及其启用状态。
type PluginEntry struct {
	ID       plugin.ID `yaml:"id"`
	Disabled bool      `yaml:"disabled,omitempty"`
}

// Load 从指定路径读取并严格校验配置。
func Load(path string) (Config, error) {
	if strings.TrimSpace(path) == "" {
		return Config{}, ErrEmptyPath
	}

	file, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("open config %q: %w", path, err)
	}
	defer func() {
		_ = file.Close()
	}()

	decoder := yaml.NewDecoder(file)
	decoder.KnownFields(true)

	var result Config
	if err := decoder.Decode(&result); err != nil {
		return Config{}, fmt.Errorf("decode config %q: %w", path, err)
	}

	var trailingDocument any
	err = decoder.Decode(&trailingDocument)
	if err == nil {
		return Config{}, ErrMultipleDocuments
	}
	if !errors.Is(err, io.EOF) {
		return Config{}, fmt.Errorf("decode trailing config document %q: %w", path, err)
	}

	if err := validate(result); err != nil {
		return Config{}, fmt.Errorf("validate config %q: %w", path, err)
	}

	return result, nil
}

func validate(value Config) error {
	if len(value.Plugins) == 0 {
		return ErrNoPlugins
	}

	ids := make(map[plugin.ID]struct{}, len(value.Plugins))

	for index, entry := range value.Plugins {
		if strings.TrimSpace(string(entry.ID)) == "" {
			return fmt.Errorf("plugin at index %d: %w", index, ErrEmptyPluginID)
		}

		if _, exists := ids[entry.ID]; exists {
			return fmt.Errorf("plugin %q: %w", entry.ID, ErrDuplicatePluginID)
		}

		ids[entry.ID] = struct{}{}
	}

	return nil
}
