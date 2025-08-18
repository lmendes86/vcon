package vcon

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
)

// Helper function to create UUID values for tests.
func newUUID() uuid.UUID {
	return uuid.New()
}

// Tests from migration_helper_functions_test.go

func TestConvertToLegacyAnalysis(t *testing.T) {
	tests := []struct {
		name     string
		analyses []Analysis
		expected int // Expected number of legacy analyses
	}{
		{
			name:     "empty analyses",
			analyses: []Analysis{},
			expected: 0,
		},
		{
			name: "single analysis",
			analyses: []Analysis{
				{
					Vendor: "test-vendor",
					Type:   "summary",
					Body:   StringPtr("Test analysis content"),
				},
			},
			expected: 1,
		},
		{
			name: "multiple analyses",
			analyses: []Analysis{
				{
					Vendor: "vendor1",
					Type:   "summary",
					Body:   StringPtr("Analysis 1"),
				},
				{
					Vendor: "vendor2",
					Type:   "transcript",
					Body:   StringPtr("Analysis 2"),
				},
			},
			expected: 2,
		},
		{
			name: "analysis with all fields",
			analyses: []Analysis{
				{
					Vendor:      "comprehensive-vendor",
					Type:        "full-analysis",
					Dialog:      IntPtr(0),
					Product:     StringPtr("analyzer-v2"),
					Schema:      StringPtr("https://example.com/schema"),
					Mediatype:   StringPtr("application/json"),
					Filename:    StringPtr("analysis.json"),
					Body:        StringPtr(`{"result": "comprehensive analysis"}`),
					Encoding:    StringPtr("json"),
					URL:         StringPtr("https://example.com/analysis"),
					ContentHash: NewContentHashSingle("sha-256:abc123"),
					Meta:        map[string]any{"quality": "high"},
				},
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertToLegacyAnalysis(tt.analyses)

			if len(result) != tt.expected {
				t.Errorf("ConvertToLegacyAnalysis() returned %d analyses, expected %d",
					len(result), tt.expected)
			}

			// Verify that each result is a map
			for i, legacyAnalysis := range result {
				if legacyAnalysis == nil {
					t.Errorf("Legacy analysis %d is nil", i)
					continue
				}

				// Check that vendor field is preserved
				if i < len(tt.analyses) {
					if vendor, ok := legacyAnalysis["vendor"]; !ok || vendor != tt.analyses[i].Vendor {
						t.Errorf("Legacy analysis %d vendor = %v, expected %v",
							i, vendor, tt.analyses[i].Vendor)
					}
				}
			}
		})
	}
}

func TestTimePtr(t *testing.T) {
	tests := []struct {
		name string
		time time.Time
	}{
		{
			name: "current time",
			time: time.Now(),
		},
		{
			name: "zero time",
			time: time.Time{},
		},
		{
			name: "specific time",
			time: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		},
		{
			name: "unix epoch",
			time: time.Unix(0, 0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TimePtr(tt.time)

			if result == nil {
				t.Error("TimePtr() returned nil")
				return
			}

			if !result.Equal(tt.time) {
				t.Errorf("TimePtr() = %v, expected %v", *result, tt.time)
			}

			// Verify it's actually a pointer to a different memory location
			if result == &tt.time {
				t.Error("TimePtr() returned pointer to original time instead of copy")
			}
		})
	}
}

func TestTimePtrIntegration(t *testing.T) {
	// Test that TimePtr works in realistic scenarios
	now := time.Now()
	ptr := TimePtr(now)

	// Modify original time - pointer should be unaffected
	now = now.Add(time.Hour)

	if ptr.Equal(now) {
		t.Error("TimePtr() result was affected by modification of original time")
	}
}

