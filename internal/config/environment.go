package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

const (
	// DefaultEnvironmentPath 是本地开发环境变量文件的默认路径。
	DefaultEnvironmentPath = ".env"

	// ConfigPathEnvironmentVariable 指定主配置文件路径。
	ConfigPathEnvironmentVariable = "PLUGGENT_CONFIG"

	// DefaultConfigPath 是没有环境变量覆盖时使用的主配置文件路径。
	DefaultConfigPath = "config/pluggent.yaml"
)

// LoadEnvironmentFile 加载可选的环境变量文件，并保留进程中已经存在的值。
func LoadEnvironmentFile(path string) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("environment file path must not be empty")
	}

	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect environment file %q: %w", path, err)
	}

	if err := godotenv.Load(path); err != nil {
		return fmt.Errorf("load environment file %q: %w", path, err)
	}

	return nil
}

// ResolveConfigPath 根据进程环境解析主配置文件路径。
func ResolveConfigPath(lookup func(string) (string, bool)) string {
	if lookup != nil {
		if value, exists := lookup(ConfigPathEnvironmentVariable); exists {
			if path := strings.TrimSpace(value); path != "" {
				return path
			}
		}
	}

	return DefaultConfigPath
}
