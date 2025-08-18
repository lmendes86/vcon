package vcon

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestIETFValidation tests the IETF-compliant validation.
func TestIETFValidation(t *testing.T) {
	tests := []struct {
		name         string
		vcon         func() *VCon
		expectError  bool
		errorMessage string
	}{
		{
			name: "valid IETF compliant vCon",
			vcon: func() *VCon {
				id := uuid.New()
				return &VCon{
					UUID:      id,
					Vcon:      "0.0.2",
					CreatedAt: time.Now().UTC(),
					Parties: []Party{
						{Name: StringPtr("Alice")},
						{Tel: StringPtr("+1234567890")},
					},
					Dialog: []Dialog{
						{
							Type:      "text",
							Start:     time.Now().UTC(),
							Parties:   NewDialogPartiesArrayPtr([]int{0, 1}),
							Body:      "Hello, how are you?",
							Encoding:  StringPtr("none"),
							Mediatype: StringPtr("text/plain"),
						},
					},
					Analysis: []Analysis{
						{
							Type:     "sentiment",
							Vendor:   "test-vendor",
							Body:     "positive",
							Encoding: StringPtr("json"),
							Dialog:   0,
						},
					},
				}
			},
			expectError: false,
		},
		{
			name: "missing UUID - now required per IETF spec",
			vcon: func() *VCon {
				return &VCon{
					Vcon:      "0.0.2",
					CreatedAt: time.Now().UTC(),
					Parties:   []Party{{Name: StringPtr("Alice")}},
				}
			},
			expectError:  true, // UUID is REQUIRED per IETF draft-03 spec
			errorMessage: "uuid is required",
		},
		{
			name: "invalid version",
			vcon: func() *VCon {
				id := uuid.New()
				return &VCon{
					UUID:      id,
					Vcon:      "0.0.1",
					CreatedAt: time.Now().UTC(),
					Parties:   []Party{{Name: StringPtr("Alice")}},
				}
			},
			expectError:  true,
			errorMessage: "vcon must be '0.0.2'",
		},
		{
			name: "party without identifier - now allowed per IETF spec",
			vcon: func() *VCon {
				id := uuid.New()
				return &VCon{
					UUID:      id,
					Vcon:      "0.0.2",
					CreatedAt: time.Now().UTC(),
					Parties: []Party{
						{}, // No identifiers - this is valid per IETF spec
					},
				}
			},
			expectError: false, // No longer an error per IETF spec
		},
		{
			name: "invalid dialog type",
			vcon: func() *VCon {
				id := uuid.New()
				return &VCon{
					UUID:      id,
					Vcon:      "0.0.2",
					CreatedAt: time.Now().UTC(),
					Parties:   []Party{{Name: StringPtr("Alice")}},
					Dialog: []Dialog{
						{
							Type:    "invalid",
							Start:   time.Now().UTC(),
							Parties: NewDialogPartiesArrayPtr([]int{0}),
						},
					},
				}
			},
			expectError:  true,
			errorMessage: "invalid dialog type: invalid",
		},
		{
			name: "incomplete dialog missing disposition",
			vcon: func() *VCon {
				id := uuid.New()
				return &VCon{
					UUID:      id,
					Vcon:      "0.0.2",
					CreatedAt: time.Now().UTC(),
					Parties:   []Party{{Name: StringPtr("Alice")}},
					Dialog: []Dialog{
						{
							Type:    "incomplete",
							Start:   time.Now().UTC(),
							Parties: NewDialogPartiesArrayPtr([]int{0}),
							// Missing disposition
						},
					},
				}
			},
			expectError:  true,
			errorMessage: "incomplete dialog must have disposition",
		},
		{
			name: "dialog with empty parties",
			vcon: func() *VCon {
				id := uuid.New()
				return &VCon{
					UUID:      id,
					Vcon:      "0.0.2",
					CreatedAt: time.Now().UTC(),
					Parties:   []Party{{Name: StringPtr("Alice")}},
					Dialog: []Dialog{
						{
							Type:    "text",
							Start:   time.Now().UTC(),
							Parties: NewDialogPartiesArrayPtr([]int{}), // Empty parties
						},
					},
				}
			},
			expectError:  true,
			errorMessage: "text dialog must have content (body/url)",
		},
		{
			name: "dialog with invalid party index",
			vcon: func() *VCon {
				id := uuid.New()
				return &VCon{
					UUID:      id,
					Vcon:      "0.0.2",
					CreatedAt: time.Now().UTC(),
					Parties:   []Party{{Name: StringPtr("Alice")}},
					Dialog: []Dialog{
						{
							Type:    "text",
							Start:   time.Now().UTC(),
							Parties: NewDialogPartiesArrayPtr([]int{5}), // Invalid index
						},
					},
				}
			},
			expectError:  true,
			errorMessage: "invalid party index 5",
		},
		{
			name: "analysis missing required fields",
			vcon: func() *VCon {
				id := uuid.New()
				return &VCon{
					UUID:      id,
					Vcon:      "0.0.2",
					CreatedAt: time.Now().UTC(),
					Parties:   []Party{{Name: StringPtr("Alice")}},
					Analysis: []Analysis{
						{
							// Missing type and vendor
							Body:     "content",
							Encoding: StringPtr("json"),
						},
					},
				}
			},
			expectError:  true,
			errorMessage: "vendor is required",
		},
		{
			name: "analysis with invalid dialog reference",
			vcon: func() *VCon {
				id := uuid.New()
				return &VCon{
					UUID:      id,
					Vcon:      "0.0.2",
					CreatedAt: time.Now().UTC(),
					Parties:   []Party{{Name: StringPtr("Alice")}},
					Analysis: []Analysis{
						{
							Type:     "sentiment",
							Vendor:   "test",
							Body:     "positive",
							Encoding: StringPtr("json"),
							Dialog:   10, // Invalid dialog index
						},
					},
				}
			},
			expectError:  true,
			errorMessage: "invalid dialog index: 10",
		},
		{
			name: "attachment with invalid party reference",
			vcon: func() *VCon {
				id := uuid.New()
				return &VCon{
					UUID:      id,
					Vcon:      "0.0.2",
					CreatedAt: time.Now().UTC(),
					Parties:   []Party{{Name: StringPtr("Alice")}},
					Attachments: []Attachment{
						{
							Type:     StringPtr("document"),
							Body:     "content",
							Encoding: StringPtr("base64url"),
							Start:    func() *time.Time { t := time.Now().UTC(); return &t }(), // Optional field per IETF spec
							Party:    IntPtr(10),                                               // Invalid party index
						},
					},
				}
			},
			expectError:  true,
			errorMessage: "invalid party index: 10",
		},
		{
			name: "non-HTTPS URL",
			vcon: func() *VCon {
				id := uuid.New()
				return &VCon{
					UUID:      id,
					Vcon:      "0.0.2",
					CreatedAt: time.Now().UTC(),
					Parties:   []Party{{Name: StringPtr("Alice")}},
					Dialog: []Dialog{
						{
							Type:    "recording",
							Start:   time.Now().UTC(),
							Parties: NewDialogPartiesArrayPtr([]int{0}),
							URL:     StringPtr("http://example.com/audio.wav"), // HTTP not HTTPS
						},
					},
				}
			},
			expectError:  true,
			errorMessage: "external URLs must use HTTPS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vcon := tt.vcon()
			err := vcon.ValidateIETF()

			if tt.expectError {
				if err == nil {
					t.Error("Expected IETF validation error but got none")
				} else if tt.errorMessage != "" {
					if !contains(err.Error(), tt.errorMessage) {
						t.Errorf("Expected error to contain '%s', got: %s", tt.errorMessage, err.Error())
					}
				}
			} else {
				if err != nil {
					t.Errorf("Expected no IETF validation error but got: %s", err.Error())
				}
			}
		})
	}
}

