package main

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/lmendes86/vcon"
)

// StringPtr returns a pointer to the given string.
func StringPtr(s string) *string {
	return &s
}

// demonstrateIETFCompliantVCon shows how to create and validate IETF-compliant vCon.
func demonstrateIETFCompliantVCon() {
	fmt.Println("=== IETF-Compliant vCon Creation ===")

	// Create a new IETF-compliant vCon
	// NewWithDefaults() now uses version "0.0.2" per IETF spec
	conv := vcon.NewWithDefaults()

	// Add parties with proper identifiers (at least one required per IETF)
	aliceIdx := conv.AddParty(vcon.Party{
		Name:   func() *string { s := "Alice Johnson"; return &s }(),
		Mailto: func() *string { s := "alice@example.com"; return &s }(),
		UUID:   func() *uuid.UUID { id := uuid.New(); return &id }(),
	})

	bobIdx := conv.AddParty(vcon.Party{
		Name: func() *string { s := "Bob Smith"; return &s }(),
		Tel:  func() *string { s := "+15550123"; return &s }(), // Valid E.164 format
	})

	// Add a text dialog
	textDialog := vcon.Dialog{
		Type:      "text",
		Start:     time.Now().UTC(),
		Parties:   vcon.NewDialogPartiesArrayPtr([]int{aliceIdx, bobIdx}),
		Body:      "Hello Bob, how are you today?",
		Mediatype: func() *string { s := "text/plain"; return &s }(),
		Encoding:  func() *string { s := "none"; return &s }(), // Required for inline content
	}
	conv.AddDialog(textDialog)

	// Add a recording dialog with IETF-compliant fields
	recordingDialog := vcon.Dialog{
		Type:        "recording",
		Start:       time.Now().UTC().Add(1 * time.Minute),
		Duration:    func() *float64 { f := 120.5; return &f }(), // 2 minutes 30 seconds
		Parties:     vcon.NewDialogPartiesArrayPtr([]int{aliceIdx, bobIdx}),
		Originator:  func() *int { i := aliceIdx; return &i }(),
		Mediatype:   func() *string { s := "audio/wav"; return &s }(),
		Filename:    func() *string { s := "conversation.wav"; return &s }(),
		URL:         func() *string { s := "https://example.com/recordings/conv123.wav"; return &s }(),                                           // HTTPS required
		ContentHash: vcon.NewContentHashSingle("sha-512:MJ7MSJwS1utMxA9QyQLytNDtd-5RGnx6m808qG1M2G-YndNbxf9JlnDaNCVbRbDP2DDoH2Bdz33FVC6TrpzXbw"), // Hash of external file in IETF format
	}
	conv.AddDialog(recordingDialog)

	// Add an incomplete dialog with required disposition
	incompleteDialog := vcon.Dialog{
		Type:        "incomplete",
		Start:       time.Now().UTC().Add(5 * time.Minute),
		Parties:     vcon.NewDialogPartiesArrayPtr([]int{aliceIdx}),
		Disposition: func() *string { s := "hung-up"; return &s }(), // Required for incomplete type - IETF compliant value
	}
	conv.AddDialog(incompleteDialog)

	// Add IETF-compliant analysis using new struct format
	sentimentAnalysis := vcon.Analysis{
		Type:     "sentiment",
		Dialog:   0, // Reference to text dialog
		Vendor:   "SentimentCorp",
		Product:  func() *string { s := "SentimentAnalyzer v2.1"; return &s }(),
		Schema:   func() *string { s := "https://sentimentcorp.com/schema/v2.1"; return &s }(),
		Body:     map[string]interface{}{"score": 0.8, "classification": "positive"},
		Encoding: StringPtr("json"),
		Meta:     map[string]interface{}{"confidence": "high", "model_version": "2.1.3"},
	}
	conv.Analysis = append(conv.Analysis, sentimentAnalysis)

	// Add transcript analysis for recording
	transcriptAnalysis := vcon.Analysis{
		Type:     "transcript",
		Dialog:   1, // Reference to recording dialog
		Vendor:   "TranscriptAI",
		Body:     "Alice: Hello Bob, how are you today?\nBob: I'm doing well, thank you!",
		Encoding: StringPtr("none"),
	}
	conv.Analysis = append(conv.Analysis, transcriptAnalysis)

	// Add IETF-compliant attachment (Type is optional)
	attachment := vcon.Attachment{
		Type:      func() *string { s := "metadata"; return &s }(),
		Body:      map[string]interface{}{"call_quality": "HD", "duration_actual": 125.3},
		Encoding:  StringPtr("json"),
		Mediatype: func() *string { s := "application/json"; return &s }(),  // Required for inline content
		Start:     func() *time.Time { t := time.Now().UTC(); return &t }(), // Optional field per IETF spec
		Party:     func() *int { i := 0; return &i }(),                      // Required field per IETF spec - reference to first party
		Dialog:    func() *int { i := 1; return &i }(),                      // Related to recording dialog
	}
	conv.AddAttachment(attachment)

	fmt.Printf("Created vCon with ID: %s\n", conv.UUID.String())
	fmt.Printf("Version: %s (IETF compliant)\n", conv.Vcon)
	fmt.Printf("Parties: %d\n", len(conv.Parties))
	fmt.Printf("Dialogs: %d\n", len(conv.Dialog))
	fmt.Printf("Analysis: %d\n", len(conv.Analysis))
	fmt.Printf("Attachments: %d\n", len(conv.Attachments))

	// Validate IETF compliance using new multi-tier validation system
	fmt.Println("\n=== Multi-Tier Validation System ===")

	// Demonstrate the 5 validation levels
	levels := []vcon.ValidationLevel{
		vcon.ValidationBasic,
		vcon.ValidationStrict,
		vcon.ValidationIETF,
		vcon.ValidationIETFStrict,
		vcon.ValidationComplete,
	}

	for _, level := range levels {
		result := conv.ValidateWithLevel(level)
		if result.Valid {
			fmt.Printf("✅ %s validation: PASSED\n", result.Level.String())
		} else {
			fmt.Printf("❌ %s validation: FAILED (%d errors)\n", result.Level.String(), len(result.Errors))
			for _, err := range result.Errors {
				fmt.Printf("   - %s: %s\n", err.Field, err.Message)
			}
		}
	}

	// Check compliance level (legacy method)
	fmt.Println("\n=== Legacy Compliance Check ===")
	level, err := conv.CheckIETFCompliance()
	if err != nil {
		fmt.Printf("Compliance issues found: %v\n", err)
	}

	switch level {
	case vcon.IETFFullyCompliant:
		fmt.Println("✅ vCon is fully IETF compliant!")
	case vcon.IETFPartiallyCompliant:
		fmt.Println("⚠️  vCon is partially IETF compliant")
	case vcon.IETFNonCompliant:
		fmt.Println("❌ vCon is not IETF compliant")
	}

	// Perform strict IETF validation (legacy method)
	if err := conv.ValidateIETF(); err != nil {
		fmt.Printf("IETF validation errors: %v\n", err)
	} else {
		fmt.Println("✅ Passed strict IETF validation!")
	}

	// Demonstrate JSON serialization
	fmt.Println("\n=== JSON Serialization ===")
	jsonData, err := conv.ToJSONIndent()
	if err != nil {
		log.Fatalf("Failed to serialize vCon: %v", err)
	}

	maxLen := len(jsonData)
	if maxLen > 500 {
		maxLen = 500
	}
	fmt.Printf("vCon JSON (first 500 chars):\n%s...\n", string(jsonData[:maxLen]))
}

