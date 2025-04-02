package shared

type ServerStatus struct {
	SendCount   uint64 `json:"send_count"`
	ActiveCount int64  `json:"active_count"`
	DoneCount   uint64 `json:"done_count"`
	FailedCount uint64 `json:"failed_count"`
	//AccountsStatus map[string]ServerAccountsStatus `json:"accounts_status"`
	GoroutineCount int   `json:"goroutine_count"`
	Interval       int64 `json:"interval"`
	StartAt        int64 `json:"start_at"`
}
