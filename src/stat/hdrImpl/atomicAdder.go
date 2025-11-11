package hdrImpl

import "sync/atomic"

type AtomicAdder struct {
	value int64
}

// GetThenReset 返回当前值并重置为 0
func (a *AtomicAdder) GetThenReset() int64 {
	for {
		// 获取当前值
		currentValue := atomic.LoadInt64(&a.value)
		// 尝试将值重置为 0
		if atomic.CompareAndSwapInt64(&a.value, currentValue, 0) {
			return currentValue // 如果成功，返回当前值
		}
		// 如果 CAS 失败，说明有其他 goroutine 修改了值，继续循环
	}
}

// Add 增加指定的值
func (a *AtomicAdder) Add(delta int64) {
	atomic.AddInt64(&a.value, delta)
}
