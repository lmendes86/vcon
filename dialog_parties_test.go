package vcon

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestDialogParties_JSON(t *testing.T) {
	tests := []struct {
		name     string
		parties  DialogParties
		expected string
	}{
		{
			name:     "single integer",
			parties:  NewDialogPartiesSingle(0),
			expected: "0",
		},
		{
			name:     "array of integers",
			parties:  NewDialogPartiesArray([]int{0, 1, 2}),
			expected: "[0,1,2]",
		},
		{
			name:     "multi-channel array",
			parties:  NewDialogPartiesMultiChannel([][]int{{0, 1}, {2}}),
			expected: "[[0,1],[2]]",
		},
		{
			name:     "empty value",
			parties:  DialogParties{value: nil},
			expected: "null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.parties)
			if err != nil {
				t.Errorf("Unexpected error during marshaling: %v", err)
				return
			}

			if string(data) != tt.expected {
				t.Errorf("Expected JSON %s, got %s", tt.expected, string(data))
			}
		})
	}
}

func TestDialogParties_JSONUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected DialogParties
		hasError bool
	}{
		{
			name:     "single integer",
			json:     "0",
			expected: NewDialogPartiesSingle(0),
		},
		{
			name:     "array of integers",
			json:     "[0,1,2]",
			expected: NewDialogPartiesArray([]int{0, 1, 2}),
		},
		{
			name:     "multi-channel array",
			json:     "[[0,1],[2]]",
			expected: NewDialogPartiesMultiChannel([][]int{{0, 1}, {2}}),
		},
		{
			name:     "null value",
			json:     "null",
			expected: DialogParties{value: nil},
		},
		{
			name:     "empty array",
			json:     "[]",
			expected: NewDialogPartiesArray([]int{}),
		},
		{
			name:     "single element array",
			json:     "[5]",
			expected: NewDialogPartiesArray([]int{5}),
		},
		{
			name:     "invalid string",
			json:     "\"invalid\"",
			hasError: true,
		},
		{
			name:     "invalid object",
			json:     "{\"invalid\": true}",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var parties DialogParties
			err := json.Unmarshal([]byte(tt.json), &parties)

			if tt.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(parties.value, tt.expected.value) {
				t.Errorf("Expected value %v, got %v", tt.expected.value, parties.value)
			}
		})
	}
}

func TestDialogParties_TypeChecks(t *testing.T) {
	singleParty := NewDialogPartiesSingle(0)
	arrayParty := NewDialogPartiesArray([]int{0, 1})
	multiChannelParty := NewDialogPartiesMultiChannel([][]int{{0}, {1}})
	nilParty := DialogParties{value: nil}

	// Test IsSingle
	if !singleParty.IsSingle() {
		t.Error("Expected single party to return true for IsSingle()")
	}
	if arrayParty.IsSingle() {
		t.Error("Expected array party to return false for IsSingle()")
	}
	if multiChannelParty.IsSingle() {
		t.Error("Expected multi-channel party to return false for IsSingle()")
	}
	if nilParty.IsSingle() {
		t.Error("Expected nil party to return false for IsSingle()")
	}

	// Test IsArray
	if singleParty.IsArray() {
		t.Error("Expected single party to return false for IsArray()")
	}
	if !arrayParty.IsArray() {
		t.Error("Expected array party to return true for IsArray()")
	}
	if multiChannelParty.IsArray() {
		t.Error("Expected multi-channel party to return false for IsArray()")
	}
	if nilParty.IsArray() {
		t.Error("Expected nil party to return false for IsArray()")
	}

	// Test IsMultiChannel
	if singleParty.IsMultiChannel() {
		t.Error("Expected single party to return false for IsMultiChannel()")
	}
	if arrayParty.IsMultiChannel() {
		t.Error("Expected array party to return false for IsMultiChannel()")
	}
	if !multiChannelParty.IsMultiChannel() {
		t.Error("Expected multi-channel party to return true for IsMultiChannel()")
	}
	if nilParty.IsMultiChannel() {
		t.Error("Expected nil party to return false for IsMultiChannel()")
	}
}

