package vcon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// HTTPClient interface for HTTP operations (for testing).
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// DefaultHTTPClient is the default HTTP client used for requests.
var DefaultHTTPClient HTTPClient = &http.Client{
	Timeout: 30 * time.Second,
}

// LoadFromURL loads a vCon from a URL.
func LoadFromURL(url string, propertyHandling string) (*VCon, error) {
	return LoadFromURLWithClient(url, propertyHandling, DefaultHTTPClient)
}

// LoadFromURLWithClient loads a vCon from a URL using a custom HTTP client.
func LoadFromURLWithClient(url string, propertyHandling string, client HTTPClient) (*VCon, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "go-vcon/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch vCon: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return BuildFromJSON(string(body), propertyHandling)
}

// BuildFromJSON creates a VCon from JSON string with property handling.
func BuildFromJSON(jsonStr string, propertyHandling string) (*VCon, error) {
	if propertyHandling == "" {
		propertyHandling = PropertyHandlingDefault
	}

	var vcon VCon

	if propertyHandling != PropertyHandlingDefault {
		handler := NewPropertyHandler(propertyHandling)
		err := handler.UnmarshalVConWithPropertyHandling([]byte(jsonStr), &vcon)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal vCon with property handling: %w", err)
		}
	} else {
		err := json.Unmarshal([]byte(jsonStr), &vcon)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal vCon: %w", err)
		}
	}

	return &vcon, nil
}

// BuildNew creates a new VCon with specified property handling.
func BuildNew(_ string) *VCon {
	vcon := NewWithDefaults()
	// Initialize empty slices to match Python implementation
	vcon.Group = []GroupObject{}
	vcon.Parties = []Party{}
	vcon.Dialog = []Dialog{}
	vcon.Attachments = []Attachment{}
	vcon.Analysis = []Analysis{}
	// Note: Extensions and MustSupport are deprecated (not in IETF draft-03) - do not initialize
	vcon.Signatures = []Signature{}
	// Appended and Redacted are now single objects, not collections - leave as nil
	vcon.Appended = nil
	vcon.Redacted = nil
	// Note: Top-level Meta field removed for IETF draft-03 compliance
	// Use object-specific meta fields (party.Meta, dialog.Meta, etc.) instead

	return vcon
}

// LoadFromFile loads a vCon from a file with property handling.
func LoadFromFile(filePath string, propertyHandling string) (*VCon, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return BuildFromJSON(string(data), propertyHandling)
}

// SaveToFile saves the vCon to a JSON file.
func (v *VCon) SaveToFile(filePath string, propertyHandling string) error {
	var data []byte
	var err error

	if propertyHandling != "" && propertyHandling != PropertyHandlingDefault {
		handler := NewPropertyHandler(propertyHandling)
		data, err = handler.MarshalVConWithPropertyHandling(v)
	} else {
		data, err = json.MarshalIndent(v, "", "  ")
	}

	if err != nil {
		return fmt.Errorf("failed to marshal vCon: %w", err)
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// PostToURL posts the vCon as JSON to a URL with optional headers.
func (v *VCon) PostToURL(url string, headers map[string]string, propertyHandling string) (*http.Response, error) {
	return v.PostToURLWithClient(url, headers, propertyHandling, DefaultHTTPClient)
}

// PostToURLWithClient posts the vCon using a custom HTTP client.
func (v *VCon) PostToURLWithClient(url string, headers map[string]string, propertyHandling string, client HTTPClient) (*http.Response, error) {
	var data []byte
	var err error

	if propertyHandling != "" && propertyHandling != PropertyHandlingDefault {
		handler := NewPropertyHandler(propertyHandling)
		data, err = handler.MarshalVConWithPropertyHandling(v)
	} else {
		data, err = json.Marshal(v)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to marshal vCon: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set default headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "go-vcon/1.0")

	// Add custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to post vCon: %w", err)
	}

	return resp, nil
}

// ToJSONString converts the vCon to a JSON string with property handling.
func (v *VCon) ToJSONString(propertyHandling string) (string, error) {
	var data []byte
	var err error

	if propertyHandling != "" && propertyHandling != PropertyHandlingDefault {
		handler := NewPropertyHandler(propertyHandling)
		data, err = handler.MarshalVConWithPropertyHandling(v)
	} else {
		data, err = v.ToJSON()
	}

	if err != nil {
		return "", fmt.Errorf("failed to marshal vCon: %w", err)
	}

	return string(data), nil
}

// ToJSONIndentString converts the vCon to an indented JSON string with property handling.
func (v *VCon) ToJSONIndentString(propertyHandling string) (string, error) {
	var data []byte
	var err error

	if propertyHandling != "" && propertyHandling != PropertyHandlingDefault {
		handler := NewPropertyHandler(propertyHandling)
		rawData, err := handler.MarshalVConWithPropertyHandling(v)
		if err != nil {
			return "", err
		}

		// Re-marshal with indentation
		var temp interface{}
		if err := json.Unmarshal(rawData, &temp); err != nil {
			return "", err
		}
		data, err = json.MarshalIndent(temp, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal indented JSON: %w", err)
		}
	} else {
		data, err = v.ToJSONIndent()
		if err != nil {
			return "", fmt.Errorf("failed to marshal vCon: %w", err)
		}
	}

	return string(data), nil
}

// Dumps is an alias for ToJSONString for Python compatibility.
func (v *VCon) Dumps(propertyHandling string) (string, error) {
	return v.ToJSONString(propertyHandling)
}

// ValidateJSONString validates a vCon from a JSON string.
func ValidateJSONString(jsonStr string) (bool, []string) {
	vcon, err := BuildFromJSON(jsonStr, PropertyHandlingDefault)
	if err != nil {
		return false, []string{fmt.Sprintf("Failed to parse JSON: %v", err)}
	}

	return vcon.IsValid()
}

// ValidateVConFile validates a vCon file.
func ValidateVConFile(filePath string) (bool, []string) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return false, []string{fmt.Sprintf("Failed to read file: %v", err)}
	}

	return ValidateJSONString(string(data))
}
