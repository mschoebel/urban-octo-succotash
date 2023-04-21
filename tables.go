package uos

import (
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

// TableData represents a table with rows and columns.
type TableData [][]interface{}

// TableSpec describes the interface every web application table must implement.
type TableSpec interface {
	// Name returns the short name of the table. The table data is available at '/tables/<name>'.
	Name() string
	// ModelName returns the name of the underlying DB model. Can return "" to indicate a custom table.
	// A returned model must be registered (using RegisterDBModels).
	ModelName() string

	// LoadData returns the table data according to the specified configuration.
	LoadData(TableConfiguration) (TableData, error)

	// ColumnInfo returns table column information for the selected columns.
	ColumnInfo([]string) []TableColumn
	// ColumnDefault returns list of default columns
	ColumnDefault() []string

	// Actions returns a set of actions related to the table and their contained items.
	Actions() *TableActions
}

// TableDisplay describes the interface a web application table must implement to
// customize the table visualization
type TableDisplay interface {
	// DisplaySettings returns custom table display settings
	DisplaySettings() TableDisplayProperties
}

// TableDelete must be implemented by a web application form to support DELETE requests.
type TableSpecDelete interface {
	// Delete removes the specified items from the database.
	Delete(ids []uint) (*ResponseAction, error)
}

// TableAction describes a function that can be triggered for a table, e.g. item deletion.
type TableAction struct {
	// (LineAwesome) icon of the action
	Icon string
	// button text
	Text string
	// button class
	ButtonClass string

	// hx-<method>
	Method string
	// action target, e.g. "/dialogs/..."
	TargetURL string
	// CSS selector for elements to include in the resulting request
	Include string
}

// TableActionButton returns a table action element with the specified icon, text and button class.
func TableActionButton(icon, text, class string) TableAction {
	return TableAction{
		Icon:        icon,
		Text:        text,
		ButtonClass: class,
	}
}

// Dialog extends the action to open the specified dialog
func (a TableAction) Dialog(target string) TableAction {
	a.Method = "get"
	a.TargetURL = fmt.Sprintf("/dialogs/%s", target)
	return a
}

// Dialog extends the action to open the specified dialog
func (a TableAction) Post(target, include string) TableAction {
	a.Method = "post"
	a.TargetURL = target
	a.Include = include
	return a
}

type TableActions []TableAction

// TableHandler returns a handler for the "/tables/" route providing the specified tables.
// The handler can be activated using RegisterAppRequestHandlers.
func TableHandler(tables ...TableSpec) AppRequestHandlerMapping {
	return AppRequestHandlerMapping{
		Route:   "/tables/",
		Handler: getTableHandlerFunc(tables),
	}
}

func getTableHandlerFunc(tables []TableSpec) AppRequestHandler {
	nameToSpec := map[string]TableSpec{}
	for _, t := range tables {
		nameToSpec[t.Name()] = t
		Log.DebugContext("register table spec", LogContext{"name": t.Name()})
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// determine table
		tableName := getElementName("tables", r.URL.Path)
		Log.DebugContext(
			"handle table",
			LogContext{
				"name":   tableName,
				"method": r.Method,
			},
		)

		tableSpec, ok := nameToSpec[tableName]
		if !ok {
			RespondNotFound(w)
			return
		}

		// prepare request processing (URL form data might be empty)
		err := r.ParseForm()
		if err != nil {
			Log.WarnError("could not parse form", err)
			RespondBadRequest(w)
			return
		}

		// process request
		switch r.Method {
		case http.MethodGet:
			renderTable(w, r, tableSpec, r.Form)
		case http.MethodPost:
			// does the table support DELETE method?
			tableDelete, ok := tableSpec.(TableSpecDelete)
			if !ok {
				RespondNotImplemented(w)
				return
			}

			// extract IDs from request
			ids := []uint{}
			for key := range r.Form {
				if strings.HasPrefix(key, "item-") {
					if id, err := strconv.Atoi(key[5:]); err == nil && id >= 0 {
						ids = append(ids, uint(id))
					}
				}
			}

			// forward to table handler
			action, err := tableDelete.Delete(ids)
			if err != nil {
				handleFormError(w, "could not delete table items", err)
				return
			}

			handleResponseAction(w, r, action)
		default:
			RespondNotImplemented(w)
		}
	}
}

// TableConfiguration describes a specific table view, e.g. a specific column selection,
// page or sort column.
type TableConfiguration struct {
	// comma separated list of selected columns
	Columns []string

	// sort column
	SortColumn string
	// sort mode (ASC or DESC or empty)
	SortMode string

	// selected page
	Page int
	// number of rows to return
	Rows int
}

func (c TableConfiguration) DBQuery() *gorm.DB {
	dbQuery := DB

	if len(c.Columns) > 0 {
		dbQuery = dbQuery.Select(append([]string{"id"}, c.Columns...))
	}
	if c.SortColumn != "" {
		dbQuery = dbQuery.Order(
			strings.TrimSpace(
				strings.Join([]string{c.SortColumn, c.SortMode}, " "),
			),
		)
	}

	// additionally order by id to get a deterministic sequence
	dbQuery = dbQuery.Order("id")

	return dbQuery.Limit(c.Rows).Offset((c.Page - 1) * c.Rows)
}

// LoadTable loads table data for the specified model.
func (c TableConfiguration) LoadTable(dest interface{}) error {
	return c.DBQuery().Find(dest).Error
}

func newTableConfiguration(t TableSpec, form url.Values) TableConfiguration {
	var (
		columns = form.Get("cols")

		sortColumn = form.Get("sort")
		sortMode   = ""
	)

	switch {
	case strings.HasPrefix(sortColumn, "+"):
		sortColumn = sortColumn[1:]
	case strings.HasPrefix(sortColumn, "-"):
		sortColumn = sortColumn[1:]
		sortMode = "DESC"
	}

	config := TableConfiguration{
		Columns:    strings.Split(columns, ","),
		SortColumn: sortColumn,
		SortMode:   sortMode,
		Page:       stringToInt(form.Get("page"), 1),
		Rows:       stringToInt(form.Get("rows"), 10),
	}

	if columns == "" {
		config.Columns = t.ColumnDefault()
	}

	return config
}

func (c TableConfiguration) isValid() bool {
	// at least one column must exist
	if len(c.Columns) == 0 {
		return false
	}

	// columns must not contain empty entries and trailing/leading whitespace
	for _, col := range c.Columns {
		if col == "" || strings.TrimSpace(col) != col {
			return false
		}
	}

	// sort column (if defined) must be included in columns list
	if c.SortColumn != "" && len(c.Columns) > 0 && !contains(c.Columns, c.SortColumn) {
		return false
	}

	// page and rows must be >= 1
	if c.Page < 1 || c.Rows < 1 {
		return false
	}

	// rows must be <= 200
	if c.Rows > 200 {
		return false
	}

	return true
}

// TableFormatFunc describes a table cell visualization function. The input parameter
// are the row id and a raw value read from the database. The returned value is visualized
// in the frontend. The return value can contain HTML.
type TableFormatFunc func(uint, interface{}) interface{}

// TableColumn describes properties of a table column.
type TableColumn struct {
	// display name
	DisplayName string
	// table supports sorting by this column
	IsSortable bool

	// function to convert raw values to visualization (e.g. for formatting)
	Format TableFormatFunc

	// is current sortfield? (determined automatically)
	IsSortField bool
	// is sorted descending? (determined automatically)
	IsSortDesc bool
	// internal name (determined automatically)
	Name string
}

// TableDisplayProperties describes basic table visualization properties.
// Default: all properties set to false.
type TableDisplayProperties struct {
	// show as full element width
	IsFullWidth bool
	// highlight table rows on hover
	IsHoverable bool
	// emphasize every second row
	IsStriped bool
	// show table content as cards on small displays
	IsMobileReady bool
	// rows can be selected
	IsSelectable bool
}

type tableRenderContext struct {
	TableID string

	Data  TableData
	Count int64

	Config  TableConfiguration
	Columns []TableColumn
	Display TableDisplayProperties

	Actions    *TableActions
	HasActions bool

	IsEmpty   bool
	IsLoading bool

	ColumnWidth  int
	TableBaseURL string

	PageCount int
}

func newTableRenderContext(t TableSpec, form url.Values) (tableRenderContext, error) {
	// get configuration from URL parameter
	config := newTableConfiguration(t, form)
	if !config.isValid() {
		return tableRenderContext{}, ErrorTableInvalidRequest
	}

	// get column information
	columns := t.ColumnInfo(config.Columns)

	// table has model? -> extended column check
	validColumns := dbColumns(t.ModelName())
	if t.ModelName() != "" && len(config.Columns) > 1 {
		for _, c := range config.Columns {
			if !contains(validColumns, c) {
				return tableRenderContext{}, ErrorTableInvalidRequest
			}
		}
	}

	// load data
	data, err := t.LoadData(config)
	if err != nil {
		return tableRenderContext{}, err
	}

	count, err := dbEntryCount(t.ModelName())
	if err != nil {
		return tableRenderContext{}, err
	}

	// call formatting functions (if defined)
	for i, row := range data {
		for j, cell := range row {
			if j == 0 {
				// first entry contains ID - not shown in table
				continue
			}

			if columns[j-1].Format != nil {
				data[i][j] = columns[j-1].Format(row[0].(uint), cell)
			}
		}
	}

	// base initialization
	context := tableRenderContext{
		TableID: fmt.Sprintf("%s-%d", t.Name(), rand.Intn(999999)),
		Data:    data,
		Count:   count,
		Config:  config,
		Columns: columns,
		Actions: t.Actions(),
	}

	// custom display settings available?
	settingsProvider, ok := t.(TableDisplay)
	if ok {
		context.Display = settingsProvider.DisplaySettings()
	}

	// evaluate some properties to simplify context handling in template
	// .. any actions defined?
	context.HasActions = context.Actions != nil && len(*context.Actions) > 0
	// .. no data available?
	context.IsEmpty = len(data) == 0
	// .. determine number of columns for table message
	context.ColumnWidth = len(context.Config.Columns)
	if context.Display.IsSelectable {
		context.ColumnWidth += 1
	}
	// .. integrate sorting info in columns info
	for i, c := range context.Config.Columns {
		if c == context.Config.SortColumn {
			context.Columns[i].IsSortField = true
			context.Columns[i].IsSortDesc = context.Config.SortMode == "DESC"
		}
		context.Columns[i].Name = c
	}
	// .. initialize table base URL - direct interaction with HTML does not change
	//    nr of rows or selected columns and resets current page to the first one
	context.TableBaseURL = fmt.Sprintf(
		"/tables/%s?rows=%d&cols=%s",
		t.Name(),
		context.Config.Rows,
		strings.Join(context.Config.Columns, ","),
	)
	// .. calculate overall page count
	context.PageCount = int(math.Ceil(float64(count) / float64(context.Config.Rows)))

	return context, nil
}

func renderTable(w http.ResponseWriter, r *http.Request, t TableSpec, form url.Values) {
	context, err := newTableRenderContext(t, form)
	if err != nil {
		handleTableError(w, "could not create table render context", err)
		return
	}

	err = renderInternalTemplate(w, r, "table", context)
	if err != nil {
		Log.ErrorContext(
			"could not render table",
			LogContext{"name": t.Name(), "error": err},
		)
		RespondInternalServerError(w)
		return
	}
}

func handleTableError(w http.ResponseWriter, message string, err error) {
	switch err {
	case ErrorTableInvalidRequest:
		RespondBadRequest(w)
		return
	}

	// all other cases: log error and respond
	Log.ErrorObj(message, err)
	RespondInternalServerError(w)
}
