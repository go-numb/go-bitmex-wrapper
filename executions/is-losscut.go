package executions

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-numb/go-bitmex"
)

type Losscut struct {
	isLosscut bool
	symbol    string
	side      int
	volume    int
	createdAt time.Time
}

func (loss Losscut) revieved(c chan Losscut) {
	c <- loss
}

func toInt(s string) int {
	side := strings.ToLower(s)
	if side == "buy" {
		return 1
	} else if side == "sell" {
		return -1
	}
	return 0
}

func toSide(i int) string {
	if 0 < i {
		return "buy"
	} else if i < 0 {
		return "sell"
	}
	return 0
}

// SetLiquidation 不利約定の集計
func (p *Execution) SetLiquidation(liqs []bitmex.Liquidation) {
	var loss Losscut
	for i := range liqs {
		loss.isLosscut = true
		loss.side += toInt(liqs[i].Side)
		loss.volume += liqs[i].LeaveQty
	}

	loss.createdAt = time.Now()

	// if gets Losscut, send to channel.
	if loss.isLosscut {
		go loss.revieved(p.l)
	}
}

func (p *Losscut) IsThere() bool {
	return p.isLosscut
}

func (p *Losscut) Side() int {
	return p.side
}

func (p *Losscut) Volume() int {
	return p.volume
}

func (p *Losscut) CreatedAt() time.Time {
	return p.createdAt
}

func (p Losscut) String() string {
	return fmt.Sprintf("%t,%s,%d,%s", p.isLosscut, toSide(p.side), p.volume, p.createdAt.Format("2006/01/02 15:04:05"))
}
