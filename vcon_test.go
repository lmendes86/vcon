package vcon

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNew(t *testing.T) {
	v := New("1.0.0")

	if v == nil {
		t.Fatal("New() returned nil")
	}

	if v.UUID == uuid.Nil {
		t.Error("UUID should not be nil")
	}

	if v.Vcon != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", v.Vcon)
	}

	if v.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}

func TestNewWithDefaults(t *testing.T) {
	v := NewWithDefaults()

	if v == nil {
		t.Fatal("NewWithDefaults() returned nil")
	}

	if v.Vcon != "0.0.2" {
		t.Errorf("Expected default version 0.0.2, got %s", v.Vcon)
	}
}

func TestVConAddParty(t *testing.T) {
	v := NewWithDefaults()

	email := "test@example.com"
	name := "Test User"
	party := Party{
		Mailto: &email,
		Name:   &name,
	}

	idx := v.AddParty(party)

	if idx != 0 {
		t.Errorf("Expected first party index to be 0, got %d", idx)
	}

	if len(v.Parties) != 1 {
		t.Errorf("Expected 1 party, got %d", len(v.Parties))
	}

	// Add another party
	phone := "+1234567890"
	party2 := Party{
		Tel: &phone,
	}

	idx2 := v.AddParty(party2)

	if idx2 != 1 {
		t.Errorf("Expected second party index to be 1, got %d", idx2)
	}

	if len(v.Parties) != 2 {
		t.Errorf("Expected 2 parties, got %d", len(v.Parties))
	}
}

func TestVConAddDialog(t *testing.T) {
	v := NewWithDefaults()

	dialog := Dialog{
		Type:    "text",
		Start:   time.Now().UTC(),
		Parties: NewDialogPartiesArrayPtr([]int{0}),
		Body:    "Hello, world!",
	}

	idx := v.AddDialog(dialog)

	if idx != 0 {
		t.Errorf("Expected first dialog index to be 0, got %d", idx)
	}

	if len(v.Dialog) != 1 {
		t.Errorf("Expected 1 dialog, got %d", len(v.Dialog))
	}
}

func TestVConAddAttachment(t *testing.T) {
	v := NewWithDefaults()

	attachment := Attachment{
		Type:     StringPtr("document"),
		Body:     "SGVsbG8gV29ybGQ=", // "Hello World" in base64
		Encoding: StringPtr("base64url"),
		Start:    func() *time.Time { t := time.Now().UTC(); return &t }(),
		Party:    IntPtr(0),
	}

	idx := v.AddAttachment(attachment)

	if idx != 0 {
		t.Errorf("Expected first attachment index to be 0, got %d", idx)
	}

	if len(v.Attachments) != 1 {
		t.Errorf("Expected 1 attachment, got %d", len(v.Attachments))
	}
}

func TestVConGetParty(t *testing.T) {
	v := NewWithDefaults()

	email := "test@example.com"
	party := Party{
		Mailto: &email,
	}

	v.AddParty(party)

	// Test valid index
	p := v.GetParty(0)
	if p == nil {
		t.Error("GetParty(0) should not return nil")
		return
	}
	if p.Mailto == nil || *p.Mailto != email {
		t.Error("GetParty(0) returned wrong party")
	}

	// Test invalid indices
	if v.GetParty(-1) != nil {
		t.Error("GetParty(-1) should return nil")
	}

	if v.GetParty(1) != nil {
		t.Error("GetParty(1) should return nil when only 1 party exists")
	}
}

func TestVConGetDialog(t *testing.T) {
	v := NewWithDefaults()

	dialog := Dialog{
		Type:    "text",
		Start:   time.Now().UTC(),
		Parties: NewDialogPartiesArrayPtr([]int{0}),
		Body:    "Test message",
	}

	v.AddDialog(dialog)

	// Test valid index
	d := v.GetDialog(0)
	if d == nil {
		t.Error("GetDialog(0) should not return nil")
		return
	}
	if d.Type != "text" {
		t.Error("GetDialog(0) returned wrong dialog")
	}

	// Test invalid indices
	if v.GetDialog(-1) != nil {
		t.Error("GetDialog(-1) should return nil")
	}

	if v.GetDialog(1) != nil {
		t.Error("GetDialog(1) should return nil when only 1 dialog exists")
	}
}

