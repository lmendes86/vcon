//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"time"

	vcon "github.com/lmendes86/vcon"
)

func main() {
	fmt.Println("=== Enhanced vCon Features Demo ===\n")

	// 1. Property Handling Demo
	fmt.Println("1. Property Handling Modes")
	demonstratePropertyHandling()

	// 2. Extension and Must-Support Management
	fmt.Println("\n2. Extension and Must-Support Management")
	demonstrateExtensionManagement()

	// 3. Tag Management System
	fmt.Println("\n3. Tag Management System")
	demonstrateTagManagement()

	// 4. Search and Utility Methods
	fmt.Println("\n4. Search and Utility Methods")
	demonstrateSearchUtilities()

	// 5. Specialized Dialog Creators
	fmt.Println("\n5. Specialized Dialog Creators")
	demonstrateSpecializedDialogs()

	// 6. Digital Signatures
	fmt.Println("\n6. Digital Signatures (JWS)")
	demonstrateDigitalSignatures()

	// 7. HTTP Operations
	fmt.Println("\n7. HTTP Operations and File I/O")
	demonstrateHTTPOperations()

	// 8. Enhanced Security Features
	fmt.Println("\n8. Enhanced Security Features")
	demonstrateEnhancedSecurity()

	// 9. Advanced Validation
	fmt.Println("\n9. Advanced Validation")
	demonstrateAdvancedValidation()

	fmt.Println("\n=== Demo Complete ===")
}

func demonstratePropertyHandling() {
	// Create vCon with custom properties
	jsonWithCustomProps := `{
		"vcon": "0.0.1",
		"uuid": "550e8400-e29b-41d4-a716-446655440000",
		"created_at": "2023-01-01T10:00:00Z",
		"parties": [{"name": "John Doe"}],
		"dialog": [],
		"custom_field": "custom_value",
		"internal_data": {"secret": "hidden"}
	}`

	// Default mode - strict adherence to schema
	fmt.Println("  Default mode (strict):")
	vconDefault, err := vcon.BuildFromJSON(jsonWithCustomProps, vcon.PropertyHandlingDefault)
	if err != nil {
		fmt.Printf("    Error: %v\n", err)
	} else {
		fmt.Printf("    Successfully parsed with %d parties\n", len(vconDefault.Parties))
	}

	// Strict mode - reject unknown properties
	fmt.Println("  Strict mode:")
	vconStrict, err := vcon.BuildFromJSON(jsonWithCustomProps, vcon.PropertyHandlingStrict)
	if err != nil {
		fmt.Printf("    Error: %v\n", err)
	} else {
		fmt.Printf("    Successfully parsed with %d parties\n", len(vconStrict.Parties))
	}

	// Meta mode - preserve unknown properties
	fmt.Println("  Meta mode (preserve unknown properties):")
	vconMeta, err := vcon.BuildFromJSON(jsonWithCustomProps, vcon.PropertyHandlingMeta)
	if err != nil {
		fmt.Printf("    Error: %v\n", err)
	} else {
		fmt.Printf("    Successfully parsed with %d parties\n", len(vconMeta.Parties))
		// Note: Meta mode preserves unknown properties in object-specific meta fields
		fmt.Printf("    Meta mode preserves unknown properties in object-specific meta fields\n")
	}
}

func demonstrateExtensionManagement() {
	fmt.Println("  Extension Management via Attachments (IETF Compliant)")

	v := vcon.NewWithDefaults()

	// Note: Per IETF draft-03, extensions and must-support are handled via attachments
	// Add extension information via attachments
	extensionInfo := map[string]interface{}{
		"extensions":   []string{"video", "encryption", "analytics"},
		"must_support": []string{"encryption", "video"},
	}

	v.AddAttachment(vcon.Attachment{
		Type:     stringPtr("extensions"),
		Body:     extensionInfo,
		Encoding: stringPtr("json"),
	})

	fmt.Printf("  Extensions stored as attachment for IETF compliance\n")

	// Find the extensions attachment
	extensionAttachment := v.FindAttachmentByType("extensions")
	if extensionAttachment != nil {
		fmt.Printf("  Extension metadata stored with type: %s\n", extensionAttachment.Type)
		if body, ok := extensionAttachment.Body.(map[string]interface{}); ok {
			if extensions, exists := body["extensions"]; exists {
				fmt.Printf("  Extensions: %v\n", extensions)
			}
			if mustSupport, exists := body["must_support"]; exists {
				fmt.Printf("  Must-support: %v\n", mustSupport)
			}
		}
	}

	fmt.Printf("  ℹ️  Note: IETF draft-03 moved extensions to attachment-based pattern\n")
}

