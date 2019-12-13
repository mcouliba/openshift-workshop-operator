package util

var OperatorStatus = struct {
	NotScheduled string
	Scheduled    string
	InProgress   string
	Installed    string
}{
	NotScheduled: "NOT SCHEDULED",
	Scheduled:    "SCHEDULED",
	InProgress:   "IN PROGRESS",
	Installed:    "INSTALLED",
}

func IsScheduled(enabled bool) string {
	result := OperatorStatus.NotScheduled
	if enabled {
		result = OperatorStatus.Scheduled
	}
	return result
}
