package executions

import (
	"fmt"
	"time"

	"github.com/go-numb/go-bitmex"
)

type Losscut struct {
	isLosscut    bool
	liquidations []bitmex.Liquidation
	createdAt    time.Time
}

// SetLiquidation 不利約定の集計
func (p *Execution) SetLiquidation(liqs []bitmex.Liquidation) {
	var loss Losscut
	loss.liquidations = make([]bitmex.Liquidation, len(liqs))

	for i := range liqs {
		loss.isLosscut = true
		loss.liquidations[i] = liqs[i]
	}

	// if gets Losscut, send to channel.
	if loss.isLosscut {
		loss.createdAt = time.Now()
		go p.received(loss)
	}
}

func (p *Losscut) IsThere() bool {
	return p.isLosscut
}

func (p *Losscut) Price() (first, last float64) {
	for i := range p.liquidations {
		if i == 0 {
			first = p.liquidations[i].Price
			last = p.liquidations[i].Price
		}
		last = p.liquidations[i].Price
	}

	return first, last
}

func (p *Losscut) Copy() []bitmex.Liquidation {
	return p.liquidations
}

func (p *Losscut) Volume() (volume int) {
	for i := range p.liquidations {
		volume = p.liquidations[i].LeavesQty
	}
	return volume
}

func (p *Losscut) CreatedAt() time.Time {
	return p.createdAt
}

func (p Losscut) String() string {
	first, last := p.Price()
	return fmt.Sprintf("%t,%.1f,%.1f,%d,%s", p.isLosscut, first, last, p.Volume(), p.createdAt.Format("2006/01/02 15:04:05"))
}
