package vcon

import (
	"strings"
	"testing"
	"time"
)

// TestValidateWithLevel tests the unified validation interface.
func TestValidateWithLevel(t *testing.T) {
	tests := []struct {
		name          string
		level         ValidationLevel
		setupVCon     func() *VCon
		expectValid   bool
		expectErrors  int
		errorContains string
	}{
		{
			name:  "ValidationBasic - valid vCon",
			level: ValidationBasic,
			setupVCon: func() *VCon {
				v := NewWithDefaults()
				v.AddParty(Party{Name: StringPtr("Test User")})
				return v
			},
			expectValid: true,
		},
		{
			name:  "ValidationBasic - invalid vCon missing UUID",
			level: ValidationBasic,
			setupVCon: func() *VCon {
				return &VCon{
					Vcon:      "0.0.2",
					CreatedAt: time.Now(),
				}
			},
			expectValid:   false,
			errorContains: "required",
		},
		{
			name:  "ValidationStrict - valid vCon",
			level: ValidationStrict,
			setupVCon: func() *VCon {
				v := NewWithDefaults()
				v.AddParty(Party{Name: StringPtr("Test User")})
				return v
			},
			expectValid: true,
		},
		{
			name:  "ValidationStrict - party without identifier",
			level: ValidationStrict,
			setupVCon: func() *VCon {
				v := NewWithDefaults()
				v.AddParty(Party{}) // No identifier
				return v
			},
			expectValid:   false,
			errorContains: "party should have at least one identifier",
		},
		{
			name:  "ValidationIETF - valid IETF vCon",
			level: ValidationIETF,
			setupVCon: func() *VCon {
				v := NewWithDefaults()
				v.Vcon = "0.0.2" // Ensure correct version
				v.AddParty(Party{Name: StringPtr("Test User")})
				return v
			},
			expectValid: true,
		},
		{
			name:  "ValidationIETF - invalid version",
			level: ValidationIETF,
			setupVCon: func() *VCon {
				v := NewWithDefaults()
				v.Vcon = "0.0.1" // Wrong version
				v.AddParty(Party{Name: StringPtr("Test User")})
				return v
			},
			expectValid:   false,
			errorContains: "vcon must be '0.0.2'",
		},
		{
			name:  "ValidationIETFStrict - valid standard vCon",
			level: ValidationIETFStrict,
			setupVCon: func() *VCon {
				v := NewWithDefaults()
				v.Vcon = "0.0.2"
				v.AddParty(Party{Name: StringPtr("Test User")})
				return v
			},
			expectValid: true,
		},
		{
			name:  "ValidationIETFStrict - vCon with extension fields",
			level: ValidationIETFStrict,
			setupVCon: func() *VCon {
				v := NewWithDefaults()
				v.Vcon = "0.0.2"
				// Extensions field removed for IETF draft-03 compliance
				// Add a party with extension fields to trigger validation error
				v.AddParty(Party{
					Name: StringPtr("Test User"),
					Stir: StringPtr("stir-value"), // This is an extension field
				})
				return v
			},
			expectValid:   false,
			errorContains: "extension field violations",
		},
		{
			name:  "ValidationComplete - valid complete vCon",
			level: ValidationComplete,
			setupVCon: func() *VCon {
				v := NewWithDefaults()
				v.Vcon = "0.0.2"
				v.AddParty(Party{Name: StringPtr("Test User")})
				return v
			},
			expectValid: true,
		},
		{
			name:  "ValidationComplete - vCon with business and IETF issues",
			level: ValidationComplete,
			setupVCon: func() *VCon {
				return &VCon{
					Vcon: "0.0.1", // Wrong version (IETF issue)
					// Extensions field removed for IETF draft-03 compliance
					// Missing UUID and CreatedAt (business issue)
				}
			},
			expectValid:   false,
			errorContains: "required", // Business validation fails first
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vcon := tt.setupVCon()
			result := vcon.ValidateWithLevel(tt.level)

			if result.Level != tt.level {
				t.Errorf("Expected level %s, got %s", tt.level.String(), result.Level.String())
			}

			if tt.expectValid && !result.Valid {
				t.Errorf("Expected valid result but got invalid: %v", result.Errors)
			}

			if !tt.expectValid && result.Valid {
				t.Error("Expected invalid result but got valid")
			}

			if tt.errorContains != "" && result.Valid {
				t.Errorf("Expected error containing '%s' but result was valid", tt.errorContains)
			}

			if tt.errorContains != "" && !result.Valid {
				found := false
				for _, err := range result.Errors {
					if strings.Contains(err.Message, tt.errorContains) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error containing '%s', but not found in: %v", tt.errorContains, result.Errors)
				}
			}
		})
	}
}