func demonstrateTagManagement() {
	v := vcon.NewWithDefaults()

	// Add tags using the tag management system
	v.AddTag("category", "customer_support")
	v.AddTag("priority", "high")
	v.AddTag("department", "sales")

	// Retrieve tags
	category := v.GetTag("category")
	priority := v.GetTag("priority")
	nonExistent := v.GetTag("non_existent")

	fmt.Printf("  Category: %v\n", stringPtrValue(category))
	fmt.Printf("  Priority: %v\n", stringPtrValue(priority))
	fmt.Printf("  Non-existent tag: %v\n", stringPtrValue(nonExistent))

	// Show tags attachment
	tagsAttachment := v.GetTags()
	if tagsAttachment != nil {
		fmt.Printf("  Tags stored as attachment type: %s\n", tagsAttachment.Type)
	}
}

func demonstrateSearchUtilities() {
	v := vcon.NewWithDefaults()

	// Add some parties
	v.AddParty(vcon.Party{
		Name:   stringPtr("John Doe"),
		Tel:    stringPtr("+1234567890"),
		Mailto: stringPtr("john@example.com"),
	})
	v.AddParty(vcon.Party{
		Name:   stringPtr("Jane Smith"),
		Mailto: stringPtr("jane@example.com"),
	})

	// Add different dialog types
	now := time.Now()
	v.AddDialog(vcon.Dialog{
		Type:    "text",
		Start:   now,
		Parties: vcon.NewDialogPartiesArrayPtr([]int{0}),
		Body:    "Hello",
	})
	v.AddDialog(vcon.Dialog{
		Type:      "recording",
		Start:     now.Add(time.Minute),
		Parties:   vcon.NewDialogPartiesArrayPtr([]int{0, 1}),
		Mediatype: stringPtr("audio/wav"),
	})

	// Search for parties
	nameIndex := v.FindPartyIndex("name", "John Doe")
	telIndex := v.FindPartyIndex("tel", "+1234567890")
	emailIndex := v.FindPartyIndex("mailto", "jane@example.com")

	fmt.Printf("  Found 'John Doe' at index: %v\n", intPtrValue(nameIndex))
	fmt.Printf("  Found '+1234567890' at index: %v\n", intPtrValue(telIndex))
	fmt.Printf("  Found 'jane@example.com' at index: %v\n", intPtrValue(emailIndex))

	// Find dialogs by type
	textDialogs := v.FindDialogsByType("text")
	recordingDialogs := v.FindDialogsByType("recording")

	fmt.Printf("  Text dialogs: %d\n", len(textDialogs))
	fmt.Printf("  Recording dialogs: %d\n", len(recordingDialogs))

	// Add and find attachments
	v.AddAttachment(vcon.Attachment{
		Type:     stringPtr("metadata"),
		Body:     map[string]interface{}{"version": "1.0"},
		Encoding: stringPtr("json"),
	})

	metadata := v.FindAttachmentByType("metadata")
	if metadata != nil {
		fmt.Printf("  Found metadata attachment with type: %s\n", metadata.Type)
	}
}

