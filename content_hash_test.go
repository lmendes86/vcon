package vcon

import (
	"encoding/json"
	"testing"
)

func TestFormatContentHash(t *testing.T) {
	tests := []struct {
		name      string
		algorithm string
		hash      string
		expected  string
	}{
		{
			name:      "sha-256 hash",
			algorithm: "sha-256",
			hash:      "abc123def456",
			expected:  "sha-256:abc123def456",
		},
		{
			name:      "sha-512 hash",
			algorithm: "sha-512",
			hash:      "xyz789uvw012",
			expected:  "sha-512:xyz789uvw012",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatContentHash(tt.algorithm, tt.hash)
			if result != tt.expected {
				t.Errorf("FormatContentHash() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseContentHash(t *testing.T) {
	tests := []struct {
		name        string
		contentHash string
		wantAlg     string
		wantHash    string
		wantErr     bool
	}{
		{
			name:        "valid sha-256 hash",
			contentHash: "sha-256:abc123def456",
			wantAlg:     "sha-256",
			wantHash:    "abc123def456",
			wantErr:     false,
		},
		{
			name:        "valid sha-512 hash",
			contentHash: "sha-512:xyz789uvw012",
			wantAlg:     "sha-512",
			wantHash:    "xyz789uvw012",
			wantErr:     false,
		},
		{
			name:        "invalid format - no colon",
			contentHash: "sha256abc123",
			wantErr:     true,
		},
		{
			name:        "invalid format - empty",
			contentHash: "",
			wantErr:     true,
		},
		{
			name:        "invalid format - only colon",
			contentHash: ":",
			wantAlg:     "",
			wantHash:    "",
			wantErr:     false, // This should parse but fail validation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alg, hash, err := ParseContentHash(tt.contentHash)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseContentHash() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseContentHash() unexpected error: %v", err)
				return
			}

			if alg != tt.wantAlg {
				t.Errorf("ParseContentHash() algorithm = %v, want %v", alg, tt.wantAlg)
			}

			if hash != tt.wantHash {
				t.Errorf("ParseContentHash() hash = %v, want %v", hash, tt.wantHash)
			}
		})
	}
}

func TestValidateContentHashFormat(t *testing.T) {
	tests := []struct {
		name        string
		contentHash string
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid sha-256 hash",
			contentHash: "sha-256:LCa0a2j_xo_5m0U8HTBBNBNCLXBkg7-g-YpeiGJm564",
			wantErr:     false,
		},
		{
			name:        "valid sha-512 hash",
			contentHash: "sha-512:MJ7MSJwS1utMxA9QyQLytNDtd-5RGnx6m808qG1M2G-YndNbxf9JlnDaNCVbRbDP2DDoH2Bdz33FVC6TrpzXbw",
			wantErr:     false,
		},
		{
			name:        "invalid algorithm",
			contentHash: "md4:abc123",
			wantErr:     true,
			errContains: "unsupported hash algorithm",
		},
		{
			name:        "invalid base64url",
			contentHash: "sha-256:abc@#$%",
			wantErr:     true,
			errContains: "invalid base64url hash",
		},
		{
			name:        "empty content hash",
			contentHash: "",
			wantErr:     true,
			errContains: "content_hash cannot be empty",
		},
		{
			name:        "invalid format",
			contentHash: "no-colon-here",
			wantErr:     true,
			errContains: "invalid content_hash format",
		},
		{
			name:        "wrong hash length for sha-256",
			contentHash: "sha-256:dGVzdA", // "test" in base64url - valid encoding but wrong length
			wantErr:     true,
			errContains: "hash length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateContentHashFormat(tt.contentHash)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateContentHashFormat() expected error but got none")
					return
				}

				if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("ValidateContentHashFormat() error = %v, want error containing %v", err, tt.errContains)
				}
			} else if err != nil {
				t.Errorf("ValidateContentHashFormat() unexpected error: %v", err)
			}
		})
	}
}

