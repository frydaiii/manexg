package vietnam

import (
	"sync"

	"github.com/banbox/banexg"
)

type Vietnam struct {
	*banexg.Exchange

	token     string
	tokenExp  int64
	tokenLock sync.RWMutex

	lookupLock      sync.RWMutex
	marketsByRawID  map[string]*banexg.Market
	marketsByTicker map[string][]*banexg.Market

	wsConnected bool
}
