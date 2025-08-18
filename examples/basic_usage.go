//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/lmendes86/vcon"
)

func main() {
	// Create a new vCon with default version
	v := vcon.NewWithDefaults()

	// Set a subject for the conversation
	subject := "Customer Support Call"
	v.Subject = &subject

	// Add participants
	// Customer
	customerEmail := "customer@example.com"
	customerName := "John Doe"
	customerIdx := v.AddParty(vcon.Party{
		Mailto: &customerEmail,
		Name:   &customerName,
		Role:   stringPtr("customer"),
	})

	// Support agent
	agentEmail := "agent@company.com"
	agentName := "Jane Smith"
	agentIdx := v.AddParty(vcon.Party{
		Mailto: &agentEmail,
		Name:   &agentName,
		Role:   stringPtr("agent"),
	})

	// Add dialog segments
	startTime := time.Now().Add(-10 * time.Minute)

	// Initial text message
	textMediatype := "text/plain"
	textEncoding := "none"
	v.AddDialog(vcon.Dialog{
		Type:       "text",
		Start:      startTime,
		Parties:    vcon.NewDialogPartiesArrayPtr([]int{customerIdx}),
		Originator: &customerIdx,
		Body:       "Hi, I need help with my account",
		Mediatype:  &textMediatype,
		Encoding:   &textEncoding,
	})

	// Agent response
	v.AddDialog(vcon.Dialog{
		Type:       "text",
		Start:      startTime.Add(30 * time.Second),
		Parties:    vcon.NewDialogPartiesArrayPtr([]int{agentIdx}),
		Originator: &agentIdx,
		Body:       "Hello! I'd be happy to help you. What seems to be the issue?",
		Mediatype:  &textMediatype,
		Encoding:   &textEncoding,
	})

	// Phone call segment
	callDuration := 300.0 // 5 minutes
	mediatype := "audio/wav"
	contentHash := vcon.NewContentHashSingle("sha-256:LCa0a2j_xo_5m0U8HTBBNBNCLXBkg7-g-YpeiGJm564")
	v.AddDialog(vcon.Dialog{
		Type:        "recording",
		Start:       startTime.Add(1 * time.Minute),
		Parties:     vcon.NewDialogPartiesArrayPtr([]int{customerIdx, agentIdx}),
		Originator:  &agentIdx,
		Duration:    &callDuration,
		Mediatype:   &mediatype,
		URL:         stringPtr("https://recordings.company.com/call123.wav"),
		ContentHash: contentHash,
	})

	// Update the timestamp to indicate when the vCon was last modified
	v.UpdateTimestamp()

	// Validate the vCon using the new multi-tier validation system
	result := v.ValidateWithLevel(vcon.ValidationComplete)
	if !result.Valid {
		log.Fatalf("vCon validation failed: %v", result.Errors)
	}
	fmt.Printf("✅ Validation passed at level: %s\n", result.Level.String())

	// Convert to JSON and display
	jsonData, err := v.ToJSONIndent()
	if err != nil {
		log.Fatalf("Failed to marshal vCon: %v", err)
	}

	fmt.Println("Generated vCon:")
	fmt.Println(string(jsonData))

	// Demonstrate round-trip serialization
	fmt.Println("\n--- Testing Round-trip Serialization ---")

	// Parse the JSON back to a vCon
	parsed, err := vcon.FromJSON(jsonData)
	if err != nil {
		log.Fatalf("Failed to parse vCon: %v", err)
	}

	// Validate the parsed vCon using legacy method (for compatibility)
	if err := parsed.Validate(); err != nil {
		log.Fatalf("Parsed vCon validation failed: %v", err)
	}

	// Also demonstrate the new validation system
	result = parsed.ValidateWithLevel(vcon.ValidationIETF)
	if !result.Valid {
		log.Printf("IETF validation warnings: %v", result.Errors)
	} else {
		fmt.Println("✅ Parsed vCon is IETF compliant")
	}

	fmt.Printf("Successfully round-tripped vCon with %d parties and %d dialogs\n",
		len(parsed.Parties), len(parsed.Dialog))

	// Show some basic statistics
	fmt.Println("\n--- vCon Statistics ---")
	fmt.Printf("UUID: %s\n", parsed.UUID.String())
	fmt.Printf("Version: %s\n", parsed.Vcon)
	fmt.Printf("Created: %s\n", parsed.CreatedAt.Format(time.RFC3339))
	if parsed.UpdatedAt != nil {
		fmt.Printf("Updated: %s\n", parsed.UpdatedAt.Format(time.RFC3339))
	}
	fmt.Printf("Subject: %s\n", *parsed.Subject)
	fmt.Printf("Parties: %d\n", len(parsed.Parties))
	fmt.Printf("Dialog segments: %d\n", len(parsed.Dialog))

	// List dialog types
	dialogTypes := make(map[string]int)
	for _, dialog := range parsed.Dialog {
		dialogTypes[dialog.Type]++
	}
	fmt.Println("Dialog types:")
	for dialogType, count := range dialogTypes {
		fmt.Printf("  %s: %d\n", dialogType, count)
	}
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
