package vcon

import (
	"testing"
	"time"
)

func TestAttachmentContentRules(t *testing.T) {
	tests := []struct {
		name          string
		setupVCon     func() *VCon
		expectError   bool
		errorContains string
	}{
		{
			name: "valid attachment with inline content and mediatype",
			setupVCon: func() *VCon {
				vcon := NewWithDefaults()
				vcon.AddParty(Party{Name: StringPtr("Test User")})
				vcon.AddAttachment(Attachment{
					Type:      StringPtr("document"),
					Body:      "document content",
					Encoding:  StringPtr("none"),
					Mediatype: StringPtr("text/plain"),
					Start:     func() *time.Time { t := time.Now().UTC(); return &t }(),
					Party:     IntPtr(0),
				})
				return vcon
			},
			expectError: false,
		},
		{
			name: "valid attachment with external content",
			setupVCon: func() *VCon {
				vcon := NewWithDefaults()
				vcon.AddParty(Party{Name: StringPtr("Test User")})
				vcon.AddAttachment(Attachment{
					Type:        StringPtr("document"),
					URL:         StringPtr("https://example.com/document.pdf"),
					ContentHash: NewContentHashSingle("sha-256:LCa0a2j_xo_5m0U8HTBBNBNCLXBkg7-g-YpeiGJm564"),
					Start:       func() *time.Time { t := time.Now().UTC(); return &t }(),
					Party:       IntPtr(0),
				})
				return vcon
			},
			expectError: false,
		},
		{
			name: "invalid attachment with inline content missing mediatype",
			setupVCon: func() *VCon {
				vcon := NewWithDefaults()
				vcon.AddParty(Party{Name: StringPtr("Test User")})
				vcon.AddAttachment(Attachment{
					Type:     StringPtr("document"),
					Body:     "document content",
					Encoding: StringPtr("none"),
					// Missing mediatype - should fail
					Start: func() *time.Time { t := time.Now().UTC(); return &t }(),
					Party: IntPtr(0),
				})
				return vcon
			},
			expectError:   true,
			errorContains: "mediatype is required for inline attachments",
		},
		{
			name: "invalid attachment with both inline and external content",
			setupVCon: func() *VCon {
				vcon := NewWithDefaults()
				vcon.AddParty(Party{Name: StringPtr("Test User")})
				vcon.AddAttachment(Attachment{
					Type:        StringPtr("document"),
					Body:        "document content",
					Encoding:    StringPtr("none"),
					Mediatype:   StringPtr("text/plain"),
					URL:         StringPtr("https://example.com/document.pdf"),
					ContentHash: NewContentHashSingle("sha-256:LCa0a2j_xo_5m0U8HTBBNBNCLXBkg7-g-YpeiGJm564"),
					Start:       func() *time.Time { t := time.Now().UTC(); return &t }(),
					Party:       IntPtr(0),
				})
				return vcon
			},
			expectError:   true,
			errorContains: "attachment cannot have both inline content (body) and external content (url)",
		},
		{
			name: "invalid attachment with no content in normal scenario",
			setupVCon: func() *VCon {
				vcon := NewWithDefaults()
				vcon.AddParty(Party{Name: StringPtr("Test User")})
				vcon.AddAttachment(Attachment{
					Type:  StringPtr("document"),
					Start: func() *time.Time { t := time.Now().UTC(); return &t }(),
					Party: IntPtr(0),
					// No body and no URL - should fail in normal scenario
				})
				return vcon
			},
			expectError:   true,
			errorContains: "attachment must have either body or url",
		},
		{
			name: "valid redacted attachment without content",
			setupVCon: func() *VCon {
				vcon := NewWithDefaults()
				vcon.AddParty(Party{Name: StringPtr("Test User")})
				// Set redacted object to enable redacted scenario
				vcon.Redacted = &RedactedObject{
					UUID: "test-uuid",
					Type: StringPtr("redacted"),
				}
				vcon.AddAttachment(Attachment{
					Type:  StringPtr("document"),
					Start: func() *time.Time { t := time.Now().UTC(); return &t }(),
					Party: IntPtr(0),
					// No content in redacted scenario - should be valid
				})
				return vcon
			},
			expectError: false,
		},
		{
			name: "invalid redacted attachment without type",
			setupVCon: func() *VCon {
				vcon := NewWithDefaults()
				vcon.AddParty(Party{Name: StringPtr("Test User")})
				// Set redacted object to enable redacted scenario
				vcon.Redacted = &RedactedObject{
					UUID: "test-uuid",
					Type: StringPtr("redacted"),
				}
				vcon.AddAttachment(Attachment{
					// Missing type in redacted scenario - should fail
					Start: func() *time.Time { t := time.Now().UTC(); return &t }(),
					Party: IntPtr(0),
				})
				return vcon
			},
			expectError:   true,
			errorContains: "redacted attachment must have type field",
		},
		{
			name: "invalid attachment with invalid mediatype",
			setupVCon: func() *VCon {
				vcon := NewWithDefaults()
				vcon.AddParty(Party{Name: StringPtr("Test User")})
				vcon.AddAttachment(Attachment{
					Type:      StringPtr("document"),
					Body:      "document content",
					Encoding:  StringPtr("none"),
					Mediatype: StringPtr("invalid-media-type"), // Invalid mediatype
					Start:     func() *time.Time { t := time.Now().UTC(); return &t }(),
					Party:     IntPtr(0),
				})
				return vcon
			},
			expectError:   true,
			errorContains: "invalid mediatype: invalid-media-type",
		},
		{
			name: "invalid attachment with invalid encoding",
			setupVCon: func() *VCon {
				vcon := NewWithDefaults()
				vcon.AddParty(Party{Name: StringPtr("Test User")})
				vcon.AddAttachment(Attachment{
					Type:      StringPtr("document"),
					Body:      "document content",
					Encoding:  StringPtr("invalid-encoding"), // Invalid encoding
					Mediatype: StringPtr("text/plain"),
					Start:     func() *time.Time { t := time.Now().UTC(); return &t }(),
					Party:     IntPtr(0),
				})
				return vcon
			},
			expectError:   true,
			errorContains: "invalid encoding: invalid-encoding",
		},
		{
			name: "valid attachment with base64url encoding",
			setupVCon: func() *VCon {
				vcon := NewWithDefaults()
				vcon.AddParty(Party{Name: StringPtr("Test User")})
				vcon.AddAttachment(Attachment{
					Type:      StringPtr("document"),
					Body:      "document content",
					Encoding:  StringPtr("base64url"),
					Mediatype: StringPtr("application/pdf"),
					Start:     func() *time.Time { t := time.Now().UTC(); return &t }(),
					Party:     IntPtr(0),
				})
				return vcon
			},
			expectError: false,
		},
		{
			name: "valid attachment with json encoding",
			setupVCon: func() *VCon {
				vcon := NewWithDefaults()
				vcon.AddParty(Party{Name: StringPtr("Test User")})
				vcon.AddAttachment(Attachment{
					Type:      StringPtr("metadata"),
					Body:      "{\"key\": \"value\"}",
					Encoding:  StringPtr("json"),
					Mediatype: StringPtr("application/json"),
					Start:     func() *time.Time { t := time.Now().UTC(); return &t }(),
					Party:     IntPtr(0),
				})
				return vcon
			},
			expectError: false,
		},
		{
			name: "invalid attachment with inline content missing encoding",
			setupVCon: func() *VCon {
				vcon := NewWithDefaults()
				vcon.AddParty(Party{Name: StringPtr("Test User")})
				vcon.AddAttachment(Attachment{
					Type:      StringPtr("document"),
					Body:      "document content",
					Mediatype: StringPtr("text/plain"),
					// Missing encoding - should fail
					Start: func() *time.Time { t := time.Now().UTC(); return &t }(),
					Party: IntPtr(0),
				})
				return vcon
			},
			expectError:   true,
			errorContains: "encoding is required when body is present",
		},
		{
			name: "valid redacted attachment with type and metadata only",
			setupVCon: func() *VCon {
				vcon := NewWithDefaults()
				vcon.AddParty(Party{Name: StringPtr("Test User")})
				// Set redacted object to enable redacted scenario
				vcon.Redacted = &RedactedObject{
					UUID: "test-uuid",
					Type: StringPtr("redacted"),
				}
				vcon.AddAttachment(Attachment{
					Type:     StringPtr("document"),
					Filename: StringPtr("redacted-document.pdf"),
					Start:    func() *time.Time { t := time.Now().UTC(); return &t }(),
					Party:    IntPtr(0),
					// Content is redacted, but type and filename are preserved
				})
				return vcon
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vcon := tt.setupVCon()
			err := vcon.ValidateIETF()

			if tt.expectError {
				if err == nil {
					t.Error("Expected validation error but got none")
				} else if tt.errorContains != "" {
					if !containsString(err.Error(), tt.errorContains) {
						t.Errorf("Expected error containing '%s', got: %v", tt.errorContains, err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation errors but got: %v", err)
				}
			}
		})
	}
}
