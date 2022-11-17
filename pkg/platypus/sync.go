package platypus

type SyncResult struct {
	NextCursor string
	HasMore    bool
	New        map[string]Transaction
	Updated    map[string]Transaction
	Deleted    []string
}