func TestVConUpdateTimestamp(t *testing.T) {
	v := NewWithDefaults()

	if v.UpdatedAt != nil {
		t.Error("UpdatedAt should be nil initially")
	}

	v.UpdateTimestamp()

	if v.UpdatedAt == nil {
		t.Error("UpdatedAt should not be nil after UpdateTimestamp()")
	}

	if v.UpdatedAt.Before(v.CreatedAt) {
		t.Error("UpdatedAt should be after CreatedAt")
	}
}

func TestVConMarshalJSON(t *testing.T) {
	v := NewWithDefaults()

	email := "test@example.com"
	v.AddParty(Party{
		Mailto: &email,
	})

	v.AddDialog(Dialog{
		Type:    "text",
		Start:   time.Now().UTC(),
		Parties: NewDialogPartiesArrayPtr([]int{0}),
		Body:    "Hello",
	})

	data, err := v.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	// Check that it's valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Errorf("MarshalJSON produced invalid JSON: %v", err)
	}

	// Check required fields are present
	if _, ok := result["uuid"]; !ok {
		t.Error("JSON missing uuid field")
	}
	if _, ok := result["vcon"]; !ok {
		t.Error("JSON missing vcon field")
	}
	if _, ok := result["created_at"]; !ok {
		t.Error("JSON missing created_at field")
	}
}

func TestVConMarshalJSONWithAdditionalProperties(t *testing.T) {
	v := NewWithDefaults()

	// Add additional properties
	v.AdditionalProperties = map[string]any{
		"custom_field": "custom_value",
		"extra_data":   123,
	}

	data, err := v.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Errorf("MarshalJSON produced invalid JSON: %v", err)
	}

	// Check additional properties are included
	if val, ok := result["custom_field"]; !ok || val != "custom_value" {
		t.Error("Additional property 'custom_field' not included correctly")
	}

	if val, ok := result["extra_data"]; !ok || val != float64(123) {
		t.Error("Additional property 'extra_data' not included correctly")
	}
}

func TestVConUnmarshalJSON(t *testing.T) {
	jsonData := `{
		"uuid": "550e8400-e29b-41d4-a716-446655440000",
		"vcon": "1.0.0",
		"created_at": "2024-01-01T00:00:00Z",
		"parties": [
			{
				"mailto": "test@example.com",
				"name": "Test User"
			}
		],
		"dialog": [
			{
				"type": "text",
				"start": "2024-01-01T00:01:00Z",
				"parties": [0],
				"body": "Hello"
			}
		],
		"custom_field": "custom_value"
	}`

	var v VCon
	if err := v.UnmarshalJSON([]byte(jsonData)); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	// Check standard fields
	if v.UUID == uuid.Nil {
		t.Error("UUID should not be nil")
	}

	if v.Vcon != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", v.Vcon)
	}

	if len(v.Parties) != 1 {
		t.Errorf("Expected 1 party, got %d", len(v.Parties))
	}

	if len(v.Dialog) != 1 {
		t.Errorf("Expected 1 dialog, got %d", len(v.Dialog))
	}

	// Check additional properties
	if v.AdditionalProperties == nil {
		t.Error("AdditionalProperties should not be nil")
	}

	if val, ok := v.AdditionalProperties["custom_field"]; !ok || val != "custom_value" {
		t.Error("Additional property 'custom_field' not parsed correctly")
	}
}

func TestVConString(t *testing.T) {
	v := NewWithDefaults()

	str := v.String()
	if str == "" {
		t.Error("String() should not return empty string")
	}

	// Should be valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(str), &result); err != nil {
		t.Errorf("String() did not produce valid JSON: %v", err)
	}
}

func TestVConWriteTo(t *testing.T) {
	v := NewWithDefaults()

	var buf bytes.Buffer
	n, err := v.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo failed: %v", err)
	}

	if n == 0 {
		t.Error("WriteTo wrote 0 bytes")
	}

	// Check that it's valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Errorf("WriteTo produced invalid JSON: %v", err)
	}
}

func TestVConWriteToIndent(t *testing.T) {
	v := NewWithDefaults()

	var buf bytes.Buffer
	n, err := v.WriteToIndent(&buf)
	if err != nil {
		t.Fatalf("WriteToIndent failed: %v", err)
	}

	if n == 0 {
		t.Error("WriteToIndent wrote 0 bytes")
	}

	// Check that output contains indentation
	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("\n  ")) {
		t.Error("WriteToIndent output should contain indentation")
	}
}