func TestDialogParties_GetMethods(t *testing.T) {
	singleParty := NewDialogPartiesSingle(5)
	arrayParty := NewDialogPartiesArray([]int{1, 2, 3})
	multiChannelParty := NewDialogPartiesMultiChannel([][]int{{0, 1}, {2}})

	// Test GetSingle
	if val, ok := singleParty.GetSingle(); !ok || val != 5 {
		t.Errorf("Expected GetSingle() to return (5, true), got (%d, %t)", val, ok)
	}
	if _, ok := arrayParty.GetSingle(); ok {
		t.Error("Expected GetSingle() to return false for array party")
	}

	// Test GetArray
	if _, ok := singleParty.GetArray(); ok {
		t.Error("Expected GetArray() to return false for single party")
	}
	if val, ok := arrayParty.GetArray(); !ok || !reflect.DeepEqual(val, []int{1, 2, 3}) {
		t.Errorf("Expected GetArray() to return correct array, got %v", val)
	}

	// Test GetMultiChannel
	if _, ok := singleParty.GetMultiChannel(); ok {
		t.Error("Expected GetMultiChannel() to return false for single party")
	}
	if val, ok := multiChannelParty.GetMultiChannel(); !ok || !reflect.DeepEqual(val, [][]int{{0, 1}, {2}}) {
		t.Errorf("Expected GetMultiChannel() to return correct array, got %v", val)
	}
}

func TestDialogParties_GetAllPartyIndices(t *testing.T) {
	tests := []struct {
		name     string
		parties  DialogParties
		expected []int
	}{
		{
			name:     "single party",
			parties:  NewDialogPartiesSingle(5),
			expected: []int{5},
		},
		{
			name:     "array party",
			parties:  NewDialogPartiesArray([]int{1, 2, 3}),
			expected: []int{1, 2, 3},
		},
		{
			name:     "multi-channel party",
			parties:  NewDialogPartiesMultiChannel([][]int{{0, 1}, {1, 2}}),
			expected: []int{0, 1, 2}, // Should be unique
		},
		{
			name:     "nil party",
			parties:  DialogParties{value: nil},
			expected: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.parties.GetAllPartyIndices()

			// Sort both slices for comparison since order doesn't matter for uniqueness
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d indices, got %d", len(tt.expected), len(result))
				return
			}

			// Check that all expected indices are present
			resultMap := make(map[int]bool)
			for _, idx := range result {
				resultMap[idx] = true
			}

			for _, expected := range tt.expected {
				if !resultMap[expected] {
					t.Errorf("Expected index %d not found in result %v", expected, result)
				}
			}
		})
	}
}

func TestDialogParties_Validate(t *testing.T) {
	tests := []struct {
		name         string
		parties      DialogParties
		totalParties int
		expectError  bool
		errorMsg     string
	}{
		{
			name:         "valid single party",
			parties:      NewDialogPartiesSingle(0),
			totalParties: 2,
			expectError:  false,
		},
		{
			name:         "valid array parties",
			parties:      NewDialogPartiesArray([]int{0, 1}),
			totalParties: 3,
			expectError:  false,
		},
		{
			name:         "valid multi-channel parties",
			parties:      NewDialogPartiesMultiChannel([][]int{{0}, {1, 2}}),
			totalParties: 3,
			expectError:  false,
		},
		{
			name:         "invalid negative index",
			parties:      NewDialogPartiesSingle(-1),
			totalParties: 2,
			expectError:  true,
			errorMsg:     "invalid party index -1",
		},
		{
			name:         "invalid high index",
			parties:      NewDialogPartiesArray([]int{0, 5}),
			totalParties: 3,
			expectError:  true,
			errorMsg:     "invalid party index 5",
		},
		{
			name:         "nil parties",
			parties:      DialogParties{value: nil},
			totalParties: 2,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.parties.Validate(tt.totalParties)

			if tt.expectError {
				if err == nil {
					t.Error("Expected validation error but got none")
					return
				}
				if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got: %s", tt.errorMsg, err.Error())
				}
			} else if err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
		})
	}
}

func TestNewDialogPartiesArrayPtr(t *testing.T) {
	// Test with non-empty array
	parties := NewDialogPartiesArrayPtr([]int{0, 1, 2})
	if parties == nil {
		t.Error("Expected non-nil pointer for non-empty array")
		return
	}

	expected := []int{0, 1, 2}
	if result, ok := parties.GetArray(); !ok || !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected array %v, got %v", expected, result)
	}

	// Test with empty array
	emptyParties := NewDialogPartiesArrayPtr([]int{})
	if emptyParties != nil {
		t.Error("Expected nil pointer for empty array")
	}
}

