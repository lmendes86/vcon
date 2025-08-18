package vcon

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestVConValidate tests basic VCon validation (business logic).
func TestVConValidate(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *VCon
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid vcon",
			setup: func() *VCon {
				v := NewWithDefaults()
				email := "test@example.com"
				v.AddParty(Party{Mailto: &email})
				v.AddDialog(Dialog{
					Type:    "text",
					Start:   time.Now(),
					Parties: NewDialogPartiesArrayPtr([]int{0}),
					Body:    "Hello",
				})
				return v
			},
			wantErr: false,
		},
		{
			name: "missing UUID",
			setup: func() *VCon {
				return &VCon{
					Vcon:      "1.0.0",
					CreatedAt: time.Now(),
				}
			},
			wantErr: true,
			errMsg:  "required",
		},
		{
			name: "missing version",
			setup: func() *VCon {
				id := uuid.New()
				return &VCon{
					UUID:      id,
					CreatedAt: time.Now(),
				}
			},
			wantErr: true,
			errMsg:  "Field validation for 'Vcon' failed on the 'required' tag",
		},
		{
			name: "missing created_at",
			setup: func() *VCon {
				id := uuid.New()
				return &VCon{
					UUID: id,
					Vcon: "1.0.0",
				}
			},
			wantErr: true,
			errMsg:  "Field validation for 'CreatedAt' failed on the 'required' tag",
		},
		{
			name: "updated_at before created_at",
			setup: func() *VCon {
				v := NewWithDefaults()
				past := v.CreatedAt.Add(-1 * time.Hour)
				v.UpdatedAt = &past
				return v
			},
			wantErr: true,
			errMsg:  "Field validation for 'UpdatedAt' failed on the 'gtfield' tag",
		},
		{
			name: "invalid party index in dialog",
			setup: func() *VCon {
				v := NewWithDefaults()
				// Add one party so we have a valid vCon base
				v.AddParty(Party{Name: StringPtr("User 0")})
				v.AddDialog(Dialog{
					Type:    "text",
					Start:   time.Now(),
					Parties: NewDialogPartiesArrayPtr([]int{0, 1}), // Party 1 doesn't exist
					Body:    "Hello",
				})
				return v
			},
			wantErr: true,
			errMsg:  "invalid party index",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := tt.setup()
			err := v.Validate()

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestPartyValidate tests Party struct validation.
func TestPartyValidate(t *testing.T) {
	tests := []struct {
		name    string
		party   Party
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid party with email",
			party: Party{
				Mailto: strPtr("test@example.com"),
			},
			wantErr: false,
		},
		{
			name: "valid party with phone",
			party: Party{
				Tel: strPtr("+1234567890"),
			},
			wantErr: false,
		},
		{
			name: "valid party with name only",
			party: Party{
				Name: strPtr("John Doe"),
			},
			wantErr: false,
		},
		{
			name: "valid party with UUID",
			party: func() Party {
				id := uuid.New()
				return Party{UUID: &id}
			}(),
			wantErr: false,
		},
		{
			name:    "party without identifier",
			party:   Party{},
			wantErr: false, // ValidateStruct doesn't check business logic - only VCon.Validate() does
		},
		{
			name: "invalid email",
			party: Party{
				Mailto: strPtr("not-an-email"),
			},
			wantErr: true,
			errMsg:  "Field validation for 'Mailto' failed on the 'mailto_uri' tag",
		},
		{
			name: "invalid phone",
			party: Party{
				Tel: strPtr("invalid-phone"), // Invalid format
			},
			wantErr: true,
			errMsg:  "Field validation for 'Tel' failed on the 'tel_uri' tag",
		},
		{
			name: "valid party with civic address",
			party: Party{
				Name: strPtr("John Doe"),
				CivicAddress: &CivicAddress{
					Country: strPtr("US"),
					A3:      strPtr("New York"),
				},
			},
			wantErr: false,
		},
		{
			name: "party with invalid civic address",
			party: Party{
				Name: strPtr("John Doe"),
				CivicAddress: &CivicAddress{
					Country: strPtr("USA"), // Should be 2 characters
				},
			},
			wantErr: true,
			errMsg:  "Field validation for 'Country' failed on the 'len' tag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(&tt.party)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestDialogValidate tests Dialog struct validation.
func TestDialogValidate(t *testing.T) {
	tests := []struct {
		name    string
		dialog  Dialog
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid text dialog",
			dialog: Dialog{
				Type:    "text",
				Start:   time.Now(),
				Parties: NewDialogPartiesArrayPtr([]int{0}),
				Body:    "Hello",
			},
			wantErr: false,
		},
		{
			name: "valid recording dialog",
			dialog: Dialog{
				Type:      "recording",
				Start:     time.Now(),
				Parties:   NewDialogPartiesArrayPtr([]int{0, 1}),
				Duration:  float64Ptr(60.5),
				Mediatype: strPtr("audio/mp3"),
			},
			wantErr: false,
		},
		{
			name: "valid recording dialog with video content",
			dialog: Dialog{
				Type:      "recording",
				Start:     time.Now(),
				Parties:   NewDialogPartiesArrayPtr([]int{0}),
				Duration:  float64Ptr(120),
				FrameRate: float64Ptr(30),
				Bitrate:   intPtr(2000),
				Mediatype: strPtr("video/mp4"),
			},
			wantErr: false,
		},
		{
			name: "missing type",
			dialog: Dialog{
				Start:   time.Now(),
				Parties: NewDialogPartiesArrayPtr([]int{0}),
			},
			wantErr: true,
			errMsg:  "Field validation for 'Type' failed on the 'required' tag",
		},
		{
			name: "invalid type",
			dialog: Dialog{
				Type:    "invalid",
				Start:   time.Now(),
				Parties: NewDialogPartiesArrayPtr([]int{0}),
			},
			wantErr: true,
			errMsg:  "Field validation for 'Type' failed on the 'oneof' tag",
		},
		{
			name: "missing start time",
			dialog: Dialog{
				Type:    "text",
				Parties: NewDialogPartiesArrayPtr([]int{0}),
				Body:    "Hello",
			},
			wantErr: true,
			errMsg:  "Field validation for 'Start' failed on the 'required' tag",
		},
		{
			name: "missing parties - now allowed per IETF spec",
			dialog: Dialog{
				Type:  "text",
				Start: time.Now(),
				Body:  "Hello",
			},
			wantErr: false, // Parties are now optional per IETF spec
		},
		{
			name: "text dialog without body",
			dialog: Dialog{
				Type:    "text",
				Start:   time.Now(),
				Parties: NewDialogPartiesArrayPtr([]int{0}),
			},
			wantErr: false, // ValidateStruct doesn't check business logic - would need Dialog.Validate() method
		},
		{
			name: "recording with negative duration",
			dialog: Dialog{
				Type:     "recording",
				Start:    time.Now(),
				Parties:  NewDialogPartiesArrayPtr([]int{0}),
				Duration: float64Ptr(-10),
			},
			wantErr: true,
			errMsg:  "Field validation for 'Duration' failed on the 'gt' tag",
		},
		{
			name: "invalid MIME type",
			dialog: Dialog{
				Type:      "recording",
				Start:     time.Now(),
				Parties:   NewDialogPartiesArrayPtr([]int{0}),
				Mediatype: strPtr("invalid"),
			},
			wantErr: true,
			errMsg:  "Field validation for 'Mediatype' failed on the 'contains' tag",
		},
		{
			name: "recording with negative frame rate",
			dialog: Dialog{
				Type:      "recording",
				Start:     time.Now(),
				Parties:   NewDialogPartiesArrayPtr([]int{0}),
				FrameRate: float64Ptr(-30),
				Mediatype: strPtr("video/mp4"),
			},
			wantErr: true,
			errMsg:  "Field validation for 'FrameRate' failed on the 'gt' tag",
		},
		{
			name: "recording with negative bitrate",
			dialog: Dialog{
				Type:      "recording",
				Start:     time.Now(),
				Parties:   NewDialogPartiesArrayPtr([]int{0}),
				Bitrate:   intPtr(-1000),
				Mediatype: strPtr("video/mp4"),
			},
			wantErr: true,
			errMsg:  "Field validation for 'Bitrate' failed on the 'gt' tag",
		},
		{
			name: "transfer without transfer fields",
			dialog: Dialog{
				Type:    "transfer",
				Start:   time.Now(),
				Parties: NewDialogPartiesArrayPtr([]int{0}),
			},
			wantErr: false, // ValidateStruct doesn't check business logic - would need Dialog.Validate() method
		},
		{
			name: "valid transfer dialog",
			dialog: Dialog{
				Type:       "transfer",
				Start:      time.Now(),
				Parties:    NewDialogPartiesArrayPtr([]int{0, 1}),
				Transferee: intPtr(1),
				Transferor: intPtr(0),
			},
			wantErr: false,
		},
		{
			name: "invalid URL",
			dialog: Dialog{
				Type:    "recording",
				Start:   time.Now(),
				Parties: NewDialogPartiesArrayPtr([]int{0}),
				URL:     strPtr("not-a-url"),
			},
			wantErr: true,
			errMsg:  "Field validation for 'URL' failed on the 'url' tag",
		},
		{
			name: "invalid encoding",
			dialog: Dialog{
				Type:     "recording",
				Start:    time.Now(),
				Parties:  NewDialogPartiesArrayPtr([]int{0}),
				Encoding: strPtr("invalid"),
			},
			wantErr: true,
			errMsg:  "Field validation for 'Encoding' failed on the 'oneof' tag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(&tt.dialog)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestAttachmentValidate tests Attachment struct validation.
func TestAttachmentValidate(t *testing.T) {
	tests := []struct {
		name    string
		att     Attachment
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid attachment",
			att: Attachment{
				Type:      StringPtr("document"),
				Body:      "SGVsbG8=",
				Encoding:  StringPtr("base64url"),
				Mediatype: StringPtr("application/octet-stream"), // Added required mediatype
				Start:     func() *time.Time { t := time.Now().UTC(); return &t }(),
				Party:     IntPtr(0), // Changed to valid party index
			},
			wantErr: false,
		},
		{
			name: "missing type - now allowed per IETF spec",
			att: Attachment{
				Body:      "data",
				Encoding:  StringPtr("base64url"),
				Mediatype: StringPtr("text/plain"), // Added required mediatype
				Start:     func() *time.Time { t := time.Now().UTC(); return &t }(),
				Party:     IntPtr(0), // Changed to valid party index
			},
			wantErr: false, // Type is optional per IETF spec
		},
		{
			name: "missing body and url - now validation error",
			att: Attachment{
				Type:     StringPtr("document"),
				Encoding: StringPtr("base64url"),
				Start:    func() *time.Time { t := time.Now().UTC(); return &t }(),
				Party:    IntPtr(0), // Changed to valid party index
				// No body or URL - should fail business logic validation
			},
			wantErr: true,
			errMsg:  "attachment must have either body or url",
		},
		{
			name: "valid attachment with only URL",
			att: Attachment{
				Type:        StringPtr("document"),
				URL:         StringPtr("https://example.com/doc.pdf"),
				ContentHash: NewContentHashSingle("sha-256:LCa0a2j_xo_5m0U8HTBBNBNCLXBkg7-g-YpeiGJm564"), // Added required content_hash
				Start:       func() *time.Time { t := time.Now().UTC(); return &t }(),
				Party:       IntPtr(0), // Changed to valid party index
				// No body or encoding - should be valid since URL is present
			},
			wantErr: false,
		},
		{
			name: "missing encoding when body present - validation error",
			att: Attachment{
				Type:  StringPtr("document"),
				Body:  "data",
				Start: func() *time.Time { t := time.Now().UTC(); return &t }(),
				Party: IntPtr(0), // Changed to valid party index
				// No encoding when body is present - should fail business logic
			},
			wantErr: true,
			errMsg:  "encoding is required when body is present",
		},
		{
			name: "invalid encoding",
			att: Attachment{
				Type:     StringPtr("document"),
				Body:     "data",
				Encoding: StringPtr("invalid"),
				Start:    func() *time.Time { t := time.Now().UTC(); return &t }(),
				Party:    IntPtr(0), // Changed to valid party index
			},
			wantErr: true,
			errMsg:  "Field validation for 'Encoding' failed on the 'oneof' tag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(&tt.att)

			// For business logic validation, also test with a VCon that includes the attachment
			v := NewWithDefaults()
			v.AddParty(Party{Name: StringPtr("Test User")})
			v.AddAttachment(tt.att)
			businessLogicErr := v.ValidateAdvanced()

			if tt.wantErr {
				// Check either struct validation or business logic validation
				if err == nil && businessLogicErr == nil {
					t.Error("Expected error but got none")
				} else if tt.errMsg != "" {
					errorFound := false
					if err != nil && strings.Contains(err.Error(), tt.errMsg) {
						errorFound = true
					}
					if businessLogicErr != nil && strings.Contains(businessLogicErr.Error(), tt.errMsg) {
						errorFound = true
					}
					if !errorFound {
						t.Errorf("Expected error containing '%s', got struct err: %v, business err: %v", tt.errMsg, err, businessLogicErr)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected struct validation error: %v", err)
				}
				if businessLogicErr != nil {
					t.Errorf("Unexpected business logic error: %v", businessLogicErr)
				}
			}
		})
	}
}

// TestVConValidateStrict tests strict business validation rules.
func TestVConValidateStrict(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *VCon
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid vcon",
			setup: func() *VCon {
				v := NewWithDefaults()
				email := "test@example.com"
				v.AddParty(Party{Mailto: &email})
				v.AddDialog(Dialog{
					Type:    "text",
					Start:   time.Now(),
					Parties: NewDialogPartiesArrayPtr([]int{0}),
					Body:    "Hello",
				})
				return v
			},
			wantErr: false,
		},
		{
			name: "duplicate party UUIDs",
			setup: func() *VCon {
				v := NewWithDefaults()
				id := uuid.New()
				v.AddParty(Party{UUID: &id})
				v.AddParty(Party{UUID: &id}) // Same UUID
				return v
			},
			wantErr: true,
			errMsg:  "duplicate UUID",
		},
		{
			name: "dialogs not in chronological order",
			setup: func() *VCon {
				v := NewWithDefaults()
				email := "test@example.com"
				v.AddParty(Party{Mailto: &email})

				now := time.Now()
				v.AddDialog(Dialog{
					Type:    "text",
					Start:   now,
					Parties: NewDialogPartiesArrayPtr([]int{0}),
					Body:    "First",
				})
				v.AddDialog(Dialog{
					Type:    "text",
					Start:   now.Add(-1 * time.Hour), // Earlier than first
					Parties: NewDialogPartiesArrayPtr([]int{0}),
					Body:    "Second",
				})
				return v
			},
			wantErr: true,
			errMsg:  "should be in chronological order",
		},
		{
			name: "invalid originator index",
			setup: func() *VCon {
				v := NewWithDefaults()
				email := "test@example.com"
				v.AddParty(Party{Mailto: &email})

				originator := 5
				v.AddDialog(Dialog{
					Type:       "text",
					Start:      time.Now(),
					Parties:    NewDialogPartiesArrayPtr([]int{0}),
					Originator: &originator, // Invalid index
					Body:       "Hello",
				})
				return v
			},
			wantErr: true,
			errMsg:  "invalid originator index",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := tt.setup()
			err := v.ValidateStrict()

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestPartyIdentifierStrictValidation tests business rule enforcement for party identifiers.
func TestPartyIdentifierStrictValidation(t *testing.T) {
	// Test that basic validation now allows empty parties (IETF compliance)
	// but strict validation still enforces the business rule

	t.Run("empty party passes basic validation", func(t *testing.T) {
		v := NewWithDefaults()
		v.AddParty(Party{}) // Empty party

		err := v.Validate()
		if err != nil {
			t.Errorf("Empty party should pass basic validation per IETF spec, got: %v", err)
		}
	})

	t.Run("empty party fails strict validation", func(t *testing.T) {
		v := NewWithDefaults()
		v.AddParty(Party{}) // Empty party

		err := v.ValidateStrict()
		if err == nil {
			t.Error("Empty party should fail strict validation (business rule)")
		}

		if !strings.Contains(err.Error(), "party should have at least one identifier") {
			t.Errorf("Expected business rule error message, got: %v", err)
		}
	})

	t.Run("party with identifier passes both validations", func(t *testing.T) {
		v := NewWithDefaults()
		v.AddParty(Party{Name: strPtr("John Doe")})

		err := v.Validate()
		if err != nil {
			t.Errorf("Party with identifier should pass basic validation, got: %v", err)
		}

		err = v.ValidateStrict()
		if err != nil {
			t.Errorf("Party with identifier should pass strict validation, got: %v", err)
		}
	})
}

// TestValidationErrors tests ValidationErrors type.
func TestValidationErrors(t *testing.T) {
	errors := ValidationErrors{
		{Field: "field1", Message: "error1"},
		{Field: "field2", Message: "error2"},
	}

	errStr := errors.Error()
	if !strings.Contains(errStr, "field1") {
		t.Error("Error string should contain field1")
	}
	if !strings.Contains(errStr, "error1") {
		t.Error("Error string should contain error1")
	}
	if !strings.Contains(errStr, "field2") {
		t.Error("Error string should contain field2")
	}

	// Test empty errors
	emptyErrors := ValidationErrors{}
	if emptyErrors.Error() != "no validation errors" {
		t.Error("Empty ValidationErrors should return 'no validation errors'")
	}
}

// TestVConIsValid tests the IsValid convenience method.
func TestVConIsValid(t *testing.T) {
	v := NewWithDefaults()
	// Add required party for IETF compliance
	v.AddParty(Party{Name: StringPtr("Test User")})
	if valid, _ := v.IsValid(); !valid {
		t.Error("New VCon should be valid")
	}

	// UUID is now optional, so test with empty parties instead
	v.Parties = nil
	if valid, _ := v.IsValid(); valid {
		t.Error("VCon without parties should not be valid")
	}
}

// TestPartyHistoryValidate tests PartyHistory struct validation.
func TestPartyHistoryValidate(t *testing.T) {
	tests := []struct {
		name    string
		ph      PartyHistory
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid party history",
			ph: PartyHistory{
				Party: 0,
				Event: "join",
				Time:  time.Now(),
			},
			wantErr: false,
		},
		{
			name: "negative party index",
			ph: PartyHistory{
				Party: -1,
				Event: "join",
				Time:  time.Now(),
			},
			wantErr: true,
			errMsg:  "Field validation for 'Party' failed on the 'min' tag",
		},
		{
			name: "missing event",
			ph: PartyHistory{
				Party: 0,
				Time:  time.Now(),
			},
			wantErr: true,
			errMsg:  "Field validation for 'Event' failed on the 'required' tag",
		},
		{
			name: "invalid event",
			ph: PartyHistory{
				Party: 0,
				Event: "invalid",
				Time:  time.Now(),
			},
			wantErr: true,
			errMsg:  "Field validation for 'Event' failed on the 'oneof' tag",
		},
		{
			name: "missing time",
			ph: PartyHistory{
				Party: 0,
				Event: "join",
			},
			wantErr: true,
			errMsg:  "Field validation for 'Time' failed on the 'required' tag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(&tt.ph)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestCivicAddressValidate tests CivicAddress struct validation.
func TestCivicAddressValidate(t *testing.T) {
	tests := []struct {
		name    string
		addr    CivicAddress
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid civic address",
			addr: CivicAddress{
				Country: strPtr("US"),
				A3:      strPtr("New York"),
				Sts:     strPtr("5th Avenue"),
				Hno:     strPtr("123"),
			},
			wantErr: false,
		},
		{
			name: "invalid country code",
			addr: CivicAddress{
				Country: strPtr("USA"), // Should be 2 characters
			},
			wantErr: true,
			errMsg:  "Field validation for 'Country' failed on the 'len' tag",
		},
		{
			name:    "empty civic address",
			addr:    CivicAddress{},
			wantErr: false, // All fields are optional
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(&tt.addr)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestEmptyPartiesValidation tests that empty parties arrays are allowed per IETF draft-03 spec.
func TestEmptyPartiesValidation(t *testing.T) {
	tests := []struct {
		name           string
		validationFunc func(*VCon) error
		expectValid    bool
	}{
		{
			name: "IETF validation with empty parties",
			validationFunc: func(v *VCon) error {
				return v.ValidateIETF()
			},
			expectValid: true,
		},
		{
			name: "Basic validation with empty parties",
			validationFunc: func(v *VCon) error {
				return v.Validate()
			},
			expectValid: true,
		},
		{
			name: "IETF strict validation with empty parties",
			validationFunc: func(v *VCon) error {
				return v.ValidateIETFStrict()
			},
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create vCon with empty parties array
			id := uuid.New()
			vcon := &VCon{
				UUID:      id,
				Vcon:      "0.0.2",
				CreatedAt: time.Now(),
				Parties:   []Party{}, // Empty parties array - should be valid per IETF draft-03
			}

			err := tt.validationFunc(vcon)
			if tt.expectValid && err != nil {
				t.Errorf("Expected validation to pass with empty parties array, but got error: %v", err)
			}
			if !tt.expectValid && err == nil {
				t.Error("Expected validation to fail, but it passed")
			}
		})
	}

	// Additional test using ValidationWithLevel
	t.Run("ValidationWithLevel IETF with empty parties", func(t *testing.T) {
		id := uuid.New()
		vcon := &VCon{
			UUID:      id,
			Vcon:      "0.0.2",
			CreatedAt: time.Now(),
			Parties:   []Party{}, // Empty parties array
		}

		result := vcon.ValidateWithLevel(ValidationIETF)
		if !result.Valid {
			t.Errorf("ValidationWithLevel(IETF) should pass with empty parties array, but got errors: %v", result.Errors)
		}
	})
}

// TestPartyURIFormatValidation tests that party tel and mailto fields accept URI formats per IETF spec.
func TestPartyURIFormatValidation(t *testing.T) {
	tests := []struct {
		name        string
		party       Party
		expectValid bool
		description string
	}{
		{
			name: "tel URI format",
			party: Party{
				Name: StringPtr("Test User"),
				Tel:  StringPtr("tel:+1234567890"),
			},
			expectValid: true,
			description: "tel: URI format should be valid per IETF spec",
		},
		{
			name: "plain tel format",
			party: Party{
				Name: StringPtr("Test User"),
				Tel:  StringPtr("+1234567890"),
			},
			expectValid: true,
			description: "Plain E.164 format should remain valid",
		},
		{
			name: "mailto URI format",
			party: Party{
				Name:   StringPtr("Test User"),
				Mailto: StringPtr("mailto:user@example.com"),
			},
			expectValid: true,
			description: "mailto: URI format should be valid per IETF spec",
		},
		{
			name: "plain email format",
			party: Party{
				Name:   StringPtr("Test User"),
				Mailto: StringPtr("user@example.com"),
			},
			expectValid: true,
			description: "Plain email format should remain valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := uuid.New()
			vcon := &VCon{
				UUID:      id,
				Vcon:      "0.0.2",
				CreatedAt: time.Now(),
				Parties:   []Party{tt.party},
			}

			err := vcon.Validate()
			if tt.expectValid && err != nil {
				t.Errorf("%s: Expected validation to pass but got error: %v", tt.description, err)
			}
			if !tt.expectValid && err == nil {
				t.Errorf("%s: Expected validation to fail but it passed", tt.description)
			}
		})
	}
}

// Helper functions.
func strPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}

// containsString checks if a string contains a substring.
func containsString(s, substr string) bool {
	return strings.Contains(s, substr)
}
