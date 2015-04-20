package pipeliner

import "sync"

// Context provides a way for running pipelines to know when
// to stop executing (via the Done channel). It also allows
// both the caller and the pipelines to cancel processing.
type Context interface {
	// Cancel may be called at any time to stop processing
	// and an error may be supplied
	Cancel(error)
	// Done is used by pipelines to determine when to stop
	// processing.
	Done() <-chan struct{}
}

type firstError struct {
	done chan struct{}
	err  error
	once sync.Once
}

// FirstError is a simple implementation of Context where
// the first value processed by the Cancel function is saved
// and returned later in Err()
//
// A different strategy might be first non-nil error to
// accommodate needing to cancel a pipeline. Maybe there
// should be a separate function for halting a pipeline
// and a function for signalling an error.
func FirstError() *firstError {
	return &firstError{
		done: make(chan struct{}),
	}
}

// Cancel causes the done channel to close and pipelines
// using this context to exit. If an error occurs while
// pipelines are being closed, then the first error
// processed by this function is retained.
// It is safe to call this function multiple times.
func (fe *firstError) Cancel(err error) {
	fe.once.Do(func() {
		fe.err = err
		close(fe.done)
	})
}

// Done returns the channel that all pipelines should listen
// against to signal cancelation
func (fe firstError) Done() <-chan struct{} {
	return fe.done
}

// Error returns the first value passed to Cancel and should
// be called after the pipeline fully closes. e.g. the drain
// has iterated over all values.
func (fe firstError) Err() error {
	return fe.err
}