// TestValidationLevel tests the ValidationLevel enum.
func TestValidationLevel(t *testing.T) {
	tests := []struct {
		level    ValidationLevel
		expected string
	}{
		{ValidationBasic, "Basic"},
		{ValidationStrict, "Strict"},
		{ValidationIETF, "IETF"},
		{ValidationIETFStrict, "IETF-Strict"},
		{ValidationComplete, "Complete"},
		{ValidationLevel(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.level.String()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestValidationResult tests the ValidationResult structure.
func TestValidationResult(t *testing.T) {
	t.Run("valid result structure", func(t *testing.T) {
		v := NewWithDefaults()
		v.AddParty(Party{Name: StringPtr("Test User")})

		result := v.ValidateWithLevel(ValidationBasic)

		if result == nil {
			t.Fatal("ValidateWithLevel returned nil result")
		}

		if result.Level != ValidationBasic {
			t.Errorf("Expected level Basic, got %s", result.Level.String())
		}

		if !result.Valid {
			t.Errorf("Expected valid result, got invalid: %v", result.Errors)
		}

		if len(result.Errors) != 0 {
			t.Errorf("Expected no errors, got %d", len(result.Errors))
		}
	})

	t.Run("invalid result structure", func(t *testing.T) {
		// Create invalid vCon
		vcon := &VCon{
			Vcon: "0.0.2",
			// Missing UUID and CreatedAt
		}

		result := vcon.ValidateWithLevel(ValidationBasic)

		if result.Valid {
			t.Error("Expected invalid result, got valid")
		}

		if len(result.Errors) == 0 {
			t.Error("Expected validation errors, got none")
		}
	})
}

// TestValidationLevelProgression tests validation level progression.
func TestValidationLevelProgression(t *testing.T) {
	// Create a vCon that passes basic validation but fails strict
	vcon := NewWithDefaults()
	vcon.AddParty(Party{}) // Empty party - passes basic, fails strict

	// Test progression: Basic should pass, Strict should fail
	basicResult := vcon.ValidateWithLevel(ValidationBasic)
	if !basicResult.Valid {
		t.Errorf("Basic validation should pass, got: %v", basicResult.Errors)
	}

	strictResult := vcon.ValidateWithLevel(ValidationStrict)
	if strictResult.Valid {
		t.Error("Strict validation should fail for empty party")
	}

	// Verify the error is about party identifier
	found := false
	for _, err := range strictResult.Errors {
		if strings.Contains(err.Message, "party should have at least one identifier") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected party identifier error, got: %v", strictResult.Errors)
	}
}

// TestCompleteStrictValidation tests comprehensive validation.
func TestCompleteStrictValidation(t *testing.T) {
	tests := []struct {
		name        string
		setupVCon   func() *VCon
		expectError bool
		errorType   string // "business", "ietf", or "none"
	}{
		{
			name: "valid vCon with no violations",
			setupVCon: func() *VCon {
				vcon := NewWithDefaults()
				vcon.AddParty(Party{Name: StringPtr("Test User")})
				vcon.AddDialog(Dialog{
					Type:      "text",
					Start:     time.Now(),
					Parties:   NewDialogPartiesArrayPtr([]int{0}),
					Body:      StringPtr("Hello world"),
					Encoding:  StringPtr("none"),
					Mediatype: StringPtr("text/plain"),
				})
				return vcon
			},
			expectError: false,
			errorType:   "none",
		},
		{
			name: "vCon with business rule violation",
			setupVCon: func() *VCon {
				vcon := &VCon{
					// Missing required fields to trigger business validation error
					Vcon: "0.0.2",
				}
				return vcon
			},
			expectError: true,
			errorType:   "business",
		},
		{
			name: "vCon with IETF extension field violations",
			setupVCon: func() *VCon {
				vcon := NewWithDefaults()
				vcon.AddParty(Party{Name: StringPtr("Test User")})

				// Add extension fields that will be flagged in strict mode
				// Extensions field removed for IETF draft-03 compliance
				// SIP field removed for IETF draft-03 compliance - using Stir instead
				vcon.Parties[0].Stir = StringPtr("stir-value") // Extension field

				return vcon
			},
			expectError: true,
			errorType:   "ietf",
		},
		{
			name: "vCon with both business and IETF violations",
			setupVCon: func() *VCon {
				vcon := &VCon{
					// Missing required fields (business violation)
					Vcon: "0.0.2",
					// Extensions field removed for IETF draft-03 compliance
				}
				return vcon
			},
			expectError: true,
			errorType:   "business", // Business validation runs first
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vcon := tt.setupVCon()
			result := vcon.ValidateWithLevel(ValidationComplete)

			if tt.expectError && result.Valid {
				t.Errorf("ValidateWithLevel(ValidationComplete) expected error (%s) but got valid result", tt.errorType)
				return
			}

			if !tt.expectError && !result.Valid {
				t.Errorf("ValidateWithLevel(ValidationComplete) expected valid result but got errors: %v", result.Errors)
			}
		})
	}
}

// TestValidationWarnings tests validation warning functionality.
func TestValidationWarnings(t *testing.T) {
	warning := ValidationWarning{
		Field:   "test_field",
		Message: "test warning message",
		Level:   "warning",
	}

	expected := "validation warning in test_field (warning): test warning message"
	if warning.String() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, warning.String())
	}
}

// TestValidationIntegration tests integration across all validation levels.
func TestValidationIntegration(t *testing.T) {
	// Integration test to ensure all validation layers work together properly
	vcon := NewWithDefaults()
	vcon.AddParty(Party{
		Name: StringPtr("Integration Test User"),
		// Add both compliant and extension fields
		Tel:  StringPtr("+1234567890"), // Compliant
		Stir: StringPtr("stir-value"),  // Extension - will be flagged in strict mode
	})

	vcon.AddDialog(Dialog{
		Type:      "text",
		Start:     time.Now(),
		Parties:   NewDialogPartiesArrayPtr([]int{0}),
		Body:      StringPtr("Test message"),
		Encoding:  StringPtr("none"),
		Mediatype: StringPtr("text/plain"),
		// Extension fields
		Campaign:  StringPtr("test-campaign"), // Extension field
		SessionID: StringPtr("session-123"),   // Extension field
	})

	// Test progression through validation levels
	basicResult := vcon.ValidateWithLevel(ValidationBasic)
	if !basicResult.Valid {
		t.Errorf("Basic validation should pass, got: %v", basicResult.Errors)
	}

	strictResult := vcon.ValidateWithLevel(ValidationStrict)
	if !strictResult.Valid {
		t.Errorf("Strict validation should pass, got: %v", strictResult.Errors)
	}

	ietfResult := vcon.ValidateWithLevel(ValidationIETF)
	if !ietfResult.Valid {
		t.Errorf("IETF validation should pass, got: %v", ietfResult.Errors)
	}

	// IETF Strict should fail due to extension fields
	ietfStrictResult := vcon.ValidateWithLevel(ValidationIETFStrict)
	if ietfStrictResult.Valid {
		t.Error("IETF Strict validation should fail due to extension fields")
	}

	// Complete should also fail (since it includes IETF strict)
	completeResult := vcon.ValidateWithLevel(ValidationComplete)
	if completeResult.Valid {
		t.Error("Complete validation should fail due to extension fields")
	}
}

// TestInvalidValidationLevel tests handling of invalid validation levels.
func TestInvalidValidationLevel(t *testing.T) {
	vcon := NewWithDefaults()
	vcon.AddParty(Party{Name: StringPtr("Test User")})

	// Test with invalid validation level
	result := vcon.ValidateWithLevel(ValidationLevel(999))

	if result.Valid {
		t.Error("Expected invalid result for unknown validation level")
	}

	if len(result.Errors) == 0 {
		t.Error("Expected error for unknown validation level")
	}

	errorFound := false
	for _, err := range result.Errors {
		if strings.Contains(err.Message, "unknown validation level") {
			errorFound = true
			break
		}
	}

	if !errorFound {
		t.Errorf("Expected 'unknown validation level' error, got: %v", result.Errors)
	}
}
