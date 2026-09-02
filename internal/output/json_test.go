package output

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestJSONRendererWritesSuccessEnvelope(t *testing.T) {
	renderer := JSONRenderer{}
	var output bytes.Buffer

	err := renderer.Render(&output, SuccessResult("transactions", []map[string]any{}, 0))
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	var decoded struct {
		Success bool             `json:"success"`
		Type    string           `json:"type"`
		Data    []map[string]any `json:"data"`
		Count   int              `json:"count"`
	}
	if err := json.Unmarshal(output.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if !decoded.Success || decoded.Type != "transactions" || decoded.Count != 0 {
		t.Fatalf("decoded envelope = %#v", decoded)
	}
	if decoded.Data == nil {
		t.Fatal("decoded data is nil, want an empty array")
	}
}

func TestJSONRendererWritesFailureEnvelope(t *testing.T) {
	renderer := JSONRenderer{}
	var output bytes.Buffer

	err := renderer.Render(&output, FailureResult("transactions", Error{
		Code:    "NOT_FOUND",
		Message: "transaction not found",
	}))
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	var decoded Result
	if err := json.Unmarshal(output.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if decoded.Success || decoded.Error == nil || decoded.Error.Code != "NOT_FOUND" {
		t.Fatalf("decoded failure = %#v", decoded)
	}
}

func TestParseFormat(t *testing.T) {
	for _, value := range []string{"table", "json"} {
		if _, err := ParseFormat(value); err != nil {
			t.Fatalf("ParseFormat(%q) returned error: %v", value, err)
		}
	}
	if _, err := ParseFormat("csv"); err == nil {
		t.Fatal("ParseFormat(csv) expected an error")
	}
}
