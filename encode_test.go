package vcon

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestToJSON(t *testing.T) {
	v := NewWithDefaults()

	data, err := v.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("ToJSON returned empty data")
	}

	// Verify it's valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Errorf("ToJSON produced invalid JSON: %v", err)
	}
}

func TestToJSONIndent(t *testing.T) {
	v := NewWithDefaults()

	data, err := v.ToJSONIndent()
	if err != nil {
		t.Fatalf("ToJSONIndent failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("ToJSONIndent returned empty data")
	}

	// Check for indentation
	if !bytes.Contains(data, []byte("\n  ")) {
		t.Error("ToJSONIndent should produce indented JSON")
	}
}

func TestEncode(t *testing.T) {
	v := NewWithDefaults()

	var buf bytes.Buffer
	err := Encode(&buf, v)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("Encode wrote no data")
	}

	// Check for indentation (Encode uses indentation)
	if !bytes.Contains(buf.Bytes(), []byte("\n  ")) {
		t.Error("Encode should produce indented JSON")
	}
}

func TestEncodeCompact(t *testing.T) {
	v := NewWithDefaults()

	var buf bytes.Buffer
	err := EncodeCompact(&buf, v)
	if err != nil {
		t.Fatalf("EncodeCompact failed: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("EncodeCompact wrote no data")
	}

	// Should not have indentation beyond the newline at the end
	content := buf.String()
	lines := bytes.Split([]byte(content), []byte("\n"))
	if len(lines) > 2 { // One line plus possible trailing newline
		t.Error("EncodeCompact should produce compact JSON")
	}
}

func TestMarshal(t *testing.T) {
	v := NewWithDefaults()

	data, err := Marshal(v)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("Marshal returned empty data")
	}

	// Should be equivalent to ToJSON
	data2, _ := v.ToJSON()
	if !bytes.Equal(data, data2) {
		t.Error("Marshal should be equivalent to ToJSON")
	}
}

func TestMarshalIndent(t *testing.T) {
	v := NewWithDefaults()

	data, err := MarshalIndent(v)
	if err != nil {
		t.Fatalf("MarshalIndent failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("MarshalIndent returned empty data")
	}

	// Should be equivalent to ToJSONIndent
	data2, _ := v.ToJSONIndent()
	if !bytes.Equal(data, data2) {
		t.Error("MarshalIndent should be equivalent to ToJSONIndent")
	}
}

func TestCompactJSON(t *testing.T) {
	indented := []byte(`{
  "uuid": "test",
  "vcon": "1.0.0",
  "created_at": "2024-01-01T00:00:00Z"
}`)

	compacted, err := CompactJSON(indented)
	if err != nil {
		t.Fatalf("CompactJSON failed: %v", err)
	}

	// Should remove whitespace
	if bytes.Contains(compacted, []byte("\n  ")) {
		t.Error("CompactJSON should remove indentation")
	}

	// Should still be valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal(compacted, &result); err != nil {
		t.Errorf("CompactJSON produced invalid JSON: %v", err)
	}
}

func TestCompactJSONInvalid(t *testing.T) {
	invalid := []byte(`{invalid json`)

	_, err := CompactJSON(invalid)
	if err == nil {
		t.Error("CompactJSON should fail on invalid JSON")
	}
}
