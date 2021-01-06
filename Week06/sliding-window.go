package Week06

import (
	"container/ring"
	"sync"
	"time"
)

type metric struct {
	Success   int64
	Failure   int64
	Timeout   int64
	Rejection int64
}

type RollingCount struct {
	buckets *ring.Ring // 桶
	lock    *sync.RWMutex
	current int64   // 当前桶的秒数
	metric  *metric // 指标
}

func NewRollingCount() *RollingCount {
	r := &RollingCount{
		current: time.Now().Unix(),
		buckets: ring.New(10),
		metric:  &metric{},
	}
	for i := 0; i < r.buckets.Len(); i++ {
		r.buckets.Value = &metric{}
		r.buckets = r.buckets.Next()
	}

	return r
}

func (r *RollingCount) decreMetric(b *metric) {
	r.metric.Success -= b.Success
	r.metric.Failure -= b.Failure
	r.metric.Timeout -= b.Timeout
	r.metric.Rejection -= b.Rejection
}

func (r *RollingCount) updateTime() {
	now := time.Now().Unix()
	r.lock.Lock()
	defer r.lock.Unlock()

	if now-r.current >= int64(10) { // 当前时间窗口已失效，清除当前窗口内数据
		for i := 0; i < r.buckets.Len(); i++ {
			r.buckets.Value = &metric{}
			r.buckets = r.buckets.Next()
		}
		r.metric = &metric{}
		r.current = now
		return
	} else if now-r.current > int64(0) { // 清除已失效桶内的数据
		for i := int64(0); i < now-r.current; i++ {
			r.buckets = r.buckets.Prev()
			b := r.buckets.Value.(*metric)
			r.decreMetric(b)
			r.buckets.Value = &metric{}
		}
		r.current = now
		return
	}
	// 还在当前桶内
	return
}

func (r *RollingCount) GetMetric() metric {
	r.updateTime()

	r.lock.RLock()
	defer r.lock.RUnlock()

	return *r.metric // 不能返回指针
}

func (r *RollingCount) IncreSuccess() {
	r.updateTime()

	r.lock.Lock()
	defer r.lock.Unlock()

	r.metric.Success++
	r.buckets.Value.(*metric).Success++
}

func (r *RollingCount) IncreFailure() {
	r.updateTime()

	r.lock.Lock()
	defer r.lock.Unlock()

	r.metric.Failure++
	r.buckets.Value.(*metric).Failure++
}

func (r *RollingCount) IncreTimeout() {
	r.updateTime()

	r.lock.Lock()
	defer r.lock.Unlock()

	r.metric.Timeout++
	r.buckets.Value.(*metric).Timeout++
}

func (r *RollingCount) IncreRejection() {
	r.updateTime()

	r.lock.Lock()
	defer r.lock.Unlock()

	r.metric.Rejection++
	r.buckets.Value.(*metric).Rejection++
}