// Test round-trip marshaling and unmarshaling.
func TestVConRoundTrip(t *testing.T) {
	original := NewWithDefaults()

	// Add some data
	email := "test@example.com"
	name := "Test User"
	original.AddParty(Party{
		Mailto: &email,
		Name:   &name,
	})

	original.AddDialog(Dialog{
		Type:    "text",
		Start:   time.Now().UTC().Round(time.Second), // Round to avoid nanosecond precision loss
		Parties: NewDialogPartiesArrayPtr([]int{0}),
		Body:    "Test message",
	})

	// Add additional properties
	original.AdditionalProperties = map[string]any{
		"custom": "value",
	}

	// Marshal to JSON
	data, err := original.MarshalJSON()
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal back
	var restored VCon
	if err := restored.UnmarshalJSON(data); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Compare key fields
	if original.UUID.String() != restored.UUID.String() {
		t.Error("UUID mismatch after round-trip")
	}

	if original.Vcon != restored.Vcon {
		t.Error("Version mismatch after round-trip")
	}

	if len(original.Parties) != len(restored.Parties) {
		t.Error("Parties count mismatch after round-trip")
	}

	if len(original.Dialog) != len(restored.Dialog) {
		t.Error("Dialog count mismatch after round-trip")
	}

	// Check additional properties
	if val, ok := restored.AdditionalProperties["custom"]; !ok || val != "value" {
		t.Error("Additional properties not preserved in round-trip")
	}
}

// Benchmark tests.
func BenchmarkVConMarshalJSON(b *testing.B) {
	v := NewWithDefaults()

	// Add some typical data
	for i := 0; i < 5; i++ {
		email := "user@example.com"
		v.AddParty(Party{Mailto: &email})
	}

	for i := 0; i < 10; i++ {
		v.AddDialog(Dialog{
			Type:    "text",
			Start:   time.Now(),
			Parties: NewDialogPartiesArrayPtr([]int{0, 1}),
			Body:    "Message content",
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = v.MarshalJSON()
	}
}

func BenchmarkVConUnmarshalJSON(b *testing.B) {
	v := NewWithDefaults()

	// Add some typical data
	for i := 0; i < 5; i++ {
		email := "user@example.com"
		v.AddParty(Party{Mailto: &email})
	}

	for i := 0; i < 10; i++ {
		v.AddDialog(Dialog{
			Type:    "text",
			Start:   time.Now(),
			Parties: NewDialogPartiesArrayPtr([]int{0, 1}),
			Body:    "Message content",
		})
	}

	data, _ := v.MarshalJSON()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var v2 VCon
		_ = v2.UnmarshalJSON(data)
	}
}

// Tests from vcon_helper_functions_test.go

func TestFindPartyIndexEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		setup func() *VCon
		field string
		value string
		want  *int
	}{
		{
			name: "empty parties array",
			setup: func() *VCon {
				vcon := NewWithDefaults()
				vcon.Parties = []Party{}
				return vcon
			},
			field: "name",
			value: "Alice",
			want:  nil,
		},
		{
			name: "nil parties array",
			setup: func() *VCon {
				vcon := NewWithDefaults()
				vcon.Parties = nil
				return vcon
			},
			field: "name",
			value: "Alice",
			want:  nil,
		},
		{
			name: "party with nil fields",
			setup: func() *VCon {
				vcon := NewWithDefaults()
				vcon.Parties = []Party{
					{}, // All fields nil
				}
				return vcon
			},
			field: "name",
			value: "Alice",
			want:  nil,
		},
		{
			name: "case sensitive search",
			setup: func() *VCon {
				vcon := NewWithDefaults()
				vcon.Parties = []Party{
					{Name: StringPtr("Alice")},
				}
				return vcon
			},
			field: "name",
			value: "alice", // Different case
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vcon := tt.setup()
			result := vcon.FindPartyIndex(tt.field, tt.value)

			if tt.want == nil {
				if result != nil {
					t.Errorf("FindPartyIndex() = %v, want nil", *result)
				}
			} else {
				if result == nil {
					t.Errorf("FindPartyIndex() = nil, want %d", *tt.want)
				} else if *result != *tt.want {
					t.Errorf("FindPartyIndex() = %d, want %d", *result, *tt.want)
				}
			}
		})
	}
}

