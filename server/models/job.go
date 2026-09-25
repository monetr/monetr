package models

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

type JobStatus string

const (
	PendingJobStatus    JobStatus = "pending"
	ProcessingJobStatus JobStatus = "processing"
	FailedJobStatus     JobStatus = "failed"
	CompletedJobStatus  JobStatus = "completed"
)

type Job struct {
	bun.BaseModel `bun:"table:jobs,alias:job"`

	JobId         ID[Job]    `json:"-" bun:"job_id,notnull,pk"`
	Priority      int64      `json:"-" bun:"priority,notnull,nullzero"`
	Queue         string     `json:"-" bun:"queue,notnull,nullzero"`
	Signature     string     `json:"-" bun:"signature,notnull,nullzero"`
	Input         string     `json:"-" bun:"input,nullzero"`
	Output        string     `json:"-" bun:"output,nullzero"`
	Status        JobStatus  `json:"-" bun:"status,notnull,nullzero"`
	SentryTraceId *string    `json:"-" bun:"sentry_trace_id"`
	SentryBaggage *string    `json:"-" bun:"sentry_baggage"`
	Attempt       int        `json:"-" bun:"attempt,nullzero"`
	CreatedAt     time.Time  `json:"-" bun:"created_at,notnull,nullzero"`
	UpdatedAt     time.Time  `json:"-" bun:"updated_at,notnull,nullzero"`
	StartedAt     *time.Time `json:"-" bun:"started_at"`
	CompletedAt   *time.Time `json:"-" bun:"completed_at"`
}

func (Job) IdentityPrefix() string {
	return "job"
}

var (
	_ bun.BeforeAppendModelHook = (*Job)(nil)
)

func (o *Job) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
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
	}

	return nil
}
