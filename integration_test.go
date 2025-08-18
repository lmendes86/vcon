package vcon

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestVConLifecycle tests the complete lifecycle of a vCon.
func TestVConLifecycle(t *testing.T) {
	// 1. Create a new vCon
	vcon := NewWithDefaults()
	vcon.Subject = StringPtr("Integration Test Call")

	// 2. Add parties
	customerParty := Party{
		Name: StringPtr("John Customer"),
		Tel:  StringPtr("+1234567890"),
		Role: StringPtr("customer"),
		Meta: map[string]any{"segment": "premium"},
	}
	vcon.AddParty(customerParty)

	agentParty := Party{
		Name:   StringPtr("Jane Agent"),
		Mailto: StringPtr("jane@company.com"),
		Role:   StringPtr("agent"),
		Meta:   map[string]any{"department": "support"},
	}
	vcon.AddParty(agentParty)

	// 3. Add dialogs
	greeting := Dialog{
		Type:    "text",
		Start:   time.Now().Add(-10 * time.Minute),
		Parties: NewDialogPartiesArrayPtr([]int{1}), // Agent
		Body:    "Hello, how can I help you today?",
	}
	vcon.AddDialog(greeting)

	response := Dialog{
		Type:    "text",
		Start:   time.Now().Add(-9 * time.Minute),
		Parties: NewDialogPartiesArrayPtr([]int{0}), // Customer
		Body:    "I need help with my account",
	}
	vcon.AddDialog(response)

	recording := Dialog{
		Type:      "recording",
		Start:     time.Now().Add(-8 * time.Minute),
		Parties:   NewDialogPartiesArrayPtr([]int{0, 1}),
		Duration:  Float64Ptr(300.5),
		Mediatype: StringPtr("audio/wav"),
		URL:       StringPtr("https://example.com/recordings/call123.wav"),
	}
	vcon.AddDialog(recording)

	// 4. Add attachments
	transcript := Attachment{
		Type:     StringPtr("transcript"),
		Body:     "Complete call transcript here...",
		Encoding: StringPtr("none"),
		Start:    func() *time.Time { t := time.Now().UTC(); return &t }(),
		Party:    IntPtr(0),
		Meta: map[string]any{
			"confidence": 0.95,
			"language":   "en-US",
		},
	}
	vcon.AddAttachment(transcript)

	// 5. Add analysis
	vcon.Analysis = []Analysis{
		{
			Type:     "sentiment",
			Vendor:   "ai-corp",
			Dialog:   float64(2), // Recording dialog
			Body:     "positive",
			Encoding: StringPtr("json"),
			Schema:   StringPtr("sentiment-v1"),
		},
		{
			Type:     "summary",
			Vendor:   "ai-corp",
			Dialog:   []interface{}{float64(0), float64(1), float64(2)},
			Body:     "Customer inquiry about account resolved successfully",
			Encoding: StringPtr("none"),
		},
	}

	// 6. Validate the vCon
	valid, errors := vcon.IsValid()
	if !valid {
		t.Fatalf("vCon should be valid, errors: %v", errors)
	}

	// 7. Test JSON serialization
	jsonData, err := vcon.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize vCon to JSON: %v", err)
	}

	// 8. Test JSON deserialization
	var restoredVCon VCon
	err = json.Unmarshal(jsonData, &restoredVCon)
	if err != nil {
		t.Fatalf("Failed to deserialize vCon from JSON: %v", err)
	}

	// 9. Verify data integrity
	if restoredVCon.UUID.String() != vcon.UUID.String() {
		t.Error("UUID should be preserved")
	}

	if len(restoredVCon.Parties) != 2 {
		t.Errorf("Expected 2 parties, got %d", len(restoredVCon.Parties))
	}

	if len(restoredVCon.Dialog) != 3 {
		t.Errorf("Expected 3 dialogs, got %d", len(restoredVCon.Dialog))
	}

	if len(restoredVCon.Attachments) != 1 {
		t.Errorf("Expected 1 attachment, got %d", len(restoredVCon.Attachments))
	}

	if len(restoredVCon.Analysis) != 2 {
		t.Errorf("Expected 2 analysis items, got %d", len(restoredVCon.Analysis))
	}
}