func TestFindAnalysisByType(t *testing.T) {
	// Create test vCon with various analyses
	vcon := NewWithDefaults()
	vcon.Analysis = []Analysis{
		{
			Type:   "summary",
			Vendor: "vendor1",
			Body:   StringPtr("Summary analysis"),
		},
		{
			Type:   "transcript",
			Vendor: "vendor2",
			Body:   StringPtr("Transcript analysis"),
		},
		{
			Type:   "sentiment",
			Vendor: "vendor3",
			Body:   StringPtr("Sentiment analysis"),
		},
		{
			Type:   "custom",
			Vendor: "vendor4",
			Body:   StringPtr("Custom analysis"),
		},
	}

	tests := []struct {
		name         string
		analysisType string
		expectedIdx  *int // Index of expected analysis, nil if not found
	}{
		{"find summary", "summary", IntPtr(0)},
		{"find transcript", "transcript", IntPtr(1)},
		{"find sentiment", "sentiment", IntPtr(2)},
		{"not found", "unknown", nil},
		{"empty type", "", nil},
		{"case sensitive", "Summary", nil}, // Different case
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vcon.FindAnalysisByType(tt.analysisType)

			if tt.expectedIdx == nil {
				if result != nil {
					t.Errorf("FindAnalysisByType(%q) = %v, expected nil", tt.analysisType, result)
				}
			} else {
				if result == nil {
					t.Errorf("FindAnalysisByType(%q) = nil, expected analysis at index %d",
						tt.analysisType, *tt.expectedIdx)
				} else {
					// Verify we got the right analysis by checking the vendor
					expectedVendor := vcon.Analysis[*tt.expectedIdx].Vendor
					if result.Vendor != expectedVendor {
						t.Errorf("FindAnalysisByType(%q) returned analysis with vendor %q, expected %q",
							tt.analysisType, result.Vendor, expectedVendor)
					}
				}
			}
		})
	}
}

func TestFindAnalysisByTypeEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		setup func() *VCon
		aType string
		want  bool // true if should find something
	}{
		{
			name: "empty analysis array",
			setup: func() *VCon {
				vcon := NewWithDefaults()
				vcon.Analysis = []Analysis{}
				return vcon
			},
			aType: "summary",
			want:  false,
		},
		{
			name: "nil analysis array",
			setup: func() *VCon {
				vcon := NewWithDefaults()
				vcon.Analysis = nil
				return vcon
			},
			aType: "summary",
			want:  false,
		},
		{
			name: "analysis with nil type",
			setup: func() *VCon {
				vcon := NewWithDefaults()
				vcon.Analysis = []Analysis{
					{Type: "", Vendor: "vendor1"}, // Empty type should fail validation
				}
				return vcon
			},
			aType: "summary",
			want:  false,
		},
		{
			name: "first match returned",
			setup: func() *VCon {
				vcon := NewWithDefaults()
				vcon.Analysis = []Analysis{
					{Type: "summary", Vendor: "vendor1"},
					{Type: "summary", Vendor: "vendor2"}, // Duplicate type
				}
				return vcon
			},
			aType: "summary",
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vcon := tt.setup()
			result := vcon.FindAnalysisByType(tt.aType)

			if tt.want {
				if result == nil {
					t.Errorf("FindAnalysisByType(%q) = nil, expected to find analysis", tt.aType)
				}
			} else {
				if result != nil {
					t.Errorf("FindAnalysisByType(%q) = %v, expected nil", tt.aType, result)
				}
			}
		})
	}
}

func TestFindAnalysisByTypeFirstMatch(t *testing.T) {
	// Test that the function returns the first match when there are duplicates
	vcon := NewWithDefaults()
	vcon.Analysis = []Analysis{
		{Type: "summary", Vendor: "first-vendor"},
		{Type: "summary", Vendor: "second-vendor"},
		{Type: "summary", Vendor: "third-vendor"},
	}

	result := vcon.FindAnalysisByType("summary")
	if result == nil {
		t.Fatal("FindAnalysisByType() returned nil, expected first analysis")
	}

	if result.Vendor != "first-vendor" {
		t.Errorf("FindAnalysisByType() returned analysis with vendor %q, expected 'first-vendor'",
			result.Vendor)
	}
}

// IntPtr returns a pointer to an int value.
func IntPtr(i int) *int {
	return &i
}
