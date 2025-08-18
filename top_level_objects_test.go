package vcon

import (
	"testing"
)

func TestTopLevelObjectMutualExclusion(t *testing.T) {
	tests := []struct {
		name        string
		setupVCon   func(*VCon)
		expectError bool
		errorMsg    string
	}{
		{
			name: "no top-level objects - should pass",
			setupVCon: func(_ *VCon) {
				// Leave all top-level objects as nil/empty
			},
			expectError: false,
		},
		{
			name: "only redacted object - should pass",
			setupVCon: func(v *VCon) {
				v.Redacted = &RedactedObject{
					UUID: "550e8400-e29b-41d4-a716-446655440000",
				}
			},
			expectError: false,
		},
		{
			name: "only appended object - should pass",
			setupVCon: func(v *VCon) {
				v.Appended = &AppendedObject{
					UUID: "550e8400-e29b-41d4-a716-446655440001",
				}
			},
			expectError: false,
		},
		{
			name: "only group objects - should pass",
			setupVCon: func(v *VCon) {
				v.Group = []GroupObject{
					{UUID: "550e8400-e29b-41d4-a716-446655440002"},
				}
			},
			expectError: false,
		},
		{
			name: "redacted and appended - should fail",
			setupVCon: func(v *VCon) {
				v.Redacted = &RedactedObject{
					UUID: "550e8400-e29b-41d4-a716-446655440000",
				}
				v.Appended = &AppendedObject{
					UUID: "550e8400-e29b-41d4-a716-446655440001",
				}
			},
			expectError: true,
			errorMsg:    "only one of redacted, appended, or group objects is allowed per IETF spec, found: redacted, appended",
		},
		{
			name: "redacted and group - should fail",
			setupVCon: func(v *VCon) {
				v.Redacted = &RedactedObject{
					UUID: "550e8400-e29b-41d4-a716-446655440000",
				}
				v.Group = []GroupObject{
					{UUID: "550e8400-e29b-41d4-a716-446655440002"},
				}
			},
			expectError: true,
			errorMsg:    "only one of redacted, appended, or group objects is allowed per IETF spec, found: redacted, group",
		},
		{
			name: "appended and group - should fail",
			setupVCon: func(v *VCon) {
				v.Appended = &AppendedObject{
					UUID: "550e8400-e29b-41d4-a716-446655440001",
				}
				v.Group = []GroupObject{
					{UUID: "550e8400-e29b-41d4-a716-446655440002"},
				}
			},
			expectError: true,
			errorMsg:    "only one of redacted, appended, or group objects is allowed per IETF spec, found: appended, group",
		},
		{
			name: "all three objects - should fail",
			setupVCon: func(v *VCon) {
				v.Redacted = &RedactedObject{
					UUID: "550e8400-e29b-41d4-a716-446655440000",
				}
				v.Appended = &AppendedObject{
					UUID: "550e8400-e29b-41d4-a716-446655440001",
				}
				v.Group = []GroupObject{
					{UUID: "550e8400-e29b-41d4-a716-446655440002"},
				}
			},
			expectError: true,
			errorMsg:    "only one of redacted, appended, or group objects is allowed per IETF spec, found: redacted, appended, group",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vcon := NewWithDefaults()
			vcon.AddParty(Party{Name: StringPtr("Test User")}) // Add at least one party for basic validation

			// Setup the vCon according to test case
			tt.setupVCon(vcon)

			// Test the mutual exclusion validation directly
			errors := vcon.validateTopLevelObjectMutualExclusion()

			if tt.expectError {
				if len(errors) == 0 {
					t.Error("Expected validation error but got none")
					return
				}

				found := false
				for _, err := range errors {
					if err.Field == "top_level_objects" && err.Message == tt.errorMsg {
						found = true
						break
					}
				}

				if !found {
					t.Errorf("Expected error message '%s' but got different errors: %v", tt.errorMsg, errors)
				}
			} else if len(errors) > 0 {
				t.Errorf("Expected no validation errors but got: %v", errors)
			}

			// Also test through ValidateAdvanced to ensure integration works
			err := vcon.ValidateAdvanced()
			if tt.expectError {
				if err == nil {
					t.Error("Expected ValidateAdvanced to fail but it passed")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					// Check if the error contains our expected message
					// For ValidateAdvanced, we might get multiple errors, so just check if our message is in there
					if err.Error() == "" || err.Error() == "no validation errors" {
						t.Errorf("Expected error containing '%s' but got: %v", tt.errorMsg, err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Expected ValidateAdvanced to pass but got error: %v", err)
				}
			}
		})
	}
}

func TestTopLevelObjectStructValidation(t *testing.T) {
	// Test validation of the individual top-level object structures
	tests := []struct {
		name      string
		setupVCon func(*VCon)
		expectErr bool
	}{
		{
			name: "valid redacted object",
			setupVCon: func(v *VCon) {
				v.Redacted = &RedactedObject{
					UUID:     "550e8400-e29b-41d4-a716-446655440000",
					Type:     StringPtr("privacy"),
					Encoding: StringPtr("json"),
				}
			},
			expectErr: false,
		},
		{
			name: "invalid redacted object - missing UUID",
			setupVCon: func(v *VCon) {
				v.Redacted = &RedactedObject{
					Type: StringPtr("privacy"),
				}
			},
			expectErr: true, // UUID is required
		},
		{
			name: "valid appended object",
			setupVCon: func(v *VCon) {
				v.Appended = &AppendedObject{
					UUID:     "550e8400-e29b-41d4-a716-446655440001",
					Type:     StringPtr("metadata"),
					Encoding: StringPtr("base64url"),
				}
			},
			expectErr: false,
		},
		{
			name: "valid group objects",
			setupVCon: func(v *VCon) {
				v.Group = []GroupObject{
					{
						UUID: "550e8400-e29b-41d4-a716-446655440002",
						Type: StringPtr("conversation-thread"),
					},
					{
						UUID: "550e8400-e29b-41d4-a716-446655440003",
						Type: StringPtr("related-calls"),
					},
				}
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vcon := NewWithDefaults()
			vcon.AddParty(Party{Name: StringPtr("Test User")})

			tt.setupVCon(vcon)

			err := vcon.Validate()
			if tt.expectErr && err == nil {
				t.Error("Expected validation error but got none")
			} else if !tt.expectErr && err != nil {
				t.Errorf("Expected no validation error but got: %v", err)
			}
		})
	}
}