func TestConvertToLegacyAnalysisWithDialogReference(t *testing.T) {
	// Test with dialog references
	analyses := []Analysis{
		{
			Vendor: "test-vendor",
			Type:   "dialog-analysis",
			Dialog: IntPtr(0),
			Body:   StringPtr("Dialog analysis"),
		},
		{
			Vendor: "test-vendor-2",
			Type:   "single-dialog",
			Dialog: IntPtr(5),
			Body:   StringPtr("Single dialog analysis"),
		},
	}

	result := ConvertToLegacyAnalysis(analyses)

	if len(result) != 2 {
		t.Fatalf("Expected 2 legacy analyses, got %d", len(result))
	}

	// Verify dialog references are properly converted
	if dialog, exists := result[0]["dialog"]; exists {
		if dialog == nil {
			t.Error("First analysis dialog reference is nil")
		}
	}

	if dialog, exists := result[1]["dialog"]; exists {
		if dialog == nil {
			t.Error("Second analysis dialog reference is nil")
		}
	}
}

func TestConvertToLegacyAnalysisEdgeCases(t *testing.T) {
	// Test edge cases
	tests := []struct {
		name     string
		analyses []Analysis
		wantLen  int
	}{
		{
			name:     "nil slice",
			analyses: nil,
			wantLen:  0,
		},
		{
			name: "analysis with only vendor",
			analyses: []Analysis{
				{Vendor: "minimal-vendor"},
			},
			wantLen: 1,
		},
		{
			name: "analysis with empty vendor",
			analyses: []Analysis{
				{Vendor: ""},
			},
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertToLegacyAnalysis(tt.analyses)
			if len(result) != tt.wantLen {
				t.Errorf("ConvertToLegacyAnalysis() len = %d, want %d", len(result), tt.wantLen)
			}
		})
	}
}

// Tests from migration_additional_test.go

