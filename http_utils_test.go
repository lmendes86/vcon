package vcon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// MockHTTPClient for testing.
type MockHTTPClient struct {
	Response    *http.Response
	Error       error
	LastRequest *http.Request
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	m.LastRequest = req
	return m.Response, m.Error
}

func TestBuildFromJSON(t *testing.T) {
	testVCon := NewWithDefaults()
	testVCon.AddParty(Party{Name: StringPtr("Test User")})

	jsonData, err := json.Marshal(testVCon)
	if err != nil {
		t.Fatalf("Failed to marshal test vCon: %v", err)
	}

	// Test normal parsing
	vcon, err := BuildFromJSON(string(jsonData), PropertyHandlingDefault)
	if err != nil {
		t.Fatalf("Failed to build from JSON: %v", err)
	}

	if len(vcon.Parties) != 1 {
		t.Errorf("Expected 1 party, got %d", len(vcon.Parties))
	}

	// Test invalid JSON
	_, err = BuildFromJSON(`{"invalid": json}`, PropertyHandlingDefault)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestBuildNew(t *testing.T) {
	vcon := BuildNew(PropertyHandlingDefault)

	// Verify all slices are initialized
	if vcon.Group == nil {
		t.Error("Group should be initialized")
	}
	if vcon.Parties == nil {
		t.Error("Parties should be initialized")
	}
	if vcon.Dialog == nil {
		t.Error("Dialog should be initialized")
	}
	if vcon.Attachments == nil {
		t.Error("Attachments should be initialized")
	}
	if vcon.Analysis == nil {
		t.Error("Analysis should be initialized")
	}
	// Extensions and MustSupport fields have been removed for IETF draft-03 compliance
	// No need to test these deprecated fields anymore
	if vcon.Signatures == nil {
		t.Error("Signatures should be initialized")
	}
	if vcon.Appended != nil {
		t.Error("Appended should be nil by default (single object, not collection)")
	}
	if vcon.Redacted != nil {
		t.Error("Redacted should be nil by default (single object, not collection)")
	}
	// Note: Top-level Meta field removed for IETF draft-03 compliance
	// Use object-specific meta fields instead

	// Verify all slices are empty
	if len(vcon.Group) != 0 {
		t.Errorf("Expected empty Group, got %d items", len(vcon.Group))
	}
	if len(vcon.Parties) != 0 {
		t.Errorf("Expected empty Parties, got %d items", len(vcon.Parties))
	}
}

func TestLoadFromURL(t *testing.T) {
	testVCon := NewWithDefaults()
	testVCon.AddParty(Party{Name: StringPtr("Test User")})

	jsonData, err := json.Marshal(testVCon)
	if err != nil {
		t.Fatalf("Failed to marshal test vCon: %v", err)
	}

	// Test successful load
	mockClient := &MockHTTPClient{
		Response: &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader(jsonData)),
		},
	}

	vcon, err := LoadFromURLWithClient("http://example.com/vcon.json", PropertyHandlingDefault, mockClient)
	if err != nil {
		t.Fatalf("Failed to load from URL: %v", err)
	}

	if len(vcon.Parties) != 1 {
		t.Errorf("Expected 1 party, got %d", len(vcon.Parties))
	}

	// Verify request headers
	if mockClient.LastRequest.Header.Get("Accept") != "application/json" {
		t.Error("Accept header not set correctly")
	}
	if mockClient.LastRequest.Header.Get("User-Agent") != "go-vcon/1.0" {
		t.Error("User-Agent header not set correctly")
	}

	// Test HTTP error
	mockClient.Response = &http.Response{
		StatusCode: 404,
		Status:     "404 Not Found",
		Body:       io.NopCloser(strings.NewReader("")),
	}

	_, err = LoadFromURLWithClient("http://example.com/notfound.json", PropertyHandlingDefault, mockClient)
	if err == nil {
		t.Error("Expected error for 404 response")
	}

	// Test network error
	mockClient.Error = fmt.Errorf("network error")
	mockClient.Response = nil

	_, err = LoadFromURLWithClient("http://example.com/error.json", PropertyHandlingDefault, mockClient)
	if err == nil {
		t.Error("Expected error for network failure")
	}
}

