package background

import (
	"context"
	"testing"
	"time"

	"github.com/gocraft/work"
	"github.com/monetr/monetr/pkg/config"
	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestNewGoCraftWorkJobEnqueuer(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		log := testutils.GetLog(t)
		redis := testutils.GetRedisPool(t)

		// Create a channel so we can send a signal from the job function to our test. This is used to make sure the job
		// actually gets run.
		ranJob := make(chan struct{}, 0)
		// Create the gocraft worker pool.
		pool := work.NewWorkerPool(struct{}{}, 1, GoCraftWorkNamespace, redis)
		// And register our testing job that will let us make sure jobs are actually being enqueued when we want them
		// to be.
		pool.Job(t.Name(), func(job *work.Job) error {
			log.Info("job was triggered!")
			ranJob <- struct{}{}
			return nil
		})
		// The pool has to be started manually.
		pool.Start()
		defer pool.Stop() // Make sure the pool get's cleaned up when we are done.

		// Create our enqueuer, this is a wrapper around the gocraft/work enqueuer that lets us more easily throw jobs
		// into our worker pool.
		enqueuer := NewGoCraftWorkJobEnqueuer(log, redis)
		// Enqueue our test job with no arguments.
		err := enqueuer.EnqueueJob(context.Background(), t.Name(), nil)
		// And make sure that it does not fail to just enqueue, if this does fail here then that means our redis testing
		// instance isn't working quite right. At the time of writing this we are using miniredis which is embedded into
		// our app, not an actual redis instance.
		assert.NoError(t, err, "must be able to enqueue job")

		// We need to make sure the job runs, so create a timer that will serve as a "timeout" channel. If we don't hear
		// from our ranJob channel in 30 seconds then this timer will fail the test.
		timeout := time.NewTimer(30 * time.Second)
		defer timeout.Stop()
		select {
		case <-timeout.C:
			t.Fatalf("timed out waiting for job to be run")
		case <-ranJob:
			// Everything worked, the job ran successfully.
		}
	})

	t.Run("enqueue job that is not registered", func(t *testing.T) {
		log := testutils.GetLog(t)
		redis := testutils.GetRedisPool(t)

		enqueuer := NewGoCraftWorkJobEnqueuer(log, redis)
		// Nothing special here, I just want to make sure that if we do try to enqueue a job that is not registered with
		// the worker pool, or if there isn't even a worker pool; the enqueue job will not fail. This is important for
		// when monetr is updating and a newer version may try to enqueue a job that the older version does not have, or
		// vice versa, the older version of monetr may enqueue a job that is no longer needed. But it is not a reason to
		// fail.
		err := enqueuer.EnqueueJob(context.Background(), t.Name(), nil)
		assert.NoError(t, err, "must not return an error even if the job is not registered")
	})
}