func TestMigrateTopLevelObjects(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *VCon
		mode    MigrationMode
		wantErr bool
	}{
		{
			name: "normal vcon with proper structs",
			setup: func() *VCon {
				vcon := NewWithDefaults()
				vcon.AddParty(Party{Name: StringPtr("Test User")})
				return vcon
			},
			mode:    MigrationModeStrict,
			wantErr: false,
		},
		{
			name: "vcon with legacy data",
			setup: func() *VCon {
				vcon := NewWithDefaults()
				vcon.AddParty(Party{Name: StringPtr("Legacy User")})
				return vcon
			},
			mode:    MigrationModeLenient,
			wantErr: false,
		},
		{
			name: "vcon with preserve mode",
			setup: func() *VCon {
				vcon := NewWithDefaults()
				vcon.AddParty(Party{Name: StringPtr("Preserve User")})
				return vcon
			},
			mode:    MigrationModePreserve,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vcon := tt.setup()
			err := vcon.MigrateTopLevelObjects(tt.mode)

			if tt.wantErr {
				if err == nil {
					t.Error("MigrateTopLevelObjects() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("MigrateTopLevelObjects() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestConvertFromLegacyAnalysisEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		legacy      []map[string]any
		mode        MigrationMode
		expectError bool
		expectCount int
	}{
		{
			name:        "empty legacy analysis",
			legacy:      []map[string]any{},
			mode:        MigrationModeStrict,
			expectError: false,
			expectCount: 0,
		},
		{
			name: "analysis with strict mode failures",
			legacy: []map[string]any{
				{
					// Missing required fields to trigger strict mode error
					"invalid": "data",
				},
			},
			mode:        MigrationModeStrict,
			expectError: true,
			expectCount: 0,
		},
		{
			name: "analysis with lenient mode - skip failures",
			legacy: []map[string]any{
				{
					// Missing required fields but lenient mode should skip
					"invalid": "data",
				},
				{
					// Valid analysis
					"type":     "valid",
					"vendor":   "test",
					"body":     "content",
					"encoding": "json",
				},
			},
			mode:        MigrationModeLenient,
			expectError: false,
			expectCount: 2, // Both are converted, invalid fields just ignored
		},
		{
			name: "mixed valid and invalid in preserve mode",
			legacy: []map[string]any{
				{
					"type":     "summary",
					"vendor":   "vendor1",
					"body":     "valid content",
					"encoding": "json",
					"extra":    "preserve this",
				},
			},
			mode:        MigrationModePreserve,
			expectError: false,
			expectCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertFromLegacyAnalysis(tt.legacy, tt.mode)

			if tt.expectError {
				if err == nil {
					t.Error("ConvertFromLegacyAnalysis() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ConvertFromLegacyAnalysis() unexpected error: %v", err)
				return
			}

			if len(result) != tt.expectCount {
				t.Errorf("ConvertFromLegacyAnalysis() expected %d results, got %d",
					tt.expectCount, len(result))
			}
		})
	}
}

func TestGetActualValueFromMigrationExtensions(t *testing.T) {
	// Test the getActualValue function from migration_extensions.go
	// Since it operates on reflect.Value, we need to use reflection here
	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{
			name:     "string input",
			input:    "test string",
			expected: "test string",
		},
		{
			name:     "int input",
			input:    42,
			expected: 42,
		},
		{
			name:     "pointer to string",
			input:    StringPtr("pointer test"),
			expected: "pointer test",
		},
		{
			name:     "pointer to int",
			input:    IntPtr(123),
			expected: 123,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert input to reflect.Value which is what getActualValue expects
			value := reflect.ValueOf(tt.input)
			result := getActualValue(value)

			// Use string comparison for simplicity in this test
			if result != tt.expected {
				t.Errorf("getActualValue() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// Tests from migration_top_level_objects_test.go

func TestConvertLegacyRedactedObject(t *testing.T) {
	tests := []struct {
		name      string
		legacy    map[string]any
		mode      MigrationMode
		expectErr bool
		expected  *RedactedObject
	}{
		{
			name:     "nil input",
			legacy:   nil,
			mode:     MigrationModeStrict,
			expected: nil,
		},
		{
			name: "valid redacted object",
			legacy: map[string]any{
				"uuid":         "550e8400-e29b-41d4-a716-446655440000",
				"type":         "privacy",
				"body":         "redacted content",
				"encoding":     "json",
				"url":          "https://example.com/redacted",
				"content_hash": "sha-256:abcdef123456",
			},
			mode: MigrationModeStrict,
			expected: &RedactedObject{
				UUID:        "550e8400-e29b-41d4-a716-446655440000",
				Type:        StringPtr("privacy"),
				Body:        StringPtr("redacted content"),
				Encoding:    StringPtr("json"),
				URL:         StringPtr("https://example.com/redacted"),
				ContentHash: NewContentHashSingle("sha-256:abcdef123456"),
			},
		},
		{
			name: "missing UUID - strict mode",
			legacy: map[string]any{
				"type": "privacy",
			},
			mode:      MigrationModeStrict,
			expectErr: true,
		},
		{
			name: "missing UUID - lenient mode",
			legacy: map[string]any{
				"type": "privacy",
			},
			mode: MigrationModeLenient,
			expected: &RedactedObject{
				UUID: "legacy-redacted-1",
				Type: StringPtr("privacy"),
			},
		},
		{
			name: "minimal valid object",
			legacy: map[string]any{
				"uuid": "550e8400-e29b-41d4-a716-446655440000",
			},
			mode: MigrationModeStrict,
			expected: &RedactedObject{
				UUID: "550e8400-e29b-41d4-a716-446655440000",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertLegacyRedactedObject(tt.legacy, tt.mode)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil && err != ErrNoMigrationNeeded {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Handle ErrNoMigrationNeeded as successful "no migration"
			if err == ErrNoMigrationNeeded {
				if tt.expected != nil {
					t.Errorf("Expected result but got ErrNoMigrationNeeded")
				}
				return
			}

			if (result == nil) != (tt.expected == nil) {
				t.Errorf("Expected nil=%v, got nil=%v", tt.expected == nil, result == nil)
				return
			}

			if tt.expected == nil {
				return
			}

			if result.UUID != tt.expected.UUID {
				t.Errorf("Expected UUID %s, got %s", tt.expected.UUID, result.UUID)
			}

			if !stringPtrEqual(result.Type, tt.expected.Type) {
				t.Errorf("Expected Type %v, got %v", ptrToStr(tt.expected.Type), ptrToStr(result.Type))
			}

			if !stringPtrEqual(result.Body, tt.expected.Body) {
				t.Errorf("Expected Body %v, got %v", ptrToStr(tt.expected.Body), ptrToStr(result.Body))
			}

			if !stringPtrEqual(result.Encoding, tt.expected.Encoding) {
				t.Errorf("Expected Encoding %v, got %v", ptrToStr(tt.expected.Encoding), ptrToStr(result.Encoding))
			}

			if !stringPtrEqual(result.URL, tt.expected.URL) {
				t.Errorf("Expected URL %v, got %v", ptrToStr(tt.expected.URL), ptrToStr(result.URL))
			}

			if !contentHashEqual(result.ContentHash, tt.expected.ContentHash) {
				t.Errorf("Expected ContentHash %v, got %v", contentHashToStr(tt.expected.ContentHash), contentHashToStr(result.ContentHash))
			}
		})
	}
}

func TestConvertLegacyAppendedObject(t *testing.T) {
	tests := []struct {
		name      string
		legacy    any
		mode      MigrationMode
		expectErr bool
		expected  *AppendedObject
	}{
		{
			name:     "nil input",
			legacy:   nil,
			mode:     MigrationModeStrict,
			expected: nil,
		},
		{
			name: "valid appended object - map format",
			legacy: map[string]any{
				"uuid":     "550e8400-e29b-41d4-a716-446655440001",
				"type":     "metadata",
				"body":     "appended content",
				"encoding": "base64url",
			},
			mode: MigrationModeStrict,
			expected: &AppendedObject{
				UUID:     "550e8400-e29b-41d4-a716-446655440001",
				Type:     StringPtr("metadata"),
				Body:     StringPtr("appended content"),
				Encoding: StringPtr("base64url"),
			},
		},
		{
			name: "valid appended object - array format",
			legacy: []any{
				map[string]any{
					"uuid": "550e8400-e29b-41d4-a716-446655440001",
					"type": "metadata",
				},
			},
			mode: MigrationModeStrict,
			expected: &AppendedObject{
				UUID: "550e8400-e29b-41d4-a716-446655440001",
				Type: StringPtr("metadata"),
			},
		},
		{
			name:     "empty array",
			legacy:   []any{},
			mode:     MigrationModeStrict,
			expected: nil,
		},
		{
			name:      "invalid array content - strict mode",
			legacy:    []any{"invalid"},
			mode:      MigrationModeStrict,
			expectErr: true,
		},
		{
			name:     "invalid array content - lenient mode",
			legacy:   []any{"invalid"},
			mode:     MigrationModeLenient,
			expected: nil,
		},
		{
			name:      "unsupported format - strict mode",
			legacy:    123,
			mode:      MigrationModeStrict,
			expectErr: true,
		},
		{
			name:     "unsupported format - lenient mode",
			legacy:   123,
			mode:     MigrationModeLenient,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertLegacyAppendedObject(tt.legacy, tt.mode)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil && err != ErrNoMigrationNeeded {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Handle ErrNoMigrationNeeded as successful "no migration"
			if err == ErrNoMigrationNeeded {
				if tt.expected != nil {
					t.Errorf("Expected result but got ErrNoMigrationNeeded")
				}
				return
			}

			if (result == nil) != (tt.expected == nil) {
				t.Errorf("Expected nil=%v, got nil=%v", tt.expected == nil, result == nil)
				return
			}

			if tt.expected == nil {
				return
			}

			if result.UUID != tt.expected.UUID {
				t.Errorf("Expected UUID %s, got %s", tt.expected.UUID, result.UUID)
			}

			if !stringPtrEqual(result.Type, tt.expected.Type) {
				t.Errorf("Expected Type %v, got %v", ptrToStr(tt.expected.Type), ptrToStr(result.Type))
			}
		})
	}
}

func TestConvertLegacyGroupObjects(t *testing.T) {
	tests := []struct {
		name      string
		legacy    []any
		mode      MigrationMode
		expectErr bool
		expected  []GroupObject
	}{
		{
			name:     "nil input",
			legacy:   nil,
			mode:     MigrationModeStrict,
			expected: nil,
		},
		{
			name:     "empty array",
			legacy:   []any{},
			mode:     MigrationModeStrict,
			expected: []GroupObject{},
		},
		{
			name: "valid group objects",
			legacy: []any{
				map[string]any{
					"uuid": "550e8400-e29b-41d4-a716-446655440002",
					"type": "conversation-thread",
				},
				map[string]any{
					"uuid": "550e8400-e29b-41d4-a716-446655440003",
					"type": "related-calls",
				},
			},
			mode: MigrationModeStrict,
			expected: []GroupObject{
				{
					UUID: "550e8400-e29b-41d4-a716-446655440002",
					Type: StringPtr("conversation-thread"),
				},
				{
					UUID: "550e8400-e29b-41d4-a716-446655440003",
					Type: StringPtr("related-calls"),
				},
			},
		},
		{
			name: "legacy UUID-only format",
			legacy: []any{
				"550e8400-e29b-41d4-a716-446655440002",
				"550e8400-e29b-41d4-a716-446655440003",
			},
			mode: MigrationModeStrict,
			expected: []GroupObject{
				{UUID: "550e8400-e29b-41d4-a716-446655440002"},
				{UUID: "550e8400-e29b-41d4-a716-446655440003"},
			},
		},
		{
			name: "mixed formats",
			legacy: []any{
				"550e8400-e29b-41d4-a716-446655440002",
				map[string]any{
					"uuid": "550e8400-e29b-41d4-a716-446655440003",
					"type": "related-calls",
				},
			},
			mode: MigrationModeStrict,
			expected: []GroupObject{
				{UUID: "550e8400-e29b-41d4-a716-446655440002"},
				{
					UUID: "550e8400-e29b-41d4-a716-446655440003",
					Type: StringPtr("related-calls"),
				},
			},
		},
		{
			name: "missing UUID - strict mode",
			legacy: []any{
				map[string]any{
					"type": "conversation-thread",
				},
			},
			mode:      MigrationModeStrict,
			expectErr: true,
		},
		{
			name: "missing UUID - lenient mode",
			legacy: []any{
				map[string]any{
					"type": "conversation-thread",
				},
			},
			mode: MigrationModeLenient,
			expected: []GroupObject{
				{
					UUID: "legacy-group-0",
					Type: StringPtr("conversation-thread"),
				},
			},
		},
		{
			name: "invalid format - strict mode",
			legacy: []any{
				123,
			},
			mode:      MigrationModeStrict,
			expectErr: true,
		},
		{
			name:     "invalid format - lenient mode",
			legacy:   []any{123},
			mode:     MigrationModeLenient,
			expected: []GroupObject{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertLegacyGroupObjects(tt.legacy, tt.mode)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil && err != ErrNoMigrationNeeded {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Handle ErrNoMigrationNeeded as successful "no migration"
			if err == ErrNoMigrationNeeded {
				if tt.expected != nil {
					t.Errorf("Expected result but got ErrNoMigrationNeeded")
				}
				return
			}

			if (result == nil) != (tt.expected == nil) {
				t.Errorf("Expected nil=%v, got nil=%v", tt.expected == nil, result == nil)
				return
			}

			if tt.expected == nil {
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d objects, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if result[i].UUID != expected.UUID {
					t.Errorf("Object %d: Expected UUID %s, got %s", i, expected.UUID, result[i].UUID)
				}

				if !stringPtrEqual(result[i].Type, expected.Type) {
					t.Errorf("Object %d: Expected Type %v, got %v", i, ptrToStr(expected.Type), ptrToStr(result[i].Type))
				}
			}
		})
	}
}

// Tests from migration_extensions_test.go

func TestMigrateExtensionsToMeta(t *testing.T) {
	tests := []struct {
		name               string
		vcon               *VCon
		expectedMigrations int
		expectedFields     []string
		expectedWarnings   int
	}{
		{
			name: "no extensions to migrate",
			vcon: &VCon{
				UUID:      newUUID(),
				Vcon:      "0.0.2",
				CreatedAt: time.Now(),
				Parties: []Party{
					{Name: StringPtr("Test User")},
				},
			},
			expectedMigrations: 0,
			expectedFields:     []string{},
			expectedWarnings:   0,
		},
		{
			name: "migrate top-level deprecated fields",
			vcon: &VCon{
				UUID:      newUUID(),
				Vcon:      "0.0.2",
				CreatedAt: time.Now(),
				Parties: []Party{
					{Name: StringPtr("Test User")},
				},
			},
			expectedMigrations: 0,          // Changed: was 2, now 0 since Extensions/MustSupport fields removed
			expectedFields:     []string{}, // Changed: was extensions/must_support, now empty
			expectedWarnings:   0,          // Changed: was 2, now 0 since no top-level migration warnings
		},
		{
			name: "migrate party extensions",
			vcon: &VCon{
				UUID:      newUUID(),
				Vcon:      "0.0.2",
				CreatedAt: time.Now(),
				Parties: []Party{
					{
						Name: StringPtr("Test User"),
						Stir: StringPtr("stir-value"),
						// Note: SIP, DID, and Timezone fields removed for IETF draft-03 compliance
					},
				},
			},
			expectedMigrations: 1,                           // Only Stir field remains as extension
			expectedFields:     []string{"parties[0].stir"}, // SIP, DID, Timezone removed from struct
			expectedWarnings:   0,
		},
		{
			name: "migrate dialog extensions",
			vcon: &VCon{
				UUID:      newUUID(),
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
						Campaign:    StringPtr("sales-2024"),
						Skill:       StringPtr("support"),
					},
				},
			},
			expectedMigrations: 5,
			expectedFields:     []string{"dialog[0].resolution", "dialog[0].frame_rate", "dialog[0].codec", "dialog[0].campaign", "dialog[0].skill"},
			expectedWarnings:   0,
		},
		{
			name: "migrate analysis extensions",
			vcon: &VCon{
				UUID:      newUUID(),
				Vcon:      "0.0.2",
				CreatedAt: time.Now(),
				Parties: []Party{
					{Name: StringPtr("Test User")},
				},
				Analysis: []Analysis{
					{
						Type:        "summary",
						Vendor:      "test-vendor",
						URL:         StringPtr("https://example.com/analysis.json"),
						ContentHash: NewContentHashSingle("sha-256:LCa0a2j_xo_5m0U8HTBBNBNCLXBkg7-g-YpeiGJm564"),
						Product:     StringPtr("analyzer-v1.0"),
						Schema:      StringPtr("https://example.com/schema"),
						Mediatype:   StringPtr("application/json"),
						Filename:    StringPtr("analysis.json"),
					},
				},
			},
			expectedMigrations: 4,
			expectedFields:     []string{"analysis[0].product", "analysis[0].schema", "analysis[0].mediatype", "analysis[0].filename"},
			expectedWarnings:   0,
		},
		{
			name: "migrate attachment extensions",
			vcon: &VCon{
				UUID:      newUUID(),
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
						URL:         StringPtr("https://example.com/document.pdf"),
						ContentHash: NewContentHashSingle("sha-256:LCa0a2j_xo_5m0U8HTBBNBNCLXBkg7-g-YpeiGJm564"),
						Dialog:      IntPtr(0),
						Mediatype:   StringPtr("application/pdf"),
						Filename:    StringPtr("document.pdf"),
					},
				},
			},
			expectedMigrations: 3,
			expectedFields:     []string{"attachments[0].dialog", "attachments[0].mediatype", "attachments[0].filename"},
			expectedWarnings:   1, // warning about dialog field
		},
		{
			name: "migrate multiple types of extensions",
			vcon: &VCon{
				UUID:      newUUID(),
				Vcon:      "0.0.2",
				CreatedAt: time.Now(),
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
						SessionID:   StringPtr("session-123"),
					},
				},
			},
			expectedMigrations: 3,                                                                         // Changed: was 4, now 3 since DID/Timezone removed
			expectedFields:     []string{"parties[0].stir", "dialog[0].campaign", "dialog[0].session_id"}, // Removed did/timezone fields
			expectedWarnings:   0,                                                                         // Changed: was 2, now 0 since no top-level field warnings
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			migratedVCon, result, err := tt.vcon.MigrateExtensionsToMeta()
			if err != nil {
				t.Errorf("MigrateExtensionsToMeta() error = %v", err)
				return
			}

			if result.MigrationsCount != tt.expectedMigrations {
				t.Errorf("Expected %d migrations, got %d", tt.expectedMigrations, result.MigrationsCount)
			}

			if len(result.MigratedFields) != len(tt.expectedFields) {
				t.Errorf("Expected %d migrated fields, got %d", len(tt.expectedFields), len(result.MigratedFields))
				t.Logf("Expected fields: %v", tt.expectedFields)
				t.Logf("Actual fields: %v", result.MigratedFields)
			}

			// Check that all expected fields were migrated
			foundFields := make(map[string]bool)
			for _, field := range result.MigratedFields {
				foundFields[field] = true
			}

			for _, expectedField := range tt.expectedFields {
				if !foundFields[expectedField] {
					t.Errorf("Expected field '%s' to be migrated but it wasn't", expectedField)
				}
			}

			if len(result.Warnings) != tt.expectedWarnings {
				t.Errorf("Expected %d warnings, got %d", tt.expectedWarnings, len(result.Warnings))
				t.Logf("Warnings: %v", result.Warnings)
			}

			// Verify that the migrated vCon passes IETF strict validation
			if result.MigrationsCount > 0 {
				if err := migratedVCon.ValidateIETFStrict(); err != nil {
					t.Errorf("Migrated vCon should pass IETF strict validation, but got: %v", err)
				}
			}

			// Verify that original vCon fails IETF strict validation if it had extensions
			if tt.expectedMigrations > 0 {
				if err := tt.vcon.ValidateIETFStrict(); err == nil {
					t.Errorf("Original vCon with extensions should fail IETF strict validation")
				}
			}
		})
	}
}

func TestMigratedFieldsAreCleared(t *testing.T) {
	// Test that extension fields are properly cleared after migration
	vcon := &VCon{
		UUID:      newUUID(),
		Vcon:      "0.0.2",
		CreatedAt: time.Now(),
		Parties: []Party{
			{
				Name: StringPtr("Test User"),
				Stir: StringPtr("stir-value"),
				// Note: DID and Timezone fields removed for IETF draft-03 compliance
			},
		},
		Dialog: []Dialog{
			{
				Type:        "recording",
				Start:       time.Now(),
				Parties:     NewDialogPartiesArrayPtr([]int{0}),
				URL:         StringPtr("https://example.com/recording.wav"),
				ContentHash: NewContentHashSingle("sha-256:LCa0a2j_xo_5m0U8HTBBNBNCLXBkg7-g-YpeiGJm564"),
				Resolution:  StringPtr("1920x1080"),
				Campaign:    StringPtr("sales-2024"),
			},
		},
	}

	migratedVCon, _, err := vcon.MigrateExtensionsToMeta()
	if err != nil {
		t.Fatalf("MigrateExtensionsToMeta() error = %v", err)
	}

	// Extensions and MustSupport fields have been removed from VCon struct
	// No need to check if they are cleared since they don't exist

	// Check that party extension fields are cleared
	party := migratedVCon.Parties[0]
	if party.Stir != nil && *party.Stir != "" {
		t.Error("Party.Stir should be cleared after migration")
	}
	// Note: DID and Timezone fields removed from Party struct for IETF draft-03 compliance
	// These fields are no longer part of the struct, so no need to check for clearing

	// Check that dialog extension fields are cleared
	dialog := migratedVCon.Dialog[0]
	if dialog.Resolution != nil && *dialog.Resolution != "" {
		t.Error("Dialog.Resolution should be cleared after migration")
	}
	if dialog.Campaign != nil && *dialog.Campaign != "" {
		t.Error("Dialog.Campaign should be cleared after migration")
	}
}

func TestMigratedDataInMeta(t *testing.T) {
	// Test that migrated data is properly stored in meta fields
	vcon := &VCon{
		UUID:      newUUID(),
		Vcon:      "0.0.2",
		CreatedAt: time.Now(),
		Parties: []Party{
			{
				Name: StringPtr("Test User"),
				// Note: DID and Timezone fields removed for IETF draft-03 compliance
				// Using Stir field instead for migration testing
				Stir: StringPtr("stir-value"),
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
			},
		},
	}

	migratedVCon, _, err := vcon.MigrateExtensionsToMeta()
	if err != nil {
		t.Fatalf("MigrateExtensionsToMeta() error = %v", err)
	}

	// Note: Top-level Meta field removed for IETF draft-03 compliance
	// Migration can no longer store deprecated extensions in top-level meta
	// Extensions are now migrated to object-specific meta fields where appropriate

	// Check party meta
	partyMeta := migratedVCon.Parties[0].Meta
	if partyMeta == nil {
		t.Fatal("Party meta should not be nil after migration")
	}

	partyExtensions, exists := partyMeta["extensions"]
	if !exists {
		t.Error("extensions should exist in party meta")
	}

	partyExtMap, ok := partyExtensions.(map[string]any)
	if !ok {
		t.Fatal("party extensions should be a map")
	}

	// Note: DID field removed from Party struct for IETF draft-03 compliance
	// Check for Stir field instead
	if partyExtMap["stir"] != "stir-value" {
		t.Errorf("Stir should be preserved in party meta. Got: %v", partyExtMap["stir"])
		t.Logf("All party extensions: %+v", partyExtMap)
	}
	// Note: Timezone field also removed from Party struct for IETF draft-03 compliance
	// Only check for remaining extension fields

	// Check dialog meta
	dialogMeta := migratedVCon.Dialog[0].Meta
	if dialogMeta == nil {
		t.Fatal("Dialog meta should not be nil after migration")
	}

	dialogExtensions, exists := dialogMeta["extensions"]
	if !exists {
		t.Error("extensions should exist in dialog meta")
	}

	dialogExtMap, ok := dialogExtensions.(map[string]any)
	if !ok {
		t.Fatal("dialog extensions should be a map")
	}

	if dialogExtMap["campaign"] != "sales-2024" {
		t.Error("Campaign should be preserved in dialog meta")
	}
	if dialogExtMap["skill"] != "support" {
		t.Error("Skill should be preserved in dialog meta")
	}
}

func TestRestoreExtensionsFromMeta(t *testing.T) {
	// Create a vCon with extensions in meta (as if it was migrated)
	vcon := &VCon{
		UUID:      newUUID(),
		Vcon:      "0.0.2",
		CreatedAt: time.Now(),
		// Note: Meta field removed for IETF draft-03 compliance
		// Test restoration warnings without actual Meta field
		Parties: []Party{
			{Name: StringPtr("Test User")},
		},
	}

	_, result, err := vcon.RestoreExtensionsFromMeta()
	if err != nil {
		t.Fatalf("RestoreExtensionsFromMeta() error = %v", err)
	}

	// Note: Restoration no longer actually restores fields since Meta, Extensions, and MustSupport removed
	// Function now issues warnings about inability to restore these deprecated fields
	if len(result.Warnings) < 3 {
		t.Errorf("Expected at least 3 warnings about removed fields, got %d", len(result.Warnings))
	}

	// Extensions and MustSupport fields have been removed from VCon struct
	// These fields can no longer be restored since they don't exist
	// The restore function should now issue warnings about this

	// Note: Meta field removed for IETF draft-03 compliance
	// Can no longer check for meta cleanup since field doesn't exist
	// Restoration function now properly issues warnings about removed fields
}

func TestMigrationPreservesOriginal(t *testing.T) {
	// Test that the original vCon is not modified by migration
	originalVCon := &VCon{
		UUID:      newUUID(),
		Vcon:      "0.0.2",
		CreatedAt: time.Now(),
		Parties: []Party{
			{
				Name: StringPtr("Test User"),
				Stir: StringPtr("stir-value"), // Using remaining extension field
			},
		},
	}

	// Store original values - using Stir since DID field removed for IETF draft-03 compliance
	originalStir := *originalVCon.Parties[0].Stir

	_, _, err := originalVCon.MigrateExtensionsToMeta()
	if err != nil {
		t.Fatalf("MigrateExtensionsToMeta() error = %v", err)
	}

	// Extensions field has been removed from VCon struct - no need to check immutability
	// DID field also removed - checking Stir field instead

	if originalVCon.Parties[0].Stir == nil || *originalVCon.Parties[0].Stir != originalStir {
		t.Error("Original vCon Party.Stir should not be modified")
	}
}

// Helper functions for testing.
func stringPtrEqual(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func ptrToStr(s *string) string {
	if s == nil {
		return "nil"
	}
	return *s
}

func contentHashEqual(a, b *ContentHashValue) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	// Compare the single hash values
	aStr := a.GetSingle()
	bStr := b.GetSingle()
	return aStr == bStr
}

func contentHashToStr(c *ContentHashValue) string {
	if c == nil {
		return "nil"
	}
	return c.GetSingle()
}
