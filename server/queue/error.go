package queue

// RetryableError is an error interface that can be returned by a job function.
// If this error interface is returned then the [Retryable] function is called
// with the number of attempts performed already, including the attempt that was
// just performed. If this method returns true, then the job is inserted back
// into the queue with the attempt counter incremented. If the method returns
// false then the job is not re-attempted.
type RetryableError interface {
	error
	Retryable(attempts int) bool
}

