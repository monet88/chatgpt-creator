package register

type StopReason string

const (
	StopReasonTargetReached       StopReason = "target_reached"
	StopReasonContextCancelled    StopReason = "context_cancelled"
	StopReasonMaxAttemptsReached  StopReason = "max_attempts_reached"
	StopReasonFailureThresholdHit StopReason = "failure_threshold_reached"
)

type BatchResult struct {
	Target         int               `json:"target"`
	Success        int64             `json:"success"`
	Attempts       int64             `json:"attempts"`
	Failures       int64             `json:"failures"`
	Elapsed        string            `json:"elapsed"`
	StopReason     StopReason        `json:"stop_reason"`
	OutputFile     string            `json:"output_file"`
	FailureSummary map[FailureKind]int64 `json:"failure_summary"`
}
