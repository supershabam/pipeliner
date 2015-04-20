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

type queueError struct {
	done   chan struct{}
	errors []error
	once   sync.Once
}

// QueueError is another type of pipeliner context that enqueues
// any errors it encounters along the way. The first Cancel call
// will signal Done and stop any pipeliner processes, and the
// error value will be added to an internal queue of errors.
//
// So, if a cancel call causes errors while terminating, you can
// get all the errors
func QueueError() *queueError {
	return &queueError{
		done:   make(chan struct{}),
		errors: []error{},
	}
}

// Cancel causes the done channel to close and also enqueues the
// provided error which will be appended to errors.
func (qe *queueError) Cancel(err error) {
	qe.errors = append(qe.errors, err)
	qe.once.Do(func() {
		close(qe.done)
	})
}

// Done returns the channel that all pipelines should listen
// against to signal cancelation
func (qe queueError) Done() <-chan struct{} {
	return qe.done
}

// Errors returns the list of all values passed to cancel in
// the order that this context processed them.
func (qe queueError) Errors() []error {
	return qe.errors
}
