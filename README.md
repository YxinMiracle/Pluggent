<h1 align="center">Pluggent</h1>
<p align="center">用 Go 从零构建一切皆插件的 Agent Harness。</p>
<p align="center"><a href="README.md">简体中文</a> · <a href="README.en.md">English</a></p>

> 当前阶段：插件内核与配置组合已可运行。Agent loop、模型、工具和 Skill 尚未实现。

## 快速开始

需要 `go.mod` 指定的 Go 版本。从仓库根目录运行：

```bash
make run
```

输出：

```text
pluggent dev
```

构建版本通过 Makefile 注入，可以这样验证整条 Provider → Service → Consumer 链路：

```bash
make run VERSION=0.1.0
# pluggent 0.1.0
```

本地 `.env` 是可选的。需要指定其他 YAML 文件时，复制示例后修改路径：

```bash
cp .env.example .env
```

## 现在怎样工作

1. 进程从 `PLUGGENT_CONFIG` 找到 YAML 文件；未设置时使用 `config/pluggent.yaml`。
2. 组合入口按 YAML 顺序，从已注册的插件工厂创建插件。
3. `Host` 为每个插件创建独立的 `Scope`，这些 Scope 共享同一个服务注册表。
4. `RuntimeInfoProviderPlugin` 提供 `RuntimeInfoService`；`VersionReporterPlugin` 获取它并输出版本。
5. 启动失败时，Host 回滚已启动插件；正常关闭时，按启动顺序的反方向清理。

```text
config/pluggent.yaml
        │ 按顺序选择插件
        ▼
      Host ─────────────── 共享的服务注册表
       ├─ RuntimeInfoProviderPlugin ─ Scope A ─ Provide(Info) ─┐
       └─ VersionReporterPlugin    ─ Scope B ─ Resolve(Info) ◀─┘
```

| 概念 | 在项目中负责什么 | 代码 |
|---|---|---|
| `Plugin` | 定义插件的 ID 和 `Apply` 方法 | [`internal/plugin/plugin.go`](internal/plugin/plugin.go) |
| `Host` | 启动插件、共享服务、回滚和关闭 | [`internal/plugin/host.go`](internal/plugin/host.go) |
| `Scope` | 持有单个插件的清理操作 | [`internal/plugin/scope.go`](internal/plugin/scope.go) |
| `Service[T]` | 用类型化键提供和获取服务 | [`internal/plugin/service.go`](internal/plugin/service.go) |
| 组合入口 | 将 YAML 插件 ID 映射为具体插件 | [`internal/app/composition.go`](internal/app/composition.go) |

## 配置插件

当前默认配置是 [`config/pluggent.yaml`](config/pluggent.yaml)：

```yaml
plugins:
  - id: runtime-info-provider
  - id: version-reporter
```

顺序有意义：提供方先启动，消费方随后才能获取服务。插件可用 `disabled: true` 禁用；启用的未知 ID、重复 ID、未知 YAML 字段和缺失的依赖服务会让启动报错。

`.env.example` 中记录了配置文件路径：

```dotenv
PLUGGENT_CONFIG=config/pluggent.yaml
```

路径优先级为 **进程环境变量 → `.env` → 代码默认值**。进程环境变量可用于临时选择配置：

```bash
PLUGGENT_CONFIG=/path/to/pluggent.yaml make run
```

YAML 目前只能选择编译时已在 [`internal/app/composition.go`](internal/app/composition.go) 登记的插件；它还不能从任意路径加载 Go 代码，也没有热重载。

## 开发命令

| 命令 | 用途 |
|---|---|
| `make run` | 从源码运行 |
| `make check` | 运行 `go vet` 和单元测试 |
| `make test-race` | 运行竞态检测 |
| `make build` | 构建 `bin/pluggent` |
| `make fmt` | 格式化 Go 源码 |
| `make help` | 查看全部命令 |

## 接下来

插件内核只是 Harness 的基础。后续会逐步加入模型服务与 Provider 注册、Agent loop、工具与 Skill、会话记录、上下文组装及压缩，并让 CLI、TUI 和 Web 作为可替换的入口接入同一套内核。

本项目同时用于学习 Go 和理解 Agent Harness 的工程设计；README 中的“当前实现”与“接下来”分别描述已有代码和计划。
