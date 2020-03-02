package executions

import (
	"fmt"
	"strings"
	"testing"
	"time"

	jsoniter "github.com/json-iterator/go"
	"golang.org/x/sync/errgroup"

	"github.com/buger/jsonparser"
	"github.com/go-numb/go-bitmex"
	"github.com/gorilla/websocket"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type tt struct {
	Op   string   `json:"op"`
	Args []string `json:"args"`
	ID   int      `json:"id,omitempty"`
}

func TestConnect(t *testing.T) {
	conn, _, err := websocket.DefaultDialer.Dial("wss://www.bitmex.com/realtime", nil)
	if err != nil {
		t.Fatal(err)
	}

	products := []string{bitmex.XBTUSD}
	channels := []string{
		"liquidation",
		// "tradeBin1d",
		"trade",

		// Privates
		// "order",
		// "execution",
		// "position",
	}

	for i := range channels {
		for _, product := range products {
			if strings.HasPrefix(channels[i], "margin") ||
				strings.HasPrefix(channels[i], "order") ||
				strings.HasPrefix(channels[i], "execution") ||
				strings.HasPrefix(channels[i], "position") ||
				strings.HasPrefix(channels[i], "wallet") {
				continue
			}
			channels[i] = fmt.Sprintf("%s:%s", channels[i], product)
		}
	}

	loss := make(chan Losscut)
	exec := New(loss)

	req := tt{
		Op:   "subscribe",
		Args: channels,
		ID:   1,
	}

	// Reading
	if err := conn.WriteJSON(req); err != nil {
		t.Fatal(err)
	}

	var eg errgroup.Group

	eg.Go(func() error {
		for {
			conn.SetReadDeadline(time.Now().Add(5 * time.Second))
			_, msg, err := conn.ReadMessage()
			if err != nil {
				fmt.Printf("websocket read: %+v\n", err)
				break
			}

			name, err := jsonparser.GetString(msg, "table")
			if err != nil {
				continue
			}

			data, _, _, err := jsonparser.Get(msg, "data")
			if err != nil {
				continue
			}

			switch name {
			case "liquidation":
				var liq []bitmex.Liquidation
				if err := json.Unmarshal(data, &liq); err != nil {
					continue
				}
				exec.SetLiquidation(liq)

			case "trade":
				var e []bitmex.Trade
				json.Unmarshal(data, &e)
				exec.Set(e)
				ask, bid := exec.Best()
				fmt.Printf("%.1f	%.1f\n", ask, bid)
			}
		}
		return fmt.Errorf("")
	})

	eg.Go(func() error {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				var l Losscut
				l.isLosscut = true
				l.liquidations = []bitmex.Liquidation{
					bitmex.Liquidation{
						Symbol: "これ送信",
					},
				}
				l.createdAt = time.Now()
				exec.received(l)

			}
		}
		return fmt.Errorf("")
	})

	eg.Go(func() error {
		for {
			select {
			case liq := <-loss:
				copy := liq.Copy()
				for i := range copy {
					fmt.Printf("%+v\n", copy[i])
				}

			}
		}
		return fmt.Errorf("")
	})

	if err := eg.Wait(); err != nil {
		t.Fatal(err)
	}
}
