package vcon

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestValidateIETFStrict(t *testing.T) {
	tests := []struct {
		name        string
		vcon        *VCon
		expectError bool
		errorCount  int
		extensions  []string
	}{
		{
			name: "valid IETF vCon with no extensions",
			vcon: &VCon{
				UUID:      newTestUUID(),
				Vcon:      "0.0.2",
				CreatedAt: time.Now(),
				Parties: []Party{
					{Name: StringPtr("Test User")},
				},
				Dialog: []Dialog{
					{
						Type:      "text",
						Start:     time.Now(),
						Parties:   NewDialogPartiesArrayPtr([]int{0}),
						Body:      "Hello world",
						Encoding:  StringPtr("none"),
						Mediatype: StringPtr("text/plain"),
					},
				},
			},
			expectError: false,
		},
		{
			name: "deprecated top-level extensions field - now IETF compliant",
			vcon: &VCon{
				UUID:      newTestUUID(),
				Vcon:      "0.0.2",
				CreatedAt: time.Now(),
				// Extensions field removed for IETF draft-03 compliance - no longer triggers validation error
				Parties: []Party{
					{Name: StringPtr("Test User")},
				},
			},
			expectError: false, // Changed: field removed, so no validation error
		},
		{
			name: "deprecated must_support field - now IETF compliant",
			vcon: &VCon{
				UUID:      newTestUUID(),
				Vcon:      "0.0.2",
				CreatedAt: time.Now(),
				// MustSupport field removed for IETF draft-03 compliance - no longer triggers validation error
				Parties: []Party{
					{Name: StringPtr("Test User")},
				},
			},
			expectError: false, // Changed: field removed, so no validation error
		},
		{
			name: "party with telephony extensions",
			vcon: &VCon{
				UUID:      newTestUUID(),
				Vcon:      "0.0.2",
				CreatedAt: time.Now(),
				Parties: []Party{
					{
						Name: StringPtr("Test User"),
						Stir: StringPtr("stir-value"),
						// Note: SIP field removed for IETF draft-03 compliance
					},
				},
			},
			expectError: true,
			errorCount:  1,                // Changed: was 2, now 1 since SIP field removed
			extensions:  []string{"stir"}, // Removed "sip"
		},
		{
			name: "dialog with video/media extensions",
			vcon: &VCon{
				UUID:      newTestUUID(),
				Vcon:      "0.0.2",
				CreatedAt: time.Now(),
				Parties: []Party{
					{Name: StringPtr("Test User")},
				},
				Dialog: []Dialog{
					{
						Type:        "recording",
						Start:       time.Now(),
						Parties:     NewDialogPartiesArrayPtr([]int{0}),
						URL:         StringPtr("https://example.com/recording.wav"),
						ContentHash: NewContentHashSingle("sha-256:LCa0a2j_xo_5m0U8HTBBNBNCLXBkg7-g-YpeiGJm564"),
						Resolution:  StringPtr("1920x1080"),
						FrameRate:   Float64Ptr(30.0),
						Codec:       StringPtr("h264"),
					},
				},
			},
			expectError: true,
			errorCount:  3,
			extensions:  []string{"resolution", "frame_rate", "codec"},
		},
		{
			name: "analysis with vendor extensions",
			vcon: &VCon{
				UUID:      newTestUUID(),
				Vcon:      "0.0.2",
				CreatedAt: time.Now(),
				Parties: []Party{
					{Name: StringPtr("Test User")},
				},
				Analysis: []Analysis{
					{
						Type:        "summary",
						Vendor:      "test-vendor",
						Product:     StringPtr("analyzer-v1.0"),
						Schema:      StringPtr("https://example.com/schema"),
						Mediatype:   StringPtr("application/json"),
						URL:         StringPtr("https://example.com/analysis.json"),
						ContentHash: NewContentHashSingle("sha-256:LCa0a2j_xo_5m0U8HTBBNBNCLXBkg7-g-YpeiGJm564"),
					},
				},
			},
			expectError: true,
			errorCount:  3,
			extensions:  []string{"product", "schema", "mediatype"},
		},
		{
			name: "attachment with content extensions",
			vcon: &VCon{
				UUID:      newTestUUID(),
				Vcon:      "0.0.2",
				CreatedAt: time.Now(),
				Parties: []Party{
					{Name: StringPtr("Test User")},
				},
				Dialog: []Dialog{
					{
						Type:      "text",
						Start:     time.Now(),
						Parties:   NewDialogPartiesArrayPtr([]int{0}),
						Body:      "Sample text",
						Encoding:  StringPtr("none"),
						Mediatype: StringPtr("text/plain"),
					},
				},
				Attachments: []Attachment{
					{
						Start:       func() *time.Time { t := time.Now(); return &t }(),
						Party:       IntPtr(0),
						Dialog:      IntPtr(0),
						Mediatype:   StringPtr("application/pdf"),
						Filename:    StringPtr("document.pdf"),
						URL:         StringPtr("https://example.com/document.pdf"),
						ContentHash: NewContentHashSingle("sha-256:LCa0a2j_xo_5m0U8HTBBNBNCLXBkg7-g-YpeiGJm564"),
						Meta:        map[string]any{"size": 1024},
					},
				},
			},
			expectError: true,
			errorCount:  3,
			extensions:  []string{"dialog", "mediatype", "filename"},
		},
		{
			name: "multiple extension categories",
			vcon: &VCon{
				UUID:      newTestUUID(),
				Vcon:      "0.0.2",
				CreatedAt: time.Now(),
				// Extensions field removed for IETF draft-03 compliance
				Parties: []Party{
					{
						Name: StringPtr("Test User"),
						// Note: DID and Timezone fields removed for IETF draft-03 compliance
						Stir: StringPtr("stir-value"), // Using remaining extension field
					},
				},
				Dialog: []Dialog{
					{
						Type:        "recording",
						Start:       time.Now(),
						Parties:     NewDialogPartiesArrayPtr([]int{0}),
						URL:         StringPtr("https://example.com/recording.wav"),
						ContentHash: NewContentHashSingle("sha-256:LCa0a2j_xo_5m0U8HTBBNBNCLXBkg7-g-YpeiGJm564"),
						Campaign:    StringPtr("sales-2024"),
						Skill:       StringPtr("support"),
						SessionID:   StringPtr("session-123"),
					},
				},
			},
			expectError: true,
			errorCount:  4,                                                   // Changed: was 5, now 4 since DID/Timezone removed
			extensions:  []string{"stir", "campaign", "skill", "session_id"}, // Removed "did", "timezone"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.vcon.ValidateIETFStrict()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}

				var strictErrors IETFStrictValidationErrors
				if !errors.As(err, &strictErrors) {
					t.Errorf("Expected IETFStrictValidationErrors but got %T: %v", err, err)
					return
				}

				if len(strictErrors) != tt.errorCount {
					t.Errorf("Expected %d errors but got %d", tt.errorCount, len(strictErrors))
					for i, e := range strictErrors {
						t.Logf("Error %d: %s", i, e.Error())
					}
				}

				// Check that all expected extensions are flagged
				foundExtensions := make(map[string]bool)
				for _, e := range strictErrors {
					foundExtensions[e.Extension] = true
				}

				for _, expectedExt := range tt.extensions {
					if !foundExtensions[expectedExt] {
						t.Errorf("Expected extension '%s' to be flagged but it wasn't", expectedExt)
					}
				}
			} else if err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestValidateIETFStrictCategories(t *testing.T) {
	tests := []struct {
		name     string
		vcon     *VCon
		category string
		field    string
	}{
		{
			name: "telephony category",
			vcon: &VCon{
				UUID:      newTestUUID(),
				Vcon:      "0.0.2",
				CreatedAt: time.Now(),
				Parties: []Party{
					{
						Name: StringPtr("Test User"),
						Stir: StringPtr("stir-value"),
					},
				},
			},
			category: "telephony",
			field:    "stir",
		},
		{
			name: "video_media category",
			vcon: &VCon{
				UUID:      newTestUUID(),
				Vcon:      "0.0.2",
				CreatedAt: time.Now(),
				Parties: []Party{
					{Name: StringPtr("Test User")},
				},
				Dialog: []Dialog{
					{
						Type:        "recording",
						Start:       time.Now(),
						Parties:     NewDialogPartiesArrayPtr([]int{0}),
						URL:         StringPtr("https://example.com/recording.wav"),
						ContentHash: NewContentHashSingle("sha-256:LCa0a2j_xo_5m0U8HTBBNBNCLXBkg7-g-YpeiGJm564"),
						Resolution:  StringPtr("1920x1080"),
					},
				},
			},
			category: "video_media",
			field:    "resolution",
		},
		{
			name: "contact_center category",
			vcon: &VCon{
				UUID:      newTestUUID(),
				Vcon:      "0.0.2",
				CreatedAt: time.Now(),
				Parties: []Party{
					{Name: StringPtr("Test User")},
				},
				Dialog: []Dialog{
					{
						Type:        "recording",
						Start:       time.Now(),
						Parties:     NewDialogPartiesArrayPtr([]int{0}),
						URL:         StringPtr("https://example.com/recording.wav"),
						ContentHash: NewContentHashSingle("sha-256:LCa0a2j_xo_5m0U8HTBBNBNCLXBkg7-g-YpeiGJm564"),
						Campaign:    StringPtr("sales-2024"),
					},
				},
			},
			category: "contact_center",
			field:    "campaign",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.vcon.ValidateIETFStrict()
			if err == nil {
				t.Errorf("Expected error but got none")
				return
			}

			var strictErrors IETFStrictValidationErrors
			if !errors.As(err, &strictErrors) {
				t.Errorf("Expected IETFStrictValidationErrors but got %T", err)
				return
			}

			found := false
			for _, e := range strictErrors {
				if e.Extension == tt.field && e.Category == tt.category {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Expected to find extension '%s' with category '%s'", tt.field, tt.category)
				for _, e := range strictErrors {
					t.Logf("Found: extension='%s', category='%s'", e.Extension, e.Category)
				}
			}
		})
	}
}

func TestValidateIETFStrictErrorMessages(t *testing.T) {
	vcon := &VCon{
		UUID:      newTestUUID(),
		Vcon:      "0.0.2",
		CreatedAt: time.Now(),
		// Extensions field removed for IETF draft-03 compliance
		Parties: []Party{
			{
				Name: StringPtr("Test User"),
				Stir: StringPtr("stir-value"), // Using remaining extension field instead of SIP
			},
		},
	}

	err := vcon.ValidateIETFStrict()
	if err == nil {
		t.Errorf("Expected error but got none")
		return
	}

	errorMsg := err.Error()

	// Check that error message contains expected information
	expectedPhrases := []string{
		"IETF strict validation failed",
		"extension field violations",
		"IETF strict validation error",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(errorMsg, phrase) {
			t.Errorf("Expected error message to contain '%s', but got: %s", phrase, errorMsg)
		}
	}
}

func TestIsZeroValue(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected bool
	}{
		{"nil pointer", (*string)(nil), true},
		{"non-nil string pointer", StringPtr("test"), false},
		{"empty string", "", true},
		{"non-empty string", "test", false},
		{"zero int", 0, true},
		{"non-zero int", 42, false},
		{"nil slice", ([]string)(nil), true},
		{"empty slice", []string{}, false},
		{"non-empty slice", []string{"test"}, false},
		{"nil map", (map[string]any)(nil), true},
		{"empty map", map[string]any{}, false},
		{"non-empty map", map[string]any{"key": "value"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := reflect.ValueOf(tt.value)
			result := isZeroValue(v)
			if result != tt.expected {
				t.Errorf("isZeroValue(%v) = %v, expected %v", tt.value, result, tt.expected)
			}
		})
	}
}

// Helper functions for test data

func newTestUUID() uuid.UUID {
	return uuid.New()
}