// TestVConSigningWorkflow tests the complete signing and verification workflow.
func TestVConSigningWorkflow(t *testing.T) {
	// 1. Create a vCon
	vcon := NewWithDefaults()
	vcon.AddParty(Party{Name: StringPtr("Test User")})
	vcon.AddDialog(Dialog{
		Type:    "text",
		Start:   time.Now(),
		Parties: NewDialogPartiesArrayPtr([]int{0}),
		Body:    "Test message",
	})

	// 2. Generate key pairs for signing
	privateKey1, publicKey1, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate first key pair: %v", err)
	}

	privateKey2, publicKey2, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate second key pair: %v", err)
	}

	// 3. Sign with first key
	err = vcon.Sign(privateKey1)
	if err != nil {
		t.Fatalf("Failed to sign vCon with first key: %v", err)
	}

	if !vcon.IsSigned() {
		t.Error("vCon should be marked as signed")
	}

	if len(vcon.Signatures) != 1 {
		t.Errorf("Expected 1 signature, got %d", len(vcon.Signatures))
	}

	// 4. Verify with correct key
	valid, err := vcon.Verify(publicKey1)
	if err != nil {
		t.Fatalf("Failed to verify signature: %v", err)
	}
	if !valid {
		t.Error("Signature should be valid with correct key")
	}

	// 5. Verify with wrong key
	valid, err = vcon.Verify(publicKey2)
	if err != nil {
		t.Fatalf("Unexpected error during verification: %v", err)
	}
	if valid {
		t.Error("Signature should be invalid with wrong key")
	}

	// 6. Test key conversion round-trip (before adding second signature)
	privatePEM, err := PrivateKeyToPEM(privateKey1)
	if err != nil {
		t.Fatalf("Failed to convert private key to PEM: %v", err)
	}

	publicPEM, err := PublicKeyToPEM(publicKey1)
	if err != nil {
		t.Fatalf("Failed to convert public key to PEM: %v", err)
	}

	_, err = PrivateKeyFromPEM(privatePEM)
	if err != nil {
		t.Fatalf("Failed to restore private key from PEM: %v", err)
	}

	restoredPublic, err := PublicKeyFromPEM(publicPEM)
	if err != nil {
		t.Fatalf("Failed to restore public key from PEM: %v", err)
	}

	// 7. Verify signature with restored keys
	valid, err = vcon.Verify(restoredPublic)
	if err != nil {
		t.Fatalf("Failed to verify with restored public key: %v", err)
	}
	if !valid {
		t.Error("Signature should be valid with restored public key")
	}

	// 8. Test JWS format (before adding second signature)
	payload, err := vcon.GetSignedPayload()
	if err != nil {
		t.Fatalf("Failed to get signed payload: %v", err)
	}

	if payload == nil {
		t.Error("Signed payload should not be nil")
		return
	}

	sig := vcon.Signatures[0]
	jwsToken := sig.Protected + "." + *payload + "." + sig.Signature

	valid, err = VerifyJWS(jwsToken, restoredPublic)
	if err != nil {
		t.Fatalf("Failed to verify JWS token: %v", err)
	}
	if !valid {
		t.Error("JWS token should be valid")
	}

	// 9. Add second signature
	err = vcon.Sign(privateKey2)
	if err != nil {
		t.Fatalf("Failed to add second signature: %v", err)
	}

	if len(vcon.Signatures) != 2 {
		t.Errorf("Expected 2 signatures, got %d", len(vcon.Signatures))
	}
}

// TestPropertyHandlingWorkflow tests the complete property handling workflow.
func TestPropertyHandlingWorkflow(t *testing.T) {
	// 1. Create vCon with custom properties
	testJSON := `{
		"uuid": "123e4567-e89b-12d3-a456-426614174000",
		"vcon": "0.0.1",
		"created_at": "2023-01-01T00:00:00Z",
		"custom_global": "global_value",
		"parties": [{
			"name": "Test User",
			"tel": "+1234567890",
			"custom_party": "party_value"
		}],
		"dialog": [{
			"type": "text",
			"start": "2023-01-01T00:00:00Z",
			"parties": [0],
			"custom_dialog": "dialog_value"
		}],
		"attachments": [{
			"type": "document",
			"body": "content",
			"encoding": "none",
			"custom_attachment": "attachment_value"
		}]
	}`

	// 2. Test default property handling
	vconDefault, err := BuildFromJSON(testJSON, PropertyHandlingDefault)
	if err != nil {
		t.Fatalf("Failed to build vCon with default handling: %v", err)
	}

	jsonDefault, err := vconDefault.ToJSONString(PropertyHandlingDefault)
	if err != nil {
		t.Fatalf("Failed to serialize with default handling: %v", err)
	}

	if !strings.Contains(jsonDefault, "custom_global") {
		t.Error("Custom properties should be preserved in default mode")
	}

	// 3. Test strict property handling
	vconStrict, err := BuildFromJSON(testJSON, PropertyHandlingStrict)
	if err != nil {
		t.Fatalf("Failed to build vCon with strict handling: %v", err)
	}

	jsonStrict, err := vconStrict.ToJSONString(PropertyHandlingStrict)
	if err != nil {
		t.Fatalf("Failed to serialize with strict handling: %v", err)
	}

	if strings.Contains(jsonStrict, "custom_global") {
		t.Error("Custom properties should be removed in strict mode")
	}

	// 4. Test meta property handling
	vconMeta, err := BuildFromJSON(testJSON, PropertyHandlingMeta)
	if err != nil {
		t.Fatalf("Failed to build vCon with meta handling: %v", err)
	}

	jsonMeta, err := vconMeta.ToJSONString(PropertyHandlingMeta)
	if err != nil {
		t.Fatalf("Failed to serialize with meta handling: %v", err)
	}

	if !strings.Contains(jsonMeta, "meta") {
		t.Error("Meta field should be created in meta mode")
	}

	// 5. Verify all versions have basic structure (legacy format may not pass IETF validation)
	for _, vcon := range []*VCon{vconDefault, vconStrict, vconMeta} {
		if vcon.UUID == uuid.Nil {
			t.Error("vCon should have UUID")
		}
		if vcon.Vcon == "" {
			t.Error("vCon should have version")
		}
		// Note: Legacy vCons may not pass IETF validation due to version requirements
	}
}

