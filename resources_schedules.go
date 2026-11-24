package corvonfse

import (
	"bytes"
	"context"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
)

// Schedules manages scheduled and recurring invoices. They are created with
// Invoices.Create and Options.Type SCHEDULE or RECURRING.
type Schedules struct {
	t *Transport
}

func scheduleListQuery(params *ScheduleListParams) url.Values {
	query := url.Values{}
	if params == nil {
		return query
	}
	queryAdd(query, "status", string(params.Status))
	queryAdd(query, "companyId", params.CompanyID)
	queryAddInt(query, "limit", params.Limit)
	queryAdd(query, "cursor", params.Cursor)
	return query
}

func schedulePath(id string) string {
	return "/schedules/" + url.PathEscape(id)
}

// List returns one page of schedules. GET /schedules — unwrap "none".
func (s *Schedules) List(ctx context.Context, params *ScheduleListParams) ([]Schedule, Pagination, *CorvoError) {
	return requestPage[Schedule](ctx, s.t, "/schedules", scheduleListQuery(params))
}

// Iterate walks every page of schedules matching params, calling fn per item
// until it returns false or all pages are consumed.
func (s *Schedules) Iterate(ctx context.Context, params *ScheduleListParams, fn func(Schedule) bool) *CorvoError {
	return iteratePages[Schedule](ctx, s.t, "/schedules", scheduleListQuery(params), fn)
}

// Get reads a schedule. GET /schedules/:id — unwrap "data".
func (s *Schedules) Get(ctx context.Context, id string) (Schedule, *CorvoError) {
	return requestJSON[Schedule](ctx, s.t, requestSpec{
		method: http.MethodGet,
		path:   schedulePath(id),
	})
}

// Update changes a schedule; it applies to the next runs only.
// PATCH /schedules/:id — unwrap "data".
func (s *Schedules) Update(ctx context.Context, id string, params ScheduleUpdateParams) (Schedule, *CorvoError) {
	return requestJSON[Schedule](ctx, s.t, requestSpec{
		method:   http.MethodPatch,
		path:     schedulePath(id),
		jsonBody: params,
	})
}

// Pause pauses a schedule. POST /schedules/:id/pause — unwrap "data".
func (s *Schedules) Pause(ctx context.Context, id string) (Schedule, *CorvoError) {
	return requestJSON[Schedule](ctx, s.t, requestSpec{
		method: http.MethodPost,
		path:   schedulePath(id) + "/pause",
	})
}

// Resume resumes a paused schedule at its next occurrence.
// POST /schedules/:id/resume — unwrap "data".
func (s *Schedules) Resume(ctx context.Context, id string) (Schedule, *CorvoError) {
	return requestJSON[Schedule](ctx, s.t, requestSpec{
		method: http.MethodPost,
		path:   schedulePath(id) + "/resume",
	})
}

// Cancel cancels a schedule for good. POST /schedules/:id/cancel — unwrap "data".
func (s *Schedules) Cancel(ctx context.Context, id string) (Schedule, *CorvoError) {
	return requestJSON[Schedule](ctx, s.t, requestSpec{
		method: http.MethodPost,
		path:   schedulePath(id) + "/cancel",
	})
}

// ListRuns returns one page of the attempts to run a schedule, newest first.
// GET /schedules/:id/runs — unwrap "none".
func (s *Schedules) ListRuns(ctx context.Context, id string, params *ListParams) ([]ScheduleRun, Pagination, *CorvoError) {
	return requestPage[ScheduleRun](ctx, s.t, schedulePath(id)+"/runs", listQuery(params))
}

// Imports issues invoices from a CSV or Excel spreadsheet: Start, wait for
// "ready", fix the rows, then Emit.
type Imports struct {
	t *Transport
}

func importPath(id string) string {
	return "/imports/" + url.PathEscape(id)
}

// Start uploads a .csv or .xlsx file (up to 2,000 rows and 10 MB) as
// multipart/form-data. companyID is optional: it issues every row that has no
// CNPJ column value. Reading is asynchronous: follow the import with Get until
// its status is "ready". An empty fileName defaults to "import.csv".
// POST /imports — unwrap "data". Returns 202.
func (i *Imports) Start(ctx context.Context, file []byte, fileName string, companyID string) (ImportStart, *CorvoError) {
	if fileName == "" {
		fileName = "import.csv"
	}
	contentType := "text/csv"
	if strings.EqualFold(filepath.Ext(fileName), ".xlsx") {
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	}

	var form bytes.Buffer
	writer := multipart.NewWriter(&form)
	buildError := func(err error) *CorvoError {
		return &CorvoError{
			Status:  0,
			Code:    "encode_error",
			Message: "Failed to build multipart body: " + err.Error(),
		}
	}
	if companyID != "" {
		if err := writer.WriteField("companyId", companyID); err != nil {
			return ImportStart{}, buildError(err)
		}
	}
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", mime.FormatMediaType("form-data", map[string]string{
		"name":     "file",
		"filename": fileName,
	}))
	header.Set("Content-Type", contentType)
	part, err := writer.CreatePart(header)
	if err != nil {
		return ImportStart{}, buildError(err)
	}
	if _, err := part.Write(file); err != nil {
		return ImportStart{}, buildError(err)
	}
	if err := writer.Close(); err != nil {
		return ImportStart{}, buildError(err)
	}

	return requestJSON[ImportStart](ctx, i.t, requestSpec{
		method:      http.MethodPost,
		path:        "/imports",
		body:        bytes.NewReader(form.Bytes()),
		contentType: writer.FormDataContentType(),
	})
}

