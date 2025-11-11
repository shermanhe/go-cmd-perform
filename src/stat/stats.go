package stat

import (
	"fmt"
	"perform-cli-framework-go/src/logger"
	"sort"
)

type Record struct {
	Key   int64
	Value int64
}

type IntervalStatistic struct {
	SendTotal  int64
	ErrorTotal int64
	Durations  int64
	SendBytes  int64
	RecvBytes  int64
	Records    []Record
}

// LogSelf 打印统计数据
func (i *IntervalStatistic) LogSelf() {
	if len(i.Records) == 0 {
		return
	}
	// 1. 按时延（Key）从小到大排序
	sort.Slice(i.Records, func(a, b int) bool {
		return i.Records[a].Key < i.Records[b].Key
	})
	// 2. 计算总请求数（用于分位数）
	total := int64(0)
	for _, r := range i.Records {
		total += r.Value
	}
	// 3. 计算QPS（ Durations 单位为微秒）
	var qps float64
	intervalTotal := int64(0)
	for _, v := range i.Records {
		intervalTotal += v.Value
	}
	if i.Durations > 0 {
		qps = float64(intervalTotal) / (float64(i.Durations) / (1000 * 1000))
	}
	// 4. 错误率
	errorRate := 0.0
	if i.SendTotal > 0 {
		errorRate = float64(i.ErrorTotal) / float64(i.SendTotal) * 100
	}
	// 5. 计算分位数（P99/P95/P90）
	calculatePercentile := func(percentile float64) int64 {
		if total == 0 {
			return 0
		}
		target := float64(total) * percentile
		count := int64(0)
		for _, r := range i.Records {
			count += r.Value
			if float64(count) >= target {
				return r.Key
			}
		}
		return i.Records[len(i.Records)-1].Key // 兜底返回最大值
	}
	p99 := calculatePercentile(0.99)
	p95 := calculatePercentile(0.95)
	p90 := calculatePercentile(0.90)
	// 6. 格式化输出
	logger.Info(
		"[Stats] QPS: %.2f | Error: %.2f%% | Send: %s /s | Recv: %s /s | Latency (us) - P99: %d, P95: %d, P90: %d",
		qps,
		errorRate,
		formatBytes(float64(i.SendBytes)/float64((i.Durations)/(1000*1000))),
		formatBytes(float64(i.RecvBytes)/float64((i.Durations)/(1000*1000))),
		p99,
		p95,
		p90,
	)
}

// 辅助函数：格式化字节为易读单位（KB/MB/GB）
func formatBytes(b float64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%fB", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", b/float64(div), "KMGTPE"[exp])
}

type Stater interface {
	AddLatency(latency int64)
	RecordBytes(value int64, isSend bool)
	RecordErr(errMsg string)
	Reset()
	GetIntervalStatistic() *IntervalStatistic
}
