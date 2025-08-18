package vcon

import (
	"encoding/json"
	"fmt"
)

// DialogParties represents the flexible parties field in a Dialog.
// It can be:
// - A single integer (party index)
// - An array of integers (multiple parties in single channel)
// - An array of arrays of integers (multi-channel with parties per channel).
type DialogParties struct {
	// Internal representation - stores the actual value
	value interface{}
}

// NewDialogPartiesSingle creates a DialogParties with a single party index.
func NewDialogPartiesSingle(index int) DialogParties {
	return DialogParties{value: index}
}

// NewDialogPartiesArray creates a DialogParties with an array of party indices.
func NewDialogPartiesArray(indices []int) DialogParties {
	return DialogParties{value: indices}
}

// NewDialogPartiesArrayPtr creates a pointer to DialogParties with an array of party indices.
// This is a convenience function for use in struct literals.
func NewDialogPartiesArrayPtr(indices []int) *DialogParties {
	if len(indices) == 0 {
		return nil
	}
	dp := NewDialogPartiesArray(indices)
	return &dp
}

// NewDialogPartiesMultiChannel creates a DialogParties with multi-channel party indices.
func NewDialogPartiesMultiChannel(channels [][]int) DialogParties {
	return DialogParties{value: channels}
}

// IsSingle returns true if this represents a single party index.
func (dp *DialogParties) IsSingle() bool {
	if dp == nil || dp.value == nil {
		return false
	}
	_, ok := dp.value.(int)
	return ok
}

// IsArray returns true if this represents an array of party indices.
func (dp *DialogParties) IsArray() bool {
	if dp == nil || dp.value == nil {
		return false
	}
	_, ok := dp.value.([]int)
	return ok
}

// IsMultiChannel returns true if this represents multi-channel party indices.
func (dp *DialogParties) IsMultiChannel() bool {
	if dp == nil || dp.value == nil {
		return false
	}
	_, ok := dp.value.([][]int)
	return ok
}

// GetSingle returns the single party index if applicable.
func (dp *DialogParties) GetSingle() (int, bool) {
	if dp == nil || dp.value == nil {
		return 0, false
	}
	val, ok := dp.value.(int)
	return val, ok
}

// GetArray returns the array of party indices if applicable.
func (dp *DialogParties) GetArray() ([]int, bool) {
	if dp == nil || dp.value == nil {
		return nil, false
	}
	val, ok := dp.value.([]int)
	return val, ok
}

// GetMultiChannel returns the multi-channel party indices if applicable.
func (dp *DialogParties) GetMultiChannel() ([][]int, bool) {
	if dp == nil || dp.value == nil {
		return nil, false
	}
	val, ok := dp.value.([][]int)
	return val, ok
}

// GetAllPartyIndices returns all unique party indices regardless of structure.
func (dp *DialogParties) GetAllPartyIndices() []int {
	if dp == nil || dp.value == nil {
		return []int{}
	}

	uniqueIndices := make(map[int]bool)

	switch v := dp.value.(type) {
	case int:
		uniqueIndices[v] = true
	case []int:
		for _, idx := range v {
			uniqueIndices[idx] = true
		}
	case [][]int:
		for _, channel := range v {
			for _, idx := range channel {
				uniqueIndices[idx] = true
			}
		}
	}

	// Convert map to slice
	result := make([]int, 0, len(uniqueIndices))
	for idx := range uniqueIndices {
		result = append(result, idx)
	}
	return result
}

// MarshalJSON implements json.Marshaler interface.
func (dp DialogParties) MarshalJSON() ([]byte, error) {
	if dp.value == nil {
		return []byte("null"), nil
	}
	return json.Marshal(dp.value)
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (dp *DialogParties) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		dp.value = nil
		return nil
	}

	// Try to unmarshal as single int first
	var singleInt int
	if err := json.Unmarshal(data, &singleInt); err == nil {
		dp.value = singleInt
		return nil
	}

	// Try to unmarshal as array of ints
	var intArray []int
	if err := json.Unmarshal(data, &intArray); err == nil {
		dp.value = intArray
		return nil
	}

	// Try to unmarshal as array of arrays of ints
	var multiChannel [][]int
	if err := json.Unmarshal(data, &multiChannel); err == nil {
		dp.value = multiChannel
		return nil
	}

	// Try as array of any to detect mixed types
	var anyArray []interface{}
	if err := json.Unmarshal(data, &anyArray); err == nil {
		// Check if it's an array that could be multi-channel
		allArrays := true
		for _, item := range anyArray {
			if _, ok := item.([]interface{}); !ok {
				allArrays = false
				break
			}
		}

		if allArrays {
			// Convert to [][]int
			multiChannel := make([][]int, len(anyArray))
			for i, item := range anyArray {
				if arr, ok := item.([]interface{}); ok {
					channel := make([]int, len(arr))
					for j, val := range arr {
						if num, ok := val.(float64); ok {
							channel[j] = int(num)
						} else {
							return fmt.Errorf("invalid party index type in multi-channel array")
						}
					}
					multiChannel[i] = channel
				}
			}
			dp.value = multiChannel
			return nil
		}

		// Try to convert to []int
		intArray := make([]int, len(anyArray))
		for i, item := range anyArray {
			if num, ok := item.(float64); ok {
				intArray[i] = int(num)
			} else {
				return fmt.Errorf("invalid party index type in array")
			}
		}
		dp.value = intArray
		return nil
	}

	return fmt.Errorf("parties field must be int, []int, or [][]int")
}

// Validate checks if the DialogParties contains valid party indices.
func (dp *DialogParties) Validate(totalParties int) error {
	if dp == nil || dp.value == nil {
		return nil // Parties field is optional
	}

	indices := dp.GetAllPartyIndices()
	for _, idx := range indices {
		if idx < 0 || idx >= totalParties {
			return fmt.Errorf("invalid party index %d (must be 0-%d)", idx, totalParties-1)
		}
	}

	return nil
}