func TestMigrateAlgSignatureToContentHash(t *testing.T) {
	tests := []struct {
		name      string
		alg       *string
		signature *string
		expected  *string
	}{
		{
			name:      "valid alg and signature",
			alg:       strPtr("sha-256"),
			signature: strPtr("abc123"),
			expected:  strPtr("sha-256:abc123"),
		},
		{
			name:      "nil alg",
			alg:       nil,
			signature: strPtr("abc123"),
			expected:  nil,
		},
		{
			name:      "nil signature",
			alg:       strPtr("sha-256"),
			signature: nil,
			expected:  nil,
		},
		{
			name:      "empty alg",
			alg:       strPtr(""),
			signature: strPtr("abc123"),
			expected:  nil,
		},
		{
			name:      "empty signature",
			alg:       strPtr("sha-256"),
			signature: strPtr(""),
			expected:  nil,
		},
		{
			name:      "both nil",
			alg:       nil,
			signature: nil,
			expected:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MigrateAlgSignatureToContentHash(tt.alg, tt.signature)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("MigrateAlgSignatureToContentHash() = %v, want nil", result)
				}
			} else {
				if result == nil {
					t.Errorf("MigrateAlgSignatureToContentHash() = nil, want %v", *tt.expected)
				} else if *result != *tt.expected {
					t.Errorf("MigrateAlgSignatureToContentHash() = %v, want %v", *result, *tt.expected)
				}
			}
		})
	}
}

func TestConvertContentHashToAlgSignature(t *testing.T) {
	tests := []struct {
		name        string
		contentHash *string
		wantAlg     *string
		wantSig     *string
		wantErr     bool
	}{
		{
			name:        "valid content hash",
			contentHash: strPtr("sha-256:abc123"),
			wantAlg:     strPtr("sha-256"),
			wantSig:     strPtr("abc123"),
			wantErr:     false,
		},
		{
			name:        "nil content hash",
			contentHash: nil,
			wantAlg:     nil,
			wantSig:     nil,
			wantErr:     false,
		},
		{
			name:        "empty content hash",
			contentHash: strPtr(""),
			wantAlg:     nil,
			wantSig:     nil,
			wantErr:     false,
		},
		{
			name:        "invalid format",
			contentHash: strPtr("invalid-format"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alg, sig, err := ConvertContentHashToAlgSignature(tt.contentHash)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ConvertContentHashToAlgSignature() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ConvertContentHashToAlgSignature() unexpected error: %v", err)
				return
			}

			// Check alg
			if tt.wantAlg == nil {
				if alg != nil {
					t.Errorf("ConvertContentHashToAlgSignature() alg = %v, want nil", alg)
				}
			} else {
				if alg == nil || *alg != *tt.wantAlg {
					t.Errorf("ConvertContentHashToAlgSignature() alg = %v, want %v", alg, *tt.wantAlg)
				}
			}

			// Check signature
			if tt.wantSig == nil {
				if sig != nil {
					t.Errorf("ConvertContentHashToAlgSignature() signature = %v, want nil", sig)
				}
			} else {
				if sig == nil || *sig != *tt.wantSig {
					t.Errorf("ConvertContentHashToAlgSignature() signature = %v, want %v", sig, *tt.wantSig)
				}
			}
		})
	}
}

func TestContentHashValue_NewContentHashSingle(t *testing.T) {
	tests := []struct {
		name     string
		hash     string
		expected bool // true if should return non-nil
	}{
		{"valid hash", "sha-256:abc123", true},
		{"empty hash", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewContentHashSingle(tt.hash)
			if tt.expected && result == nil {
				t.Errorf("Expected non-nil result for hash %q", tt.hash)
			}
			if !tt.expected && result != nil {
				t.Errorf("Expected nil result for hash %q", tt.hash)
			}
			if result != nil && result.GetSingle() != tt.hash {
				t.Errorf("Expected hash %q, got %q", tt.hash, result.GetSingle())
			}
		})
	}
}

func TestContentHashValue_NewContentHashArray(t *testing.T) {
	tests := []struct {
		name     string
		hashes   []string
		expected bool // true if should return non-nil
		want     []string
	}{
		{"valid array", []string{"sha-256:abc", "sha-512:def"}, true, []string{"sha-256:abc", "sha-512:def"}},
		{"empty array", []string{}, false, nil},
		{"array with empty strings", []string{"", "sha-256:abc", ""}, true, []string{"sha-256:abc"}},
		{"all empty strings", []string{"", ""}, false, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewContentHashArray(tt.hashes)
			if tt.expected && result == nil {
				t.Errorf("Expected non-nil result for hashes %v", tt.hashes)
			}
			if !tt.expected && result != nil {
				t.Errorf("Expected nil result for hashes %v", tt.hashes)
			}
			if result != nil {
				got := result.GetArray()
				if len(got) != len(tt.want) {
					t.Errorf("Expected %d hashes, got %d", len(tt.want), len(got))
				}
				for i, hash := range tt.want {
					if i >= len(got) || got[i] != hash {
						t.Errorf("Expected hash[%d] = %q, got %q", i, hash, got[i])
					}
				}
			}
		})
	}
}

