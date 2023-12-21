package models

import "time"

type CronJob struct {
	tableName string `pg:"cron_jobs"`

	Queue        string     `json:"-" pg:"queue,notnull,pk"`
	CronSchedule string     `json:"-" pg:"cron_schedule,notnull"`
	LastRunAt    *time.Time `json:"-" pg:"last_run_at"`
	NextRunAt    time.Time  `json:"-" pg:"next_run_at"`
}
