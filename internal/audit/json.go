package audit

import "encoding/json"

func mustJSON(v any) string {
	if v == nil {
		return "{}"
	}
	b, _ := json.Marshal(v)
	return string(b)
}

func unmarshalJSON(b []byte) map[string]any {
	if len(b) == 0 {
		return map[string]any{}
	}
	out := map[string]any{}
	if err := json.Unmarshal(b, &out); err != nil {
		return map[string]any{}
	}
	return out
}
