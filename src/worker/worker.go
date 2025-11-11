package worker

import (
	"context"
	"errors"
	"fmt"
	"os"
	"perform-cli-framework-go/src/conf"
	"perform-cli-framework-go/src/logger"
	"perform-cli-framework-go/src/utils"
	"sync"
)

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

var m = sync.Mutex{}
var b = false
var f = false

func ResetStatus() {
	f = false
	b = false
}

var ExitError = errors.New("exit worker")

type Proxy struct {
	workerHandler Worker
}

func (w *Proxy) SetupGlobal(c context.Context, config conf.BenchConfig) error {
	m.Lock()
	defer func() {
		m.Unlock()
	}()
	if !b {
		err := w.workerHandler.SetupGlobal(c, config)
		if err != nil {
			return err
		}
		b = true
	}
	return nil
}

func (w *Proxy) PostGlobal(c context.Context, config conf.BenchConfig) error {
	m.Lock()
	defer func() {
		m.Unlock()
	}()
	if !f {
		err := w.workerHandler.PostGlobal(c, config)
		if err != nil {
			return err
		}
		f = true
	}
	return nil
}

func (w *Proxy) DefaultConfig() string {
	return w.workerHandler.DefaultConfig()
}

func (w *Proxy) Setup(data *conf.GoData) error {
	return w.workerHandler.Setup(data)
}

func (w *Proxy) DoWorker(data *conf.GoData) error {
	defer func() {
		data.SendTotal++
		if p := recover(); p != nil {
			data.StaterI.RecordErr(fmt.Sprintf("do work err %v", p))
		}
	}()
	begin := utils.GetTimeUs()
	err := w.workerHandler.DoWorker(data)
	if err != nil {
		if err == ExitError {
			return ExitError
		}
		if data.Cfg.PError {
			logger.Error("Do worker with err: %v", err)
		}
		data.StaterI.RecordErr(fmt.Sprintf("do work err %v", err))
		return nil
	}
	after := utils.GetTimeUs()
	data.StaterI.AddLatency(after - begin)
	return nil
}

func (w *Proxy) Post(data *conf.GoData) {
	w.workerHandler.Post(data)
}

func (w *Proxy) NewInstance() Worker {
	return w.NewInstance()
}

func (w *Proxy) Clone() Worker {
	return &Proxy{
		workerHandler: w.workerHandler.Clone(),
	}
}

// NewWorker 根据需要返回实现得worker
func NewWorker(name string) *Proxy {
	var w Worker
	if value, ok := workers[name]; ok {
		w = value
	} else {
		fmt.Printf("can not found worker %s\n", name)
		os.Exit(-1)
	}
	return &Proxy{
		workerHandler: w.NewInstance(),
	}
}

func GetAllWorkers() map[string]Worker {
	return workers
}
