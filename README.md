# perform-cli-framework-go

## 项目概述

`perform-cli-framework-go` 是一个用 Go 语言编写的高性能基准测试框架，旨在提供一个灵活、可扩展的基准测试解决方案。该框架支持本地执行和远程 gRPC 控制，允许用户通过命令行参数或远程 API 来启动和管理基准测试。

## 主要特性

- **命令行参数支持**：通过命令行参数配置测试参数，如工作线程数、测试时长、请求速率等。
- **gRPC 服务器支持**：支持通过 gRPC 服务进行远程控制，包括启动、停止和收集统计信息。
- **插件式架构**：支持通过插件方式扩展不同的测试类型，每个测试类型实现特定的 `Worker` 接口。
- **统计信息收集**：提供详细的性能统计信息，包括延迟、吞吐量、错误率等。
- **配置管理**：支持通过 JSON 格式的配置文件进行配置。
- **错误处理**：提供详细的错误日志和错误处理机制。
- **并发控制**：支持多线程并发执行测试，提高测试效率。

## 项目结构

```
src/
├── conf/                # 配置文件相关
│   └── config.go        # 配置结构定义
├── logger/              # 日志相关
│   └── logger.go        # 日志功能实现
├── main.go              # 主程序入口
├── perform_pb/          # gRPC 协议定义
│   ├── perform_pb.go    # gRPC 服务定义
│   └── perform.proto    # gRPC 协议文件
├── runner/              # 基准测试执行器
│   └── benchmarkRunner.go # 基准测试执行逻辑
├── service/             # 服务相关
│   └── grcService.go    # gRPC 服务实现
├── stat/                # 统计信息相关
│   ├── stats.go         # 统计接口定义
│   └── hdrImpl/         # HDR 直方图实现
│       ├── hdrhistogramStat.go
│       ├── recorder.go
│       └── atomicAdder.go
├── utils/               # 工具函数
│   ├── timeUtils.go     # 时间工具函数
│   └── registrationUtils.go # 注册工具函数
└── worker/              # 工作器相关
    ├── init.go          # 工作器注册
    ├── worker.go        # 工作器接口定义
    └── exampleWorker.go # 示例工作器实现
```

## 使用方法

### 1. 安装依赖

```bash
go mod tidy
```

### 2. 编译项目

```bash
go build -o perform-cli-framework-go src/main.go
```

### 3. 运行基准测试

```bash
./perform-cli-framework-go -w 10 -d 600 -r 500 -s 1000
```

### 4. 启动 gRPC 服务器

```bash
./perform-cli-framework-go -w 10 -d 600 -r 500 -s 1000 -D -P 5052 -R 127.0.0.1:8080 -G test_group -N test_executor -L 127.0.0.1
```

### 5. 查看支持的工作器

```bash
./perform-cli-framework-go -list_worker
```

## 命令行参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-w` | 工作线程数 | 10 |
| `-d` | 测试时长（秒） | 600 |
| `-t` | 请求超时（秒） | 30 |
| `-r` | 请求速率（每秒） | 500 |
| `-s` | 请求总数 | 0 |
| `-p` | 是否打印错误详情 | false |
| `-D` | 是否启动 gRPC 服务器 | false |
| `-P` | gRPC 服务器端口 | 5052 |
| `-R` | 远程控制器端点 | 空 |
| `-G` | 执行器组名 | 空 |
| `-N` | 执行器名称 | 空 |
| `-L` | 本地 IP 地址 | 空 |
| `-n` | 执行器工作名称 | 空 |
| `-c` | 工作器配置值 | `{}` |
| `-list_worker` | 打印支持的工作器 | false |

## gRPC 服务

### 1. 启动服务

```bash
./perform-cli-framework-go -D -P 5052 -R 127.0.0.1:8080 -G test_group -N test_executor -L 127.0.0.1
```

### 2. API 调用

- **StartPerform**：启动基准测试
  - 请求：`{"json": "{\"workers\":10,\"duration\":600,\"rate\":500,\"nums\":0,\"pError\":false,\"grpcCfg\":{\"enable\":true,\"port\":5052,\"registrationCtEndpoint\":\"127.0.0.1:8080\",\"groupName\":\"test_group\",\"name\":\"test_executor\",\"localIp\":\"127.0.0.1\"},\"workerName\":\"ExampleWorker\"}"}`

- **StopPerform**：停止基准测试
  - 请求：`{}`

- **CollectStats**：收集统计信息
  - 请求：`{}`

- **KeepAlive**：保持连接
  - 请求：`{}`

## 插件式架构

### 1. 定义工作器接口

```go
type Worker interface {
    SetupGlobal(c context.Context, config conf.BenchConfig) error
    Setup(data *conf.GoData) error
    DoWorker(data *conf.GoData) error
    Post(data *conf.GoData)
    PostGlobal(c context.Context, config conf.BenchConfig) error
    NewInstance() Worker
    DefaultConfig() string
    Clone() Worker
}
```

### 2. 实现工作器

```go
type ExampleWorker struct{}

func (w *ExampleWorker) SetupGlobal(c context.Context, config conf.BenchConfig) error {
    // 全局初始化
    return nil
}

func (w *ExampleWorker) Setup(data *conf.GoData) error {
    // 初始化
    return nil
}

func (w *ExampleWorker) DoWorker(data *conf.GoData) error {
    // 执行测试
    return nil
}

func (w *ExampleWorker) Post(data *conf.GoData) {
    // 后置处理
}

func (w *ExampleWorker) PostGlobal(c context.Context, config conf.BenchConfig) error {
    // 全局后置处理
    return nil
}

func (w *ExampleWorker) NewInstance() Worker {
    return &ExampleWorker{}
}

func (w *ExampleWorker) DefaultConfig() string {
    return "{}"
}

func (w *ExampleWorker) Clone() Worker {
    return &ExampleWorker{}
}
```

### 3. 注册工作器

```go
func init() {
    workers["ExampleWorker"] = &ExampleWorker{}
}
```

## 统计信息

### 1. 统计信息结构

```go
type IntervalStatistic struct {
    Durations   []float64
    ErrorTotal  int64
    SendTotal   int64
    SendBytes   int64
    RecvBytes   int64
    Records     []*Record
}
```

### 2. 统计信息收集

- **延迟**：使用 HDR Histogram 收集延迟数据
- **吞吐量**：计算每秒请求数
- **错误率**：统计错误请求数

## 项目依赖

- `go.uber.org/ratelimit`：限速库
- `github.com/HdrHistogram/hdrhistogram-go`：HDR 直方图库
- `google.golang.org/grpc`：gRPC 库

## 开发者指南

### 1. 添加新的工作器

1. 创建新的工作器实现
2. 在 `init()` 函数中注册工作器
3. 实现 `Worker` 接口的所有方法

### 2. 添加新的统计信息

1. 在 `stat` 包中添加新的统计信息类型
2. 实现统计信息收集逻辑
3. 在 `IntervalStatistic` 结构中添加新的字段

### 3. 优化性能

1. 使用并发编程提高性能
2. 优化数据结构和算法
3. 使用缓存减少重复计算

## 贡献指南

1. 提交 Pull Request 前请确保代码风格一致
2. 添加必要的单元测试
3. 更新文档以反映新功能
4. 提供详细的变更说明

## 许可证

MIT License

## 联系方式

如有任何问题，请联系 [email protected]