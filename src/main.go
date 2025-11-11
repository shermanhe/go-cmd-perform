package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"perform-cli-framework-go/src/conf"
	"perform-cli-framework-go/src/logger"
	"perform-cli-framework-go/src/runner"
	"perform-cli-framework-go/src/service"
	"perform-cli-framework-go/src/utils"
	"perform-cli-framework-go/src/worker"
	"runtime"
	"syscall"
)

var cfg conf.BenchConfig
var pressCount int

// parseArg 解析命令行参数
func parseArg() error {
	// 定义支持短参数的命令行参数
	flag.Int64Var(&cfg.Workers, "w", 10, "Number of workers")
	flag.Int64Var(&cfg.Duration, "d", 600, "Test duration in seconds")
	flag.Int64Var(&cfg.Timeout, "t", 30, "Request timeout in seconds")
	flag.Int64Var(&cfg.Rate, "r", 500, "Request rate per second")
	flag.Int64Var(&cfg.Nums, "s", 0, "Request nums")
	flag.BoolVar(&cfg.PError, "p", false, "Print error details or not")
	flag.BoolVar(&cfg.GrpcCfg.Enable, "D", false, "Start with grpc server")
	flag.IntVar(&cfg.GrpcCfg.Port, "P", 5052, "Grpc server port")
	flag.StringVar(&cfg.GrpcCfg.RegistrationCtEndpoint, "R", "", "The remote controller endpoint")
	flag.StringVar(&cfg.GrpcCfg.GroupName, "G", "", "Executor group name")
	flag.StringVar(&cfg.WorkerName, "n", "", "Executor worker name")
	flag.StringVar(&cfg.GrpcCfg.Name, "N", "", "Executor name")
	flag.StringVar(&cfg.GrpcCfg.LocalIp, "L", "", "Local IP address")
	flag.StringVar(&cfg.WorkerConfig, "c", "{}", "Worker config value")
	flag.BoolVar(&cfg.ListWorker, "list_worker", false, "Print supported workers")
	flag.Parse()
	if cfg.ListWorker {
		for wN, w := range worker.GetAllWorkers() {
			// 获取默认配置的 JSON 字符串
			jsonConfig := w.DefaultConfig()

			// 定义一个变量来存储格式化后的 JSON
			var prettyJSON string
			var jsonData interface{}

			// 将 JSON 字符串解码为一个通用接口
			err := json.Unmarshal([]byte(jsonConfig), &jsonData)
			if err != nil {
				logger.Error("Error unmarshalling JSON: %v", err)
				continue
			}

			// 使用 MarshalIndent 格式化 JSON
			prettyJSONBytes, err := json.MarshalIndent(jsonData, "", "    ") // 4个空格的缩进
			if err != nil {
				logger.Error("Error formatting JSON: %v", err)
				continue
			}
			prettyJSON = string(prettyJSONBytes)

			// 记录工作名称和格式化后的 JSON
			logger.Info("Worker: %s, with default config = %s", wN, prettyJSON)
		}
		os.Exit(0)
	}
	// 检查配置
	if err := checkConfig(); err != nil {
		return err
	}
	return nil
}

// checkConfig 检查配置的有效性
func checkConfig() error {
	if cfg.WorkerName == "" {
		return fmt.Errorf("you must specify an executor worker name")
	}
	if cfg.GrpcCfg.Enable {
		if cfg.GrpcCfg.Port <= 0 {
			return fmt.Errorf("invalid gRPC server port: %d", cfg.GrpcCfg.Port)
		}
		if cfg.GrpcCfg.RegistrationCtEndpoint == "" {
			return fmt.Errorf("you must specify a remote controller endpoint with -R when using gRPC")
		}
		if cfg.GrpcCfg.LocalIp == "" {
			return fmt.Errorf("you must specify a local IP address with -L when using gRPC")
		}
		if cfg.GrpcCfg.GroupName == "" {
			return fmt.Errorf("you must specify a group name with -G when using gRPC")
		}
		if cfg.GrpcCfg.Name == "" {
			return fmt.Errorf("you must specify an executor name with -N when using gRPC")
		}
	}
	return nil
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU() + 2)
	reg := utils.NewRegistrationUtils()
	err := parseArg()
	if err != nil {
		logger.Fatal("Parse args err: %v", err)
	}
	benchmarkRunner := runner.NewBenchRunner(cfg.Timeout * 1000 * 1000)
	signal.Ignore(os.Interrupt, syscall.SIGINT)
	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGALRM)
	go func() {
		for {
			select {
			case si := <-sigs:
				pressCount++
				if pressCount > 1 {
					logger.Warning("recev signal %s, force to exit\n", si)
					cancel()
					return
				} else {
					logger.Info("recev signal %s, waiting to exit, press Ctr + c force exit\n", si)
					benchmarkRunner.Stop()
					err := reg.Unregister()
					if err != nil {
						logger.Error("Unregister with err: %v\n", err)
					}
					service.StopGrpcServer()
				}
			}
		}
	}()
	if cfg.GrpcCfg.Enable {
		w := worker.NewWorker(cfg.WorkerName)
		cfg.WorkerConfig = w.DefaultConfig()
		err = reg.Register(cfg)
		if err != nil {
			logger.Fatal("Can not connect to remote ctl %s with err: %v", cfg.GrpcCfg.RegistrationCtEndpoint, err)
		}
		err := service.StartGrpcServer(cfg.GrpcCfg.Port, benchmarkRunner)
		if err != nil {
			logger.Fatal("Start grpc failed with err: %v", err)
		}
	} else {
		err := benchmarkRunner.Start(ctx, cfg)
		if err != nil {
			logger.Fatal("Run benchmark err: %v", err)
		}
	}
}
