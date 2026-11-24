package corvonfse

// ScheduleStatus is the state of a schedule.
type ScheduleStatus string

const (
	ScheduleActive    ScheduleStatus = "active"
	SchedulePaused    ScheduleStatus = "paused"
	ScheduleCompleted ScheduleStatus = "completed"
	ScheduleCancelled ScheduleStatus = "cancelled"
)

// Schedule is a scheduled (once) or recurring invoice. Its fields are
// snake_case on the wire and are kept as they are. The note only exists when
// the schedule runs (09:00 America/Sao_Paulo, Monday to Friday; a weekend
// date runs on Monday): follow the runs with Schedules.ListRuns.
type Schedule struct {
	AnchorOn  string `json:"anchor_on"`
	Attempts  int    `json:"attempts"`
	CompanyID string `json:"company_id"`
	CreatedAt string `json:"created_at"`
	EndsOn    string `json:"ends_on"`
	// Frequency is WEEKLY, BIWEEKLY or MONTHLY; empty for a one-off schedule.
	Frequency string `json:"frequency"`
	ID        string `json:"id"`
	// Invoice is the invoice body each run issues.
	Invoice map[string]any `json:"invoice"`
	// Kind is "once" or "recurring".
	Kind      string `json:"kind"`
	LastRunOn string `json:"last_run_on"`
	NextRunOn string `json:"next_run_on"`
	// PauseReason is "customer" or "failures" when the schedule is paused.
	PauseReason string         `json:"pause_reason"`
	Status      ScheduleStatus `json:"status"`
}

// ScheduleListParams are the filters accepted by Schedules.List and
// Schedules.Iterate. A nil pointer is treated as the zero value.
type ScheduleListParams struct {
	Status    ScheduleStatus
	CompanyID string
	Limit     int
	Cursor    string
}

// ScheduleUpdateParams are the fields accepted by Schedules.Update; the change
// applies to the next runs only. Invoice is merged over the stored body and
// validated like an emission. Null[string]() clears EndsOn.
type ScheduleUpdateParams struct {
	AnchorOn  string              `json:"anchorOn,omitempty"`
	EndsOn    *Nullable[string]   `json:"endsOn,omitempty"`
	Frequency RecurrenceFrequency `json:"frequency,omitempty"`
	Invoice   map[string]any      `json:"invoice,omitempty"`
}

// ScheduleRun is one attempt to run a schedule. Outcome is "issued" or
// "failed".
type ScheduleRun struct {
	Attempt      int    `json:"attempt"`
	CreatedAt    string `json:"created_at"`
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
	ID           string `json:"id"`
	InvoiceID    string `json:"invoice_id"`
	OccurrenceOn string `json:"occurrence_on"`
	Outcome      string `json:"outcome"`
	RanOn        string `json:"ran_on"`
}

// ImportStatus is the state of an import.
type ImportStatus string

const (
	ImportProcessing ImportStatus = "processing"
	ImportReady      ImportStatus = "ready"
	ImportEmitting   ImportStatus = "emitting"
	ImportDone       ImportStatus = "done"
	ImportFailed     ImportStatus = "failed"
)

// ImportRowStatus is the state of one row of an import.
type ImportRowStatus string

const (
	ImportRowValid    ImportRowStatus = "valid"
	ImportRowInvalid  ImportRowStatus = "invalid"
	ImportRowQueued   ImportRowStatus = "queued"
	ImportRowAccepted ImportRowStatus = "accepted"
	ImportRowRejected ImportRowStatus = "rejected"
)

// ImportCounters are the counters of an import (snake_case on the wire).
type ImportCounters struct {
	Accepted         int     `json:"accepted"`
	Authorized       int     `json:"authorized"`
	AuthorizedAmount float64 `json:"authorized_amount"`
	AwaitingAmount   float64 `json:"awaiting_amount"`
	AwaitingCredits  int     `json:"awaiting_credits"`
	Failed           int     `json:"failed"`
	FailedAmount     float64 `json:"failed_amount"`
	Invalid          int     `json:"invalid"`
	Pending          int     `json:"pending"`
	PendingAmount    float64 `json:"pending_amount"`
	Total            int     `json:"total"`
	Valid            int     `json:"valid"`
	ValidAmount      float64 `json:"valid_amount"`
}

// ImportJob is an import and its counters (snake_case on the wire).
type ImportJob struct {
	CompanyID      string         `json:"company_id"`
	Counters       ImportCounters `json:"counters"`
	CreatedAt      string         `json:"created_at"`
	EmitFinishedAt string         `json:"emit_finished_at"`
	EmitStartedAt  string         `json:"emit_started_at"`
	ErrorCode      string         `json:"error_code"`
	FileName       string         `json:"file_name"`
	ID             string         `json:"id"`
	RowsExpected   *int           `json:"rows_expected"`
	Status         ImportStatus   `json:"status"`
	UnknownColumns []string       `json:"unknown_columns"`
}

// ImportStart is the answer of Imports.Start. DuplicateOf is the id of an
// earlier import of the same file, when there is one.
type ImportStart struct {
	DuplicateOf string       `json:"duplicate_of"`
	ID          string       `json:"id"`
	Status      ImportStatus `json:"status"`
}

// ImportRow is one row of an import.
type ImportRow struct {
	Data         map[string]any  `json:"data"`
	ErrorCode    string          `json:"error_code"`
	ErrorMessage string          `json:"error_message"`
	InvoiceID    string          `json:"invoice_id"`
	Line         int             `json:"line"`
	Status       ImportRowStatus `json:"status"`
}

// ImportRowListParams are the filters accepted by Imports.ListRows and
// Imports.IterateRows. Limit is 1 to 200 (default 50).
type ImportRowListParams struct {
	Status ImportRowStatus
	Limit  int
	Cursor string
}

// ImportEmitParams are the fields accepted by Imports.Emit. SendEmail sends
// each issued invoice to its customer.
type ImportEmitParams struct {
	SendEmail *bool `json:"sendEmail,omitempty"`
}
