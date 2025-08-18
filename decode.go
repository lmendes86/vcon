package vcon

import (
	"encoding/json"
	"fmt"
	"io"
)

// FromJSON unmarshals a VCon from JSON bytes.
func FromJSON(data []byte) (*VCon, error) {
	var v VCon
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, fmt.Errorf("failed to unmarshal vcon: %w", err)
	}
	return &v, nil
}

// ReadFrom reads and unmarshals a VCon from the provided reader.
// It implements a similar pattern to io.ReaderFrom.
func ReadFrom(r io.Reader) (*VCon, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}
	return FromJSON(data)
}

// Decode reads a VCon from a reader using a JSON decoder.
// This is more efficient for streaming large VCons.
func Decode(r io.Reader) (*VCon, error) {
	var v VCon
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&v); err != nil {
		return nil, fmt.Errorf("failed to decode vcon: %w", err)
	}
	return &v, nil
}

// Unmarshal is a convenience function that unmarshals a VCon from JSON bytes.
func Unmarshal(data []byte) (*VCon, error) {
	return FromJSON(data)
}

// ParseString parses a VCon from a JSON string.
func ParseString(s string) (*VCon, error) {
	return FromJSON([]byte(s))
}

// ValidateJSON performs basic JSON validation on the provided data.
// It checks if the data is valid JSON and can be unmarshaled into a VCon.
func ValidateJSON(data []byte) error {
	var v VCon
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("invalid vcon JSON: %w", err)
	}
	return nil
}

// DecodeAndValidate decodes a VCon from a reader and performs validation.
func DecodeAndValidate(r io.Reader) (*VCon, error) {
	v, err := Decode(r)
	if err != nil {
		return nil, err
	}

	// Perform basic validation
	if err := v.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return v, nil
}
