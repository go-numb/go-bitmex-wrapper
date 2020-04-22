package orders

import (
	"context"
	"math"

	"github.com/go-numb/go-bitmex"
)

const (
	MINSIZE = 0.0025
)

type TradePrice struct {
	Instruments map[string]bitmex.Instrument
}

func NewTradePrice() *TradePrice {
	c := bitmex.NewAPIClient(bitmex.NewConfiguration())
	o := &bitmex.InstrumentGetOpts{}
	o.Filter.Set(`{"state": "Open"}`)
	// o.StartTime.Set(time.Now().UTC().Add(-24 * 180 * time.Hour))
	// o.EndTime.Set(time.Now().UTC())
	inst, res, err := c.InstrumentApi.InstrumentGet(context.Background(), o)
	if err != nil {
		return nil
	}
	defer res.Body.Close()

	dict := make(map[string]bitmex.Instrument)
	for i := range inst {
		dict[inst[i].Symbol] = inst[i]
	}

	return &TradePrice{
		Instruments: dict,
	}
}

func (p *TradePrice) TickPrice(product string) float64 {
	tick, ok := p.Instruments[product]
	if !ok {
		return 0
	}
	return tick.TickSize
}

const SATOSHI = 100000000

func (p *TradePrice) Decimal(product string, price float64) float64 {
	tick, ok := p.Instruments[product]
	if !ok {
		return price
	}

	// half := tick.TickSize * 0.5

	n, f := math.Modf(price)

	// 余りを取引通貨歩み値で割り、
	addN := math.Max(0, math.RoundToEven(f/tick.TickSize))

	return math.Trunc((n+(tick.TickSize*addN))*SATOSHI) / SATOSHI
}

// MinQty is bitmex min size
func MinQty(ltp float64) int {
	return int(math.Ceil(ltp * MINSIZE))
}
