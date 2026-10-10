package pagination

import (
	"encoding/json"
	"errors"
	"net/url"
	"testing"
)

func TestParse_Defaults(t *testing.T) {
	p, err := Parse(url.Values{})
	if err != nil {
		t.Fatal(err)
	}
	if p.Page != 1 || p.PerPage != DefaultPerPage {
		t.Fatalf("got %+v", p)
	}
	if p.Offset() != 0 || p.Limit() != DefaultPerPage {
		t.Fatalf("offset/limit wrong: %d/%d", p.Offset(), p.Limit())
	}
}

func TestParse_ExplicitValues(t *testing.T) {
	p, err := Parse(url.Values{"page": {"3"}, "per_page": {"10"}})
	if err != nil {
		t.Fatal(err)
	}
	if p.Offset() != 20 || p.Limit() != 10 {
		t.Fatalf("offset/limit wrong: %d/%d", p.Offset(), p.Limit())
	}
}

func TestParse_RejectsBadValues(t *testing.T) {
	cases := []struct {
		name  string
		query url.Values
		field string
	}{
		{"zero page", url.Values{"page": {"0"}}, "page"},
		{"negative page", url.Values{"page": {"-1"}}, "page"},
		{"non-numeric page", url.Values{"page": {"abc"}}, "page"},
		{"zero per_page", url.Values{"per_page": {"0"}}, "per_page"},
		{"per_page over max", url.Values{"per_page": {"101"}}, "per_page"},
		{"huge page", url.Values{"page": {"99999999999999999999"}}, "page"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse(tc.query)
			var perr *Error
			if !errors.As(err, &perr) {
				t.Fatalf("expected *Error, got %v", err)
			}
			if perr.Field != tc.field {
				t.Fatalf("field = %q, want %q", perr.Field, tc.field)
			}
		})
	}
}

func TestNew_TotalPages(t *testing.T) {
	cases := []struct{ total, perPage, want int }{
		{0, 20, 0}, {1, 20, 1}, {20, 20, 1}, {21, 20, 2}, {95, 20, 5},
	}
	for _, tc := range cases {
		got := New([]int{}, Params{Page: 1, PerPage: tc.perPage}, tc.total).TotalPages
		if got != tc.want {
			t.Errorf("total=%d perPage=%d: got %d pages, want %d", tc.total, tc.perPage, got, tc.want)
		}
	}
}

func TestNew_NilItemsEncodesAsEmptyArray(t *testing.T) {
	out, err := json.Marshal(New[string](nil, Params{Page: 1, PerPage: 20}, 0))
	if err != nil {
		t.Fatal(err)
	}
	want := `{"items":[],"page":1,"per_page":20,"total":0,"total_pages":0}`
	if string(out) != want {
		t.Fatalf("got %s, want %s", out, want)
	}
}