// TestIETFComplianceLevel tests the compliance level checking.
func TestIETFComplianceLevel(t *testing.T) {
	tests := []struct {
		name          string
		vcon          func() *VCon
		expectedLevel IETFComplianceLevel
		expectError   bool
	}{
		{
			name: "fully compliant",
			vcon: func() *VCon {
				id := uuid.New()
				return &VCon{
					UUID:      id,
					Vcon:      "0.0.2",
					CreatedAt: time.Now().UTC(),
					Parties:   []Party{{Name: StringPtr("Alice")}},
				}
			},
			expectedLevel: IETFFullyCompliant,
			expectError:   false,
		},
		{
			name: "partially compliant - minor issues",
			vcon: func() *VCon {
				id := uuid.New()
				return &VCon{
					UUID:      id,
					Vcon:      "0.0.2",
					CreatedAt: time.Now().UTC(),
					Parties:   []Party{{Name: StringPtr("Alice")}},
					Dialog: []Dialog{
						{
							Type:    "incomplete",
							Start:   time.Now().UTC(),
							Parties: NewDialogPartiesArrayPtr([]int{0}),
							// Missing disposition (minor issue)
						},
					},
				}
			},
			expectedLevel: IETFPartiallyCompliant,
			expectError:   true,
		},
		{
			name: "non-compliant - critical issues",
			vcon: func() *VCon {
				return &VCon{
					// Missing UUID (critical)
					Vcon:      "0.0.1", // Wrong version (critical)
					CreatedAt: time.Now().UTC(),
					Parties:   []Party{{Name: StringPtr("Alice")}},
				}
			},
			expectedLevel: IETFNonCompliant,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vcon := tt.vcon()
			level, err := vcon.CheckIETFCompliance()

			if level != tt.expectedLevel {
				t.Errorf("Expected compliance level %d, got %d", tt.expectedLevel, level)
			}

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %s", err.Error())
			}
		})
	}
}

