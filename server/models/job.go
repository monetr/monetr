package models

import "time"

type Job struct {
	tableName string `pg:"jobs"`

	JobId      string                 `json:"jobId" pg:"job_id,notnull,pk"`
	AccountId  uint64                 `json:"-" pg:"account_id,on_delete:CASCADE"`
	Account    *Account               `json:"-" pg:"rel:has-one"`
	Name       string                 `json:"name" pg:"name,notnull"`
	Args       map[string]interface{} `json:"arguments" pg:"arguments,type:jsonb"`
	EnqueuedAt time.Time              `json:"enqueuedAt" pg:"enqueued_at,notnull"`
	StartedAt  *time.Time             `json:"startedAt" pg:"started_at"`
	FinishedAt *time.Time             `json:"finishedAt" pg:"finished_at"`
	Retries    int                    `json:"retries" pg:"retries,use_zero"`
}