func TestDialogParties_Integration(t *testing.T) {
	// Test that DialogParties works in Dialog struct with JSON
	dialog := Dialog{
		Type:    "recording",
		Parties: NewDialogPartiesArrayPtr([]int{0, 1}),
	}

	// Marshal to JSON
	data, err := json.Marshal(dialog)
	if err != nil {
		t.Errorf("Error marshaling dialog: %v", err)
		return
	}

	// Unmarshal from JSON
	var unmarshaled Dialog
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Errorf("Error unmarshaling dialog: %v", err)
		return
	}

	// Check that parties are correctly preserved
	if unmarshaled.Parties == nil {
		t.Error("Expected parties to be preserved after JSON round-trip")
		return
	}

	result := unmarshaled.Parties.GetAllPartyIndices()
	expected := []int{0, 1}

	if len(result) != len(expected) {
		t.Errorf("Expected %d parties, got %d", len(expected), len(result))
		return
	}

	// Check indices (order doesn't matter)
	resultMap := make(map[int]bool)
	for _, idx := range result {
		resultMap[idx] = true
	}

	for _, expectedIdx := range expected {
		if !resultMap[expectedIdx] {
			t.Errorf("Expected party index %d not found", expectedIdx)
		}
	}
}

// Tests from dialog_parties_additional_test.go

func TestDialogPartiesUnmarshalJSONEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		jsonData    string
		expectError bool
		description string
	}{
		{
			name:        "invalid boolean value",
			jsonData:    `true`,
			expectError: true,
			description: "boolean values should be rejected",
		},
		{
			name:        "invalid float value",
			jsonData:    `3.14`,
			expectError: true,
			description: "float values should be rejected",
		},
		{
			name:        "array with invalid string element",
			jsonData:    `["not-a-number"]`,
			expectError: true,
			description: "string elements in array should be rejected",
		},
		{
			name:        "array with mixed valid and invalid types",
			jsonData:    `[1, "invalid", 2]`,
			expectError: true,
			description: "arrays with mixed types should be rejected",
		},
		{
			name:        "array with nested arrays",
			jsonData:    `[[1, 2], [3, 4]]`,
			expectError: false,
			description: "nested arrays should be handled (multi-channel)",
		},
		{
			name:        "deeply nested array structure",
			jsonData:    `[[[1]], [[2]]]`,
			expectError: true,
			description: "deeply nested arrays should be rejected",
		},
		{
			name:        "array with null elements",
			jsonData:    `[1, null, 2]`,
			expectError: false, // null elements are actually handled
			description: "null elements in array are handled",
		},
		{
			name:        "object with invalid structure",
			jsonData:    `{"invalid": "object"}`,
			expectError: true,
			description: "objects should be rejected",
		},
		{
			name:        "large valid integer",
			jsonData:    `999999`,
			expectError: false,
			description: "large integers should be handled",
		},
		{
			name:        "valid multi-channel with different sizes",
			jsonData:    `[[1, 2, 3], [4], [5, 6]]`,
			expectError: false,
			description: "multi-channel arrays of different sizes should work",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var dp DialogParties
			err := json.Unmarshal([]byte(tt.jsonData), &dp)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", tt.description)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for %s: %v", tt.description, err)
				}
			}
		})
	}
}

func TestDialogPartiesUnmarshalJSONComplexTypes(t *testing.T) {
	// Test with complex JSON that exercises different code paths
	complexTests := []struct {
		name     string
		jsonData string
		validate func(*testing.T, DialogParties)
	}{
		{
			name:     "multi-channel with empty channel",
			jsonData: `[[1, 2], [], [3]]`,
			validate: func(t *testing.T, dp DialogParties) {
				if !dp.IsMultiChannel() {
					t.Error("Should be multi-channel")
				}
				channels, ok := dp.GetMultiChannel()
				if !ok {
					t.Error("Failed to get multi-channel data")
					return
				}
				if len(channels) != 3 {
					t.Errorf("Expected 3 channels, got %d", len(channels))
				}
				if len(channels[1]) != 0 {
					t.Errorf("Expected empty channel, got %v", channels[1])
				}
			},
		},
		{
			name:     "single large number",
			jsonData: `999999999`,
			validate: func(t *testing.T, dp DialogParties) {
				if dp.IsSingle() {
					party, ok := dp.GetSingle()
					if !ok {
						t.Error("Failed to get single party")
						return
					}
					if party != 999999999 {
						t.Errorf("Expected 999999999, got %d", party)
					}
				} else {
					t.Error("Should be single party")
				}
			},
		},
	}

	for _, tt := range complexTests {
		t.Run(tt.name, func(t *testing.T) {
			var dp DialogParties
			err := json.Unmarshal([]byte(tt.jsonData), &dp)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			tt.validate(t, dp)
		})
	}
}

// Uses existing helper functions from other test files