func TestContentHashValue_JSONMarshaling(t *testing.T) {
	tests := []struct {
		name     string
		value    *ContentHashValue
		expected string
	}{
		{"nil value", nil, "null"},
		{"single hash", NewContentHashSingle("sha-256:abc123"), `"sha-256:abc123"`},
		{"single item array", NewContentHashArray([]string{"sha-256:abc123"}), `"sha-256:abc123"`},
		{"multiple hashes", NewContentHashArray([]string{"sha-256:abc", "sha-512:def"}), `["sha-256:abc","sha-512:def"]`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := json.Marshal(tt.value)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if string(result) != tt.expected {
				t.Errorf("Expected JSON %q, got %q", tt.expected, string(result))
			}
		})
	}
}

func TestContentHashValue_JSONUnmarshaling(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		validate func(t *testing.T, result *ContentHashValue)
	}{
		{
			name:    "single hash string",
			input:   `"sha-256:abc123"`,
			wantErr: false,
			validate: func(t *testing.T, result *ContentHashValue) {
				if result.IsArray() {
					t.Error("Expected single hash, got array")
				}
				if result.GetSingle() != "sha-256:abc123" {
					t.Errorf("Expected hash sha-256:abc123, got %q", result.GetSingle())
				}
			},
		},
		{
			name:    "array of hashes",
			input:   `["sha-256:abc", "sha-512:def"]`,
			wantErr: false,
			validate: func(t *testing.T, result *ContentHashValue) {
				if !result.IsArray() {
					t.Error("Expected array, got single hash")
				}
				array := result.GetArray()
				if len(array) != 2 {
					t.Errorf("Expected 2 hashes, got %d", len(array))
				}
				if len(array) >= 1 && array[0] != "sha-256:abc" {
					t.Errorf("Expected first hash sha-256:abc, got %q", array[0])
				}
				if len(array) >= 2 && array[1] != "sha-512:def" {
					t.Errorf("Expected second hash sha-512:def, got %q", array[1])
				}
			},
		},
		{
			name:    "null value",
			input:   `null`,
			wantErr: false,
			validate: func(t *testing.T, result *ContentHashValue) {
				if !result.IsEmpty() {
					t.Error("Expected empty value for null")
				}
			},
		},
		{
			name:    "invalid type",
			input:   `42`,
			wantErr: true,
			validate: func(_ *testing.T, _ *ContentHashValue) {
				// Should not be called on error
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result ContentHashValue
			err := json.Unmarshal([]byte(tt.input), &result)

			if tt.wantErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.wantErr && err == nil {
				tt.validate(t, &result)
			}
		})
	}
}

func TestContentHashValue_IsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		value    *ContentHashValue
		expected bool
	}{
		{"nil value", nil, true},
		{"single empty hash", NewContentHashSingle(""), true},
		{"single valid hash", NewContentHashSingle("sha-256:abc"), false},
		{"empty array", NewContentHashArray([]string{}), true}, // NewContentHashArray returns nil for empty
		{"valid array", NewContentHashArray([]string{"sha-256:abc"}), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.value.IsEmpty()
			if result != tt.expected {
				t.Errorf("Expected IsEmpty() = %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestContentHashValue_GetMethods(t *testing.T) {
	t.Run("GetSingle from array", func(t *testing.T) {
		value := NewContentHashArray([]string{"sha-256:first", "sha-512:second"})
		result := value.GetSingle()
		if result != "sha-256:first" {
			t.Errorf("Expected first hash from array, got %q", result)
		}
	})

	t.Run("GetArray from single", func(t *testing.T) {
		value := NewContentHashSingle("sha-256:abc")
		result := value.GetArray()
		if len(result) != 1 || result[0] != "sha-256:abc" {
			t.Errorf("Expected single-item array [sha-256:abc], got %v", result)
		}
	})

	t.Run("nil value methods", func(t *testing.T) {
		var value *ContentHashValue
		if value.GetSingle() != "" {
			t.Error("Expected empty string from nil GetSingle()")
		}
		if value.GetArray() != nil {
			t.Error("Expected nil from nil GetArray()")
		}
		if !value.IsEmpty() {
			t.Error("Expected true from nil IsEmpty()")
		}
		if value.IsArray() {
			t.Error("Expected false from nil IsArray()")
		}
	})
}

// containsString is already defined in advanced_validation_test.go
