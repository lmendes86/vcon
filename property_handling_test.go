package vcon

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func TestNewPropertyHandler(t *testing.T) {
	tests := []struct {
		name     string
		mode     string
		expected string
	}{
		{"default mode", PropertyHandlingDefault, PropertyHandlingDefault},
		{"strict mode", PropertyHandlingStrict, PropertyHandlingStrict},
		{"meta mode", PropertyHandlingMeta, PropertyHandlingMeta},
		{"invalid mode", "invalid", PropertyHandlingDefault},
		{"empty mode", "", PropertyHandlingDefault},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewPropertyHandler(tt.mode)
			if handler.Mode != tt.expected {
				t.Errorf("Expected mode %s, got %s", tt.expected, handler.Mode)
			}
		})
	}
}

func TestProcessPropertiesDefault(t *testing.T) {
	handler := NewPropertyHandler(PropertyHandlingDefault)

	input := map[string]interface{}{
		"uuid":         "valid-uuid",
		"vcon":         "0.0.1",
		"custom_field": "custom_value",
		"extra_data":   map[string]interface{}{"nested": "value"},
	}

	result := handler.ProcessProperties(input, AllowedVConProperties)

	// Should keep all properties in default mode
	if len(result) != 4 {
		t.Errorf("Expected 4 properties, got %d", len(result))
	}

	if result["uuid"] != "valid-uuid" {
		t.Error("Standard property 'uuid' should be preserved")
	}

	if result["custom_field"] != "custom_value" {
		t.Error("Non-standard property 'custom_field' should be preserved in default mode")
	}

	if result["extra_data"] == nil {
		t.Error("Non-standard property 'extra_data' should be preserved in default mode")
	}
}

func TestProcessPropertiesStrict(t *testing.T) {
	handler := NewPropertyHandler(PropertyHandlingStrict)

	input := map[string]interface{}{
		"uuid":         "valid-uuid",
		"vcon":         "0.0.1",
		"custom_field": "custom_value",
		"extra_data":   map[string]interface{}{"nested": "value"},
	}

	result := handler.ProcessProperties(input, AllowedVConProperties)

	// Should only keep standard properties in strict mode
	if len(result) != 2 {
		t.Errorf("Expected 2 properties, got %d", len(result))
	}

	if result["uuid"] != "valid-uuid" {
		t.Error("Standard property 'uuid' should be preserved")
	}

	if result["vcon"] != "0.0.1" {
		t.Error("Standard property 'vcon' should be preserved")
	}

	if result["custom_field"] != nil {
		t.Error("Non-standard property 'custom_field' should be removed in strict mode")
	}

	if result["extra_data"] != nil {
		t.Error("Non-standard property 'extra_data' should be removed in strict mode")
	}
}

func TestProcessPropertiesMeta(t *testing.T) {
	handler := NewPropertyHandler(PropertyHandlingMeta)

	input := map[string]interface{}{
		"uuid":         "valid-uuid",
		"vcon":         "0.0.1",
		"custom_field": "custom_value",
		"extra_data":   map[string]interface{}{"nested": "value"},
	}

	result := handler.ProcessProperties(input, AllowedVConProperties)

	// Should move non-standard properties to meta
	if result["uuid"] != "valid-uuid" {
		t.Error("Standard property 'uuid' should be preserved")
	}

	if result["vcon"] != "0.0.1" {
		t.Error("Standard property 'vcon' should be preserved")
	}

	meta, ok := result["meta"].(map[string]interface{})
	if !ok {
		t.Fatal("Meta field should be created and be a map")
	}

	if meta["custom_field"] != "custom_value" {
		t.Error("Non-standard property 'custom_field' should be moved to meta")
	}

	if meta["extra_data"] == nil {
		t.Error("Non-standard property 'extra_data' should be moved to meta")
	}
}

func TestProcessPropertiesMetaExistingMeta(t *testing.T) {
	handler := NewPropertyHandler(PropertyHandlingMeta)

	input := map[string]interface{}{
		"uuid":         "valid-uuid",
		"meta":         map[string]interface{}{"existing": "data"},
		"custom_field": "custom_value",
	}

	result := handler.ProcessProperties(input, AllowedVConProperties)

	meta, ok := result["meta"].(map[string]interface{})
	if !ok {
		t.Fatal("Meta field should be a map")
	}

	if meta["existing"] != "data" {
		t.Error("Existing meta data should be preserved")
	}

	if meta["custom_field"] != "custom_value" {
		t.Error("Non-standard property should be added to existing meta")
	}
}