func demonstrateSpecializedDialogs() {
	v := vcon.NewWithDefaults()

	// Add some parties first
	v.AddParty(vcon.Party{Name: stringPtr("Caller")})
	v.AddParty(vcon.Party{Name: stringPtr("Agent")})
	v.AddParty(vcon.Party{Name: stringPtr("Supervisor")})

	now := time.Now()

	// Add transfer dialog
	transferData := map[string]any{
		"reason": "Escalation to supervisor",
		"from":   "+1234567890",
		"to":     "+1987654321",
	}
	metadata := map[string]any{
		"system": "PBX",
		"queue":  "support",
	}

	transferIdx := v.AddTransferDialog(now, transferData, []int{0, 1, 2}, metadata)
	fmt.Printf("  Added transfer dialog at index: %d\n", transferIdx)

	// Add incomplete dialog
	incompleteDetails := map[string]any{
		"ringDuration": 45000,
		"reason":       "No answer",
		"attempts":     1,
	}

	incompleteIdx := v.AddIncompleteDialog(
		now.Add(5*time.Minute),
		"NO_ANSWER",
		incompleteDetails,
		[]int{0},
		map[string]any{"system": "auto-dialer"},
	)
	fmt.Printf("  Added incomplete dialog at index: %d\n", incompleteIdx)

	// Add analysis
	v.AddAnalysis(
		"sentiment",      // analysisType
		[]int{0, 1},      // dialog
		"acme_analytics", // vendor
		map[string]any{"score": 0.8, "confidence": "high"}, // body
		"json",                              // encoding
		stringPtr("SentimentAnalyzer v2.0"), // product
		stringPtr("https://acme.com/schema/sentiment/v2.0"), // schema
		map[string]any{"model": "transformer_v3"},           // meta
	)

	analysis := v.FindAnalysisByType("sentiment")
	if analysis != nil {
		fmt.Printf("  Added sentiment analysis by vendor: %s\n", analysis.Vendor)
	}
}

func demonstrateDigitalSignatures() {
	v := vcon.NewWithDefaults()
	v.AddParty(vcon.Party{Name: stringPtr("Test User")})

	// Generate RSA key pair
	privateKey, publicKey, err := vcon.GenerateKeyPair()
	if err != nil {
		log.Printf("  Error generating keys: %v", err)
		return
	}
	fmt.Println("  Generated RSA key pair (2048 bits)")

	// Sign the vCon
	err = v.Sign(privateKey)
	if err != nil {
		log.Printf("  Error signing vCon: %v", err)
		return
	}
	fmt.Printf("  Signed vCon (has %d signatures)\n", len(v.Signatures))

	// Verify signature
	valid, err := v.Verify(publicKey)
	if err != nil {
		log.Printf("  Error verifying signature: %v", err)
		return
	}
	fmt.Printf("  Signature verification: %t\n", valid)

	// Get signed payload
	payload, err := v.GetSignedPayload()
	if err != nil {
		log.Printf("  Error getting signed payload: %v", err)
		return
	}
	fmt.Printf("  Signed payload length: %d characters\n", len(*payload))

	// Demonstrate key conversion
	privatePEM, err := vcon.PrivateKeyToPEM(privateKey)
	if err != nil {
		log.Printf("  Error converting private key to PEM: %v", err)
		return
	}
	fmt.Printf("  Private key PEM length: %d bytes\n", len(privatePEM))

	publicPEM, err := vcon.PublicKeyToPEM(publicKey)
	if err != nil {
		log.Printf("  Error converting public key to PEM: %v", err)
		return
	}
	fmt.Printf("  Public key PEM length: %d bytes\n", len(publicPEM))
}

func demonstrateHTTPOperations() {
	v := vcon.NewWithDefaults()
	v.AddParty(vcon.Party{Name: stringPtr("Demo User")})

	// Save to file
	tmpFile := "/tmp/demo_vcon.json"
	err := v.SaveToFile(tmpFile, vcon.PropertyHandlingDefault)
	if err != nil {
		log.Printf("  Error saving to file: %v", err)
		return
	}
	fmt.Printf("  Saved vCon to: %s\n", tmpFile)

	// Load from file
	loaded, err := vcon.LoadFromFile(tmpFile, vcon.PropertyHandlingDefault)
	if err != nil {
		log.Printf("  Error loading from file: %v", err)
		return
	}
	fmt.Printf("  Loaded vCon with %d parties\n", len(loaded.Parties))

	// Convert to JSON strings
	jsonStr, err := v.ToJSONString(vcon.PropertyHandlingDefault)
	if err != nil {
		log.Printf("  Error converting to JSON string: %v", err)
		return
	}
	fmt.Printf("  JSON string length: %d characters\n", len(jsonStr))

	indentedStr, err := v.ToJSONIndentString(vcon.PropertyHandlingDefault)
	if err != nil {
		log.Printf("  Error converting to indented JSON: %v", err)
		return
	}
	fmt.Printf("  Indented JSON length: %d characters\n", len(indentedStr))

	// Clean up
	os.Remove(tmpFile)

	// Note: HTTP operations like LoadFromURL and PostToURL would require
	// actual HTTP endpoints, so we're just demonstrating the file operations
	fmt.Println("  HTTP operations available: LoadFromURL, PostToURL")
}

