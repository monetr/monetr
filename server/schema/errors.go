package schema

import (
	"fmt"

	"github.com/Oudwins/zog"
)

type Error interface {
	error
	Issues() zog.ZogIssueList
}

var (
	_ Error = &issueError{}
)

type issueError struct {
	issues zog.ZogIssueList
}

// Error implements [Error].
func (i *issueError) Error() string {
	// TODO Make this better!
	return fmt.Sprintf("failed to validate schema:\n%s", zog.Issues.Prettify(i.issues))
}

// Issues implements [Error].
func (i *issueError) Issues() zog.ZogIssueList {
	return i.issues
}

func NewIssueError(issues zog.ZogIssueList) error {
	return &issueError{issues}
}