// List returns the recent imports, newest first. limit is 1 to 20 (default
// 10; 0 uses the default). Not paginated. GET /imports — unwrap "data".
func (i *Imports) List(ctx context.Context, limit int) ([]ImportJob, *CorvoError) {
	query := url.Values{}
	queryAddInt(query, "limit", limit)
	return requestJSON[[]ImportJob](ctx, i.t, requestSpec{
		method: http.MethodGet,
		path:   "/imports",
		query:  query,
	})
}

// Get returns the status and counters of an import.
// GET /imports/:id — unwrap "data".
func (i *Imports) Get(ctx context.Context, id string) (ImportJob, *CorvoError) {
	return requestJSON[ImportJob](ctx, i.t, requestSpec{
		method: http.MethodGet,
		path:   importPath(id),
	})
}

// Emit sends every valid row to be issued; it is safe to call twice. params is
// optional. POST /imports/:id/emit — unwrap "data". Returns 202.
func (i *Imports) Emit(ctx context.Context, id string, params *ImportEmitParams) (ImportJob, *CorvoError) {
	spec := requestSpec{
		method: http.MethodPost,
		path:   importPath(id) + "/emit",
	}
	if params != nil {
		spec.jsonBody = params
	}
	return requestJSON[ImportJob](ctx, i.t, spec)
}

func importRowQuery(params *ImportRowListParams) url.Values {
	query := url.Values{}
	if params == nil {
		return query
	}
	queryAdd(query, "status", string(params.Status))
	queryAddInt(query, "limit", params.Limit)
	queryAdd(query, "cursor", params.Cursor)
	return query
}

// ListRows returns one page of the rows of an import (default 50, at most
// 200). GET /imports/:id/rows — unwrap "none".
func (i *Imports) ListRows(ctx context.Context, id string, params *ImportRowListParams) ([]ImportRow, Pagination, *CorvoError) {
	return requestPage[ImportRow](ctx, i.t, importPath(id)+"/rows", importRowQuery(params))
}

// IterateRows walks every page of the rows of an import, calling fn per row
// until it returns false or all pages are consumed.
func (i *Imports) IterateRows(ctx context.Context, id string, params *ImportRowListParams, fn func(ImportRow) bool) *CorvoError {
	return iteratePages[ImportRow](ctx, i.t, importPath(id)+"/rows", importRowQuery(params), fn)
}

// UpdateRow corrects one row (line is the line of the file) and checks it
// again; only while the import is "ready". item holds the fields of the invoice
// to change. PATCH /imports/:id/rows/:line — unwrap "data".
func (i *Imports) UpdateRow(ctx context.Context, id string, line int, item map[string]any) (ImportRow, *CorvoError) {
	return requestJSON[ImportRow](ctx, i.t, requestSpec{
		method: http.MethodPatch,
		path:   importPath(id) + "/rows/" + strconv.Itoa(line),
		jsonBody: struct {
			Item map[string]any `json:"item,omitempty"`
		}{Item: item},
	})
}

// Discard drops an import that has not been emitted.
// DELETE /imports/:id — 204 on success.
func (i *Imports) Discard(ctx context.Context, id string) *CorvoError {
	_, err := requestJSON[any](ctx, i.t, requestSpec{
		method: http.MethodDelete,
		path:   importPath(id),
	})
	return err
}

// DownloadTemplate downloads the spreadsheet template (CSV).
// GET /imports/template — raw binary.
func (i *Imports) DownloadTemplate(ctx context.Context) (DownloadedFile, *CorvoError) {
	return requestFile(ctx, i.t, http.MethodGet, "/imports/template")
}

// DownloadErrors downloads the CSV with the rows that need attention.
// GET /imports/:id/errors.csv — raw binary.
func (i *Imports) DownloadErrors(ctx context.Context, id string) (DownloadedFile, *CorvoError) {
	return requestFile(ctx, i.t, http.MethodGet, importPath(id)+"/errors.csv")
}