func TestLoadFromFile(t *testing.T) {
	testVCon := NewWithDefaults()
	testVCon.AddParty(Party{Name: StringPtr("Test User")})

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "test_vcon_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write test data to file
	jsonData, err := json.Marshal(testVCon)
	if err != nil {
		t.Fatalf("Failed to marshal test vCon: %v", err)
	}

	_, err = tmpFile.Write(jsonData)
	if err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	tmpFile.Close()

	// Test loading from file
	vcon, err := LoadFromFile(tmpFile.Name(), PropertyHandlingDefault)
	if err != nil {
		t.Fatalf("Failed to load from file: %v", err)
	}

	if len(vcon.Parties) != 1 {
		t.Errorf("Expected 1 party, got %d", len(vcon.Parties))
	}

	// Test loading non-existent file
	_, err = LoadFromFile("/nonexistent/file.json", PropertyHandlingDefault)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestSaveToFile(t *testing.T) {
	testVCon := NewWithDefaults()
	testVCon.AddParty(Party{Name: StringPtr("Test User")})

	// Create temporary file path
	tmpFile, err := os.CreateTemp("", "test_save_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	// Test saving to file
	err = testVCon.SaveToFile(tmpFile.Name(), PropertyHandlingDefault)
	if err != nil {
		t.Fatalf("Failed to save to file: %v", err)
	}

	// Verify file was written
	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	// Verify content
	var loadedVCon VCon
	err = json.Unmarshal(data, &loadedVCon)
	if err != nil {
		t.Fatalf("Failed to unmarshal saved vCon: %v", err)
	}

	if len(loadedVCon.Parties) != 1 {
		t.Errorf("Expected 1 party in saved vCon, got %d", len(loadedVCon.Parties))
	}
}

func TestPostToURL(t *testing.T) {
	testVCon := NewWithDefaults()
	testVCon.AddParty(Party{Name: StringPtr("Test User")})

	// Test successful post
	mockClient := &MockHTTPClient{
		Response: &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader("OK")),
		},
	}

	headers := map[string]string{
		"Authorization": "Bearer token123",
		"X-Custom":      "custom-value",
	}

	resp, err := testVCon.PostToURLWithClient("http://example.com/vcons", headers, PropertyHandlingDefault, mockClient)
	if err != nil {
		t.Fatalf("Failed to post to URL: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify request was made correctly
	if mockClient.LastRequest.Method != "POST" {
		t.Errorf("Expected POST method, got %s", mockClient.LastRequest.Method)
	}

	if mockClient.LastRequest.Header.Get("Content-Type") != "application/json" {
		t.Error("Content-Type header not set correctly")
	}

	if mockClient.LastRequest.Header.Get("Authorization") != "Bearer token123" {
		t.Error("Authorization header not set correctly")
	}

	if mockClient.LastRequest.Header.Get("X-Custom") != "custom-value" {
		t.Error("Custom header not set correctly")
	}

	// Verify request body
	body, err := io.ReadAll(mockClient.LastRequest.Body)
	if err != nil {
		t.Fatalf("Failed to read request body: %v", err)
	}

	var sentVCon VCon
	err = json.Unmarshal(body, &sentVCon)
	if err != nil {
		t.Fatalf("Failed to unmarshal sent vCon: %v", err)
	}

	if len(sentVCon.Parties) != 1 {
		t.Errorf("Expected 1 party in sent vCon, got %d", len(sentVCon.Parties))
	}
}

func TestToJSONString(t *testing.T) {
	testVCon := NewWithDefaults()
	testVCon.AddParty(Party{Name: StringPtr("Test User")})

	// Test normal JSON conversion
	jsonStr, err := testVCon.ToJSONString(PropertyHandlingDefault)
	if err != nil {
		t.Fatalf("Failed to convert to JSON string: %v", err)
	}

	var parsed VCon
	err = json.Unmarshal([]byte(jsonStr), &parsed)
	if err != nil {
		t.Fatalf("Failed to parse generated JSON: %v", err)
	}

	if len(parsed.Parties) != 1 {
		t.Errorf("Expected 1 party in parsed vCon, got %d", len(parsed.Parties))
	}

	// Test indented JSON conversion
	indentedStr, err := testVCon.ToJSONIndentString(PropertyHandlingDefault)
	if err != nil {
		t.Fatalf("Failed to convert to indented JSON string: %v", err)
	}

	if !strings.Contains(indentedStr, "\n") {
		t.Error("Indented JSON should contain newlines")
	}

	// Test Dumps alias
	dumpsStr, err := testVCon.Dumps(PropertyHandlingDefault)
	if err != nil {
		t.Fatalf("Failed to use Dumps method: %v", err)
	}

	if dumpsStr != jsonStr {
		t.Error("Dumps should return same result as ToJSONString")
	}
}

func TestValidation(t *testing.T) {
	// Test valid vCon
	validVCon := NewWithDefaults()
	validVCon.AddParty(Party{Name: StringPtr("Test User")})

	jsonStr, err := json.Marshal(validVCon)
	if err != nil {
		t.Fatalf("Failed to marshal valid vCon: %v", err)
	}

	isValid, errors := ValidateJSONString(string(jsonStr))
	if !isValid {
		t.Errorf("Valid vCon should pass validation, errors: %v", errors)
	}

	// Test invalid JSON
	isValid, errors = ValidateJSONString(`{"invalid": json}`)
	if isValid {
		t.Error("Invalid JSON should fail validation")
	}
	if len(errors) == 0 {
		t.Error("Should have validation errors for invalid JSON")
	}

	// Create temporary file for file validation test
	tmpFile, err := os.CreateTemp("", "test_validate_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write(jsonStr)
	if err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	tmpFile.Close()

	// Test file validation
	isValid, errors = ValidateVConFile(tmpFile.Name())
	if !isValid {
		t.Errorf("Valid vCon file should pass validation, errors: %v", errors)
	}

	// Test validating non-existent file
	isValid, _ = ValidateVConFile("/nonexistent/file.json")
	if isValid {
		t.Error("Non-existent file should fail validation")
	}
}

func TestLoadFromURLDirect(t *testing.T) {
	// Test the direct LoadFromURL function (0% coverage)
	testVCon := NewWithDefaults()
	testVCon.AddParty(Party{Name: StringPtr("Test User")})

	jsonData, err := json.Marshal(testVCon)
	if err != nil {
		t.Fatalf("Failed to marshal test vCon: %v", err)
	}

	// Mock the default client by temporarily replacing it
	originalClient := DefaultHTTPClient
	defer func() { DefaultHTTPClient = originalClient }()

	mockClient := &MockHTTPClient{
		Response: &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader(jsonData)),
		},
	}
	DefaultHTTPClient = mockClient

	vcon, err := LoadFromURL("http://example.com/vcon.json", PropertyHandlingDefault)
	if err != nil {
		t.Fatalf("Failed to load from URL: %v", err)
	}

	if len(vcon.Parties) != 1 {
		t.Errorf("Expected 1 party, got %d", len(vcon.Parties))
	}
}

func TestPostToURLDirect(t *testing.T) {
	// Test the direct PostToURL function (0% coverage)
	testVCon := NewWithDefaults()
	testVCon.AddParty(Party{Name: StringPtr("Test User")})

	// Mock the default client
	originalClient := DefaultHTTPClient
	defer func() { DefaultHTTPClient = originalClient }()

	mockClient := &MockHTTPClient{
		Response: &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader("OK")),
		},
	}
	DefaultHTTPClient = mockClient

	headers := map[string]string{"Authorization": "Bearer token"}

	resp, err := testVCon.PostToURL("http://example.com/vcons", headers, PropertyHandlingDefault)
	if err != nil {
		t.Fatalf("Failed to post to URL: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestToJSONIndentStringEdgeCases(t *testing.T) {
	// Test ToJSONIndentString with different property handling modes (35.3% coverage)
	testVCon := NewWithDefaults()
	testVCon.AddParty(Party{Name: StringPtr("Test User")})

	// Test with non-default property handling
	indentedStr, err := testVCon.ToJSONIndentString(PropertyHandlingStrict)
	if err != nil {
		t.Fatalf("Failed to convert to indented JSON string with strict mode: %v", err)
	}

	if !strings.Contains(indentedStr, "\n") {
		t.Error("Indented JSON should contain newlines")
	}

	// Test with meta property handling
	indentedStr, err = testVCon.ToJSONIndentString(PropertyHandlingMeta)
	if err != nil {
		t.Fatalf("Failed to convert to indented JSON string with meta mode: %v", err)
	}

	if !strings.Contains(indentedStr, "\n") {
		t.Error("Indented JSON should contain newlines")
	}
}

func TestBuildFromJSONPropertyHandling(t *testing.T) {
	// Test BuildFromJSON with different property handling modes
	testJSON := `{
		"uuid": "123e4567-e89b-12d3-a456-426614174000",
		"vcon": "0.0.1",
		"created_at": "2023-01-01T00:00:00Z",
		"custom_field": "should_be_handled"
	}`

	// Test with strict property handling
	vcon, err := BuildFromJSON(testJSON, PropertyHandlingStrict)
	if err != nil {
		t.Fatalf("Failed to build from JSON with strict mode: %v", err)
	}

	if vcon.UUID == uuid.Nil {
		t.Error("UUID should be preserved")
	}

	// Test with meta property handling
	vcon, err = BuildFromJSON(testJSON, PropertyHandlingMeta)
	if err != nil {
		t.Fatalf("Failed to build from JSON with meta mode: %v", err)
	}

	if vcon.UUID == uuid.Nil {
		t.Error("UUID should be preserved")
	}

	// Test with empty property handling (should default)
	vcon, err = BuildFromJSON(testJSON, "")
	if err != nil {
		t.Fatalf("Failed to build from JSON with empty mode: %v", err)
	}

	if vcon.UUID == uuid.Nil {
		t.Error("UUID should be preserved")
	}
}

func TestSaveToFilePropertyHandling(t *testing.T) {
	// Test SaveToFile with different property handling modes
	testVCon := NewWithDefaults()
	testVCon.AddParty(Party{Name: StringPtr("Test User")})

	// Test with strict property handling
	tmpFile, err := os.CreateTemp("", "test_save_strict_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	err = testVCon.SaveToFile(tmpFile.Name(), PropertyHandlingStrict)
	if err != nil {
		t.Fatalf("Failed to save to file with strict mode: %v", err)
	}

	// Verify file was written
	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	var loadedVCon VCon
	err = json.Unmarshal(data, &loadedVCon)
	if err != nil {
		t.Fatalf("Failed to unmarshal saved vCon: %v", err)
	}

	// Test with meta property handling
	tmpFile2, err := os.CreateTemp("", "test_save_meta_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile2.Close()
	defer os.Remove(tmpFile2.Name())

	err = testVCon.SaveToFile(tmpFile2.Name(), PropertyHandlingMeta)
	if err != nil {
		t.Fatalf("Failed to save to file with meta mode: %v", err)
	}
}

func TestPostToURLPropertyHandling(t *testing.T) {
	// Test PostToURLWithClient with different property handling modes
	testVCon := NewWithDefaults()
	testVCon.AddParty(Party{Name: StringPtr("Test User")})

	mockClient := &MockHTTPClient{
		Response: &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader("OK")),
		},
	}

	// Test with strict property handling
	resp, err := testVCon.PostToURLWithClient("http://example.com/vcons", nil, PropertyHandlingStrict, mockClient)
	if err != nil {
		t.Fatalf("Failed to post with strict mode: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Test with meta property handling
	resp, err = testVCon.PostToURLWithClient("http://example.com/vcons", nil, PropertyHandlingMeta, mockClient)
	if err != nil {
		t.Fatalf("Failed to post with meta mode: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestToJSONStringPropertyHandling(t *testing.T) {
	// Test ToJSONString with different property handling modes
	testVCon := NewWithDefaults()
	testVCon.AddParty(Party{Name: StringPtr("Test User")})

	// Test with strict property handling
	jsonStr, err := testVCon.ToJSONString(PropertyHandlingStrict)
	if err != nil {
		t.Fatalf("Failed to convert to JSON string with strict mode: %v", err)
	}

	var parsed VCon
	err = json.Unmarshal([]byte(jsonStr), &parsed)
	if err != nil {
		t.Fatalf("Failed to parse generated JSON: %v", err)
	}

	// Test with meta property handling
	jsonStr, err = testVCon.ToJSONString(PropertyHandlingMeta)
	if err != nil {
		t.Fatalf("Failed to convert to JSON string with meta mode: %v", err)
	}

	err = json.Unmarshal([]byte(jsonStr), &parsed)
	if err != nil {
		t.Fatalf("Failed to parse generated JSON: %v", err)
	}
}

func TestHTTPErrorScenarios(t *testing.T) {
	// Test various HTTP error scenarios
	tests := []struct {
		name          string
		response      *http.Response
		expectError   bool
		errorContains string
	}{
		{
			name: "404 Not Found",
			response: &http.Response{
				StatusCode: 404,
				Status:     "404 Not Found",
				Body:       io.NopCloser(strings.NewReader("")),
			},
			expectError:   true,
			errorContains: "HTTP error: 404",
		},
		{
			name: "500 Internal Server Error",
			response: &http.Response{
				StatusCode: 500,
				Status:     "500 Internal Server Error",
				Body:       io.NopCloser(strings.NewReader("")),
			},
			expectError:   true,
			errorContains: "HTTP error: 500",
		},
		{
			name: "403 Forbidden",
			response: &http.Response{
				StatusCode: 403,
				Status:     "403 Forbidden",
				Body:       io.NopCloser(strings.NewReader("")),
			},
			expectError:   true,
			errorContains: "HTTP error: 403",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{Response: tt.response}

			_, err := LoadFromURLWithClient("http://example.com/test.json", PropertyHandlingDefault, mockClient)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestInvalidJSONResponses(t *testing.T) {
	// Test handling of invalid JSON responses
	tests := []struct {
		name        string
		json        string
		expectError bool
	}{
		{"invalid_syntax", `{"invalid": json}`, true},
		{"incomplete", `{incomplete json`, true},
		{"null", `null`, false}, // null is valid JSON but might not be a valid vCon
		{"string", `"not an object"`, true},
		{"array", `[]`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{
				Response: &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(tt.json)),
				},
			}

			_, err := LoadFromURLWithClient("http://example.com/test.json", PropertyHandlingDefault, mockClient)
			if tt.expectError && err == nil {
				t.Error("Expected error for invalid JSON response")
			} else if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
