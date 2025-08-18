package vcon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// ToJSON marshals the VCon to JSON bytes.
func (v *VCon) ToJSON() ([]byte, error) {
	return json.Marshal(v)
}

// ToJSONIndent marshals the VCon to indented JSON bytes for better readability.
func (v *VCon) ToJSONIndent() ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

// WriteTo writes the VCon as JSON to the provided writer.
// It implements the io.WriterTo interface.
func (v *VCon) WriteTo(w io.Writer) (int64, error) {
	data, err := v.ToJSON()
	if err != nil {
		return 0, fmt.Errorf("failed to marshal vcon: %w", err)
	}

	n, err := w.Write(data)
	return int64(n), err
}

// WriteToIndent writes the VCon as indented JSON to the provided writer.
func (v *VCon) WriteToIndent(w io.Writer) (int64, error) {
	data, err := v.ToJSONIndent()
	if err != nil {
		return 0, fmt.Errorf("failed to marshal vcon: %w", err)
	}

	n, err := w.Write(data)
	return int64(n), err
}

// String returns the VCon as a JSON string.
func (v *VCon) String() string {
	data, err := v.ToJSONIndent()
	if err != nil {
		return fmt.Sprintf("error marshaling vcon: %v", err)
	}
	return string(data)
}

// Encode writes the VCon to a writer using a JSON encoder.
// This is more efficient for streaming large VCons.
func Encode(w io.Writer, v *VCon) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

// EncodeCompact writes the VCon to a writer using a JSON encoder without indentation.
func EncodeCompact(w io.Writer, v *VCon) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(v)
}

// Marshal is a convenience function that marshals a VCon to JSON bytes.
func Marshal(v *VCon) ([]byte, error) {
	return v.ToJSON()
}

// MarshalIndent is a convenience function that marshals a VCon to indented JSON bytes.
func MarshalIndent(v *VCon) ([]byte, error) {
	return v.ToJSONIndent()
}

// CompactJSON removes unnecessary whitespace from JSON data.
func CompactJSON(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	if err := json.Compact(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to compact JSON: %w", err)
	}
	return buf.Bytes(), nil
}
