package hdrImpl

import (
	"perform-cli-framework-go/src/logger"
	"perform-cli-framework-go/src/stat"
	"perform-cli-framework-go/src/utils"
	"sync/atomic"

	"github.com/HdrHistogram/hdrhistogram-go"
)

type HdrHistogramStat struct {
	IntervalSendBytes AtomicAdder
	IntervalRecvBytes AtomicAdder
	SendTotal         atomic.Int64
	SendErr           atomic.Int64
	HdrHistogram      *hdrhistogram.Histogram
	Recorder          *Recorder
	timePoint         int64
}

func New(timeUs int64) *HdrHistogramStat {

	return &HdrHistogramStat{
		IntervalSendBytes: AtomicAdder{value: 0},
		IntervalRecvBytes: AtomicAdder{value: 0},
		SendErr:           atomic.Int64{},
		SendTotal:         atomic.Int64{},
		HdrHistogram:      hdrhistogram.New(1, timeUs, 5),
		Recorder:          NewRecorder(timeUs),
		timePoint:         utils.GetTimeUs(),
	}
}

func (h *HdrHistogramStat) Reset() {
	h.IntervalRecvBytes.GetThenReset()
	h.IntervalRecvBytes.GetThenReset()
	h.SendTotal.Store(0)
	h.SendErr.Store(0)
	// 输出基础统计
	logger.Info("Latency Statistics (us):")
	logger.Info("  Min      : %d", h.HdrHistogram.Min())
	logger.Info("  Max      : %d", h.HdrHistogram.Max())
	logger.Info("  Mean     : %.2f", h.HdrHistogram.Mean())
	logger.Info("  StdDev   : %.2f", h.HdrHistogram.StdDev())

	// 输出分位数统计（P50, P90, P95, P99）
	logger.Info("Percentiles:")
	logger.Info("  P90    : %d", h.HdrHistogram.ValueAtPercentile(90.0))
	logger.Info("  P95    : %d", h.HdrHistogram.ValueAtPercentile(95.0))
	logger.Info("  P99    : %d", h.HdrHistogram.ValueAtPercentile(99.0))
	h.HdrHistogram.Reset()
}

func (h *HdrHistogramStat) AddLatency(latency int64) {
	h.SendTotal.Add(1)
	h.Recorder.RecordValue(latency)
	err := h.HdrHistogram.RecordValue(latency)
	if err != nil {
		return
	}
}

func (h *HdrHistogramStat) RecordBytes(value int64, isSend bool) {
	if isSend {
		h.IntervalSendBytes.Add(value)
	} else {
		h.IntervalRecvBytes.Add(value)
	}
}
func (h *HdrHistogramStat) RecordErr(errMsg string) {
	h.SendErr.Add(1)
}
func (h *HdrHistogramStat) GetIntervalStatistic() *stat.IntervalStatistic {
	// 获取一定时间间隔的统计数据
	hdr := h.Recorder.GetIntervalHistogram()
	sendTotal := h.SendTotal.Load()
	sendErr := h.SendErr.Load()
	recvBytes := h.IntervalRecvBytes.GetThenReset()
	sendBytes := h.IntervalSendBytes.GetThenReset()
	now := utils.GetTimeUs()
	d := now - h.timePoint
	h.timePoint = now
	records := make([]stat.Record, 0)
	sn := hdr.Export()
	for k, v := range sn.Counts {
		if v == 0 {
			continue
		}
		records = append(records, stat.Record{
			Key:   int64(k),
			Value: v,
		})
	}
	return &stat.IntervalStatistic{
		SendTotal:  sendTotal,
		SendBytes:  sendBytes,
		ErrorTotal: sendErr,
		RecvBytes:  recvBytes,
		Durations:  d,
		Records:    records,
	}
}
