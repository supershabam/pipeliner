package main

import (
	"fmt"

	"github.com/supershabam/pipeliner"
)

func source(ctx pipeliner.Context, start, end int) <-chan int {
	done := ctx.Done()
	out := make(chan int)
	go func() {
		defer close(out)
		for i := start; i < end; i++ {
			select {
			case <-done:
				return
			case out <- i:
			}
		}
	}()
	return out
}

//go:generate pipeliner -file=batch_ints.go -from=int -func=batchInts -operator=batch
func main() {
	ctx := pipeliner.FirstError()
	sourceCh := source(ctx, 0, 9000)
	batches := batchInts(ctx, 320, sourceCh)
	for batch := range batches {
		fmt.Printf("received batch of length: %d\n", len(batch))
	}
}