// TestHTTPWorkflow tests the complete HTTP workflow.
func TestHTTPWorkflow(t *testing.T) {
	// 1. Create test vCon
	vcon := NewWithDefaults()
	vcon.AddParty(Party{Name: StringPtr("Test User")})
	vcon.AddDialog(Dialog{
		Type:    "text",
		Start:   time.Now(),
		Parties: NewDialogPartiesArrayPtr([]int{0}),
		Body:    "Test message",
	})

	// 2. Mock HTTP client for testing
	testData, err := json.Marshal(vcon)
	if err != nil {
		t.Fatalf("Failed to marshal test vCon: %v", err)
	}

	mockClient := &MockHTTPClient{
		Response: &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(string(testData))),
		},
	}

	// 3. Test loading from URL
	loadedVCon, err := LoadFromURLWithClient("http://example.com/vcon.json", PropertyHandlingDefault, mockClient)
	if err != nil {
		t.Fatalf("Failed to load vCon from URL: %v", err)
	}

	if len(loadedVCon.Parties) != 1 {
		t.Errorf("Expected 1 party in loaded vCon, got %d", len(loadedVCon.Parties))
	}

	// 4. Test posting to URL
	mockClient.Response = &http.Response{
		StatusCode: 201,
		Body:       io.NopCloser(strings.NewReader("Created")),
	}

	headers := map[string]string{
		"Authorization": "Bearer test-token",
		"Content-Type":  "application/json",
	}

	resp, err := vcon.PostToURLWithClient("http://example.com/vcons", headers, PropertyHandlingDefault, mockClient)
	if err != nil {
		t.Fatalf("Failed to post vCon to URL: %v", err)
	}

	if resp.StatusCode != 201 {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	// 5. Test file operations
	tmpFile, err := os.CreateTemp("", "integration_test_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Save to file
	err = vcon.SaveToFile(tmpFile.Name(), PropertyHandlingDefault)
	if err != nil {
		t.Fatalf("Failed to save vCon to file: %v", err)
	}

	// Load from file
	loadedFromFile, err := LoadFromFile(tmpFile.Name(), PropertyHandlingDefault)
	if err != nil {
		t.Fatalf("Failed to load vCon from file: %v", err)
	}

	if loadedFromFile.UUID.String() != vcon.UUID.String() {
		t.Error("UUID should be preserved in file round-trip")
	}
}

// TestAdvancedValidationWorkflow tests the advanced validation features.
func TestAdvancedValidationWorkflow(t *testing.T) {
	// 1. Create vCon with various validation scenarios
	vcon := NewWithDefaults()

	// Add parties
	vcon.AddParty(Party{Name: StringPtr("Customer")})
	vcon.AddParty(Party{Name: StringPtr("Agent")})

	// 2. Test valid dialogs
	vcon.AddDialog(Dialog{
		Type:      "recording",
		Start:     time.Now().Add(-10 * time.Minute),
		Parties:   NewDialogPartiesArrayPtr([]int{0, 1}),
		Mediatype: StringPtr("audio/wav"),
		Body:      "audio data",           // Added required content
		Encoding:  StringPtr("base64url"), // Added required encoding
		Duration:  Float64Ptr(300.0),
	})

	vcon.AddDialog(Dialog{
		Type:        "incomplete",
		Start:       time.Now().Add(-5 * time.Minute),
		Parties:     NewDialogPartiesArrayPtr([]int{0}),
		Disposition: StringPtr("busy"),
	})

	vcon.AddDialog(Dialog{
		Type:  "transfer",
		Start: time.Now(),
		// Removed Parties - transfer dialogs shouldn't have parties
		Transferor:     IntPtr(1),
		Transferee:     IntPtr(0),
		TransferTarget: IntPtr(0), // Added required transfer-target
		Original:       IntPtr(0), // Added required original
		TargetDialog:   IntPtr(1), // Added required target_dialog
	})

	// 3. Add valid attachments
	vcon.AddAttachment(Attachment{
		Type:      StringPtr("transcript"),
		Body:      "Call transcript",
		Encoding:  StringPtr("none"),
		Mediatype: StringPtr("text/plain"), // Added required mediatype
		Start:     func() *time.Time { t := time.Now().UTC(); return &t }(),
		Party:     IntPtr(0),
	})

	// 4. Add valid analysis
	vcon.Analysis = []Analysis{
		{
			Type:     "sentiment",
			Dialog:   float64(0),
			Vendor:   "ai-vendor",
			Body:     "positive",
			Encoding: StringPtr("json"),
		},
	}

	// 5. Test basic validation
	err := vcon.Validate()
	if err != nil {
		t.Errorf("Basic validation should pass: %v", err)
	}

	// 6. Test strict validation
	err = vcon.ValidateStrict()
	if err != nil {
		t.Errorf("Strict validation should pass: %v", err)
	}

	// 7. Test advanced validation
	err = vcon.ValidateAdvanced()
	if err != nil {
		t.Errorf("Advanced validation should pass: %v", err)
	}

	// 8. Test validation with errors
	invalidVCon := NewWithDefaults()
	invalidVCon.AddDialog(Dialog{
		Type:    "incomplete",
		Start:   time.Now(),
		Parties: NewDialogPartiesArrayPtr([]int{0}), // Invalid party index
		// Missing disposition
	})

	err = invalidVCon.ValidateAdvanced()
	if err == nil {
		t.Error("Advanced validation should fail for invalid vCon")
	}
}

// TestPerformanceWorkflow tests performance-related functionality.
func TestPerformanceWorkflow(t *testing.T) {
	// 1. Create a large vCon
	vcon := NewWithDefaults()

	// Add many parties
	for i := 0; i < 100; i++ {
		party := Party{
			Name: StringPtr(fmt.Sprintf("User %d", i)),
			Tel:  StringPtr(fmt.Sprintf("+1234567890%02d", i)),
		}
		vcon.AddParty(party)
	}

	// Add many dialogs
	for i := 0; i < 1000; i++ {
		dialog := Dialog{
			Type:    "text",
			Start:   time.Now().Add(time.Duration(i) * time.Second),
			Parties: NewDialogPartiesArrayPtr([]int{i % 100}), // Cycle through parties
			Body:    fmt.Sprintf("Message %d", i),
		}
		vcon.AddDialog(dialog)
	}

	// 2. Test serialization performance
	start := time.Now()
	jsonData, err := vcon.ToJSON()
	serializationTime := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to serialize large vCon: %v", err)
	}

	t.Logf("Serialization time for large vCon: %v", serializationTime)

	// 3. Test deserialization performance
	start = time.Now()
	var restoredVCon VCon
	err = json.Unmarshal(jsonData, &restoredVCon)
	deserializationTime := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to deserialize large vCon: %v", err)
	}

	t.Logf("Deserialization time for large vCon: %v", deserializationTime)

	// 4. Test validation performance
	start = time.Now()
	valid, _ := restoredVCon.IsValid()
	validationTime := time.Since(start)

	if !valid {
		// Get validation errors for debugging
		_, errors := restoredVCon.IsValid()
		t.Errorf("Large vCon should be valid, errors: %v", errors)
	}

	t.Logf("Validation time for large vCon: %v", validationTime)

	// 5. Verify performance is reasonable (these are generous thresholds)
	if serializationTime > 5*time.Second {
		t.Errorf("Serialization took too long: %v", serializationTime)
	}

	if deserializationTime > 5*time.Second {
		t.Errorf("Deserialization took too long: %v", deserializationTime)
	}

	if validationTime > 10*time.Second {
		t.Errorf("Validation took too long: %v", validationTime)
	}
}

// Helper functions for integration tests.
func Float64Ptr(f float64) *float64 {
	return &f
}