func TestProcessPropertiesNilInput(t *testing.T) {
	handler := NewPropertyHandler(PropertyHandlingDefault)

	result := handler.ProcessProperties(nil, AllowedVConProperties)

	if result != nil {
		t.Error("Processing nil input should return nil")
	}
}

func TestProcessVCon(t *testing.T) {
	handler := NewPropertyHandler(PropertyHandlingStrict)

	input := map[string]interface{}{
		"uuid":          "test-uuid",
		"vcon":          "0.0.1",
		"custom_global": "should_be_removed",
		"parties": []interface{}{
			map[string]interface{}{
				"name":         "Test User",
				"custom_party": "should_be_removed",
				"tel":          "+1234567890",
			},
		},
		"dialog": []interface{}{
			map[string]interface{}{
				"type":          "text",
				"custom_dialog": "should_be_removed",
				"start":         "2023-01-01T00:00:00Z",
				"parties":       []interface{}{0},
			},
		},
		"attachments": []interface{}{
			map[string]interface{}{
				"type":              "document",
				"custom_attachment": "should_be_removed",
				"body":              "test content",
				"encoding":          "none",
			},
		},
		"analysis": []interface{}{
			map[string]interface{}{
				"type":            "sentiment",
				"custom_analysis": "should_be_removed",
				"vendor":          "test",
				"body":            "positive",
				"encoding":        "none",
				"dialog":          0,
			},
		},
	}

	result := handler.ProcessVCon(input)

	// Check global level
	if result["custom_global"] != nil {
		t.Error("Custom global property should be removed in strict mode")
	}

	// Check parties
	parties, ok := result["parties"].([]interface{})
	if !ok {
		t.Fatal("Parties should be preserved as array")
	}

	party, ok := parties[0].(map[string]interface{})
	if !ok {
		t.Fatal("Expected party to be map[string]interface{}")
	}
	if party["custom_party"] != nil {
		t.Error("Custom party property should be removed in strict mode")
	}
	if party["name"] != "Test User" {
		t.Error("Standard party property should be preserved")
	}

	// Check dialogs
	dialogs, ok := result["dialog"].([]interface{})
	if !ok {
		t.Fatal("Dialogs should be preserved as array")
	}

	dialog, ok := dialogs[0].(map[string]interface{})
	if !ok {
		t.Fatal("Expected dialog to be map[string]interface{}")
	}
	if dialog["custom_dialog"] != nil {
		t.Error("Custom dialog property should be removed in strict mode")
	}
	if dialog["type"] != "text" {
		t.Error("Standard dialog property should be preserved")
	}

	// Check attachments
	attachments, ok := result["attachments"].([]interface{})
	if !ok {
		t.Fatal("Attachments should be preserved as array")
	}

	attachment, ok := attachments[0].(map[string]interface{})
	if !ok {
		t.Fatal("Expected attachment to be map[string]interface{}")
	}
	if attachment["custom_attachment"] != nil {
		t.Error("Custom attachment property should be removed in strict mode")
	}
	if attachment["type"] != "document" {
		t.Error("Standard attachment property should be preserved")
	}

	// Check analysis
	analysis, ok := result["analysis"].([]interface{})
	if !ok {
		t.Fatal("Analysis should be preserved as array")
	}

	analysisItem, ok := analysis[0].(map[string]interface{})
	if !ok {
		t.Fatal("Expected analysis item to be map[string]interface{}")
	}
	if analysisItem["custom_analysis"] != nil {
		t.Error("Custom analysis property should be removed in strict mode")
	}
	if analysisItem["type"] != "sentiment" {
		t.Error("Standard analysis property should be preserved")
	}
}

func TestProcessVConNonMapItems(t *testing.T) {
	handler := NewPropertyHandler(PropertyHandlingDefault)

	input := map[string]interface{}{
		"uuid": "test-uuid",
		"parties": []interface{}{
			"not-a-map", // Should be preserved as-is
			map[string]interface{}{
				"name": "Test User",
			},
		},
		"dialog": []interface{}{
			123, // Should be preserved as-is
		},
	}

	result := handler.ProcessVCon(input)

	parties, ok := result["parties"].([]interface{})
	if !ok {
		t.Fatal("Expected parties to be []interface{}")
	}
	if parties[0] != "not-a-map" {
		t.Error("Non-map party items should be preserved as-is")
	}

	dialogs, ok := result["dialog"].([]interface{})
	if !ok {
		t.Fatal("Expected dialogs to be []interface{}")
	}
	if dialogs[0] != 123 {
		t.Error("Non-map dialog items should be preserved as-is")
	}
}

