package vcon

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// ContentHashValue represents a content hash that can be either a single string or an array of strings.
// Per IETF spec, content_hash can be a string for a single hash or an array for multiple hashes.
// Examples: "sha-256:abc..." or ["sha-256:abc...", "sha-512:def..."].
type ContentHashValue struct {
	value interface{} // Can be string or []string
}

// NewContentHashSingle creates a ContentHashValue with a single hash string.
func NewContentHashSingle(hash string) *ContentHashValue {
	if hash == "" {
		return nil
	}
	return &ContentHashValue{value: hash}
}

// NewContentHashArray creates a ContentHashValue with multiple hash strings.
func NewContentHashArray(hashes []string) *ContentHashValue {
	if len(hashes) == 0 {
		return nil
	}
	// Filter out empty hashes
	var filtered []string
	for _, h := range hashes {
		if h != "" {
			filtered = append(filtered, h)
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	return &ContentHashValue{value: filtered}
}

// IsEmpty returns true if the ContentHashValue is empty or nil.
func (c *ContentHashValue) IsEmpty() bool {
	if c == nil {
		return true
	}
	switch v := c.value.(type) {
	case string:
		return v == ""
	case []string:
		return len(v) == 0
	default:
		return true
	}
}

// GetSingle returns the hash as a single string. If it's an array, returns the first hash.
func (c *ContentHashValue) GetSingle() string {
	if c == nil {
		return ""
	}
	switch v := c.value.(type) {
	case string:
		return v
	case []string:
		if len(v) > 0 {
			return v[0]
		}
		return ""
	default:
		return ""
	}
}

// GetArray returns the hash(es) as a string slice.
func (c *ContentHashValue) GetArray() []string {
	if c == nil {
		return nil
	}
	switch v := c.value.(type) {
	case string:
		if v == "" {
			return nil
		}
		return []string{v}
	case []string:
		return v
	default:
		return nil
	}
}

// IsArray returns true if the ContentHashValue contains multiple hashes.
func (c *ContentHashValue) IsArray() bool {
	if c == nil {
		return false
	}
	_, ok := c.value.([]string)
	return ok
}

// MarshalJSON implements custom JSON marshaling for ContentHashValue.
// Single hash: "sha-256:abc..."
// Multiple hashes: ["sha-256:abc...", "sha-512:def..."].
func (c *ContentHashValue) MarshalJSON() ([]byte, error) {
	if c == nil {
		return []byte("null"), nil
	}

	switch v := c.value.(type) {
	case string:
		return json.Marshal(v)
	case []string:
		if len(v) == 1 {
			// Single item arrays are serialized as strings for backwards compatibility
			return json.Marshal(v[0])
		}
		return json.Marshal(v)
	default:
		return []byte("null"), nil
	}
}

// UnmarshalJSON implements custom JSON unmarshaling for ContentHashValue.
// Accepts both string and array formats.
func (c *ContentHashValue) UnmarshalJSON(data []byte) error {
	// First try to unmarshal as string
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		c.value = str
		return nil
	}

	// Then try to unmarshal as array
	var arr []string
	if err := json.Unmarshal(data, &arr); err == nil {
		c.value = arr
		return nil
	}

	return fmt.Errorf("content_hash must be a string or array of strings")
}

// FormatContentHash creates a content hash string in the IETF spec format: "algorithm:hash".
func FormatContentHash(algorithm, hash string) string {
	return fmt.Sprintf("%s:%s", algorithm, hash)
}

// ParseContentHash parses a content hash string into algorithm and hash components.
func ParseContentHash(contentHash string) (algorithm, hash string, err error) {
	parts := strings.SplitN(contentHash, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid content_hash format: expected 'algorithm:hash', got '%s'", contentHash)
	}
	return parts[0], parts[1], nil
}

// ValidateContentHashFormat validates the format and content of a content hash string.
func ValidateContentHashFormat(contentHash string) error {
	if contentHash == "" {
		return fmt.Errorf("content_hash cannot be empty")
	}

	algorithm, hash, err := ParseContentHash(contentHash)
	if err != nil {
		return err
	}

	// Validate algorithm - IETF spec supports common hash algorithms
	// Define algorithm categories for enhanced security validation
	strongAlgs := map[string]bool{
		"sha-256":  true,
		"sha-512":  true, // Always supported per IETF requirement
		"sha-384":  true,
		"sha3-256": true, // Additional SHA-3 support
		"sha3-512": true, // Additional SHA-3 support
		"blake2b":  true, // Modern hash algorithm
		"blake2s":  true, // Modern hash algorithm
	}

	weakAlgs := map[string]bool{
		"sha1": true, // Legacy support (deprecated)
		"md5":  true, // Legacy support (strongly discouraged)
	}

	// Check if algorithm is supported
	if !strongAlgs[algorithm] && !weakAlgs[algorithm] {
		return fmt.Errorf("unsupported hash algorithm: %s (supported: sha-256, sha-512, sha-384, sha3-256, sha3-512, blake2b, blake2s, sha1, md5)", algorithm)
	}

	// Warning for weak algorithms (per IETF security requirements)
	if weakAlgs[algorithm] {
		return fmt.Errorf("weak hash algorithm detected: %s - consider using stronger algorithms like sha-256 or sha-512", algorithm)
	}

	// Validate hash is valid base64url encoding
	if _, err := base64.RawURLEncoding.DecodeString(hash); err != nil {
		return fmt.Errorf("invalid base64url hash: %v", err)
	}

	// Basic length validation for common algorithms
	expectedLengths := map[string]int{
		"sha-256":  43, // 32 bytes * 4/3 (base64) ≈ 43 chars
		"sha-512":  86, // 64 bytes * 4/3 (base64) ≈ 86 chars
		"sha-384":  64, // 48 bytes * 4/3 (base64) ≈ 64 chars
		"sha3-256": 43, // 32 bytes * 4/3 (base64) ≈ 43 chars
		"sha3-512": 86, // 64 bytes * 4/3 (base64) ≈ 86 chars
		"blake2b":  86, // 64 bytes * 4/3 (base64) ≈ 86 chars
		"blake2s":  43, // 32 bytes * 4/3 (base64) ≈ 43 chars
		"sha1":     27, // 20 bytes * 4/3 (base64) ≈ 27 chars
		"md5":      22, // 16 bytes * 4/3 (base64) ≈ 22 chars
	}

	if expectedLen, exists := expectedLengths[algorithm]; exists {
		// Allow some variance for padding differences
		if len(hash) < expectedLen-2 || len(hash) > expectedLen+2 {
			return fmt.Errorf("hash length %d is invalid for algorithm %s (expected ~%d)", len(hash), algorithm, expectedLen)
		}
	}

	return nil
}

// MigrateAlgSignatureToContentHash converts legacy alg+signature to content_hash format.
func MigrateAlgSignatureToContentHash(alg, signature *string) *string {
	if alg == nil || signature == nil || *alg == "" || *signature == "" {
		return nil
	}

	contentHash := FormatContentHash(*alg, *signature)
	return &contentHash
}

// ConvertContentHashToAlgSignature converts content_hash back to legacy alg+signature format.
// This is used for backward compatibility during migration.
func ConvertContentHashToAlgSignature(contentHash *string) (alg, signature *string, err error) {
	if contentHash == nil || *contentHash == "" {
		return nil, nil, nil
	}

	algorithm, hash, parseErr := ParseContentHash(*contentHash)
	if parseErr != nil {
		return nil, nil, parseErr
	}

	return &algorithm, &hash, nil
}
