package hdrImpl

import (
	"sync"

	"github.com/HdrHistogram/hdrhistogram-go"
)

// Recorder 结构体，用于记录延迟
type Recorder struct {
	histogramC *hdrhistogram.Histogram
	histogramB *hdrhistogram.Histogram
	mu         sync.RWMutex
}

// NewRecorder 创建一个新的 Recorder
func NewRecorder(timeoutMicros int64) *Recorder {
	// 创建一个新的 HdrHistogram
	return &Recorder{
		histogramC: hdrhistogram.New(1, timeoutMicros, 5),
		histogramB: hdrhistogram.New(1, timeoutMicros, 5),
	}
}

// RecordValue 记录延迟值
func (r *Recorder) RecordValue(value int64) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	err := r.histogramC.RecordValue(value)
	if err != nil {
		return
	}
}

// GetIntervalHistogram 获取当前的直方图
func (r *Recorder) GetIntervalHistogram() *hdrhistogram.Histogram {
	r.mu.Lock()
	defer r.mu.Unlock()
	tmp := r.histogramC
	r.histogramB.Reset()        // 重置 histogramB
	r.histogramC = r.histogramB // 将 histogramC 指向 histogramB
	r.histogramB = tmp          // 将 histogramB 指向原来的 histogramC
	return tmp
}
