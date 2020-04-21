package orders_test

import (
	"fmt"
	"testing"

	"github.com/go-numb/go-bitmex-wrapper/orders"
)

func TestDecimal(t *testing.T) {
	toPrice := orders.NewTradePrice()
	for k := range toPrice.Instruments {
		fmt.Printf("%+v	%f\n", toPrice.Instruments[k].Symbol, toPrice.Instruments[k].TickSize)
	}

	number := 7200.0
	count := 100

	for i := 0; i < count; i++ {
		f := number + float64(i)*0.01
		fmt.Printf("BTC:	%.3f	->	%+v\n", f, toPrice.Decimal("XBTUSD", f))
	}

	number = 170
	count = 1000

	for i := 0; i < count; i++ {
		f := number + float64(i)*0.001
		fmt.Printf("ETH:	%.3f	->	%+v\n", f, toPrice.Decimal("ETHUSD", f))
	}
}
