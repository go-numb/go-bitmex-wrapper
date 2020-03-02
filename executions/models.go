package executions

import (
	"math"
	"sync"
	"time"

	"github.com/go-numb/go-bitmex"
)

type Execution struct {
	sync.RWMutex

	isBuy  bool
	length int

	price float64
	ltp   float64
	ask   float64
	bid   float64

	buySize  int
	sellSize int

	// 1配信の
	// 加重平均を求めるため、intをfloatとして保持
	prices  []float64
	volumes []float64

	// Losscuts
	Event chan interface{}

	delay time.Duration
}

// New is new Executes
func New() *Execution {
	return &Execution{

		prices:  make([]float64, 0),
		volumes: make([]float64, 0),

		Event: make(chan interface{}),
	}
}

func (p *Execution) received(in interface{}) {
	p.Event <- in
}

// Set price/ltp(before1ws), bestbid/ask, volume, delay
// benchmark: 7-25μs by Macbook Pro i7 2015 Late
func (p *Execution) Set(ex []bitmex.Trade) {
	p.Lock()
	defer p.Unlock()

	var (
		wg sync.WaitGroup
		l  = len(ex)
	)

	p.length = l
	// 1配信毎の Reset
	p.buySize = 0
	p.sellSize = 0
	p.prices = []float64{}
	p.volumes = []float64{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if l != 0 {
			p.delay = time.Now().Sub(ex[l-1].Timestamp)
		}
	}()

	wg.Add(1)
	go func() { // 約定を保存
		defer wg.Done()
		prices := make([]float64, len(ex))
		volumes := make([]float64, len(ex))
		for i := range ex { // EMAをつくる
			prices[i] = ex[i].Price
			volumes[i] = float64(ex[i].Size)
		}

		if len(prices) != len(volumes) {
			return
		}
		p.prices = prices
		p.volumes = volumes
	}()

	// == より、strings.HasPrefixのほうが早い
	wg.Add(1)
	go func() { // 約定ベースのBest値をとっていく
		defer wg.Done()

		// 一配信前の価格を退避
		p.ltp = p.price

		for i := range ex {
			if ex[i].Side == bitmex.BUY {
				// 配信内初回約定
				p.buySize += ex[i].Size
				p.best(true, ex[i].Price)

			} else if ex[i].Side == bitmex.SELL {
				// 配信内初回約定
				p.sellSize += ex[i].Size
				p.best(false, ex[i].Price)

			}
		}
	}()

	wg.Wait()
}

// HighPerformanceSet price/ltp(before1ws), bestbid/ask, volume, delay
// benchmark: 1.7-11μs by Macbook Pro i7 2015 Late
func (p *Execution) HighPerformanceSet(ex []bitmex.Trade) {
	p.Lock()
	defer p.Unlock()

	l := len(ex)
	p.length = l
	// 1配信毎の Reset
	p.buySize = 0
	p.sellSize = 0
	p.prices = []float64{}
	p.volumes = []float64{}

	if l != 0 {
		p.delay = time.Now().Sub(ex[l-1].Timestamp)
	}

	// 一配信前の価格を退避
	p.ltp = p.price

	prices := make([]float64, len(ex))
	volumes := make([]float64, len(ex))

	for i := range ex {
		prices[i] = ex[i].Price
		volumes[i] = float64(ex[i].Size)

		if ex[i].Side == bitmex.BUY {
			// 配信内初回約定
			p.buySize += ex[i].Size
			p.best(true, ex[i].Price)

		} else if ex[i].Side == bitmex.SELL {
			// 配信内初回約定
			p.sellSize += ex[i].Size
			p.best(false, ex[i].Price)

		}
	}

	if len(prices) != len(volumes) {
		return
	}
	p.prices = prices
	p.volumes = volumes
}

func (p *Execution) best(isBuy bool, price float64) {
	if !isBuy {
		p.price = price
		p.bid = price
		p.isBuy = false
		return
	}

	p.price = price
	p.ask = price
	p.isBuy = true
}

func (p *Execution) IsBuy() bool {
	p.RLock()
	defer p.RUnlock()
	return p.isBuy
}

func (p *Execution) Lenght() int {
	p.RLock()
	defer p.RUnlock()
	return p.length
}

func (p *Execution) LTP() float64 {
	p.RLock()
	defer p.RUnlock()
	return p.price
}

// Volume 1配信中の出来高
// 正の場合は買い成が強く、負の場合は売り成が強い
func (p *Execution) Volume() (sum, buy, sell int) {
	p.RLock()
	defer p.RUnlock()
	return p.buySize + p.sellSize, p.buySize, p.sellSize
}

func (p *Execution) Spread() float64 {
	p.RLock()
	defer p.RUnlock()
	return math.Max(0, p.ask-p.bid)
}

// Best get bestask and bestbid
func (p *Execution) Best() (ask, bid float64) {
	p.RLock()
	defer p.RUnlock()
	return p.ask, p.bid
}

func (p *Execution) Copy() (prices, volumes []float64) {
	p.RLock()
	defer p.RUnlock()
	return p.prices, p.volumes
}

func (p *Execution) Delay() time.Duration {
	p.RLock()
	defer p.RUnlock()
	return p.delay
}
