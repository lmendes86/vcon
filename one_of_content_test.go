package vcon

import (
	"testing"
	"time"
)

func TestOneOfContentValidation(t *testing.T) {
	tests := []struct {
		name          string
		setupVCon     func() *VCon
		expectError   bool
		errorContains string
	}{
		{
			name: "valid dialog with inline content",
			setupVCon: func() *VCon {
				vcon := NewWithDefaults()
				vcon.AddParty(Party{Name: StringPtr("Test User")})
				vcon.AddDialog(Dialog{
					Type:      "recording",
					Start:     time.Now(),
					Parties:   NewDialogPartiesArrayPtr([]int{0}),
					Body:      "audio data",
					Encoding:  StringPtr("base64url"),
					Mediatype: StringPtr("audio/wav"),
				})
				return vcon
			},
			expectError: false,
		},
		{
			name: "valid dialog with external content",
			setupVCon: func() *VCon {
				vcon := NewWithDefaults()
				vcon.AddParty(Party{Name: StringPtr("Test User")})
				vcon.AddDialog(Dialog{
					Type:        "recording",
					Start:       time.Now(),
					Parties:     NewDialogPartiesArrayPtr([]int{0}),
					URL:         StringPtr("https://example.com/audio.wav"),
					ContentHash: NewContentHashSingle("sha-256:LCa0a2j_xo_5m0U8HTBBNBNCLXBkg7-g-YpeiGJm564"),
				})
				return vcon
			},
			expectError: false,
		},
		{
			name: "invalid dialog with both inline and external content",
			setupVCon: func() *VCon {
				vcon := NewWithDefaults()
				vcon.AddParty(Party{Name: StringPtr("Test User")})
				vcon.AddDialog(Dialog{
					Type:        "recording",
					Start:       time.Now(),
					Parties:     NewDialogPartiesArrayPtr([]int{0}),
					Body:        "audio data",
					Encoding:    StringPtr("base64url"),
					Mediatype:   StringPtr("audio/wav"),
					URL:         StringPtr("https://example.com/audio.wav"),
					ContentHash: NewContentHashSingle("sha-256:LCa0a2j_xo_5m0U8HTBBNBNCLXBkg7-g-YpeiGJm564"),
				})
				return vcon
			},
			expectError:   true,
			errorContains: "dialog cannot have both inline content (body) and external content (url)",
		},
		{
			name: "invalid dialog with inline content missing encoding",
			setupVCon: func() *VCon {
				vcon := NewWithDefaults()
				vcon.AddParty(Party{Name: StringPtr("Test User")})
				vcon.AddDialog(Dialog{
					Type:      "recording",
					Start:     time.Now(),
					Parties:   NewDialogPartiesArrayPtr([]int{0}),
					Body:      "audio data",
					Mediatype: StringPtr("audio/wav"),
					// Missing encoding - should fail
				})
				return vcon
			},
			expectError:   true,
			errorContains: "encoding is required when body is present",
		},
		{
			name: "valid analysis with inline content",
			setupVCon: func() *VCon {
				vcon := NewWithDefaults()
				vcon.AddParty(Party{Name: StringPtr("Test User")})
				vcon.Analysis = []Analysis{
					{
						Type:     "sentiment",
						Vendor:   "test-vendor",
						Body:     "positive",
						Encoding: StringPtr("json"),
					},
				}
				return vcon
			},
			expectError: false,
		},
		{
			name: "valid analysis with external content",
			setupVCon: func() *VCon {
				vcon := NewWithDefaults()
				vcon.AddParty(Party{Name: StringPtr("Test User")})
				vcon.Analysis = []Analysis{
					{
						Type:        "sentiment",
						Vendor:      "test-vendor",
						URL:         StringPtr("https://example.com/analysis.json"),
						ContentHash: NewContentHashSingle("sha-256:LCa0a2j_xo_5m0U8HTBBNBNCLXBkg7-g-YpeiGJm564"),
					},
				}
				return vcon
			},
			expectError: false,
		},
		{
			name: "invalid analysis with both inline and external content",
			setupVCon: func() *VCon {
				vcon := NewWithDefaults()
				vcon.AddParty(Party{Name: StringPtr("Test User")})
				vcon.Analysis = []Analysis{
					{
						Type:        "sentiment",
						Vendor:      "test-vendor",
						Body:        "positive",
						Encoding:    StringPtr("json"),
						URL:         StringPtr("https://example.com/analysis.json"),
						ContentHash: NewContentHashSingle("sha-256:LCa0a2j_xo_5m0U8HTBBNBNCLXBkg7-g-YpeiGJm564"),
					},
				}
				return vcon
			},
			expectError:   true,
			errorContains: "analysis cannot have both inline content (body) and external content (url)",
		},
		{
			name: "invalid analysis with no content",
			setupVCon: func() *VCon {
				vcon := NewWithDefaults()
				vcon.AddParty(Party{Name: StringPtr("Test User")})
				vcon.Analysis = []Analysis{
					{
						Type:   "sentiment",
						Vendor: "test-vendor",
						// No body and no URL - should fail
					},
				}
				return vcon
			},
			expectError:   true,
			errorContains: "analysis must have either body+encoding or url+content_hash",
		},
		{
			name: "valid attachment with inline content",
			setupVCon: func() *VCon {
				vcon := NewWithDefaults()
				vcon.AddParty(Party{Name: StringPtr("Test User")})
				vcon.AddAttachment(Attachment{
					Type:      StringPtr("document"),
					Body:      "document content",
					Encoding:  StringPtr("none"),
					Mediatype: StringPtr("text/plain"), // Added required mediatype
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
			name: "invalid attachment with both inline and external content",
			setupVCon: func() *VCon {
				vcon := NewWithDefaults()
				vcon.AddParty(Party{Name: StringPtr("Test User")})
				vcon.AddAttachment(Attachment{
					Type:        StringPtr("document"),
					Body:        "document content",
					Encoding:    StringPtr("none"),
					Mediatype:   StringPtr("text/plain"), // Added required mediatype
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
			name: "invalid encoding values",
			setupVCon: func() *VCon {
				vcon := NewWithDefaults()
				vcon.AddParty(Party{Name: StringPtr("Test User")})
				vcon.AddDialog(Dialog{
					Type:      "recording",
					Start:     time.Now(),
					Parties:   NewDialogPartiesArrayPtr([]int{0}),
					Body:      "audio data",
					Encoding:  StringPtr("invalid-encoding"), // Invalid encoding
					Mediatype: StringPtr("audio/wav"),
				})
				return vcon
			},
			expectError:   true,
			errorContains: "invalid encoding: invalid-encoding",
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
