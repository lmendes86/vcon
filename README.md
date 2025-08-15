# go-vcon

[![Go Reference](https://pkg.go.dev/badge/github.com/yourusername/go-vcon.svg)](https://pkg.go.dev/github.com/yourusername/go-vcon)
[![Go Report Card](https://goreportcard.com/badge/github.com/yourusername/go-vcon)](https://goreportcard.com/report/github.com/yourusername/go-vcon)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Go library for working with vCon (virtual conversation) format - a standardized way to capture, store, and exchange conversation data including voice calls, video calls, text messages, and metadata.

## Features

- ✅ Complete vCon specification compliance
- ✅ JSON marshaling/unmarshaling with proper validation
- ✅ Support for all dialog types: recording, text, transfer, incomplete, video
- ✅ Comprehensive validation with custom error types
- ✅ Extensible design with additional properties support
- ✅ High test coverage (>95%)
- ✅ Zero external dependencies (except UUID)

## Installation

```bash
go get github.com/yourusername/go-vcon
```

## Quick Start

```go
package main

import (
	"fmt"
	"time"

	"github.com/yourusername/go-vcon/vcon"
)

func main() {
	// Create a new vCon
	v := vcon.NewWithDefaults()
	
	// Add a participant
	email := "alice@example.com"
	name := "Alice Smith"
	aliceIdx := v.AddParty(vcon.Party{
		Mailto: &email,
		Name:   &name,
	})
	
	// Add a text dialog
	v.AddDialog(vcon.Dialog{
		Type:    "text",
		Start:   time.Now(),
		Parties: []int{aliceIdx},
		Body:    "Hello, this is a test message",
	})
	
	// Convert to JSON
	jsonData, err := v.ToJSONIndent()
	if err != nil {
		panic(err)
	}
	
	fmt.Println(string(jsonData))
}
```

## Core Types

### VCon
The top-level container that holds all conversation data:

```go
type VCon struct {
	UUID        *uuid.UUID    `json:"uuid"`
	Vcon        string        `json:"vcon"`         // Version
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   *time.Time    `json:"updated_at,omitempty"`
	Subject     *string       `json:"subject,omitempty"`
	Parties     []Party       `json:"parties,omitempty"`
	Dialog      []Dialog      `json:"dialog,omitempty"`
	Attachments []Attachment  `json:"attachments,omitempty"`
	// ... additional fields
}
```

### Party
Represents conversation participants:

```go
type Party struct {
	Tel          *string      `json:"tel,omitempty"`          // Phone number
	Mailto       *string      `json:"mailto,omitempty"`       // Email address
	Name         *string      `json:"name,omitempty"`         // Display name
	UUID         *uuid.UUID   `json:"uuid,omitempty"`         // Participant UUID
	CivicAddress *CivicAddress `json:"civicaddress,omitempty"` // Physical address
	// ... additional fields
}
```

### Dialog
Represents conversation segments:

```go
type Dialog struct {
	Type     string    `json:"type"`      // "recording", "text", "transfer", "incomplete", "video"
	Start    time.Time `json:"start"`     // When this segment started
	Parties  []int     `json:"parties"`   // Participant indices
	Body     any       `json:"body,omitempty"`     // Content (text, etc.)
	Duration *float64  `json:"duration,omitempty"` // Length in seconds
	// ... type-specific fields
}
```

## Usage Examples

### Working with Text Messages

```go
v := vcon.NewWithDefaults()

// Add participants
alice := v.AddParty(vcon.Party{
	Name: stringPtr("Alice"),
	Tel:  stringPtr("+1234567890"),
})

bob := v.AddParty(vcon.Party{
	Name: stringPtr("Bob"),
	Tel:  stringPtr("+0987654321"),
})

// Add text messages
v.AddDialog(vcon.Dialog{
	Type:    "text",
	Start:   time.Now(),
	Parties: []int{alice},
	Body:    "Hi Bob!",
})

v.AddDialog(vcon.Dialog{
	Type:    "text",
	Start:   time.Now().Add(30 * time.Second),
	Parties: []int{bob},
	Body:    "Hey Alice, how are you?",
})
```

### Working with Audio Recordings

```go
v := vcon.NewWithDefaults()

// Add participants
participants := []int{0, 1} // Assuming parties already added

// Add recording dialog
mimetype := "audio/wav"
duration := 120.5
v.AddDialog(vcon.Dialog{
	Type:     "recording",
	Start:    time.Now(),
	Parties:  participants,
	Mimetype: &mimetype,
	Duration: &duration,
	URL:      stringPtr("https://example.com/recording.wav"),
})
```

### Working with Video Calls

```go
// Add video dialog with metadata
resolution := "1920x1080"
frameRate := 30.0
bitrate := 2500
v.AddDialog(vcon.Dialog{
	Type:       "video",
	Start:      time.Now(),
	Parties:    []int{0, 1},
	Resolution: &resolution,
	FrameRate:  &frameRate,
	Bitrate:    &bitrate,
	Duration:   floatPtr(1800), // 30 minutes
})
```

### Adding Attachments

```go
// Add a document attachment
v.AddAttachment(vcon.Attachment{
	Type:     "document",
	Body:     "SGVsbG8gV29ybGQ=", // "Hello World" in base64
	Encoding: "base64",
	Metadata: map[string]any{
		"filename": "document.pdf",
		"size":     1024,
	},
})
```

## Encoding and Decoding

### JSON Marshaling

```go
// Marshal to compact JSON
data, err := v.ToJSON()

// Marshal to indented JSON
data, err := v.ToJSONIndent()

// Use package-level functions
data, err := vcon.Marshal(v)
data, err := vcon.MarshalIndent(v)
```

### JSON Unmarshaling

```go
// From JSON bytes
v, err := vcon.FromJSON(jsonData)

// From string
v, err := vcon.ParseString(jsonString)

// From reader
v, err := vcon.ReadFrom(reader)
v, err := vcon.Decode(reader)

// With validation
v, err := vcon.DecodeAndValidate(reader)
```

### Streaming

```go
// Write to stream
err := vcon.Encode(writer, v)
err := vcon.EncodeCompact(writer, v)

// Read from stream
v, err := vcon.Decode(reader)
```

## Validation

The library provides comprehensive validation:

```go
// Basic validation
err := v.Validate()
if err != nil {
	fmt.Printf("Validation failed: %v\n", err)
}

// Strict validation (additional business rules)
err = v.ValidateStrict()

// Check if valid
if v.IsValid() {
	fmt.Println("vCon is valid")
}

// JSON validation
err = vcon.ValidateJSON(jsonData)
```

### Validation Features

- Required field validation
- Data type and format validation
- Cross-reference validation (party indices, etc.)
- Chronological order validation
- Duplicate detection
- Custom error types with detailed messages

## Error Handling

The library uses custom error types for detailed error information:

```go
err := v.Validate()
if err != nil {
	switch e := err.(type) {
	case vcon.ValidationErrors:
		// Multiple validation errors
		for _, validationErr := range e {
			fmt.Printf("Field %s: %s\n", validationErr.Field, validationErr.Message)
		}
	case vcon.ValidationError:
		// Single validation error
		fmt.Printf("Field %s: %s\n", e.Field, e.Message)
	default:
		// Other errors
		fmt.Printf("Error: %v\n", err)
	}
}
```

## Testing

Run the test suite:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run benchmarks
go test -bench=. ./...
```

## Development

### Prerequisites

- Go 1.21 or later
- Make (optional, for using Makefile)

### Building

```bash
# Build the library
go build ./...

# Run tests
go test ./...

# Run linter
golangci-lint run

# Format code
go fmt ./...
```

### Quality Tools

Install development tools:

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install golang.org/x/tools/cmd/godoc@latest
```

Run quality checks:

```bash
# Linting
golangci-lint run ./...

# Security scan
gosec ./...

# Generate documentation
godoc -http=:6060
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for your changes
5. Ensure all tests pass (`go test ./...`)
6. Run linting (`golangci-lint run ./...`)
7. Commit your changes (`git commit -m 'Add amazing feature'`)
8. Push to the branch (`git push origin feature/amazing-feature`)
9. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Specification

This library implements the vCon specification. For more information about the vCon format, see:

- [vCon Specification](https://datatracker.ietf.org/doc/draft-ietf-vcon-vcon-container/)
- [CDDL Schema](schema/vcon.cddl)

## Roadmap

- [ ] Support for additional encoding formats
- [ ] vCon validation against CDDL schema
- [ ] CLI tools for vCon manipulation
- [ ] Streaming parser for large vCons
- [ ] Integration with common telephony systems

## Support

- 📖 [Documentation](https://pkg.go.dev/github.com/yourusername/go-vcon)
- 🐛 [Issue Tracker](https://github.com/yourusername/go-vcon/issues)
- 💬 [Discussions](https://github.com/yourusername/go-vcon/discussions)

---

*Note: Replace `yourusername` with your actual GitHub username before publishing.*