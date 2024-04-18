package models

import (
	"time"

	"github.com/monetr/monetr/server/identifier"
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

	JobId       identifier.ID `json:"-" pg:"job_id,notnull,pk,type:'bigserial'"`
	Queue       string        `json:"-" pg:"queue,notnull"`
	Signature   string        `json:"-" pg:"signature,notnull"`
	Input       string        `json:"-" pg:"input"`
	Output      string        `json:"-" pg:"output"`
	Status      JobStatus     `json:"-" pg:"status,notnull"`
	CreatedAt   time.Time     `json:"-" pg:"created_at,notnull"`
	UpdatedAt   time.Time     `json:"-" pg:"updated_at,notnull"`
	StartedAt   *time.Time    `json:"-" pg:"started_at"`
	CompletedAt *time.Time    `json:"-" pg:"completed_at"`
}
