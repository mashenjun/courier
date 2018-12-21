package service

import (
	"context"
	"github.com/go-kit/kit/metrics"
	"time"
)
// InstrumentMiddleware provide metric related value
type InstrumentMiddleware struct {
	RequestCount   metrics.Counter
	RequestLatency metrics.Histogram
	Next           Service
}


func (mw InstrumentMiddleware) Ping(ctx context.Context, _ struct{}) (*PingResp, error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "ping"}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return mw.Next.Ping(ctx, struct{}{})
}

func (mw InstrumentMiddleware) Subscribe(ctx context.Context, req SubscribeReq) (*SubscribeResp, error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "subscribe"}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return mw.Next.Subscribe(ctx, req)
}

func (mw InstrumentMiddleware) Send(ctx context.Context, req SendReq) (*SendResp, error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "send"}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return mw.Next.Send(ctx, req)
}

func (mw InstrumentMiddleware) Close(ctx context.Context, req CloseReq)(*CloseResp, error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "close"}
		mw.RequestCount.With(lvs...).Add(1)
		mw.RequestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return mw.Next.Close(ctx, req)
}

