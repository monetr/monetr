package models

import (
	"time"
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

	JobId         uint64     `json:"-" pg:"job_id,notnull,pk,type:'bigserial'"`
	Queue         string     `json:"-" pg:"queue,notnull"`
	Signature     string     `json:"-" pg:"signature,notnull"`
	Input         string     `json:"-" pg:"input"`
	Output        string     `json:"-" pg:"output"`
	Status        JobStatus  `json:"-" pg:"status,notnull"`
	Attempt       int        `json:"-" pg:"attempt,notnull"`
	Timestamp     time.Time  `json:"-" pg:"timestamp,notnull"`
	PreviousJobId *uint64    `json:"-" pg:"previous_job_id"`
	PreviousJob   *Job       `json:"-" pg:"rel:has-one"`
	CreatedAt     time.Time  `json:"-" pg:"created_at,notnull"`
	UpdatedAt     time.Time  `json:"-" pg:"updated_at,notnull"`
	StartedAt     *time.Time `json:"-" pg:"started_at"`
	CompletedAt   *time.Time `json:"-" pg:"completed_at"`
}