// demonstrateMigration shows how to migrate legacy vCon to IETF format.
func demonstrateMigration() {
	fmt.Println("\n=== Legacy vCon Migration ===")

	// Create a legacy-style vCon (version 0.0.1, map-based analysis)
	conv := vcon.New("0.0.1") // Old version

	// Add party without proper identifier
	conv.AddParty(vcon.Party{
		// No name, tel, mailto, or uuid - not IETF compliant
	})

	// Add incomplete dialog without disposition
	conv.AddDialog(vcon.Dialog{
		Type:    "incomplete",
		Start:   time.Now().UTC(),
		Parties: vcon.NewDialogPartiesArrayPtr([]int{0}),
		// Missing disposition - not IETF compliant
	})

	fmt.Printf("Legacy vCon version: %s\n", conv.Vcon)

	// Check compliance before migration
	level, err := conv.CheckIETFCompliance()
	fmt.Printf("Pre-migration compliance: %v, error: %v\n", level, err)

	// Migrate to IETF compliance using lenient mode
	fmt.Println("Migrating to IETF compliance...")
	err = vcon.MigrateVConToIETF(conv, vcon.MigrationModeLenient)
	if err != nil {
		fmt.Printf("Migration failed: %v\n", err)
		return
	}

	fmt.Printf("Post-migration version: %s\n", conv.Vcon)

	// Check compliance after migration
	level, err = conv.CheckIETFCompliance()
	fmt.Printf("Post-migration compliance: %v, error: %v\n", level, err)

	// Show what was automatically fixed
	if len(conv.Parties) > 0 && conv.Parties[0].Name != nil {
		fmt.Printf("✅ Added default name to party: %s\n", *conv.Parties[0].Name)
	}

	if len(conv.Dialog) > 0 && conv.Dialog[0].Disposition != nil {
		fmt.Printf("✅ Added default disposition to incomplete dialog: %s\n", *conv.Dialog[0].Disposition)
	}
}

