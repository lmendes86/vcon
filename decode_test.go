package vcon

import (
	"bytes"
	"strings"
	"testing"
)

func TestFromJSON(t *testing.T) {
	jsonData := []byte(`{
		"uuid": "550e8400-e29b-41d4-a716-446655440000",
		"vcon": "1.0.0",
		"created_at": "2024-01-01T00:00:00Z",
		"parties": [
			{
				"mailto": "test@example.com"
			}
		]
	}`)

	v, err := FromJSON(jsonData)
	if err != nil {
		t.Fatalf("FromJSON failed: %v", err)
	}

	if v == nil {
		t.Fatal("FromJSON returned nil")
	}

	if v.Vcon != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", v.Vcon)
	}

	if len(v.Parties) != 1 {
		t.Errorf("Expected 1 party, got %d", len(v.Parties))
	}
}

func TestFromJSONInvalid(t *testing.T) {
	invalidJSON := []byte(`{invalid json}`)

	_, err := FromJSON(invalidJSON)
	if err == nil {
		t.Error("FromJSON should fail on invalid JSON")
	}

	if !strings.Contains(err.Error(), "failed to unmarshal") {
		t.Errorf("Expected unmarshal error, got: %v", err)
	}
}

func TestReadFrom(t *testing.T) {
	jsonData := `{
		"uuid": "550e8400-e29b-41d4-a716-446655440000",
		"vcon": "1.0.0",
		"created_at": "2024-01-01T00:00:00Z"
	}`

	reader := strings.NewReader(jsonData)
	v, err := ReadFrom(reader)
	if err != nil {
		t.Fatalf("ReadFrom failed: %v", err)
	}

	if v == nil {
		t.Fatal("ReadFrom returned nil")
	}

	if v.Vcon != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", v.Vcon)
	}
}

func TestReadFromError(t *testing.T) {
	// Test with a reader that fails
	reader := &failingReader{}
	_, err := ReadFrom(reader)
	if err == nil {
		t.Error("ReadFrom should fail when reader fails")
	}
}

func TestDecode(t *testing.T) {
	jsonData := `{
		"uuid": "550e8400-e29b-41d4-a716-446655440000",
		"vcon": "1.0.0",
		"created_at": "2024-01-01T00:00:00Z"
	}`

	reader := strings.NewReader(jsonData)
	v, err := Decode(reader)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if v == nil {
		t.Fatal("Decode returned nil")
	}

	if v.Vcon != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", v.Vcon)
	}
}

func TestDecodeInvalid(t *testing.T) {
	invalidJSON := `{invalid json}`

	reader := strings.NewReader(invalidJSON)
	_, err := Decode(reader)
	if err == nil {
		t.Error("Decode should fail on invalid JSON")
	}

	if !strings.Contains(err.Error(), "failed to decode") {
		t.Errorf("Expected decode error, got: %v", err)
	}
}

func TestUnmarshal(t *testing.T) {
	jsonData := []byte(`{
		"uuid": "550e8400-e29b-41d4-a716-446655440000",
		"vcon": "1.0.0",
		"created_at": "2024-01-01T00:00:00Z"
	}`)

	v, err := Unmarshal(jsonData)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if v == nil {
		t.Fatal("Unmarshal returned nil")
	}

	// Should be equivalent to FromJSON
	v2, _ := FromJSON(jsonData)
	if v.Vcon != v2.Vcon {
		t.Error("Unmarshal should be equivalent to FromJSON")
	}
}

func TestParseString(t *testing.T) {
	jsonString := `{
		"uuid": "550e8400-e29b-41d4-a716-446655440000",
		"vcon": "1.0.0",
		"created_at": "2024-01-01T00:00:00Z"
	}`

	v, err := ParseString(jsonString)
	if err != nil {
		t.Fatalf("ParseString failed: %v", err)
	}

	if v == nil {
		t.Fatal("ParseString returned nil")
	}

	if v.Vcon != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", v.Vcon)
	}
}

func TestValidateJSON(t *testing.T) {
	validJSON := []byte(`{
		"uuid": "550e8400-e29b-41d4-a716-446655440000",
		"vcon": "1.0.0",
		"created_at": "2024-01-01T00:00:00Z"
	}`)

	err := ValidateJSON(validJSON)
	if err != nil {
		t.Errorf("ValidateJSON failed on valid JSON: %v", err)
	}

	invalidJSON := []byte(`{invalid json}`)
	err = ValidateJSON(invalidJSON)
	if err == nil {
		t.Error("ValidateJSON should fail on invalid JSON")
	}
}

func TestDecodeAndValidate(t *testing.T) {
	// Test with valid VCon
	validJSON := `{
		"uuid": "550e8400-e29b-41d4-a716-446655440000",
		"vcon": "0.0.2",
		"created_at": "2024-01-01T00:00:00Z",
		"parties": [
			{"name": "Test User"}
		]
	}`

	reader := strings.NewReader(validJSON)
	v, err := DecodeAndValidate(reader)
	if err != nil {
		t.Fatalf("DecodeAndValidate failed on valid VCon: %v", err)
	}

	if v == nil {
		t.Fatal("DecodeAndValidate returned nil")
	}

	// Test with invalid VCon (missing required fields)
	invalidVCon := `{
		"vcon": "1.0.0"
	}`

	reader2 := strings.NewReader(invalidVCon)
	_, err = DecodeAndValidate(reader2)
	if err == nil {
		t.Error("DecodeAndValidate should fail on invalid VCon")
	}

	if !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("Expected validation error, got: %v", err)
	}
}

// Test helper - a reader that always fails.
type failingReader struct{}

func (f *failingReader) Read(_ []byte) (n int, err error) {
	return 0, bytes.ErrTooLarge
}
