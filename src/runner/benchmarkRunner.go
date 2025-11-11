package runner

import (
	"context"
	"errors"
	"perform-cli-framework-go/src/conf"
	"perform-cli-framework-go/src/logger"
	"perform-cli-framework-go/src/stat"
	"perform-cli-framework-go/src/stat/hdrImpl"
	"perform-cli-framework-go/src/utils"
	"perform-cli-framework-go/src/worker"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/ratelimit"
)

type BenchMarkRunner struct {
	running   atomic.Bool
	sendM     sync.Mutex
	startM    sync.Mutex
	sendCount int64
	stater    stat.Stater
	call      *context.CancelFunc
	c         *stat.IntervalStatistic
}

func NewBenchRunner(t int64) *BenchMarkRunner {
	return &BenchMarkRunner{
		stater: hdrImpl.New(t),
	}
}

func (b *BenchMarkRunner) IsRunning() bool {
	return b.running.Load() == true
}

func (b *BenchMarkRunner) CachedStatistics() *stat.IntervalStatistic {
	return b.c
}

func (b *BenchMarkRunner) GetStatistics() *stat.IntervalStatistic {
	b.c = b.stater.GetIntervalStatistic()
	return b.c
}

func (b *BenchMarkRunner) mainLoop(c context.Context, data *conf.GoData, r ratelimit.Limiter, workerHand worker.Worker, wait *sync.WaitGroup) {
	// 获取自己实现的Worker
	defer func() {
		defer func() {
			if p := recover(); p != nil {
				logger.Error("Worker with unknown error: %v,exit benchmark", p)
				b.Stop()
			}
		}()
		workerHand.Post(data)
		wait.Done()
	}()
	err := workerHand.Setup(data)
	if err != nil {
		logger.Error("Setup err: %v", err)
		return
	}
	for b.running.Load() {
		select {
		case <-c.Done():
			return
		default:
			if data.Cfg.Rate > 0 && data.RateLimiter != nil {
				r.Take()
			}
			if data.Cfg.Nums > 0 {
				b.sendM.Lock()
				if b.sendCount >= data.Cfg.Nums {
					b.sendM.Unlock()
					return
				}
				b.sendCount++
				b.sendM.Unlock()
			}
			err1 := workerHand.DoWorker(data)
			if err1 == worker.ExitError {
				return
			}
		}
	}
}

func (b *BenchMarkRunner) StartAsync(config conf.BenchConfig) {
	ctx, cc := context.WithCancel(context.Background())
	b.call = &cc
	go func(config conf.BenchConfig) {
		// 服务形式的不能挂
		defer func() {
			if p := recover(); p != nil {
				logger.Error("Run with Fatal err: %v", p)
				b.running.Store(false)
			}
		}()
		err := b.Start(ctx, config)
		if err != nil {
			logger.Error("Run benchmark err: %v", err)
		}
	}(config)
}

// Start 启动压测
func (b *BenchMarkRunner) Start(ctx context.Context, cfg conf.BenchConfig) error {
	worker.ResetStatus()
	b.startM.Lock()
	start := utils.GetTimeUs()
	goDataS := make([]*conf.GoData, cfg.Workers)
	var r ratelimit.Limiter
	if cfg.Rate > 0 {
		r = ratelimit.New(int(cfg.Rate))
	}
	var wait sync.WaitGroup
	wait.Add(int(cfg.Workers))
	b.sendCount = 0
	workerHand := worker.NewWorker(cfg.WorkerName)
	if b.running.Load() {
		// 已经在执行则释放锁退出
		logger.Warning("Benchmark started, ignore")
		b.startM.Unlock()
		return errors.New("benchmark started, ignore")
	}
	b.running.Store(true)
	b.startM.Unlock()
	// 执行全局前置
	err := workerHand.SetupGlobal(ctx, cfg)
	if err != nil {
		return err
	}
	defer func() {
		b.running.Store(false)
		b.stater.Reset()
	}()
	defer func() {
		// 全局后置
		err := workerHand.PostGlobal(ctx, cfg)
		if err != nil {
			logger.Error("Run PostGlobal err: %v", err)
			return
		}
	}()
	for i := int64(0); i < cfg.Workers; i++ {
		data := &conf.GoData{
			Cfg:         cfg,
			RateLimiter: &r,
			SendTotal:   0,
		}
		// 这里这个context是用来做强制退出的的一般网络库都会一个ctx给客户端做主动退出
		data.Ctx = ctx
		goDataS[i] = data
		data.StaterI = b.stater
		// 保证全局初始化的数据全部传输成功
		tmpW := workerHand.Clone()
		go b.mainLoop(ctx, goDataS[i], r, tmpW, &wait)
	}
	if cfg.Nums > 0 {
		logger.Info("Running %d Nums test @%d", cfg.Nums, start)
	} else {
		logger.Info("Running %d s test @%d", cfg.Duration, start)
	}
	logger.Info("  %d goroutines", cfg.Workers)
	go func() { // 若是没有指定发送的数据就指定时间
		if cfg.Nums == 0 {
			time.Sleep(time.Duration(cfg.Duration) * time.Second)
			b.Stop()
		}
	}()
	// 启动一个协程打印临时的压测统计
	go func() {
		for b.running.Load() {
			time.Sleep(time.Second * 30)
			// 未启动grpc服务的时候，自己打印压测统计到控制台
			var ss *stat.IntervalStatistic
			if !cfg.GrpcCfg.Enable {
				ss = b.GetStatistics()
			} else {
				ss = b.CachedStatistics()
			}
			if ss != nil {
				ss.LogSelf()
			}
		}
	}()
	wait.Wait()
	complete := int64(0)
	for _, v := range goDataS {
		complete += v.SendTotal
	}
	runtimeUs := utils.GetTimeUs() - start
	runtimeS := runtimeUs / 1000000.0
	if runtimeS == 0 {
		// 没有运行1s按1s算
		runtimeS = 1
	}
	reqPerS := complete / runtimeS
	logger.Info("Complete request: %d", complete)
	logger.Info("Test durations: %d s", runtimeS)
	logger.Info("Requests/sec: %d", reqPerS)
	logger.Info("Bye. @%d", utils.GetTimeUs())
	return nil
}

func (b *BenchMarkRunner) Stop() {
	b.running.Store(false)
	time.Sleep(1 * time.Second)
	if b.call != nil {
		(*b.call)()
		b.call = nil
	}
}
