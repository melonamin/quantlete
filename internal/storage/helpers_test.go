package storage

import (
	"testing"
)

func TestExpandSliceParams(t *testing.T) {
	tests := []struct {
		name        string
		sql         string
		args        []any
		wantSQL     string
		wantArgs    []any
		wantErr     bool
		errContains string
	}{
		{
			name:     "no slice params",
			sql:      "SELECT * FROM users WHERE id = ?1",
			args:     []any{123},
			wantSQL:  "SELECT * FROM users WHERE id = ?1",
			wantArgs: []any{123},
		},
		{
			name:     "single slice param",
			sql:      "SELECT * FROM users WHERE id IN (/*SLICE:ids*/?1)",
			args:     []any{[]int{1, 2, 3}},
			wantSQL:  "SELECT * FROM users WHERE id IN (?, ?, ?)",
			wantArgs: []any{1, 2, 3},
		},
		{
			name:     "slice with non-slice param",
			sql:      "SELECT * FROM users WHERE status = ?1 AND id IN (/*SLICE:ids*/?2)",
			args:     []any{"active", []int{1, 2, 3}},
			wantSQL:  "SELECT * FROM users WHERE status = ?1 AND id IN (?, ?, ?)",
			wantArgs: []any{"active", 1, 2, 3},
		},
		{
			name:     "multiple different slice params",
			sql:      "SELECT * FROM users WHERE id IN (/*SLICE:ids*/?1) OR name IN (/*SLICE:names*/?2)",
			args:     []any{[]int{1, 2}, []string{"alice", "bob", "charlie"}},
			wantSQL:  "SELECT * FROM users WHERE id IN (?, ?) OR name IN (?, ?, ?)",
			wantArgs: []any{1, 2, "alice", "bob", "charlie"},
		},
		{
			name:     "single element slice",
			sql:      "SELECT * FROM users WHERE id IN (/*SLICE:ids*/?1)",
			args:     []any{[]int{42}},
			wantSQL:  "SELECT * FROM users WHERE id IN (?)",
			wantArgs: []any{42},
		},
		{
			name:     "string slice",
			sql:      "SELECT * FROM users WHERE name IN (/*SLICE:names*/?1)",
			args:     []any{[]string{"alice", "bob"}},
			wantSQL:  "SELECT * FROM users WHERE name IN (?, ?)",
			wantArgs: []any{"alice", "bob"},
		},
		{
			name:     "int64 slice",
			sql:      "SELECT * FROM activities WHERE id IN (/*SLICE:ids*/?1)",
			args:     []any{[]int64{1000000000001, 1000000000002}},
			wantSQL:  "SELECT * FROM activities WHERE id IN (?, ?)",
			wantArgs: []any{int64(1000000000001), int64(1000000000002)},
		},
		{
			name:     "float64 slice",
			sql:      "SELECT * FROM data WHERE value IN (/*SLICE:vals*/?1)",
			args:     []any{[]float64{1.5, 2.5}},
			wantSQL:  "SELECT * FROM data WHERE value IN (?, ?)",
			wantArgs: []any{1.5, 2.5},
		},
		{
			name:        "empty slice",
			sql:         "SELECT * FROM users WHERE id IN (/*SLICE:ids*/?1)",
			args:        []any{[]int{}},
			wantErr:     true,
			errContains: "empty slice not allowed",
		},
		{
			name:        "slice exceeds max length",
			sql:         "SELECT * FROM users WHERE id IN (/*SLICE:ids*/?1)",
			args:        []any{make([]int, MaxSliceParamLength+1)},
			wantErr:     true,
			errContains: "exceeds maximum",
		},
		{
			name:        "param out of range",
			sql:         "SELECT * FROM users WHERE id IN (/*SLICE:ids*/?5)",
			args:        []any{[]int{1, 2, 3}},
			wantErr:     true,
			errContains: "out of range",
		},
		{
			name:        "non-slice arg for slice placeholder",
			sql:         "SELECT * FROM users WHERE id IN (/*SLICE:ids*/?1)",
			args:        []any{123},
			wantErr:     true,
			errContains: "expected slice",
		},
		{
			name:        "repeated slice param",
			sql:         "SELECT * FROM users WHERE id IN (/*SLICE:ids*/?1) OR backup_id IN (/*SLICE:ids*/?1)",
			args:        []any{[]int{1, 2, 3}},
			wantErr:     true,
			errContains: "used more than once",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSQL, gotArgs, err := expandSliceParams(tt.sql, tt.args)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.errContains)
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("error = %q, want containing %q", err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if gotSQL != tt.wantSQL {
				t.Errorf("SQL = %q, want %q", gotSQL, tt.wantSQL)
			}

			if len(gotArgs) != len(tt.wantArgs) {
				t.Errorf("args length = %d, want %d", len(gotArgs), len(tt.wantArgs))
				return
			}

			for i, want := range tt.wantArgs {
				if gotArgs[i] != want {
					t.Errorf("args[%d] = %v (%T), want %v (%T)", i, gotArgs[i], gotArgs[i], want, want)
				}
			}
		})
	}
}

func TestExpandSliceParams_MaxLength(t *testing.T) {
	// Test exactly at the limit should succeed
	slice := make([]int, MaxSliceParamLength)
	for i := range slice {
		slice[i] = i
	}

	sql := "SELECT * FROM users WHERE id IN (/*SLICE:ids*/?1)"
	gotSQL, gotArgs, err := expandSliceParams(sql, []any{slice})

	if err != nil {
		t.Fatalf("at max length should succeed: %v", err)
	}

	if len(gotArgs) != MaxSliceParamLength {
		t.Errorf("args length = %d, want %d", len(gotArgs), MaxSliceParamLength)
	}

	// Verify SQL has correct number of placeholders
	placeholderCount := 0
	for _, c := range gotSQL {
		if c == '?' {
			placeholderCount++
		}
	}
	if placeholderCount != MaxSliceParamLength {
		t.Errorf("placeholder count = %d, want %d", placeholderCount, MaxSliceParamLength)
	}
}

func TestSliceLength(t *testing.T) {
	tests := []struct {
		name    string
		arg     any
		want    int
		wantErr bool
	}{
		{"[]int", []int{1, 2, 3}, 3, false},
		{"[]int64", []int64{1, 2}, 2, false},
		{"[]string", []string{"a", "b"}, 2, false},
		{"[]float64", []float64{1.1, 2.2}, 2, false},
		{"[]any", []any{1, "two", 3.0}, 3, false},
		{"empty []int", []int{}, 0, false},
		{"non-slice", 123, 0, true},
		{"string", "not a slice", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sliceLength(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("sliceLength() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("sliceLength() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestExpandSlice(t *testing.T) {
	tests := []struct {
		name    string
		arg     any
		want    []any
		wantErr bool
	}{
		{"[]int", []int{1, 2, 3}, []any{1, 2, 3}, false},
		{"[]int64", []int64{10, 20}, []any{int64(10), int64(20)}, false},
		{"[]string", []string{"a", "b"}, []any{"a", "b"}, false},
		{"[]float64", []float64{1.5, 2.5}, []any{1.5, 2.5}, false},
		{"[]any passthrough", []any{1, "two"}, []any{1, "two"}, false},
		{"non-slice", 123, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := expandSlice(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("expandSlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("expandSlice() length = %d, want %d", len(got), len(tt.want))
					return
				}
				for i := range got {
					if got[i] != tt.want[i] {
						t.Errorf("expandSlice()[%d] = %v, want %v", i, got[i], tt.want[i])
					}
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	if substr == "" {
		return true
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
