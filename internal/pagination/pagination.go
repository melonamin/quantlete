// Package pagination provides utilities for consistent pagination across repositories.
package pagination

import (
	"net/url"
	"strconv"
)

// Pagination defaults and limits.
const (
	DefaultPage    = 1
	DefaultPerPage = 50
	MaxPerPage     = 200
	MaxPage        = 10000
)

// Params holds normalized pagination parameters.
type Params struct {
	Page    int
	PerPage int
}

// NewParams creates normalized pagination parameters from raw input values.
// It applies default values and enforces min/max constraints.
func NewParams(page, perPage int) Params {
	p := Params{
		Page:    page,
		PerPage: perPage,
	}
	p.Normalize()
	return p
}

// Normalize applies default values and enforces constraints.
func (p *Params) Normalize() {
	if p.Page < DefaultPage {
		p.Page = DefaultPage
	}
	if p.Page > MaxPage {
		p.Page = MaxPage
	}
	if p.PerPage < 1 {
		p.PerPage = DefaultPerPage
	}
	if p.PerPage > MaxPerPage {
		p.PerPage = MaxPerPage
	}
}

// Offset calculates the SQL OFFSET value for the current page.
func (p Params) Offset() int {
	return (p.Page - 1) * p.PerPage
}

// TotalPages calculates the total number of pages given the total item count.
func (p Params) TotalPages(total int) int {
	if total <= 0 {
		return 1
	}
	pages := (total + p.PerPage - 1) / p.PerPage
	if pages < 1 {
		return 1
	}
	return pages
}

// Result holds the pagination metadata for a paginated response.
type Result struct {
	Total      int `json:"total"`
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalPages int `json:"total_pages"`
}

// NewResult creates a Result from Params and total count.
func NewResult(p Params, total int) Result {
	return Result{
		Total:      total,
		Page:       p.Page,
		PerPage:    p.PerPage,
		TotalPages: p.TotalPages(total),
	}
}

// OrderDirection represents a sort direction.
type OrderDirection string

const (
	OrderAsc  OrderDirection = "ASC"
	OrderDesc OrderDirection = "DESC"
)

// ParseOrderDir parses a direction string into an OrderDirection.
// Returns OrderAsc for "asc" and OrderDesc for "desc" or unknown values.
func ParseOrderDir(dir string) OrderDirection {
	if dir == "asc" {
		return OrderAsc
	}
	return OrderDesc
}

// ValidateOrderBy checks if the given column is in the allowed set and returns
// the safe column name. Returns empty string and false if not valid.
func ValidateOrderBy(column string, allowed map[string]string) (string, bool) {
	if col, ok := allowed[column]; ok {
		return col, true
	}
	return "", false
}

// BuildOrderClause builds a safe ORDER BY clause from column and direction.
// Returns default if column validation fails.
func BuildOrderClause(column, direction string, allowed map[string]string, defaultClause string) string {
	col, ok := ValidateOrderBy(column, allowed)
	if !ok {
		return defaultClause
	}
	return col + " " + string(ParseOrderDir(direction))
}

// OrderColumn defines a sortable column with optional NULL handling.
type OrderColumn struct {
	Column    string // SQL column expression
	NullsLast bool   // Append NULLS LAST to the ORDER BY clause
}

// BuildOrderClauseExt builds a safe ORDER BY clause with extended column options.
// Supports NULLS LAST for columns that may contain NULL values.
func BuildOrderClauseExt(column, direction string, allowed map[string]OrderColumn, defaultClause string) string {
	oc, ok := allowed[column]
	if !ok {
		return defaultClause
	}
	clause := oc.Column + " " + string(ParseOrderDir(direction))
	if oc.NullsLast {
		clause += " NULLS LAST"
	}
	return clause
}

// ParseFromQuery parses pagination parameters from URL query values.
// Returns normalized Params with defaults applied and constraints enforced.
func ParseFromQuery(q url.Values) Params {
	page := DefaultPage
	perPage := DefaultPerPage

	if p := q.Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil {
			page = parsed
		}
	}
	if pp := q.Get("per_page"); pp != "" {
		if parsed, err := strconv.Atoi(pp); err == nil {
			perPage = parsed
		}
	}

	return NewParams(page, perPage)
}

// ParseSortFromQuery parses sorting parameters from URL query values.
// Returns order_by and order_dir values.
func ParseSortFromQuery(q url.Values) (orderBy, orderDir string) {
	return q.Get("order_by"), q.Get("order_dir")
}

// QueryParams combines pagination and sorting parameters.
// This struct can be embedded into filter types for consolidated parsing.
type QueryParams struct {
	Page     int
	PerPage  int
	OrderBy  string
	OrderDir string
}

// ParseQueryParams parses pagination and sorting from URL query values.
// Returns all common filter parameters in a single call.
func ParseQueryParams(q url.Values) QueryParams {
	p := ParseFromQuery(q)
	return QueryParams{
		Page:     p.Page,
		PerPage:  p.PerPage,
		OrderBy:  q.Get("order_by"),
		OrderDir: q.Get("order_dir"),
	}
}