func demonstrateEnhancedSecurity() {
	fmt.Println("  🔒 Enhanced Security Features")

	// 1. Modern Hash Algorithm Support
	fmt.Println("    Modern Hash Algorithms:")
	modernHashes := []string{
		"sha-256:dGVzdA",                   // Strong algorithm
		"sha-512:dGVzdERhdGFGb3JUZXN0aW5n", // Strong algorithm
		"sha3-256:dGVzdA",                  // SHA-3 support
		"blake2b:dGVzdERhdGFGb3JUZXN0aW5n", // Modern algorithm
	}

	for _, hash := range modernHashes {
		err := vcon.ValidateContentHashFormat(hash)
		if err != nil {
			fmt.Printf("      ❌ %s: %v\n", hash, err)
		} else {
			fmt.Printf("      ✅ %s: Valid modern algorithm\n", hash)
		}
	}

	// 2. Weak Algorithm Warnings
	fmt.Println("    Weak Algorithm Detection:")
	weakHashes := []string{
		"sha1:dGVzdA", // Deprecated
		"md5:dGVzdA",  // Strongly discouraged
	}

	for _, hash := range weakHashes {
		err := vcon.ValidateContentHashFormat(hash)
		if err != nil {
			fmt.Printf("      ⚠️  %s: %v\n", hash, err)
		} else {
			fmt.Printf("      ✅ %s: Valid but weak\n", hash)
		}
	}

	// 3. Content Hash Array Support
	fmt.Println("    Content Hash Arrays:")
	v := vcon.NewWithDefaults()

	// Create content hash arrays with multiple algorithms
	multipleHashes := []string{
		"sha-256:dGVzdERhdGE",
		"sha-512:dGVzdERhdGFGb3JUZXN0aW5nRGF0YUZvckhhc2hpbmc",
		"sha3-256:dGVzdERhdGE",
	}

	contentHashValue := vcon.NewContentHashArray(multipleHashes)

	// Add dialog with multiple content hashes
	now := time.Now()
	v.AddDialog(vcon.Dialog{
		Type:        "recording",
		Start:       now,
		Parties:     vcon.NewDialogPartiesArrayPtr([]int{0}),
		ContentHash: contentHashValue,
		Mediatype:   stringPtr("audio/wav"),
	})

	fmt.Printf("      Multi-algorithm hash support: %d algorithms\n", len(contentHashValue.GetArray()))
	fmt.Printf("      Primary hash: %s\n", contentHashValue.GetSingle())

	// 4. Security-focused Analysis
	fmt.Println("    Security Analysis:")

	// Add analysis with secure hash
	v.AddAnalysis(
		"security_scan",
		[]int{0}, // Dialog reference
		"security_vendor",
		map[string]any{"threats_detected": 0, "risk_score": "low"},
		"json",
		stringPtr("SecurityScanner v3.0"),
		stringPtr("https://example.com/schema/security/v3.0"),
		map[string]any{"scan_duration": "15s", "algorithm": "behavioral_analysis"},
	)

	// Get the last added analysis and add secure content hash
	if len(v.Analysis) > 0 {
		analysis := &v.Analysis[len(v.Analysis)-1]
		// Add secure content hash to analysis
		secureHash := "sha-512:c2VjdXJpdHlBbmFseXNpc1Jlc3VsdERhdGFGb3JUZXN0aW5nSGFzaA"
		analysis.ContentHash = vcon.NewContentHashSingle(secureHash)
		fmt.Printf("      Security analysis with hash: %s\n", analysis.ContentHash.GetSingle())
	}

	// 5. Hash Format Validation
	fmt.Println("    Hash Format Validation:")
	testHashes := []struct {
		hash   string
		desc   string
		expect string
	}{
		{"sha-256:validBase64Hash", "Valid format", "✅"},
		{"invalid-format", "Missing colon", "❌"},
		{"unknown-algo:dGVzdA", "Unknown algorithm", "❌"},
		{"sha-256:", "Empty hash", "❌"},
		{"sha-256:invalid base64!", "Invalid base64", "❌"},
	}

	for _, test := range testHashes {
		err := vcon.ValidateContentHashFormat(test.hash)
		status := "✅"
		if err != nil {
			status = "❌"
		}
		fmt.Printf("      %s %s: %s\n", status, test.desc, test.hash)
	}
}

