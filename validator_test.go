package vcon

import (
	"encoding/json"
	"testing"
	"time"
)

func TestJSONValidationErrorsMarshalJSON(t *testing.T) {
	// Test the MarshalJSON method (0% coverage)
	vcon := &VCon{
		// Missing required fields to trigger validation errors
	}

	err := ValidateStruct(vcon)
	if err == nil {
		t.Fatal("Expected validation errors for invalid VCon")
	}

	jsonErr, ok := err.(JSONValidationErrors)
	if !ok {
		t.Fatalf("Expected JSONValidationErrors, got %T", err)
	}

	data, marshalErr := jsonErr.MarshalJSON()
	if marshalErr != nil {
		t.Fatalf("Failed to marshal JSONValidationErrors: %v", marshalErr)
	}

	// Verify the JSON structure
	var result map[string]interface{}
	unmarshalErr := json.Unmarshal(data, &result)
	if unmarshalErr != nil {
		t.Fatalf("Failed to unmarshal JSONValidationErrors JSON: %v", unmarshalErr)
	}

	// Should contain validation error details
	if len(result) == 0 {
		t.Error("Expected validation errors in JSON result")
	}
}

func TestToJSONValidationErrors(t *testing.T) {
	// Test the ToJSONValidationErrors function
	vcon := &VCon{
		// Missing required fields to trigger validation errors
	}

	// Get validation errors
	structErr := SchemaValidator.Struct(vcon)
	if structErr == nil {
		t.Fatal("Expected validation errors for invalid VCon")
	}

	// Convert to JSON validation errors
	jsonErr := ToJSONValidationErrors(structErr)
	if jsonErr == nil {
		t.Fatal("Expected non-nil error from ToJSONValidationErrors")
	}

	// Verify it's the correct type
	_, ok := jsonErr.(JSONValidationErrors)
	if !ok {
		t.Errorf("Expected JSONValidationErrors, got %T", jsonErr)
	}

	// Test with non-validation error
	nonValidationErr := &ValidationError{Field: "test", Message: "test"}
	result := ToJSONValidationErrors(nonValidationErr)
	if result != nonValidationErr {
		t.Error("Non-validation errors should be returned as-is")
	}
}

func TestValidateStructJSON(t *testing.T) {
	// Test the ValidateStructJSON function (0% coverage)

	// Test valid struct
	validVCon := NewWithDefaults()
	validVCon.AddParty(Party{Name: StringPtr("Test User")})

	data, err := ValidateStructJSON(validVCon)
	if err != nil {
		t.Errorf("Expected no error for valid VCon, got: %v", err)
	}
	if data != nil {
		t.Error("Expected nil data for valid struct")
	}

	// Test invalid struct
	invalidVCon := &VCon{
		// Missing required fields
	}

	data, err = ValidateStructJSON(invalidVCon)
	if err != nil {
		t.Errorf("ValidateStructJSON should not return error, got: %v", err)
	}
	if data == nil {
		t.Error("Expected validation error data for invalid struct")
	}

	// Verify the returned data is valid JSON
	if data != nil {
		var result map[string]interface{}
		unmarshalErr := json.Unmarshal(data, &result)
		if unmarshalErr != nil {
			t.Errorf("Returned data is not valid JSON: %v", unmarshalErr)
		}
	}

	// Test with nil struct
	data, err = ValidateStructJSON(nil)
	if err != nil {
		t.Errorf("ValidateStructJSON should not return error for nil, got: %v", err)
	}
	if data == nil {
		t.Error("Expected validation error data for nil struct")
	}
}

func TestValidateStructJSONWithDifferentTypes(t *testing.T) {
	// Test ValidateStructJSON with different struct types

	// Test with valid Party struct
	validParty := &Party{Name: StringPtr("Test User")}
	data, err := ValidateStructJSON(validParty)
	if err != nil {
		t.Errorf("Expected no error for valid Party, got: %v", err)
	}
	if data != nil {
		t.Error("Expected nil data for valid Party")
	}

	// Test with invalid Party (missing required identifier)
	invalidParty := &Party{} // No identifiers
	_, err = ValidateStructJSON(invalidParty)
	if err != nil {
		t.Errorf("ValidateStructJSON should not return error, got: %v", err)
	}
	// Note: This might not trigger validation errors since Party validation is done by business logic, not struct tags

	// Test with Attachment struct
	validAttachment := &Attachment{
		Type:     StringPtr("document"),
		Body:     "content",
		Encoding: StringPtr("none"),
		Start:    func() *time.Time { t := time.Now().UTC(); return &t }(),
		Party:    IntPtr(1),
	}
	data, err = ValidateStructJSON(validAttachment)
	if err != nil {
		t.Errorf("Expected no error for valid Attachment, got: %v", err)
	}
	if data != nil {
		t.Error("Expected nil data for valid Attachment")
	}

	// Test with invalid Attachment (missing required fields)
	invalidAttachment := &Attachment{} // Missing required fields
	data, err = ValidateStructJSON(invalidAttachment)
	if err != nil {
		t.Errorf("ValidateStructJSON should not return error, got: %v", err)
	}
	if data == nil {
		t.Error("Expected validation error data for invalid Attachment")
	}
}

func TestJSONValidationErrorsRoundTrip(t *testing.T) {
	// Test that JSONValidationErrors can be marshaled and unmarshaled correctly
	invalidVCon := &VCon{} // Missing required fields

	err := ValidateStruct(invalidVCon)
	if err == nil {
		t.Fatal("Expected validation errors for invalid VCon")
	}

	jsonErr, ok := err.(JSONValidationErrors)
	if !ok {
		t.Fatalf("Expected JSONValidationErrors, got %T", err)
	}

	// Marshal to JSON
	data, marshalErr := jsonErr.MarshalJSON()
	if marshalErr != nil {
		t.Fatalf("Failed to marshal JSONValidationErrors: %v", marshalErr)
	}

	// Verify it's valid JSON
	var result map[string]interface{}
	unmarshalErr := json.Unmarshal(data, &result)
	if unmarshalErr != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", unmarshalErr)
	}

	// Should contain validation error details
	if len(result) == 0 {
		t.Error("Expected validation errors in JSON result")
	}
}

func TestValidatorStructIntegration(t *testing.T) {
	// Test the integration between validator functions and struct validation

	// Create a VCon with validation errors
	vcon := &VCon{
		// Missing required UUID
		Vcon: "", // Invalid empty version
		// Missing required CreatedAt
	}

	// Test struct validation
	err := ValidateStruct(vcon)
	if err == nil {
		t.Error("Expected validation errors for invalid VCon")
	}

	// Test ValidateStructJSON
	data, jsonErr := ValidateStructJSON(vcon)
	if jsonErr != nil {
		t.Errorf("ValidateStructJSON should not return error, got: %v", jsonErr)
	}
	if data == nil {
		t.Error("Expected validation error data for invalid VCon")
	}

	// Verify the JSON result is valid
	if data != nil {
		var result map[string]interface{}
		parseErr := json.Unmarshal(data, &result)
		if parseErr != nil {
			t.Errorf("Invalid JSON result: %v", parseErr)
		}
	}
}
