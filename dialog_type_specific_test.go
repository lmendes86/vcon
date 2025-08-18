package vcon

import (
	"testing"
	"time"
)

func TestDialogTypeSpecificValidation(t *testing.T) {
	tests := []struct {
		name          string
		dialog        Dialog
		expectError   bool
		errorContains string
	}{
		{
			name: "recording dialog with inline content requires mediatype",
			dialog: Dialog{
				Type:    "recording",
				Start:   time.Now(),
				Parties: NewDialogPartiesArrayPtr([]int{0}),
				Body:    "inline audio data",
				// Missing mediatype - should fail IETF validation
			},
			expectError:   true,
			errorContains: "recording dialog with inline content requires mediatype",
		},
		{
			name: "text dialog with inline content requires mediatype",
			dialog: Dialog{
				Type:    "text",
				Start:   time.Now(),
				Parties: NewDialogPartiesArrayPtr([]int{0}),
				Body:    "Hello world",
				// Missing mediatype - should fail IETF validation
			},
			expectError:   true,
			errorContains: "text dialog with inline content requires mediatype",
		},
		{
			name: "recording dialog with external URL - no mediatype required",
			dialog: Dialog{
				Type:        "recording",
				Start:       time.Now(),
				Parties:     NewDialogPartiesArrayPtr([]int{0}),
				URL:         StringPtr("https://example.com/audio.wav"),
				ContentHash: NewContentHashSingle("sha-256:LCa0a2j_xo_5m0U8HTBBNBNCLXBkg7-g-YpeiGJm564"),
				// No mediatype required for external content, but content_hash is required
			},
			expectError: false,
		},
		{
			name: "transfer dialog cannot have parties",
			dialog: Dialog{
				Type:           "transfer",
				Start:          time.Now(),
				Parties:        NewDialogPartiesArrayPtr([]int{0}), // Should be forbidden
				Original:       IntPtr(0),                          // Dialog index
				TargetDialog:   IntPtr(1),                          // Dialog index
				Transferee:     IntPtr(1),
				Transferor:     IntPtr(2),
				TransferTarget: IntPtr(3),
			},
			expectError:   true,
			errorContains: "transfer dialog must not have parties",
		},
		{
			name: "transfer dialog cannot have originator",
			dialog: Dialog{
				Type:           "transfer",
				Start:          time.Now(),
				Originator:     IntPtr(0), // Should be forbidden
				Original:       IntPtr(0), // Dialog index
				TargetDialog:   IntPtr(1), // Dialog index
				Transferee:     IntPtr(1),
				Transferor:     IntPtr(2),
				TransferTarget: IntPtr(3),
			},
			expectError:   true,
			errorContains: "transfer dialog must not have originator",
		},
		{
			name: "transfer dialog cannot have mediatype",
			dialog: Dialog{
				Type:           "transfer",
				Start:          time.Now(),
				Mediatype:      StringPtr("audio/wav"), // Should be forbidden
				Original:       IntPtr(0),              // Dialog index
				TargetDialog:   IntPtr(1),              // Dialog index
				Transferee:     IntPtr(1),
				Transferor:     IntPtr(2),
				TransferTarget: IntPtr(3),
			},
			expectError:   true,
			errorContains: "transfer dialog must not have mediatype",
		},
		{
			name: "transfer dialog cannot have filename",
			dialog: Dialog{
				Type:           "transfer",
				Start:          time.Now(),
				Filename:       StringPtr("audio.wav"), // Should be forbidden
				Original:       IntPtr(0),              // Dialog index
				TargetDialog:   IntPtr(1),              // Dialog index
				Transferee:     IntPtr(1),
				Transferor:     IntPtr(2),
				TransferTarget: IntPtr(3),
			},
			expectError:   true,
			errorContains: "transfer dialog must not have filename",
		},
		{
			name: "transfer dialog must have original",
			dialog: Dialog{
				Type:  "transfer",
				Start: time.Now(),
				// Missing Original - should fail
				TargetDialog:   IntPtr(1), // Dialog index
				Transferee:     IntPtr(1),
				Transferor:     IntPtr(2),
				TransferTarget: IntPtr(3),
			},
			expectError:   true,
			errorContains: "transfer dialog must have original dialog reference",
		},
		{
			name: "transfer dialog must have target_dialog",
			dialog: Dialog{
				Type:     "transfer",
				Start:    time.Now(),
				Original: IntPtr(0), // Dialog index
				// Missing TargetDialog - should fail
				Transferee:     IntPtr(1),
				Transferor:     IntPtr(2),
				TransferTarget: IntPtr(3),
			},
			expectError:   true,
			errorContains: "transfer dialog must have target-dialog reference",
		},
		{
			name: "incomplete dialog cannot have mediatype",
			dialog: Dialog{
				Type:        "incomplete",
				Start:       time.Now(),
				Parties:     NewDialogPartiesArrayPtr([]int{0}),
				Disposition: StringPtr("busy"),
				Mediatype:   StringPtr("audio/wav"), // Should be forbidden
			},
			expectError:   true,
			errorContains: "incomplete dialog must not have mediatype",
		},
		{
			name: "incomplete dialog cannot have filename",
			dialog: Dialog{
				Type:        "incomplete",
				Start:       time.Now(),
				Parties:     NewDialogPartiesArrayPtr([]int{0}),
				Disposition: StringPtr("busy"),
				Filename:    StringPtr("audio.wav"), // Should be forbidden
			},
			expectError:   true,
			errorContains: "incomplete dialog must not have filename",
		},
		{
			name: "valid transfer dialog with all required fields",
			dialog: Dialog{
				Type:           "transfer",
				Start:          time.Now(),
				Original:       IntPtr(0), // References first dummy dialog
				TargetDialog:   IntPtr(1), // References second dummy dialog
				Transferee:     IntPtr(1),
				Transferor:     IntPtr(2),
				TransferTarget: IntPtr(3),
			},
			expectError: false,
		},
		{
			name: "valid incomplete dialog",
			dialog: Dialog{
				Type:        "incomplete",
				Start:       time.Now(),
				Parties:     NewDialogPartiesArrayPtr([]int{0}),
				Disposition: StringPtr("busy"),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vcon := NewWithDefaults()
			vcon.AddParty(Party{Name: StringPtr("Test User")})
			vcon.AddParty(Party{Name: StringPtr("Test User 2")})
			vcon.AddParty(Party{Name: StringPtr("Test User 3")})
			vcon.AddParty(Party{Name: StringPtr("Test User 4")})

			// For transfer dialog tests, we need multiple dialogs to reference
			if tt.dialog.Type == "transfer" {
				// Create a dummy dialog at index 0 for reference
				dummyDialog1 := Dialog{
					Type:      "text",
					Start:     time.Now().Add(-1 * time.Hour), // Earlier time
					Parties:   NewDialogPartiesArrayPtr([]int{0}),
					Body:      "First dialog",
					Encoding:  StringPtr("none"),
					Mediatype: StringPtr("text/plain"),
				}
				// Create a dummy dialog at index 1 for reference
				dummyDialog2 := Dialog{
					Type:      "text",
					Start:     time.Now().Add(-30 * time.Minute), // Earlier time
					Parties:   NewDialogPartiesArrayPtr([]int{1}),
					Body:      "Second dialog",
					Encoding:  StringPtr("none"),
					Mediatype: StringPtr("text/plain"),
				}
				vcon.Dialog = []Dialog{dummyDialog1, dummyDialog2, tt.dialog}
			} else {
				vcon.Dialog = []Dialog{tt.dialog}
			}

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
