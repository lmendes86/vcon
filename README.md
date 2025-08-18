# vcon

[![Go Reference](https://pkg.go.dev/badge/github.com/lmendes86/vcon.svg)](https://pkg.go.dev/github.com/lmendes86/vcon)
[![Go Report Card](https://goreportcard.com/badge/github.com/lmendes86/vcon)](https://goreportcard.com/report/github.com/lmendes86/vcon)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Go library for working with vCon (virtual conversation) format - a standardized way to capture, store, and exchange conversation data including voice calls, video calls, text messages, and metadata.

🎯 **IETF Specification Compliant** - Fully implements [draft-ietf-vcon-vcon-container-03](https://datatracker.ietf.org/doc/draft-ietf-vcon-vcon-container/) version 0.0.2 with strict compliance validation

## Features

### Core Features

- ✅ **IETF vCon Specification v0.0.2** - Full compliance with latest IETF draft
- ✅ **Strict IETF Validation** - `ValidateIETF()` for specification compliance
- ✅ **Migration Support** - Convert legacy vCons to IETF format
- ✅ **Compliance Assessment** - Check compliance level (Full/Partial/Non-compliant)
- ✅ **JSON marshaling/unmarshaling** with proper validation
- ✅ **IETF Dialog Types**: recording, text, transfer, incomplete (video removed per spec)
- ✅ **Flexible Analysis Types** - Standard types + custom types per IETF "SHOULD" language
- ✅ **HTTPS Security** - External URLs must use HTTPS per spec
- ✅ **RFC3339 Timestamps** - Strict timestamp format validation
- ✅ **Party Validation** - At least one identifier required (tel, mailto, name, uuid)
- ✅ **Dialog Content Validation** - Recording/text MUST have content, incomplete/transfer MUST NOT
- ✅ **Parties Field Required** - Parties array is required per IETF specification
- ✅ **Comprehensive Test Coverage** (89.0%+)
- ✅ **Zero external dependencies** (except UUID)

### Enhanced Features

- 🔐 **Digital Signatures (JWS)** - Sign and verify vCons with RSA keys
- 🌐 **HTTP Utilities** - Load from URLs, post to endpoints, file I/O operations
- ⚙️ **Property Handling** - Three modes: default, strict, meta for non-standard fields
- 🏷️ **Tag Management** - Organized tagging system with attachment-based storage
- 🔍 **Search Utilities** - Find parties, dialogs, attachments with advanced search
- 📚 **Extension Management** - Dynamic extension and must-support field handling
- 🎯 **Specialized Dialogs** - Helper functions for transfer and incomplete dialogs
- 🔬 **Multi-Tier Validation Architecture** - Unified validation interface with 5 levels
- 🏗️ **Clean Architecture** - Modular validation with low cyclomatic complexity (<30)
- ⚡ **Performance Optimized** - Efficient validation with focused method responsibilities
- ✅ **Strict IETF Content Rules** - Enforces dialog content requirements per specification
- 🎯 **Enhanced Error Messages** - Detailed IETF specification references in validation errors

## Installation

```bash
go get github.com/lmendes86/vcon
```

## Quick Start

```go
package main

import (
	"fmt"
	"time"

	"github.com/lmendes86/vcon"
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

## IETF Compliance

This library fully implements the [IETF vCon specification](https://datatracker.ietf.org/doc/draft-ietf-vcon-vcon-container/) version 0.0.2.

### Multi-Tier Validation System

```go
// Create IETF-compliant vCon (uses version "0.0.2")
vcon := vcon.NewWithDefaults()
vcon.AddParty(vcon.Party{Name: vcon.StringPtr("Test User")})

// Unified validation interface with 5 levels
result := vcon.ValidateWithLevel(vcon.ValidationBasic)
if !result.Valid {
    fmt.Printf("Validation errors: %v\n", result.Errors)
}

// Available validation levels
result = vcon.ValidateWithLevel(vcon.ValidationBasic)       // Basic business rules
result = vcon.ValidateWithLevel(vcon.ValidationStrict)      // Strict business rules
result = vcon.ValidateWithLevel(vcon.ValidationIETF)        // IETF compliance
result = vcon.ValidateWithLevel(vcon.ValidationIETFStrict)  // IETF + extension detection
result = vcon.ValidateWithLevel(vcon.ValidationComplete)    // All validation layers

// Check compliance level
level, err := vcon.CheckIETFCompliance()
switch level {
case vcon.IETFFullyCompliant:
    fmt.Println("✅ Fully IETF compliant")
case vcon.IETFPartiallyCompliant:
    fmt.Println("⚠️ Partially compliant")
case vcon.IETFNonCompliant:
    fmt.Println("❌ Not compliant")
}

// Legacy validation methods still available
if err := vcon.ValidateIETF(); err != nil {
    fmt.Printf("IETF validation errors: %v\n", err)
}
```

### Migration from Legacy Format

```go
// Migrate legacy vCon to IETF compliance
err := vcon.MigrateVConToIETF(legacyVCon, vcon.MigrationModeLenient)

// Convert legacy Analysis maps to new struct format
analyses, err := vcon.ConvertFromLegacyAnalysis(legacyMaps, vcon.MigrationModePreserve)
```

### Implementation Notes: IETF Specification Ambiguities

**Conservative Compliance Approach**: This library takes a conservative, practical approach to IETF vCon draft-03 compliance, prioritizing real-world interoperability over literal interpretation of ambiguous specification language.

#### UUID Field Requirement

**Our Implementation**: UUIDs are **required** in all vCon objects.

**IETF Specification Ambiguity**: The IETF vCon draft-03 specification contains internal inconsistencies regarding the UUID field:

- Section 4.1.2 marks UUID as `"String" (optional)`
- Other sections include UUID in required field lists
- The specification states "The UUID MUST be globally unique" (implying presence)
- Signed and encrypted forms "SHOULD" include UUIDs

**Our Rationale**: We require UUIDs because:

- **Practical Necessity**: UUIDs are essential for vCon identification, referencing, and security operations
- **Real-World Usage**: Most vCon implementations expect UUIDs to be present
- **Conservative Safety**: When specifications are ambiguous, stricter compliance ensures broader interoperability
- **Security Best Practices**: Many cryptographic and integrity operations depend on reliable UUID presence

#### Analysis Type Field

**Our Implementation**: Analysis `type` field is **required** and allows both standard types (`summary`, `transcript`, `translation`, `sentiment`, `tts`) and custom types per IETF "SHOULD" language.

**IETF Specification**: Uses "SHOULD be one of" language, which we interpret as a strong recommendation with allowance for custom types.

**Custom Analysis Types**: Fully supported per IETF "SHOULD" language. Examples include:
- **Standard Types**: `summary`, `transcript`, `translation`, `sentiment`, `tts`  
- **Custom Types**: `emotion_detection`, `keyword_extraction`, `voice_analytics`, `compliance_check`, `ai_insights`
- **Vendor-Specific**: Any meaningful analysis type identifier

**Interoperability**: This implementation ensures maximum compatibility with other IETF vCon implementations that may use custom analysis types.

**Note**: This implementation prioritizes **practical interoperability** and **security** over strict literal compliance with ambiguous specification language. For production use, having required UUIDs and typed analysis objects ensures reliable vCon handling across different systems.

### Key IETF Requirements

- **Version**: Must be "0.0.2"
- **Parties Field**: Required array (cannot be empty)
- **Party Identifiers**: At least one of `tel`, `mailto`, `name`, or `uuid` required
- **Dialog Types**: Only `recording`, `text`, `transfer`, `incomplete` (video removed per spec)
- **Dialog Content Rules**:
  - Recording/text dialogs MUST have content (body or URL)
  - Incomplete/transfer dialogs MUST NOT have content
- **Incomplete Dialogs**: Must have `disposition` field
- **External URLs**: Must use HTTPS
- **Timestamps**: Must be RFC3339 format
- **Field Names**: `mediatype` (not `mimetype`) per specification
- **Analysis Structure**: Structured `Analysis` type with required `vendor`, `type`, `encoding` per IETF spec
- **MIME Types**: Accept any valid MIME format per RFC2046

## Core Types

### VCon

The top-level container that holds all conversation data:

```go
type VCon struct {
	UUID        uuid.UUID     `json:"uuid"`                 // Required per IETF spec
	Vcon        string        `json:"vcon"`                 // Version
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   *time.Time    `json:"updated_at,omitempty"`
	Subject     *string       `json:"subject,omitempty"`
	Parties     []Party       `json:"parties"`              // Required per IETF spec
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
	Type     string    `json:"type"`      // "recording", "text", "transfer", "incomplete" per IETF spec
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

// Add recording dialog with required content
mediatype := "audio/wav"
duration := 120.5
v.AddDialog(vcon.Dialog{
	Type:      "recording",
	Start:     time.Now(),
	Parties:   participants,
	Mediatype: &mediatype,  // Note: mediatype not mimetype per IETF spec
	Duration:  &duration,
	URL:       stringPtr("https://example.com/recording.wav"), // Required content
})
```

### Working with Video Recordings

```go
// Add video recording dialog (use "recording" type with video mediatype per IETF spec)
mediatype := "video/mp4"
resolution := "1920x1080"
frameRate := 30.0
bitrate := 2500
v.AddDialog(vcon.Dialog{
	Type:       "recording",  // Use "recording" type for video per IETF spec
	Start:      time.Now(),
	Parties:    []int{0, 1},
	Mediatype:  &mediatype,   // Video content specified via mediatype
	URL:        stringPtr("https://example.com/video.mp4"), // Required content
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

## Enhanced Features

### Digital Signatures (JWS)

Sign and verify vCons using RSA keys with JSON Web Signature (JWS) standard:

```go
// Generate RSA key pair
privateKey, publicKey, err := vcon.GenerateKeyPair()
if err != nil {
	panic(err)
}

// Sign a vCon
v := vcon.NewWithDefaults()
v.AddParty(vcon.Party{Name: stringPtr("Test User")})
err = v.Sign(privateKey)
if err != nil {
	panic(err)
}

// Verify signature
valid, err := v.Verify(publicKey)
if err != nil {
	panic(err)
}
fmt.Printf("Signature valid: %t\n", valid)

// Check if vCon is signed
if v.IsSigned() {
	fmt.Printf("vCon has %d signatures\n", len(v.Signatures))
}

// Convert keys to PEM format
privatePEM, err := vcon.PrivateKeyToPEM(privateKey)
publicPEM, err := vcon.PublicKeyToPEM(publicKey)

// Load keys from PEM
loadedPrivateKey, err := vcon.PrivateKeyFromPEM(privatePEM)
loadedPublicKey, err := vcon.PublicKeyFromPEM(publicPEM)
```

### HTTP Utilities

Load vCons from URLs, post to endpoints, and handle file operations:

```go
// Load from URL
v, err := vcon.LoadFromURL("https://example.com/vcon.json", vcon.PropertyHandlingDefault)

// Load from file
v, err := vcon.LoadFromFile("/path/to/vcon.json", vcon.PropertyHandlingDefault)

// Save to file
err = v.SaveToFile("/path/to/output.json", vcon.PropertyHandlingDefault)

// Post to URL with headers
headers := map[string]string{
	"Authorization": "Bearer token123",
	"X-Custom":      "custom-value",
}
resp, err := v.PostToURL("https://api.example.com/vcons", headers, vcon.PropertyHandlingDefault)

// Convert to JSON strings
jsonStr, err := v.ToJSONString(vcon.PropertyHandlingDefault)
indentedStr, err := v.ToJSONIndentString(vcon.PropertyHandlingDefault)

// Validate JSON strings and files
valid, errors := vcon.ValidateJSONString(jsonString)
valid, errors = vcon.ValidateVConFile("/path/to/vcon.json")
```

### Property Handling

Handle non-standard properties in three different modes:

```go
// Default mode - strict schema compliance
v, err := vcon.BuildFromJSON(jsonData, vcon.PropertyHandlingDefault)

// Strict mode - reject unknown properties
v, err := vcon.BuildFromJSON(jsonData, vcon.PropertyHandlingStrict)

// Meta mode - preserve unknown properties in meta field
v, err := vcon.BuildFromJSON(jsonData, vcon.PropertyHandlingMeta)

// Build new vCon with property handling
v := vcon.BuildNew(vcon.PropertyHandlingMeta)
```

### Extension and Must-Support Management

Manage vCon extensions and must-support requirements:

```go
v := vcon.NewWithDefaults()

// Extension management
v.AddExtension("video")
v.AddExtension("encryption")
extensions := v.GetExtensions() // ["video", "encryption"]
v.RemoveExtension("video")

// Must-support management
v.AddMustSupport("encryption")
v.AddMustSupport("signatures")
mustSupport := v.GetMustSupport() // ["encryption", "signatures"]
v.RemoveMustSupport("encryption")
```

### Tag Management System

Organize vCons with a flexible tagging system:

```go
v := vcon.NewWithDefaults()

// Add tags
v.AddTag("category", "customer_support")
v.AddTag("priority", "high")
v.AddTag("department", "sales")

// Retrieve tags
category := v.GetTag("category")     // *string "customer_support"
priority := v.GetTag("priority")     // *string "high"
missing := v.GetTag("non_existent")  // nil

// Get tags attachment (where tags are stored)
tagsAttachment := v.GetTags() // *Attachment with type "tags"
```

### Search Utilities

Find parties, dialogs, and attachments efficiently:

```go
v := vcon.NewWithDefaults()

// Add some data
v.AddParty(vcon.Party{Name: stringPtr("John"), Tel: stringPtr("+1234567890")})
v.AddParty(vcon.Party{Name: stringPtr("Jane"), Mailto: stringPtr("jane@example.com")})

// Find parties by field
nameIndex := v.FindPartyIndex("name", "John")     // *int 0
telIndex := v.FindPartyIndex("tel", "+1234567890") // *int 0
emailIndex := v.FindPartyIndex("mailto", "jane@example.com") // *int 1

// Find dialogs by type
textDialogs := v.FindDialogsByType("text")
recordingDialogs := v.FindDialogsByType("recording")

// Find attachments by type
metadata := v.FindAttachmentByType("metadata")
```

### Specialized Dialog Creators

Helper functions for specific dialog types:

```go
v := vcon.NewWithDefaults()
v.AddParty(vcon.Party{Name: stringPtr("Caller")})
v.AddParty(vcon.Party{Name: stringPtr("Agent")})

// Create transfer dialog
transferData := map[string]any{
	"reason": "Escalation to supervisor",
	"from":   "+1234567890",
	"to":     "+1987654321",
}
metadata := map[string]any{"system": "PBX"}
index := v.AddTransferDialog(time.Now(), transferData, []int{0, 1}, metadata)

// Create incomplete dialog
incompleteDetails := map[string]any{
	"ringDuration": 45000,
	"reason":       "No answer",
}
index = v.AddIncompleteDialog(time.Now(), "NO_ANSWER", incompleteDetails, []int{0}, metadata)

// Add analysis data
v.AddAnalysis(
	"sentiment",                    // type
	[]int{0, 1},                   // dialog references
	"acme_analytics",              // vendor
	map[string]any{"score": 0.8},  // body
	"json",                        // encoding
	map[string]any{"version": "2.0"}, // schema
	map[string]any{"model": "v3"},    // meta
)

// Find analysis by type
analysis := v.FindAnalysisByType("sentiment")

// Add custom analysis types (fully supported per IETF spec)
v.AddAnalysis(
	"emotion_detection",            // custom type
	[]int{0, 1},                   // dialog references
	"EmotionAI Corp",              // vendor
	map[string]any{               // analysis result
		"primary_emotion": "happy",
		"confidence":      0.95,
		"secondary":       []string{"excited", "satisfied"},
	},
	"json",                        // encoding
	map[string]any{"model": "v4.2"}, // schema
	map[string]any{"lang": "en"},    // meta
)

// Custom analysis types are fully compatible with IETF validation
result := v.ValidateWithLevel(vcon.ValidationIETF)
if result.Valid {
	fmt.Println("✅ Custom analysis types pass IETF validation")
}
```

### Multi-Tier Validation Architecture

Enhanced validation with unified interface and modular architecture:

```go
v := vcon.NewWithDefaults()
v.AddParty(vcon.Party{Name: vcon.StringPtr("Test User")})

// Unified validation interface
result := v.ValidateWithLevel(vcon.ValidationComplete)
if !result.Valid {
	for _, err := range result.Errors {
		fmt.Printf("Validation error in %s: %s\n", err.Field, err.Message)
	}
}

// Test validation level progression
basicResult := v.ValidateWithLevel(vcon.ValidationBasic)
strictResult := v.ValidateWithLevel(vcon.ValidationStrict)
ietfResult := v.ValidateWithLevel(vcon.ValidationIETF)

// Legacy validation methods still available
valid, errors := v.IsValid()
err := v.ValidateAdvanced()
err = v.ValidateIETF()
err = v.ValidateStrict()
```

**Validation Architecture:**

- **Multi-Tier System**: 5 validation levels (Basic → Strict → IETF → IETF-Strict → Complete)
- **Unified Interface**: Single `ValidateWithLevel()` method for all validation types
- **Modular Architecture**: Validation split into focused, maintainable files
- **Low Complexity**: Each validation method has low cyclomatic complexity (<30)
- **Progressive Validation**: Higher levels include all lower level checks
- **Comprehensive Coverage**: Business rules, IETF compliance, and extension detection
- **IETF Content Rules**: Enforces recording/text MUST have content, incomplete/transfer MUST NOT
- **Parties Validation**: Ensures parties array is present and not empty per IETF spec
- **Field Name Compliance**: Validates `mediatype` field name per IETF specification
- **Analysis Type Flexibility**: Supports both standard and custom analysis types per IETF "SHOULD" language
- **Clear Error Messages**: Detailed error reporting with IETF specification references
- **Performance Optimized**: Efficient validation with minimal overhead

**Validation Files Structure:**

- `validation_business.go` → Business logic validation
- `validation_ietf_compliance.go` → IETF specification compliance
- `validation_ietf_extensions.go` → Extension field detection
- `validation_common.go` → Unified validation interface

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

The library provides comprehensive multi-tier validation:

```go
// Unified validation interface (recommended)
result := v.ValidateWithLevel(vcon.ValidationComplete)
if !result.Valid {
	fmt.Printf("Validation failed: %v\n", result.Errors)
}

// Legacy validation methods
err := v.Validate()          // Basic business rules
err = v.ValidateStrict()     // Strict business rules
err = v.ValidateIETF()       // IETF specification compliance
err = v.ValidateAdvanced()   // Advanced business rules

// Check if valid (legacy)
if valid, _ := v.IsValid(); valid {
	fmt.Println("vCon is valid")
}

// JSON validation
err = vcon.ValidateJSON(jsonData)
```

### Validation Features

- **Multi-Tier Architecture**: 5 progressive validation levels with unified interface
- **Required Field Validation**: Ensures all mandatory fields are present
- **Data Type and Format Validation**: Validates data types and formats (RFC3339, MIME types)
- **Cross-Reference Validation**: Validates party indices and dialog references
- **Chronological Order Validation**: Ensures dialogs are in proper sequence
- **Duplicate Detection**: Prevents duplicate UUIDs and references
- **Custom Error Types**: Detailed error messages with field-specific feedback
- **Modular Architecture**: Low-complexity validation methods for maintainability
- **IETF Compliance**: Full validation against IETF vCon specification draft-03 v0.0.2
- **Extension Detection**: Identifies non-standard fields with smart migration suggestions
- **Enhanced Content Validation**: Strict dialog content requirements per IETF specification
- **Migration Support**: Validate legacy vCons and guide migration to IETF format

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

- Go 1.24 or later (required for latest features)
- Make (optional, for using Makefile)
- golangci-lint (for code quality checks)

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
# Linting (passes with 0 issues)
golangci-lint run ./...

# Security scan
gosec ./...

# Generate documentation
godoc -http=:6060

# Coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**Quality Metrics:**

- ✅ **89.0%+ Test Coverage** - Comprehensive test suite including IETF compliance tests
- ✅ **Zero Linting Issues** - Clean code with optimized cyclomatic complexity (<30)
- ✅ **IETF Draft-03 Compliant** - Full specification compliance with strict validation
- ✅ **Multi-Tier Validation** - 5-level validation architecture with unified interface
- ✅ **Modular Architecture** - Well-organized validation files with clear separation of concerns
- ✅ **Enhanced Content Validation** - Enforces dialog content requirements
- ✅ **Production Ready** - Battle-tested with real-world scenarios

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

## Examples

Comprehensive examples are available in the [examples/](examples/) directory:

### Available Examples

- **[basic_usage.go](examples/basic_usage.go)** - Basic vCon creation and usage
- **[call_center.go](examples/call_center.go)** - Call center conversation example
- **[sms_conversation.go](examples/sms_conversation.go)** - SMS/text message example
- **[enhanced_features_demo.go](examples/enhanced_features_demo.go)** - All enhanced features demonstration
- **[digital_signatures_demo.go](examples/digital_signatures_demo.go)** - Complete digital signature workflow
- **[ietf_compliant_vcon.go](examples/ietf_compliant_vcon.go)** - IETF specification compliance demonstration
- **[customer_service_call.go](examples/customer_service_call.go)** - Real-world customer service scenario

### Running Examples

The examples are designed to be run after installing the library:

```bash
# Install the library first
go get github.com/lmendes86/vcon

# Copy examples to your project directory and run
cp examples/enhanced_features_demo.go /path/to/your/project/
go run enhanced_features_demo.go

# Or run the basic examples directly (they use relative imports)
go run examples/basic_usage.go
go run examples/call_center.go
go run examples/sms_conversation.go
```

**Note**: The enhanced examples (`enhanced_features_demo.go` and `digital_signatures_demo.go`) require the library to be installed as they demonstrate the full API. The basic examples can be run directly as they use relative imports.

## Specification

This library implements the IETF vCon specification draft-03. For more information about the vCon format, see:

- [IETF vCon Specification Draft-03](https://datatracker.ietf.org/doc/draft-ietf-vcon-vcon-container/)
- [Specification Document](https://www.ietf.org/archive/id/draft-ietf-vcon-vcon-container-03.txt)

## Support

- 📖 [Documentation](https://pkg.go.dev/github.com/lmendes86/vcon)
- 🐛 [Issue Tracker](https://github.com/lmendes86/vcon/issues)
- 💬 [Discussions](https://github.com/lmendes86/vcon/discussions)
