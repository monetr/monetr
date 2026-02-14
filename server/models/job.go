package models

import (
	"context"
	"time"

	"github.com/go-pg/pg/v10"
)

type JobStatus string

const (
	PendingJobStatus    JobStatus = "pending"
	ProcessingJobStatus JobStatus = "processing"
	FailedJobStatus     JobStatus = "failed"
	CompletedJobStatus  JobStatus = "completed"
)

type Job struct {
	tableName string `pg:"jobs"`

	JobId         ID[Job]    `json:"-" pg:"job_id,notnull,pk"`
	Priority      uint64     `json:"-" pg:"priority,notnull"`
	Queue         string     `json:"-" pg:"queue,notnull"`
	Signature     string     `json:"-" pg:"signature,notnull"`
	Input         string     `json:"-" pg:"input"`
	Output        string     `json:"-" pg:"output"`
	Status        JobStatus  `json:"-" pg:"status,notnull"`
	SentryTraceId *string    `json:"-" pg:"sentry_trace_id"`
	SentryBaggage *string    `json:"-" pg:"sentry_baggage"`
	CreatedAt     time.Time  `json:"-" pg:"created_at,notnull"`
	UpdatedAt     time.Time  `json:"-" pg:"updated_at,notnull"`
	StartedAt     *time.Time `json:"-" pg:"started_at"`
	CompletedAt   *time.Time `json:"-" pg:"completed_at"`
}

func (Job) IdentityPrefix() string {
	return "job"
}

var (
	_ pg.BeforeInsertHook = (*Job)(nil)
)

func (o *Job) BeforeInsert(ctx context.Context) (context.Context, error) {
	if o.JobId.IsZero() {
		o.JobId = NewID[Job]()
	}

	now := time.Now()
	if o.CreatedAt.IsZero() {
		o.CreatedAt = now
	}

	if o.UpdatedAt.IsZero() {
		o.UpdatedAt = now
	}

	return ctx, nil
}