// TestMigrationHelpers tests the migration helper functions.
func TestMigrationHelpers(t *testing.T) {
	t.Run("AnalysisMapToStruct", func(t *testing.T) {
		analysisMap := map[string]any{
			"type":     "sentiment",
			"vendor":   "test-vendor",
			"body":     "positive",
			"encoding": "json",
			"dialog":   float64(0),
			"product":  "sentiment-analyzer",
			"custom":   "value", // Non-standard field
		}

		// Test strict mode
		analysis, err := AnalysisMapToStruct(analysisMap, MigrationModeStrict)
		if err != nil {
			t.Fatalf("Unexpected error in strict mode: %v", err)
		}

		if analysis.Type != "sentiment" {
			t.Errorf("Expected Type 'sentiment', got '%v'", analysis.Type)
		}
		if analysis.Vendor != "test-vendor" {
			t.Errorf("Expected Vendor 'test-vendor', got '%s'", analysis.Vendor)
		}
		if analysis.Product == nil || *analysis.Product != "sentiment-analyzer" {
			t.Errorf("Expected Product 'sentiment-analyzer', got %v", analysis.Product)
		}

		// Test preserve mode
		analysis, err = AnalysisMapToStruct(analysisMap, MigrationModePreserve)
		if err != nil {
			t.Fatalf("Unexpected error in preserve mode: %v", err)
		}

		if val, ok := analysis.AdditionalProperties["custom"]; !ok || val != "value" {
			t.Errorf("Expected custom field to be preserved, got %v", analysis.AdditionalProperties)
		}
	})

	t.Run("AnalysisStructToMap", func(t *testing.T) {
		product := "test-product"
		analysis := &Analysis{
			Type:     "sentiment",
			Vendor:   "test-vendor",
			Body:     "positive",
			Encoding: StringPtr("json"),
			Dialog:   0,
			Product:  &product,
			AdditionalProperties: map[string]any{
				"custom": "value",
			},
		}

		result := AnalysisStructToMap(analysis)

		if result["type"] != "sentiment" {
			t.Errorf("Expected type 'sentiment', got %v", result["type"])
		}
		if result["product"] != "test-product" {
			t.Errorf("Expected product 'test-product', got %v", result["product"])
		}
		if result["custom"] != "value" {
			t.Errorf("Expected custom field, got %v", result["custom"])
		}
	})

	t.Run("MigrateVConToIETF", func(t *testing.T) {
		id := uuid.New()
		vcon := &VCon{
			UUID:      id,
			Vcon:      "0.0.1", // Old version
			CreatedAt: time.Now().UTC(),
			Parties: []Party{
				{}, // Party without identifier
			},
			Dialog: []Dialog{
				{
					Type:    "incomplete",
					Start:   time.Now().UTC(),
					Parties: NewDialogPartiesArrayPtr([]int{0}),
					// Missing disposition
				},
			},
		}

		// Test lenient mode
		err := MigrateVConToIETF(vcon, MigrationModeLenient)
		if err != nil {
			t.Fatalf("Unexpected error in lenient mode: %v", err)
		}

		// Check that version was updated
		if vcon.Vcon != "0.0.2" {
			t.Errorf("Expected version '0.0.2', got '%s'", vcon.Vcon)
		}

		// Check that party got a name
		if vcon.Parties[0].Name == nil {
			t.Error("Expected party to get a name")
		}

		// Check that dialog got disposition
		if vcon.Dialog[0].Disposition == nil {
			t.Error("Expected incomplete dialog to get disposition")
		}

		// Test strict mode with issues
		vcon.Vcon = "0.0.1" // Reset version
		err = MigrateVConToIETF(vcon, MigrationModeStrict)
		if err == nil {
			t.Error("Expected error in strict mode")
		}
	})
}

// TestValidMIMETypes tests the relaxed MIME type validation.
func TestValidMIMETypes(t *testing.T) {
	tests := []struct {
		name     string
		mimeType string
		expected bool
	}{
		{"standard audio", "audio/mpeg", true},
		{"standard video", "video/mp4", true},
		{"standard text", "text/plain", true},
		{"standard application", "application/json", true},
		{"custom type", "application/vnd.company.format", true},
		{"experimental", "text/x-custom", true},
		{"with parameters", "text/html; charset=utf-8", false}, // Basic validation doesn't handle parameters
		{"invalid format", "invalid", false},
		{"missing subtype", "audio/", false},
		{"missing type", "/mp4", false},
		{"empty string", "", false},
		{"no slash", "audiompeg", false},
		{"multiple slashes", "audio/mp3/extra", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidMIMEType(tt.mimeType)
			if result != tt.expected {
				t.Errorf("Expected %v for '%s', got %v", tt.expected, tt.mimeType, result)
			}
		})
	}
}

