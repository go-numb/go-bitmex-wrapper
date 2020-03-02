package orders

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"

	"github.com/go-numb/go-bitmex"
)

// Managed is orders/positions struct
type Managed struct {
	Orders    *Orders
	Cancels   *Orders
	Positions *Orders

	Result chan string
}

func New() *Managed {
	return &Managed{

		Orders:    new(Orders),
		Cancels:   new(Orders),
		Positions: new(Orders),
	}
}

type Orders struct {
	m sync.Map
}

type StatusType int

const (
	NotExist StatusType = iota
	OnBoard
	Partial
	Completed
	Canceled
	Expired
)

// Order informations
type Order bitmex.Execution

func toInt(side string) int {
	if strings.HasPrefix(side, bitmex.BUY) {
		return 1
	} else if strings.HasPrefix(side, bitmex.SELL) {
		return -1
	}
	return 0
}

// Switch
func (p *Managed) Switch(childs []bitmex.Execution) {
	for i := range childs {
		// OrdStatus: https://www.onixs.biz/fix-dictionary/5.0.SP2/msgType_8_8.html
		switch childs[i].OrdStatus {
		case "New":
			p.Orders.Set(Order(childs[i]))

		case "Rejected":
			p.Orders.Delete(childs[i].OrderID)
			p.Cancels.Delete(childs[i].OrderID)

		case "Filled": // 完全約定
			p.executed(childs[i])

		case "Partially": // 部分約定
			p.executed(childs[i])

		case "Canceled":
			p.cancel(childs[i].OrderID)

		case "Expired": // Expired
			p.cancel(childs[i].OrderID)

		case "Done": // Expired
			p.cancel(childs[i].OrderID)
		}
	}
}

func (p *Managed) executed(e bitmex.Execution) StatusType {
	o, ok := p.Orders.IsThere(e.OrderID)
	if !ok {
		return NotExist
	}

	if strings.HasPrefix(e.Side, bitmex.BUY) {
		// Qtyは正, e.Sizeは正
		o.OrderQty -= e.OrderQty
		if 0 < o.OrderQty { // 買建玉が残る部分約定
			return p.partial(o, e.OrderQty)
		}

		o.OrderQty = e.OrderQty
		return p.complete(o)

	} else if strings.HasPrefix(e.Side, bitmex.SELL) {
		// Qtyは負, e.Sizeは正
		o.OrderQty += e.OrderQty
		if o.OrderQty < 0 { // 売建玉が残る部分約定
			return p.partial(o, e.OrderQty)
		}

		o.OrderQty = e.OrderQty
		return p.complete(o)

	}

	return NotExist
}

func (p *Orders) Set(o Order) {
	o.OrderQty = o.OrderQty * toInt(o.Side)
	p.m.Store(o.OrderID, o)
}

func (p *Orders) Delete(uuid interface{}) {
	p.m.Delete(uuid)
}

// Deletes 保有注文/建玉/キャンセルを一部削除する
// 削除ルールはorderIDをストリングソートし、古い方から引数分だけ
func (p *Orders) Deletes(parcent int) (deleteCount int) {
	var keys []string
	p.m.Range(func(k, v interface{}) bool {
		keys = append(keys, fmt.Sprintf("%v", k))
		return true
	})

	// 古いものが先頭
	sort.Strings(keys)
	stop := float64(len(keys)) * float64(parcent) / float64(100)
	for range keys {
		p.m.Delete(keys[deleteCount])

		deleteCount++
		if stop < float64(deleteCount) {
			break
		}
	}
	return deleteCount
}

func (p *Orders) IsThere(uuid interface{}) (o Order, isThere bool) {
	v, ok := p.m.Load(uuid)
	if !ok {
		return o, false
	}
	return assert(v)
}

func assert(in interface{}) (o Order, ok bool) {
	o, ok = in.(Order)
	if !ok {
		return o, false
	}
	return o, true
}

func (p *Orders) Sum() (length, sum int) {
	p.m.Range(func(key, value interface{}) bool {
		o, ok := p.IsThere(key)
		if !ok {
			return false
		}

		length++
		sum += o.OrderQty

		return true
	})
	return length, sum
}

// Check 約定情報を引数に、mapに保有したordersから約定/部分約定を確認
// 確認後positionsへ移動する
func (p *Managed) Check(isCancel bool, uuid interface{}, side string, qty int) (status StatusType) {
	if isCancel {
		return p.cancel(uuid)
	}

	o, ok := p.Orders.IsThere(uuid)
	if !ok {
		return NotExist
	}

	if strings.HasPrefix(side, bitmex.BUY) {
		// Qtyは正, qtyは正
		o.OrderQty -= qty
		if 0 < o.OrderQty { // 買建玉が残る部分約定
			return p.partial(o, qty)
		}

		o.OrderQty = qty
		return p.complete(o)

	} else if strings.HasPrefix(side, bitmex.SELL) {
		// Qtyは負, qtyは正
		o.OrderQty += qty
		if o.OrderQty < 0 { // 売建玉が残る部分約定
			return p.partial(o, qty)
		}

		o.OrderQty = qty
		return p.complete(o)

	}

	// sideが合わないなど稀有な例
	return NotExist
}

func (p *Managed) partial(o Order, qty int) StatusType {
	if o.OrdStatus == "Partially" { // 部分約定ならば前約定と合算
		pos, ok := p.Positions.IsThere(o.OrderID)
		if !ok {
			return NotExist
		}
		// 残注文
		p.Orders.m.Store(o.OrderID, o)
		// 約定 -> 建玉
		o.OrderQty = int(math.Abs(float64(pos.OrderQty))+math.Abs(float64(qty))) * toInt(o.Side)
	} else {
		// 残注文
		// o.Status = Partial
		p.Orders.m.Store(o.OrderID, o)

		// 約定 -> 建玉
		o.OrderQty = int(math.Abs(float64(qty))) * toInt(o.Side)
	}

	// o.Status = Partial
	p.Positions.m.Store(o.OrderID, o)
	return Partial
}

func (p *Managed) complete(o Order) StatusType {
	p.Orders.m.Delete(o.OrderID)

	if o.OrdStatus == "Partially" { // 部分約定ならば前約定と合算
		pos, ok := p.Positions.IsThere(o.OrderID)
		if !ok {
			return NotExist
		}
		o.OrderQty = int(math.Abs(float64(pos.OrderQty))+math.Abs(float64(o.OrderQty))) * toInt(o.Side)
	} else {
		o.OrderQty = int(math.Abs(float64(o.OrderQty))) * toInt(o.Side)
	}

	// o.Status = Completed
	p.Positions.m.Store(o.OrderID, o)
	return Completed
}

func (p *Managed) cancel(uuid interface{}) StatusType {
	p.Orders.m.Delete(uuid)
	p.Cancels.m.Store(uuid, Order{})
	return Canceled
}
