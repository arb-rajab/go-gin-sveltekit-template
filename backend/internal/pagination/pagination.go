// Package pagination parses page/per_page query parameters and builds the one
// list envelope every paginated endpoint returns:
//
//	{"items": [...], "page": 2, "per_page": 20, "total": 95, "total_pages": 5}
//
// It has no HTTP framework dependency, so it works in any handler or job.
package pagination

import (
	"fmt"
	"net/url"
	"strconv"
)

const (
	// DefaultPerPage is used when the caller sends no per_page.
	DefaultPerPage = 20
	// MaxPerPage caps per_page so a client cannot request an unbounded page.
	MaxPerPage = 100
)

// Params is a validated page request. The zero value is not valid; use Parse.
type Params struct {
	Page    int
	PerPage int
}

// Error names the query parameter that failed validation.
type Error struct {
	Field   string
	Message string
}

func (e *Error) Error() string { return e.Field + ": " + e.Message }

// Parse reads page (1-based, default 1) and per_page (default DefaultPerPage,
// at most MaxPerPage) from query. Anything else is an *Error, never silently
// clamped, so a client bug shows up instead of returning surprising data.
func Parse(query url.Values) (Params, error) {
	page, err := intParam(query, "page", 1, 1<<31-1)
	if err != nil {
		return Params{}, err
	}
	perPage, err := intParam(query, "per_page", DefaultPerPage, MaxPerPage)
	if err != nil {
		return Params{}, err
	}
	return Params{Page: page, PerPage: perPage}, nil
}

func intParam(query url.Values, name string, def, limit int) (int, error) {
	raw := query.Get(name)
	if raw == "" {
		return def, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 {
		return 0, &Error{Field: name, Message: name + " must be a positive integer"}
	}
	if n > limit {
		return 0, &Error{Field: name, Message: fmt.Sprintf("%s must be at most %d", name, limit)}
	}
	return n, nil
}

// Limit is the SQL LIMIT value.
func (p Params) Limit() int { return p.PerPage }

// Offset is the SQL OFFSET value. The page number is capped at parse time so
// Page*PerPage cannot overflow an int on 64-bit platforms.
func (p Params) Offset() int { return (p.Page - 1) * p.PerPage }

// Page is the list envelope.
type Page[T any] struct {
	Items      []T `json:"items"`
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// New builds the envelope for one page of items out of total matching rows.
// A nil items slice becomes an empty one so it encodes as [] rather than null.
func New[T any](items []T, p Params, total int) Page[T] {
	if items == nil {
		items = []T{}
	}
	totalPages := 0
	if total > 0 {
		totalPages = (total + p.PerPage - 1) / p.PerPage
	}
	return Page[T]{Items: items, Page: p.Page, PerPage: p.PerPage, Total: total, TotalPages: totalPages}
}
