# 📄 Go 优雅关机组件（Graceful Shutdown）设计与使用文档

## 🧩 简介

在构建长期运行的服务程序（如 Web 服务、后台任务、微服务等）时，我们经常需要处理一些 **关闭前的清理逻辑** 。例如：

- 关闭数据库连接池
- 提交未完成的任务
- 保存缓存状态
- 通知其他服务节点下线

为了实现这一目标，我们设计了一个 **模块化、可扩展、支持多场景的优雅关机组件 `shutdown`** ，它允许你注册多个回调函数，并在接收到关闭信号时统一执行这些清理操作，从而保证程序安全退出。

## 🎯 核心目标

1. ✅ 支持多种触发源（POSIX 信号、AWS Auto Scaling、Kubernetes PreStop 等）
2. ✅ 模块化设计，易于扩展
3. ✅ 回调并发执行，提升效率
4. ✅ 统一错误处理机制
5. ✅ 可集成进各类云原生项目

## 🧱 架构设计

### 1. 核心组件

| 组件               | 说明                                            |
| ------------------ | ----------------------------------------------- |
| `ShutdownCallback` | 定义关闭时要执行的回调函数接口                  |
| `ShutdownManager`  | 表示一种关闭触发器（如监听信号、监听 AWS 消息） |
| `GracefulShutdown` | 主控类，协调所有回调和管理器的执行流程          |

### 2. 接口定义

#### `ShutdownCallback`

```go
type ShutdownCallback interface {
	 OnShutdown(managerName string) error
}
```

#### `ShutdownFunc`（辅助类型）

```go
type ShutdownFunc func(managerName string) error
```

#### `ShutdownManager`

```go
type ShutdownManager interface {
    GetName() string
    Start(gs GSInterface) error
    ShutdownStart() error=
    ShutdownFinish() error
}
```

#### `GSInterface`

```go
type GSInterface interface {
	StartShutdown(manager ShutdownManager)
	ReportError(err error)
	AddShutdownCallback(cb ShutdownCallback)
}
```



## 🛠️ 使用方法

### 1. 初始化 `GracefulShutdown`

```go
gs := shutdown.New()
```



### 2. 添加关闭管理器（例如 POSIX 信号监听）

```go
gs.AddShutdownManager(posixsignal.NewPosixSignalManager())
```

> 后续还可以添加 AWS、K8s 等管理器 

### 3. 添加关闭回调函数

```go
gs.AddShutdownCallback(shutdown.ShutdownFunc(func(managerName string) error {
	fmt.Println("正在保存状态数据...")
	time.Sleep(500 * time.Millisecond)
	fmt.Println("状态数据已保存")
	return nil
}))
```



### 4. 设置错误处理器（可选）

```go
gs.SetErrorHandler(func(err error) {
	log.Printf("发生错误: %v", err)
})
```



### 5. 启动监听

```go
if err := gs.Start(); err != nil {
	panic(err)
}
```



## 🔁 执行流程详解

当程序收到关闭信号（如 `Ctrl+C` 或 AWS 发送终止消息）后，会按如下顺序执行：

1. 触发 `ShutdownManager` 的 `StartShutdown`
2. 调用 `ShutdownStart()` 方法（通常为空或打印日志）
3. 并发执行所有 `ShutdownCallback.OnShutdown()`
4. 所有回调完成后，调用 `ShutdownFinish()`
5. 程序退出

## 🔌 可扩展性说明

本框架具有良好的可扩展性，你可以轻松添加新的 `ShutdownManager` 实现，比如：

- `awsmanager`: 监听 AWS SQS 消息并配合 Auto Scaling API
- `k8smanager`: 监听 Kubernetes 的 `/shutdown` 健康检查端点
- `httpmanager`: 通过 HTTP 请求手动触发关闭流程

每个新功能只需实现 `ShutdownManager` 接口，然后通过 `AddShutdownManager()` 注册即可。

## 📌 最佳实践建议

| 场景       | 建议做法                                                 |
| ---------- | -------------------------------------------------------- |
| 开发调试   | 启用`PosixSignalManager`，便于本地测试                   |
| 生产环境   | 结合`AwsManager`或`K8sManager`实现自动扩缩容下的优雅关机 |
| 日志记录   | 在`ErrorHandler`中加入日志记录                           |
| 错误恢复   | 在回调中实现重试机制或回滚逻辑                           |
| 多回调顺序 | 不依赖回调执行顺序，确保幂等性                           |

## 📚 总结

这个 `shutdown` 模块提供了一种 **灵活、可扩展、结构清晰的优雅关机机制** ，适用于各种 Go 项目，特别是需要在关闭前执行清理逻辑的场景。它不仅提高了系统的稳定性，还为未来接入更多平台（如 AWS、Kubernetes）提供了良好的扩展基础。

## 📢 后续计划（可选）

如果你希望继续完善该模块，可以考虑以下方向：

| 功能            | 描述                                   |
| --------------- | -------------------------------------- |
| 超时控制        | 如果回调超过一定时间仍未完成，强制退出 |
| 日志模块集成    | 将关键事件写入日志系统                 |
| Prometheus 指标 | 记录每次关闭事件的耗时和成功率         |
| 自动测试        | 编写单元测试验证各模块行为             |
| 文档生成        | 使用 godoc 生成在线文档                |