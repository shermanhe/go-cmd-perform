package worker

import (
	"context"
	"perform-cli-framework-go/src/conf"
	"perform-cli-framework-go/src/logger"
	"time"
)

type ExampleWorker struct {
}

func (w *ExampleWorker) NewInstance() Worker {
	return &ExampleWorker{}
}

func (w *ExampleWorker) DefaultConfig() string {
	return "{}"
}

func (w *ExampleWorker) Clone() Worker {
	// 深浅取决于需求，思路是先实例化一个worker，在SetupGlobal后使用该实例进行克隆
	return &ExampleWorker{}
}

func (w *ExampleWorker) Setup(data *conf.GoData) error {
	return nil
}
func (w *ExampleWorker) SetupGlobal(c context.Context, config conf.BenchConfig) error {
	// 这里传入worker的自定义配置，从命令行传入或者从web控制台传入进来的
	wc := config.WorkerConfig
	logger.Info("Worker Config: %s", wc)
	return nil
}

func (w *ExampleWorker) PostGlobal(c context.Context, config conf.BenchConfig) error {
	return nil
}

func (w *ExampleWorker) DoWorker(data *conf.GoData) error {
	// use anyData
	time.Sleep(time.Millisecond * 5)
	// 发送的字节数
	data.StaterI.RecordBytes(10, true)
	// 接受的字节数
	data.StaterI.RecordBytes(20, false)
	return nil
}

func (w *ExampleWorker) Post(data *conf.GoData) {
}
