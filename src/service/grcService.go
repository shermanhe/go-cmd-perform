package service

import (
	"context"
	"encoding/json"
	"fmt"
	"perform-cli-framework-go/src/conf"
	"perform-cli-framework-go/src/logger"
	"perform-cli-framework-go/src/runner"

	"google.golang.org/grpc"

	"net"
	"perform-cli-framework-go/src/perform_pb"
)

var grpcServer *grpc.Server

type server struct {
	perform_pb.UnsafePerformServiceServer
	Runner *runner.BenchMarkRunner
}

// StartPerform 实现 PerformService 的 StartPerform 方法
func (s *server) StartPerform(ctx context.Context, req *perform_pb.StartMessage) (*perform_pb.CmRespMessage, error) {
	// 处理 StartPerform 请求
	jsonBodyStr := req.GetJson()
	benchConfig := conf.BenchConfig{}
	err := json.Unmarshal(jsonBodyStr, &benchConfig)
	if err != nil {
		return &perform_pb.CmRespMessage{Code: -1, Message: []byte(fmt.Sprintf("Failed parse json str: %s",
			jsonBodyStr))}, nil
	}
	// 这里要标记下时grpc服务启动的
	benchConfig.GrpcCfg.Enable = true
	logger.Info("Start benchmark with conf: %v", benchConfig)
	s.Runner.StartAsync(benchConfig)
	return &perform_pb.CmRespMessage{Code: 0, Message: []byte("success")}, nil
}

// StopPerform 实现 PerformService 的 StopPerform 方法
func (s *server) StopPerform(ctx context.Context, req *perform_pb.EmptyMessage) (*perform_pb.PerformMessage, error) {
	// 处理 StopPerform 请求
	logger.Info("Received StopPerform request")
	s.Runner.Stop()
	return &perform_pb.PerformMessage{Code: 0}, nil
}

// CollectStats 实现 PerformService 的 CollectStats 方法
func (s *server) CollectStats(ctx context.Context, req *perform_pb.EmptyMessage) (*perform_pb.PerformMessage, error) {
	// 处理 CollectStats 请求
	statistic := s.Runner.GetStatistics()
	if statistic == nil {
		return &perform_pb.PerformMessage{Code: -1}, nil
	}
	records := make([]*perform_pb.Record, 0)
	for _, v := range statistic.Records {
		records = append(records, &perform_pb.Record{
			Key:   v.Key,
			Value: v.Value,
		})
	}
	stats := &perform_pb.PerformStats{
		Duration:  statistic.Durations,
		ErrCount:  statistic.ErrorTotal,
		SendCount: statistic.SendTotal,
		SendBytes: statistic.SendBytes,
		RecvBytes: statistic.RecvBytes,
		Latency:   records,
		ErrMsgs:   make([][]byte, 0),
	}
	return &perform_pb.PerformMessage{Code: 0, Stats: stats}, nil
}

// KeepAlive 实现 PerformService 的 KeepAlive 方法
func (s *server) KeepAlive(ctx context.Context, req *perform_pb.EmptyMessage) (*perform_pb.ExecutorStatus, error) {
	// 处理 KeepAlive 请求
	if s.Runner.IsRunning() {
		return &perform_pb.ExecutorStatus{Status: perform_pb.Status_STATUS_RUNNING}, nil
	}
	return &perform_pb.ExecutorStatus{Status: perform_pb.Status_STATUS_IDLE}, nil
}

func StartGrpcServer(port int, r *runner.BenchMarkRunner) error {
	// 创建 gRPC 服务器
	grpcServer = grpc.NewServer()
	perform_pb.RegisterPerformServiceServer(grpcServer, &server{Runner: r})

	// 监听端口
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	logger.Info("Starting gRPC server on :%d", port)
	if err := grpcServer.Serve(lis); err != nil {
		return err
	}
	return nil
}

func StopGrpcServer() {
	if grpcServer != nil {
		grpcServer.Stop()
	}
}
