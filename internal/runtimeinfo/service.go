package runtimeinfo

import "github.com/yxinmiracle/pluggent/internal/plugin"

// Info 保存当前 Pluggent 进程的构建和运行信息。
type Info struct {
	Version string
}

// RuntimeInfoService 定义运行信息服务的类型和唯一身份。
var RuntimeInfoService = plugin.DefineService[Info]("runtime-info")
