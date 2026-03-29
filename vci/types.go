package vci

import (
	"sync"

	"github.com/banbox/banexg"
)

type VCI struct {
	*banexg.Exchange

	lookupLock      sync.RWMutex
	marketsByRawID  map[string]*banexg.Market
	marketsByTicker map[string][]*banexg.Market
}
