package conf

import (
	"context"
	"perform-cli-framework-go/src/stat"

	"go.uber.org/ratelimit"
)

type GrpcConf struct {
	Enable                 bool
	Port                   int
	Name                   string
	RegistrationCtEndpoint string
	LocalIp                string
	GroupName              string
}

type BenchConfig struct {
	Workers      int64  `json:"workers"`
	Duration     int64  `json:"duration"`
	Timeout      int64  `json:"timeout"`
	Rate         int64  `json:"rate"`
	Nums         int64  `json:"nums"`
	PError       bool   `json:"pError"`
	WorkerName   string `json:"workerName"`
	WorkerConfig string `json:"workerConfig"`
	ListWorker   bool
	GrpcCfg      GrpcConf `json:"-"`
}

type GoData struct {
	Cfg         BenchConfig
	RateLimiter *ratelimit.Limiter
	Ctx         context.Context
	StaterI     stat.Stater
	SendTotal   int64
}
