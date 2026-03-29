package vci

import (
	"context"
	"sync"
	"testing"

	"github.com/banbox/banexg"
)

var (
	vciTestReqMu sync.Mutex
	vciTestReqFn func(ctx context.Context, endpoint string, params map[string]interface{}, retryNum int, readCache, writeCache bool) *banexg.HttpRes
)

func setVCITestRequest(t *testing.T, fn func(ctx context.Context, endpoint string, params map[string]interface{}, retryNum int, readCache, writeCache bool) *banexg.HttpRes) {
	t.Helper()
	vciTestReqMu.Lock()
	vciTestReqFn = fn
	vciTestReqMu.Unlock()
	t.Cleanup(func() {
		vciTestReqMu.Lock()
		vciTestReqFn = nil
		vciTestReqMu.Unlock()
	})
}

func (e *VCI) RequestApiRetryAdv(ctx context.Context, endpoint string, params map[string]interface{}, retryNum int, readCache, writeCache bool) *banexg.HttpRes {
	vciTestReqMu.Lock()
	fn := vciTestReqFn
	vciTestReqMu.Unlock()
	if fn != nil {
		return fn(ctx, endpoint, params, retryNum, readCache, writeCache)
	}
	return e.Exchange.RequestApiRetryAdv(ctx, endpoint, params, retryNum, readCache, writeCache)
}