func demonstrateAdvancedValidation() {
	fmt.Println("  🔬 Multi-Tier Validation Architecture")

	// Create a vCon with some validation issues for testing different levels
	v := vcon.NewWithDefaults()

	// Add party without proper identifier (will fail strict validation)
	v.Parties = append(v.Parties, vcon.Party{
		// No name, tel, mailto, or uuid - fails strict business rules but passes basic
	})

	// Add valid party
	v.AddParty(vcon.Party{Name: stringPtr("Valid User")})

	// Add dialog with invalid party reference
	v.Dialog = append(v.Dialog, vcon.Dialog{
		Type:    "text",
		Start:   time.Now(),
		Parties: vcon.NewDialogPartiesArrayPtr([]int{999}), // Invalid party index
		Body:    "Test message",
	})

	// Add extension via attachment to test IETF strict validation
	v.AddAttachment(vcon.Attachment{
		Type:     stringPtr("custom-extension"),
		Body:     map[string]interface{}{"test": "data"},
		Encoding: stringPtr("json"),
	})

	fmt.Println("    🆕 New Multi-Tier Validation System:")

	// Test all validation levels
	validationLevels := []struct {
		level vcon.ValidationLevel
		desc  string
	}{
		{vcon.ValidationBasic, "Basic business rules"},
		{vcon.ValidationStrict, "Strict business rules"},
		{vcon.ValidationIETF, "IETF specification compliance"},
		{vcon.ValidationIETFDraft03, "IETF Draft-03 strict compliance"},
		{vcon.ValidationIETFStrict, "IETF + extension detection"},
		{vcon.ValidationComplete, "All validation layers"},
	}

	for _, test := range validationLevels {
		result := v.ValidateWithLevel(test.level)
		if result.Valid {
			fmt.Printf("      ✅ %s: PASSED\n", result.Level.String())
		} else {
			fmt.Printf("      ❌ %s: FAILED (%d errors)\n", result.Level.String(), len(result.Errors))
			// Show first error as example
			if len(result.Errors) > 0 {
				fmt.Printf("         • %s: %s\n", result.Errors[0].Field, result.Errors[0].Message)
			}
		}
	}

	fmt.Println("    📚 Legacy Validation Methods:")

	// Test basic validation (legacy)
	fmt.Println("      Basic validation (legacy):")
	valid, errors := v.IsValid()
	fmt.Printf("        Valid: %t\n", valid)
	if !valid && len(errors) > 0 {
		fmt.Printf("        Sample error: %s\n", errors[0])
	}

	// Test advanced validation (legacy)
	fmt.Println("      Advanced validation (legacy):")
	err := v.ValidateAdvanced()
	if err != nil {
		fmt.Printf("        Errors: %v\n", err)
	} else {
		fmt.Println("        Passed")
	}

	// Test IETF validation (legacy)
	fmt.Println("      IETF validation (legacy):")
	err = v.ValidateIETF()
	if err != nil {
		fmt.Printf("        Errors: %v\n", err)
	} else {
		fmt.Println("        Passed")
	}

	// Test validation of JSON strings
	fmt.Println("    📄 JSON String Validation:")
	validJSON := `{"vcon":"0.0.2","uuid":"550e8400-e29b-41d4-a716-446655440000","created_at":"2023-01-01T10:00:00Z","parties":[{"name":"Test"}],"dialog":[]}`
	invalidJSON := `{"invalid": json}`

	isValid, _ := vcon.ValidateJSONString(validJSON)
	fmt.Printf("      Valid IETF v0.0.2 JSON validates: %t\n", isValid)

	isValid, _ = vcon.ValidateJSONString(invalidJSON)
	fmt.Printf("      Invalid JSON validates: %t\n", isValid)
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func stringPtrValue(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}

func intPtrValue(i *int) string {
	if i == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%d", *i)
}
