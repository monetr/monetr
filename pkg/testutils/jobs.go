package testutils

import (
	"github.com/brianvoe/gofakeit/v6"
)

type MockJobManager struct {
}

func NewMockJobManager() *MockJobManager {
	return &MockJobManager{}
}

func (m *MockJobManager) TriggerPullInitialTransactions(accountId, userId, linkId uint64) (jobId string, err error) {
	return gofakeit.UUID(), nil
}

func (m *MockJobManager) Close() error {
	return nil
}