// demonstrateValidationModes shows different validation approaches.
func demonstrateValidationModes() {
	fmt.Println("\n=== Validation Architecture Comparison ===")

	// Create a vCon with potential issues for testing validation levels
	conv := vcon.NewWithDefaults()
	conv.AddParty(vcon.Party{Name: func() *string { s := "Alice"; return &s }()})

	// Add recording with custom MIME type and required content
	conv.AddDialog(vcon.Dialog{
		Type:        "recording",
		Start:       time.Now().UTC(),
		Parties:     vcon.NewDialogPartiesArrayPtr([]int{0}),
		Mediatype:   func() *string { s := "audio/x-custom-format"; return &s }(),                     // Custom but valid MIME
		URL:         func() *string { s := "https://example.com/demo.wav"; return &s }(),              // Required content for recording
		ContentHash: vcon.NewContentHashSingle("sha-256:LCa0a2j_xo_5m0U8HTBBNBNCLXBkg7-g-YpeiGJm564"), // Required for external URLs
	})

	// Note: Extensions field has been removed for IETF draft-03 compliance

	fmt.Println("🔬 New Multi-Tier Validation System:")

	// Test all validation levels with the unified interface
	validationLevels := []struct {
		level vcon.ValidationLevel
		desc  string
	}{
		{vcon.ValidationBasic, "Basic business rules"},
		{vcon.ValidationStrict, "Strict business rules"},
		{vcon.ValidationIETF, "IETF specification compliance"},
		{vcon.ValidationIETFDraft03, "IETF Draft-03 strict compliance"},
		{vcon.ValidationIETFStrict, "IETF + extension field detection"},
		{vcon.ValidationComplete, "All validation layers"},
	}

	for _, test := range validationLevels {
		result := conv.ValidateWithLevel(test.level)
		if result.Valid {
			fmt.Printf("  ✅ %s (%s): PASSED\n", result.Level.String(), test.desc)
		} else {
			fmt.Printf("  ❌ %s (%s): FAILED with %d errors\n", result.Level.String(), test.desc, len(result.Errors))
			for i, err := range result.Errors {
				if i < 2 { // Show first 2 errors to avoid clutter
					fmt.Printf("     • %s: %s\n", err.Field, err.Message)
				}
			}
			if len(result.Errors) > 2 {
				fmt.Printf("     • ... and %d more errors\n", len(result.Errors)-2)
			}
		}
	}

	fmt.Println("\n📚 Legacy Validation Methods (for compatibility):")

	// Basic validation (lenient)
	fmt.Println("Basic validation:")
	if err := conv.Validate(); err != nil {
		fmt.Printf("  ❌ Errors: %v\n", err)
	} else {
		fmt.Println("  ✅ Passed basic validation")
	}

	// IETF validation (strict)
	fmt.Println("IETF validation:")
	if err := conv.ValidateIETF(); err != nil {
		fmt.Printf("  ❌ Errors: %v\n", err)
	} else {
		fmt.Println("  ✅ Passed IETF validation")
	}

	// Strict validation
	fmt.Println("Strict validation:")
	if err := conv.ValidateStrict(); err != nil {
		fmt.Printf("  ❌ Errors: %v\n", err)
	} else {
		fmt.Println("  ✅ Passed strict validation")
	}
}

func main() {
	demonstrateIETFCompliantVCon()
	demonstrateMigration()
	demonstrateValidationModes()
}