// Helper function to check if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && s[:len(substr)] == substr) ||
		(len(s) > len(substr) && s[len(s)-len(substr):] == substr) ||
		findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestCustomAnalysisTypes verifies that custom analysis types are accepted per IETF specification.
func TestCustomAnalysisTypes(t *testing.T) {
	v := NewWithDefaults()
	v.AddParty(Party{Name: StringPtr("Test User")})

	customTypes := []string{
		"emotion_detection",
		"keyword_extraction",
		"voice_analytics",
		"compliance_check",
		"custom_vendor_analysis",
		"ai_insights",
	}

	for _, customType := range customTypes {
		analysis := Analysis{
			Type:     customType,
			Vendor:   "TestVendor",
			Body:     map[string]any{"result": "test_data"},
			Encoding: StringPtr("json"), // Required when body is present
		}

		v.Analysis = append(v.Analysis, analysis)

		// Validate that custom types are accepted
		result := v.ValidateWithLevel(ValidationIETF)
		if !result.Valid {
			t.Errorf("Custom analysis type %q should be valid per IETF spec, but validation failed: %v",
				customType, result.Errors)
		}
	}
}

// TestStandardAnalysisTypes verifies that standard types still work.
func TestStandardAnalysisTypes(t *testing.T) {
	v := NewWithDefaults()
	v.AddParty(Party{Name: StringPtr("Test User")})

	standardTypes := []string{
		"summary",
		"transcript",
		"translation",
		"sentiment",
		"tts",
	}

	for _, standardType := range standardTypes {
		analysis := Analysis{
			Type:     standardType,
			Vendor:   "TestVendor",
			Body:     map[string]any{"result": "test_data"},
			Encoding: StringPtr("json"), // Required when body is present
		}

		v.Analysis = append(v.Analysis, analysis)

		// Validate that standard types are still accepted
		result := v.ValidateWithLevel(ValidationIETF)
		if !result.Valid {
			t.Errorf("Standard analysis type %q should be valid, but validation failed: %v",
				standardType, result.Errors)
		}
	}
}

// TestAnalysisTypeRequired verifies that empty/missing type is still rejected.
func TestAnalysisTypeRequired(t *testing.T) {
	v := NewWithDefaults()
	v.AddParty(Party{Name: StringPtr("Test User")})

	// Test empty type - should fail validation
	v.Analysis = []Analysis{
		{
			Type:     "", // Empty should fail
			Vendor:   "TestVendor",
			Body:     map[string]any{"result": "test_data"},
			Encoding: StringPtr("json"),
		},
	}

	result := v.ValidateWithLevel(ValidationIETF)
	if result.Valid {
		t.Error("Empty analysis type should fail validation")
	}

	// Test with valid custom type - should pass
	v.Analysis[0].Type = "custom_type"
	result = v.ValidateWithLevel(ValidationIETF)
	if !result.Valid {
		t.Errorf("Non-empty analysis type should pass validation: %v", result.Errors)
	}
}

// TestIETFComplianceWithCustomAnalysis verifies IETF compliance with custom analysis types.
func TestIETFComplianceWithCustomAnalysis(t *testing.T) {
	v := &VCon{
		UUID:      uuid.New(),
		Vcon:      "0.0.2",
		CreatedAt: time.Now(),
		Parties:   []Party{{Name: StringPtr("Test User")}},
		Analysis: []Analysis{
			{
				Type:     "emotion_detection", // Custom type
				Vendor:   "EmotionAI Inc",
				Body:     map[string]any{"emotion": "happy", "confidence": 0.95},
				Encoding: StringPtr("json"),
			},
			{
				Type:     "summary", // Standard type
				Vendor:   "SummaryBot",
				Body:     "This conversation was about testing custom analysis types.",
				Encoding: StringPtr("none"),
			},
		},
	}

	// Should pass IETF validation with mixed standard and custom types
	result := v.ValidateWithLevel(ValidationIETF)
	if !result.Valid {
		t.Errorf("vCon with custom analysis types should pass IETF validation: %v", result.Errors)
	}

	// Check compliance level
	level, err := v.CheckIETFCompliance()
	if err != nil {
		t.Errorf("IETF compliance check should succeed: %v", err)
	}
	if level != IETFFullyCompliant {
		t.Errorf("Expected fully compliant, got %v", level)
	}
}
