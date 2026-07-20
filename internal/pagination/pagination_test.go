package pagination

import "testing"

func TestParseOrderDirCaseInsensitive(t *testing.T) {
	tests := []struct {
		input string
		want  OrderDirection
	}{
		{input: "asc", want: OrderAsc},
		{input: "ASC", want: OrderAsc},
		{input: "aSc", want: OrderAsc},
		{input: "desc", want: OrderDesc},
		{input: "DESC", want: OrderDesc},
		{input: "unknown", want: OrderDesc},
		{input: "", want: OrderDesc},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			if got := ParseOrderDir(test.input); got != test.want {
				t.Errorf("ParseOrderDir(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}
