package vietnam

import (
	"context"
	"sync"
	"testing"

	"github.com/banbox/banexg"
)

var (
	vietnamTestReqMu sync.Mutex
	vietnamTestReqFn func(ctx context.Context, endpoint string, params map[string]interface{}, retryNum int, readCache, writeCache bool) *banexg.HttpRes
)

func setVietnamTestRequest(t *testing.T, fn func(ctx context.Context, endpoint string, params map[string]interface{}, retryNum int, readCache, writeCache bool) *banexg.HttpRes) {
	t.Helper()
	vietnamTestReqMu.Lock()
	vietnamTestReqFn = fn
	vietnamTestReqMu.Unlock()
	t.Cleanup(func() {
		vietnamTestReqMu.Lock()
		vietnamTestReqFn = nil
		vietnamTestReqMu.Unlock()
	})
}

func (e *Vietnam) RequestApiRetryAdv(ctx context.Context, endpoint string, params map[string]interface{}, retryNum int, readCache, writeCache bool) *banexg.HttpRes {
	vietnamTestReqMu.Lock()
	fn := vietnamTestReqFn
	vietnamTestReqMu.Unlock()
	if fn != nil {
		return fn(ctx, endpoint, params, retryNum, readCache, writeCache)
	}
	return e.Exchange.RequestApiRetryAdv(ctx, endpoint, params, retryNum, readCache, writeCache)
}
