// AUTOMATICALLY GENERATED FILE - DO NOT EDIT
// generated by pipeliner@v0.1.0
package main

// batchInts is a generated implementation of the batch operator
func batchInts(done <-chan struct{}, max int, in <-chan int) <-chan []int {
	if max <= 0 {
		panic("batch max must be greater than zero")
	}
	out := make(chan []int)
	go func() {
		defer close(out)
	Start:
		batch := []int{}
		for {
			if len(batch) >= max {
				select {
				case <-done:
					return
				case out <- batch:
				}
				goto Start
			}
			item, active := <-in
			if !active {
				if len(batch) > 0 {
					select {
					case <-done:
						return
					case out <- batch:
					}
				}
				return
			}
			batch = append(batch, item)
		}
	}()
	return out
}