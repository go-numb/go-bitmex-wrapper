package orders

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// 実行した回数: 1000000000
// １回あたりの実行に掛かった時間(ns/op)

// BenchmarkHasPrefix
// buy/sell: 0.000087 ns/op
// 16byte: 0.000099 ns/op

// BenchmarkEqual
// buy/sell:0.000070 ns/op
// 16byte:  0.000087 ns/op

func BenchmarkTimeA(t *testing.B) {
	// 	BenchmarkTimeA-8   	19396.51
	// 19393.56
	// 33080.64
	// 23244.48
	// 31200.19
	// 1000000000	         0.000087 ns/op
	count := 100

	var avgs int64

	t.ResetTimer()
	func() {
		for i := 0; i < count; i++ {
			start := time.Now()
			defer func() {
				end := time.Now()
				avgs += int64(end.Sub(start))
			}()

			if strings.HasPrefix("0193e879-cb6f-2891-d099-2c4eb40fee21", "0193e879-cb6f-2891-d099-2c4eb40fee21") {

			} else if strings.HasPrefix("0193e879-cb6f-2891-d099-2c4eb40fee21", "1193e879-cb6f-2891-d099-2c4eb40fee21") {

			}
		}
	}()

	fmt.Printf("%+v\n", float64(avgs)/float64(count))

}

func BenchmarkTimeB(t *testing.B) {
	count := 100

	var avgs int64

	t.ResetTimer()
	func() {
		for i := 0; i < count; i++ {
			start := time.Now()
			defer func() {
				end := time.Now()
				avgs += int64(end.Sub(start))
			}()

			if "0193e879-cb6f-2891-d099-2c4eb40fee21" == "0193e879-cb6f-2891-d099-2c4eb40fee21" {

			} else if "0193e879-cb6f-2891-d099-2c4eb40fee21" == "1193e879-cb6f-2891-d099-2c4eb40fee21" {

			}
		}
	}()

	fmt.Printf("%+v\n", float64(avgs)/float64(count))
}
