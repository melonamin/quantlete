package storage

import (
	"encoding/json"
	"fmt"
)

func decodeFloat64Array(raw json.RawMessage) ([]float64, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var anySlice []any
	if err := json.Unmarshal(raw, &anySlice); err != nil {
		return nil, fmt.Errorf("decoding json array: %w", err)
	}
	out := make([]float64, 0, len(anySlice))
	for _, v := range anySlice {
		switch n := v.(type) {
		case float64:
			out = append(out, n)
		case int:
			out = append(out, float64(n))
		default:
			// Ignore unexpected types.
		}
	}
	return out, nil
}

func DecodeFloat64Array(raw json.RawMessage) ([]float64, error) {
	return decodeFloat64Array(raw)
}
