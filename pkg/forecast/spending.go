package forecast

import (
	"time"

	"github.com/monetr/monetr/pkg/models"
	"github.com/monetr/monetr/pkg/util"
)

type SpendingEvent struct {
	Date              time.Time
	Amount            int64
	RollingAllocation int64
	Funding           []FundingEvent
	SpendingId        uint64
}

var (
	_ SpendingInstructions = &spendingInstructionBase{}
)

type SpendingInstructions interface {
	GetNextNSpendingEventsAfter(n int, input time.Time, timezone *time.Location) []SpendingEvent
	GetNextSpendingEventAfter(input time.Time, timezone *time.Location) *SpendingEvent
}

type spendingInstructionBase struct {
	spending models.Spending
	funding  FundingInstructions
}

func (s spendingInstructionBase) GetNextNSpendingEventsAfter(n int, input time.Time, timezone *time.Location) []SpendingEvent {
	//TODO implement me
	panic("implement me")
}

func (s *spendingInstructionBase) GetRecurrencesBetween(start, end time.Time, timezone *time.Location) []time.Time {
	switch s.spending.SpendingType {
	case models.SpendingTypeExpense:
		dtMidnight := util.MidnightInLocal(start, timezone)
		rule := s.spending.RecurrenceRule.RRule
		rule.DTStart(dtMidnight)
		return rule.Between(start, end, false)
	case models.SpendingTypeGoal:
		if s.spending.NextRecurrence.After(start) && s.spending.NextRecurrence.Before(end) {
			return []time.Time{s.spending.NextRecurrence}
		}
		fallthrough
	default:
		return nil
	}
}

func (s *spendingInstructionBase) GetNextSpendingEventAfter(input time.Time, timezone *time.Location) *SpendingEvent {
	//TODO implement me
	panic("implement me")
}