func TestMarshalVConWithPropertyHandling(t *testing.T) {
	// Create a VCon with additional properties
	vcon := NewWithDefaults()
	vcon.AddParty(Party{Name: StringPtr("Test User")})

	// Test marshaling with different property handling modes
	tests := []struct {
		name string
		mode string
	}{
		{"default mode", PropertyHandlingDefault},
		{"strict mode", PropertyHandlingStrict},
		{"meta mode", PropertyHandlingMeta},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewPropertyHandler(tt.mode)

			data, err := handler.MarshalVConWithPropertyHandling(vcon)
			if err != nil {
				t.Fatalf("Failed to marshal vCon: %v", err)
			}

			// Verify it's valid JSON
			var result map[string]interface{}
			err = json.Unmarshal(data, &result)
			if err != nil {
				t.Fatalf("Failed to unmarshal result: %v", err)
			}

			if result["uuid"] == nil {
				t.Error("UUID should be preserved")
			}

			if result["vcon"] == nil {
				t.Error("VCon version should be preserved")
			}
		})
	}
}

func TestMarshalVConWithPropertyHandlingError(t *testing.T) {
	// Create a VCon that will cause marshal error
	handler := NewPropertyHandler(PropertyHandlingDefault)

	// This will cause an error because we can't marshal channels
	// Note: Top-level Meta field removed for IETF draft-03 compliance
	// Test with invalid data in party meta instead
	invalidVCon := &VCon{
		Parties: []Party{
			{
				Name: StringPtr("Test"),
				Meta: map[string]interface{}{
					"invalid": make(chan int),
				},
			},
		},
	}

	_, err := handler.MarshalVConWithPropertyHandling(invalidVCon)
	if err == nil {
		t.Error("Expected error when marshaling invalid vCon")
	}
}

func TestUnmarshalVConWithPropertyHandling(t *testing.T) {
	// Create test JSON with custom properties
	testJSON := `{
		"uuid": "123e4567-e89b-12d3-a456-426614174000",
		"vcon": "0.0.1",
		"created_at": "2023-01-01T00:00:00Z",
		"custom_field": "custom_value",
		"parties": [{
			"name": "Test User",
			"custom_party_field": "party_custom"
		}]
	}`

	tests := []struct {
		name string
		mode string
	}{
		{"default mode", PropertyHandlingDefault},
		{"strict mode", PropertyHandlingStrict},
		{"meta mode", PropertyHandlingMeta},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewPropertyHandler(tt.mode)
			var vcon VCon

			err := handler.UnmarshalVConWithPropertyHandling([]byte(testJSON), &vcon)
			if err != nil {
				t.Fatalf("Failed to unmarshal vCon: %v", err)
			}

			if vcon.UUID == uuid.Nil {
				t.Error("UUID should be preserved")
			}

			if vcon.Vcon != "0.0.1" {
				t.Error("VCon version should be preserved")
			}

			if len(vcon.Parties) != 1 {
				t.Error("Parties should be preserved")
			}

			if vcon.Parties[0].Name == nil || *vcon.Parties[0].Name != "Test User" {
				t.Error("Party name should be preserved")
			}
		})
	}
}

func TestUnmarshalVConWithPropertyHandlingError(t *testing.T) {
	handler := NewPropertyHandler(PropertyHandlingDefault)
	var vcon VCon

	// Test invalid JSON
	err := handler.UnmarshalVConWithPropertyHandling([]byte("{invalid json}"), &vcon)
	if err == nil {
		t.Error("Expected error when unmarshaling invalid JSON")
	}
}

func TestPropertyHandlingRoundTrip(t *testing.T) {
	// Test that marshal -> unmarshal preserves data correctly
	original := NewWithDefaults()
	original.AddParty(Party{Name: StringPtr("Test User")})

	handler := NewPropertyHandler(PropertyHandlingDefault)

	// Marshal
	data, err := handler.MarshalVConWithPropertyHandling(original)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal
	var restored VCon
	err = handler.UnmarshalVConWithPropertyHandling(data, &restored)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Compare key fields
	if restored.UUID.String() != original.UUID.String() {
		t.Error("UUID should be preserved in round trip")
	}

	if restored.Vcon != original.Vcon {
		t.Error("VCon version should be preserved in round trip")
	}

	if len(restored.Parties) != len(original.Parties) {
		t.Error("Parties count should be preserved in round trip")
	}

	if *restored.Parties[0].Name != *original.Parties[0].Name {
		t.Error("Party name should be preserved in round trip")
	}
}