func TestNewGoCraftWorkJobProcessor(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		log := testutils.GetLog(t)
		configuration := config.BackgroundJobs{
			Engine:      config.BackgroundJobEngineGoCraftWork,
			Scheduler:   config.BackgroundJobSchedulerInternal,
			JobSchedule: map[string]string{},
		}
		redis := testutils.GetRedisPool(t)
		enqueuer := NewGoCraftWorkJobEnqueuer(log, redis)

		// We want to make sure that even when gocraft/work is wrapped in our job processor, jobs will still be run as
		// they should and our wrapper code is functional.
		processor := NewGoCraftWorkJobProcessor(log, configuration, redis, enqueuer)

		// Same as before, we will use a channel to indicate whether the job has run.
		ranJob := make(chan struct{}, 0)
		testHandler := NewTestJobHandler(t, func(t *testing.T, ctx context.Context, data []byte) error {
			log.Info("running test job!")
			var input string
			// On top of running the job, we also want to make sure that we can pass some arbitrary data to the job
			// from the enqueuer. In this test we are just passing a string (the name of the test) to the job to make
			// sure that it can be read back.
			assert.NoError(t, DefaultJobUnmarshaller(data, &input), "must unmarshal data from job")
			assert.Equal(t, t.Name(), input, "unmarshalled string must be the test name")
			ranJob <- struct{}{}
			log.Info("finished running test job!")
			return nil
		})

		// Jobs are registered a bit different with the processor wrapper than they are with the gocraft/work pool
		// directly.
		assert.NoError(t, processor.RegisterJob(context.Background(), testHandler), "must be able to register job")
		// Start the job processor.
		assert.NoError(t, processor.Start(), "must be able to start job processor")
		defer func() {
			// And stop the job processor. This won't fail naturally, but can fail if the job processor fails to drain
			// all of its workers.
			assert.NoError(t, processor.Close(), "must close processor gracefully at the end of the test")
		}()

		// Enqueue our test handler with the test name as the argument. We are going to make sure the processor works.
		assert.NoError(t, enqueuer.EnqueueJob(context.Background(), testHandler.QueueName(), t.Name()), "must be able to enqueue job")

		// Same as before, setup a timer to be a "timeout" for our job. We want to make sure the job is actually run in
		// this window.
		timeout := time.NewTimer(10 * time.Second)
		defer timeout.Stop()
		select {
		case <-timeout.C:
			t.Fatalf("timed out waiting for job to be run")
		case <-ranJob:
			// The job ran successfully!
		}
	})

	t.Run("job panics gracefully", func(t *testing.T) {
		log := testutils.GetLog(t)
		configuration := config.BackgroundJobs{
			Engine:      config.BackgroundJobEngineGoCraftWork,
			Scheduler:   config.BackgroundJobSchedulerInternal,
			JobSchedule: map[string]string{},
		}
		redis := testutils.GetRedisPool(t)
		enqueuer := NewGoCraftWorkJobEnqueuer(log, redis)

		processor := NewGoCraftWorkJobProcessor(log, configuration, redis, enqueuer)

		ranJob := make(chan struct{}, 0)
		testHandler := NewTestJobHandler(t, func(t *testing.T, ctx context.Context, data []byte) error {
			log.Info("running test job!")
			var input string
			assert.NoError(t, DefaultJobUnmarshaller(data, &input), "must unmarshal data from job")
			assert.Equal(t, t.Name(), input, "unmarshalled string must be the test name")

			ranJob <- struct{}{}
			log.Info("finished running test job!")

			panic("what happens if i have a breakdown")
		})

		assert.NoError(t, processor.RegisterJob(context.Background(), testHandler), "must be able to register job")
		assert.NoError(t, processor.Start(), "must be able to start job processor")
		defer func() {
			assert.NoError(t, processor.Close(), "must close processor gracefully at the end of the test")
		}()

		// Enqueue our test handler with the test name as the argument. We are going to make sure the processor works.
		assert.NoError(t, enqueuer.EnqueueJob(context.Background(), testHandler.QueueName(), t.Name()), "must be able to enqueue job")

		timeout := time.NewTimer(10 * time.Second)
		select {
		case <-timeout.C:
			t.Fatalf("timed out waiting for job to be run")
		case <-ranJob:
		}
	})
}

