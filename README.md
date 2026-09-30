# Pluggent

一个使用 Go 实现的一切皆插件的 Agent Harness。

## 当前状态

Pluggent 正处于初始架构阶段，目前包含：

- 薄进程启动器；
- 应用启动边界；
- 最小插件协议；
- 支持逆序清理的插件作用域；
- Makefile 开发命令；
- 单元测试和竞态检查。

## 开发命令

```bash
make help
make check
make test-race
make build
make run
```
