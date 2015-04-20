package main

import "fmt"

func source(done <-chan struct{}, start, end int) <-chan int {
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
	done := make(chan struct{})
	sourcec := source(done, 0, 9000)
	batches := batchInts(done, 320, sourcec)
	for batch := range batches {
		fmt.Printf("received batch of length: %d\n", len(batch))
	}
}
