// AUTOMATICALLY GENERATED FILE - DO NOT EDIT
// generated by pipeliner@v0.1.0
package main

import "sync"

// filenamesToPeople is a generated implementation of the flatMap operator
func filenamesToPeople(outerDone <-chan struct{}, concurrency int, in <-chan string, fn func(<-chan struct{}, string) (<-chan Person, <-chan error)) (<-chan Person, <-chan error) {
	out := make(chan Person)
	errc := make(chan error, 1)
	go func() {
		defer close(out)
		var outerErr error
		defer func() {
			errc <- outerErr
		}()
		// create our own done function that we have the ability to
		// close upon error. Wrap a once function around closing
		// the done function to protect ourselves.
		done := make(chan struct{})
		once := sync.Once{}
		stop := func(err error) {
			once.Do(func() {
				if err != nil {
					outerErr = err
				}
				close(done)
			})
		}
		// outerDone signals stop
		go func() {
			select {
			case <-outerDone:
				stop(nil)
			case <-done:
			}
		}()
		// stop must be called at least once to ensure the
		// goroutine exits.
		defer stop(nil)

		wg := sync.WaitGroup{}
		wg.Add(concurrency)
		for i := 0; i < concurrency; i++ {
			go func() {
				defer wg.Done()
				for v := range in {
					vch, verrc := fn(done, v)
					for w := range vch {
						select {
						case <-done:
							return
						case out <- w:
						}
					}
					err := <-verrc
					if err != nil {
						stop(err)
						return
					}
				}
			}()
		}
		wg.Wait()
	}()
	return out, errc
}