func TestGoCraftWorkJobProcessor_Close(t *testing.T) {
	t.Run("does not time out", func(t *testing.T) {
		log := testutils.GetLog(t)
		configuration := config.BackgroundJobs{
			Engine:      config.BackgroundJobEngineGoCraftWork,
			Scheduler:   config.BackgroundJobSchedulerInternal,
			JobSchedule: map[string]string{},
		}
		redis := testutils.GetRedisPool(t)
		enqueuer := NewGoCraftWorkJobEnqueuer(log, redis)

		// We want to make sure that even when gocraft/work is wrapped in our job processor, jobs will still be run as
		// they should and our wrapper code is functional.
		processor := NewGoCraftWorkJobProcessor(log, configuration, redis, enqueuer)

		// Same as before, we will use a channel to indicate whether the job has run.
		ranJob := make(chan struct{}, 1)
		testHandler := NewTestJobHandler(t, func(t *testing.T, ctx context.Context, data []byte) error {
			log.Info("running test job!")

			// Close() has a 30-second timeout, so make sure our job takes just slightly longer than that to trigger the
			// failure case.
			time.Sleep(10 * time.Second)

			ranJob <- struct{}{}
			log.Info("finished running test job!")
			return nil
		})

		// Jobs are registered a bit different with the processor wrapper than they are with the gocraft/work pool
		// directly.
		assert.NoError(t, processor.RegisterJob(context.Background(), testHandler), "must be able to register job")
		// Start the job processor.
		assert.NoError(t, processor.Start(), "must be able to start job processor")

		// Enqueue our job, we are going to make sure that the processor will fail to drain if our job takes too long to
		// finish.
		assert.NoError(t, enqueuer.EnqueueJob(context.Background(), testHandler.QueueName(), nil), "must be able to enqueue job")
		// Wait just a moment.
		time.Sleep(2 * time.Second)
		assert.NoError(t, processor.Close(), "should not timeout closing this time")

		// Same as before, set up a timer to be a "timeout" for our job. This is so we can wait for the job to finish
		// even after the processor has been closed (note: a processor can be closed without being completely drained).
		timeout := time.NewTimer(30 * time.Second)
		defer timeout.Stop()
		select {
		case <-timeout.C:
			t.Fatalf("timed out waiting for job to be run")
		case <-ranJob:
			// The job ran successfully!
		}
	})

	t.Run("times out", func(t *testing.T) {
		log := testutils.GetLog(t)
		configuration := config.BackgroundJobs{
			Engine:      config.BackgroundJobEngineGoCraftWork,
			Scheduler:   config.BackgroundJobSchedulerInternal,
			JobSchedule: map[string]string{},
		}
		redis := testutils.GetRedisPool(t)
		enqueuer := NewGoCraftWorkJobEnqueuer(log, redis)

		// We want to make sure that even when gocraft/work is wrapped in our job processor, jobs will still be run as
		// they should and our wrapper code is functional.
		processor := NewGoCraftWorkJobProcessor(log, configuration, redis, enqueuer)

		// Same as before, we will use a channel to indicate whether the job has run.
		ranJob := make(chan struct{}, 1)
		testHandler := NewTestJobHandler(t, func(t *testing.T, ctx context.Context, data []byte) error {
			log.Info("running test job!")

			// Close() has a 30-second timeout, so make sure our job takes just slightly longer than that to trigger the
			// failure case.
			time.Sleep(40 * time.Second)

			ranJob <- struct{}{}
			log.Info("finished running test job!")
			return nil
		})

		// Jobs are registered a bit different with the processor wrapper than they are with the gocraft/work pool
		// directly.
		assert.NoError(t, processor.RegisterJob(context.Background(), testHandler), "must be able to register job")
		// Start the job processor.
		assert.NoError(t, processor.Start(), "must be able to start job processor")

		// Enqueue our job, we are going to make sure that the processor will fail to drain if our job takes too long to
		// finish.
		assert.NoError(t, enqueuer.EnqueueJob(context.Background(), testHandler.QueueName(), nil), "must be able to enqueue job")
		// Wait just a moment.
		time.Sleep(2 * time.Second)
		// Then try to close the job processor, this should block for at most 30 seconds.
		assert.EqualError(t, processor.Close(), "timeout while draining gocraft/work queue")

		// Same as before, set up a timer to be a "timeout" for our job. This is so we can wait for the job to finish
		// even after the processor has been closed (note: a processor can be closed without being completely drained).
		timeout := time.NewTimer(40 * time.Second)
		defer timeout.Stop()
		select {
		case <-timeout.C:
			t.Fatalf("timed out waiting for job to be run")
		case <-ranJob:
			// The job ran successfully!
		}
	})

	t.Run("already closed", func(t *testing.T) {
		log := testutils.GetLog(t)
		configuration := config.BackgroundJobs{
			Engine:      config.BackgroundJobEngineGoCraftWork,
			Scheduler:   config.BackgroundJobSchedulerInternal,
			JobSchedule: map[string]string{},
		}
		redis := testutils.GetRedisPool(t)
		enqueuer := NewGoCraftWorkJobEnqueuer(log, redis)
		processor := NewGoCraftWorkJobProcessor(log, configuration, redis, enqueuer)

		// Start the job processor.
		assert.NoError(t, processor.Start(), "must be able to start job processor")

		// Pause a bit between starting the workers and closing them below.
		time.Sleep(500 * time.Millisecond)

		assert.NoError(t, processor.Close(), "the first time we close the processor it should succeed")
		// The second time we close it should complain though.
		assert.EqualError(t, processor.Close(), "gocraft/work job processor is either already closed, or is in an invalid state")
	})
}
