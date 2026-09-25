package models

import (
	"time"

	"github.com/uptrace/bun"
)

type CronJob struct {
	bun.BaseModel `bun:"table:cron_jobs,alias:cron_job"`

	Queue        string     `json:"-" bun:"queue,notnull,pk"`
	CronSchedule string     `json:"-" bun:"cron_schedule,notnull,nullzero"`
	LastRunAt    *time.Time `json:"-" bun:"last_run_at"`
	NextRunAt    time.Time  `json:"-" bun:"next_run_at,nullzero"`
}
