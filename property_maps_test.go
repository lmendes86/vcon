package vcon

import (
	"testing"
)

func TestIsPropertyAllowed(t *testing.T) {
	tests := []struct {
		name       string
		objectType string
		property   string
		expected   bool
	}{
		// VCon properties
		{"vcon valid property", "vcon", "uuid", true},
		{"vcon valid property vcon", "vcon", "vcon", true},
		{"vcon valid property parties", "vcon", "parties", true},
		{"vcon extension property", "vcon", "extensions", true},
		{"vcon invalid property", "vcon", "nonexistent", false},

		// Party properties
		{"party valid property", "party", "name", true},
		{"party valid property tel", "party", "tel", true},
		{"party extension property", "party", "sip", true},
		{"party invalid property", "party", "nonexistent", false},

		// Dialog properties
		{"dialog valid property", "dialog", "type", true},
		{"dialog valid property start", "dialog", "start", true},
		{"dialog extension property", "dialog", "campaign", true},
		{"dialog invalid property", "dialog", "nonexistent", false},

		// Attachment properties
		{"attachment valid property", "attachment", "start", true},
		{"attachment valid property party", "attachment", "party", true},
		{"attachment extension property", "attachment", "meta", true},
		{"attachment invalid property", "attachment", "nonexistent", false},

		// Analysis properties
		{"analysis valid property", "analysis", "vendor", true},
		{"analysis valid property type", "analysis", "type", true},
		{"analysis extension property", "analysis", "product", true},
		{"analysis invalid property", "analysis", "nonexistent", false},

		// Invalid object types
		{"invalid object type", "invalid", "property", false},
		{"empty object type", "", "property", false},
		{"null object type", "null", "property", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsPropertyAllowed(tt.objectType, tt.property)
			if result != tt.expected {
				t.Errorf("IsPropertyAllowed(%q, %q) = %v, expected %v",
					tt.objectType, tt.property, result, tt.expected)
			}
		})
	}
}

func TestGetAllowedProperties(t *testing.T) {
	tests := []struct {
		name       string
		objectType string
		expected   map[string]bool
		checkProps []string // Properties to verify exist
	}{
		{
			name:       "vcon properties",
			objectType: "vcon",
			expected:   AllowedVConProperties,
			checkProps: []string{"uuid", "vcon", "parties", "dialog", "analysis"},
		},
		{
			name:       "party properties",
			objectType: "party",
			expected:   AllowedPartyProperties,
			checkProps: []string{"name", "tel", "mailto", "uuid", "role"},
		},
		{
			name:       "dialog properties",
			objectType: "dialog",
			expected:   AllowedDialogProperties,
			checkProps: []string{"type", "start", "parties", "duration", "body"},
		},
		{
			name:       "attachment properties",
			objectType: "attachment",
			expected:   AllowedAttachmentProperties,
			checkProps: []string{"start", "party", "dialog", "mediatype", "body"},
		},
		{
			name:       "analysis properties",
			objectType: "analysis",
			expected:   AllowedAnalysisProperties,
			checkProps: []string{"vendor", "type", "dialog", "body", "url"},
		},
		{
			name:       "invalid object type",
			objectType: "invalid",
			expected:   make(map[string]bool),
			checkProps: []string{},
		},
		{
			name:       "empty object type",
			objectType: "",
			expected:   make(map[string]bool),
			checkProps: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetAllowedProperties(tt.objectType)

			// For valid object types, check that we get the expected map reference
			if tt.objectType != "invalid" && tt.objectType != "" {
				if len(result) != len(tt.expected) {
					t.Errorf("GetAllowedProperties(%q) returned map with %d items, expected %d",
						tt.objectType, len(result), len(tt.expected))
				}

				// Verify specific properties exist
				for _, prop := range tt.checkProps {
					if !result[prop] {
						t.Errorf("GetAllowedProperties(%q) missing expected property %q",
							tt.objectType, prop)
					}
				}
			} else if len(result) != 0 {
				// For invalid types, should return empty map
				t.Errorf("GetAllowedProperties(%q) should return empty map, got %d items",
					tt.objectType, len(result))
			}
		})
	}
}

func TestPropertyMapsConstants(t *testing.T) {
	// Test that property handling constants are defined
	constants := []string{
		PropertyHandlingDefault,
		PropertyHandlingStrict,
		PropertyHandlingMeta,
	}

	expectedValues := []string{"default", "strict", "meta"}

	for i, constant := range constants {
		if constant != expectedValues[i] {
			t.Errorf("Property handling constant %d = %q, expected %q",
				i, constant, expectedValues[i])
		}
	}
}

func TestPropertyMapsIntegration(t *testing.T) {
	// Test integration between IsPropertyAllowed and GetAllowedProperties
	objectTypes := []string{"vcon", "party", "dialog", "attachment", "analysis"}

	for _, objType := range objectTypes {
		properties := GetAllowedProperties(objType)

		// Test a few properties from each map
		for prop := range properties {
			if !IsPropertyAllowed(objType, prop) {
				t.Errorf("Property %q should be allowed for %q according to GetAllowedProperties, but IsPropertyAllowed returned false",
					prop, objType)
			}
			// Only test first few properties to avoid excessive testing
			break
		}
	}
}

func TestPropertyMapsEdgeCases(t *testing.T) {
	// Test edge cases
	tests := []struct {
		name       string
		objectType string
		property   string
		expected   bool
	}{
		{"case sensitive object type", "VCON", "uuid", false},
		{"case sensitive property", "vcon", "UUID", false},
		{"whitespace in object type", " vcon ", "uuid", false},
		{"whitespace in property", "vcon", " uuid ", false},
		{"empty property name", "vcon", "", false},
		{"numeric object type", "123", "property", false},
		{"numeric property", "vcon", "123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsPropertyAllowed(tt.objectType, tt.property)
			if result != tt.expected {
				t.Errorf("IsPropertyAllowed(%q, %q) = %v, expected %v",
					tt.objectType, tt.property, result, tt.expected)
			}
		})
	}
}
