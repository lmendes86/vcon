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
	// Create a vCon for an SMS conversation
	v := vcon.NewWithDefaults()

	subject := "SMS Conversation"
	v.Subject = &subject

	// Add participants
	alice := v.AddParty(vcon.Party{
		Tel:  stringPtr("+1234567890"),
		Name: stringPtr("Alice Johnson"),
	})

	bob := v.AddParty(vcon.Party{
		Tel:  stringPtr("+1987654321"),
		Name: stringPtr("Bob Wilson"),
	})

	// Simulate an SMS conversation over time
	baseTime := time.Now().Add(-2 * time.Hour)

	messages := []struct {
		sender  int
		content string
		delay   time.Duration
	}{
		{alice, "Hey Bob! Are we still on for lunch today?", 0},
		{bob, "Yes! How about 12:30 at the cafe?", 2 * time.Minute},
		{alice, "Perfect! See you there 😊", 30 * time.Second},
		{bob, "Great! I'll grab us a table", 5 * time.Minute},
		{alice, "Thanks! I'm running about 5 mins late", 25 * time.Minute},
		{bob, "No worries, I'm here already", 1 * time.Minute},
		{alice, "Just arrived! Looking for you", 5 * time.Minute},
		{bob, "I'm in the corner by the window", 30 * time.Second},
	}

	currentTime := baseTime
	for _, msg := range messages {
		currentTime = currentTime.Add(msg.delay)

		mediatype := "text/plain"
		encoding := "none"
		v.AddDialog(vcon.Dialog{
			Type:       "text",
			Start:      currentTime,
			Parties:    vcon.NewDialogPartiesArrayPtr([]int{msg.sender}),
			Originator: &msg.sender,
			Body:       msg.content,
			Mediatype:  &mediatype,
			Encoding:   &encoding,
			Meta: map[string]any{
				"transport":  "sms",
				"message_id": fmt.Sprintf("msg_%d_%d", msg.sender, currentTime.Unix()),
			},
		})
	}

	// Add conversation metadata to the first dialog's meta field
	// Note: Top-level Meta field removed for IETF draft-03 compliance
	if len(v.Dialog) > 0 && v.Dialog[0].Meta == nil {
		v.Dialog[0].Meta = make(map[string]any)
	}
	if len(v.Dialog) > 0 {
		v.Dialog[0].Meta["conversation_metadata"] = map[string]any{
			"conversation_type": "sms",
			"duration_minutes":  int(currentTime.Sub(baseTime).Minutes()),
			"message_count":     len(messages),
		}
	}

	// Validate the vCon using the new multi-tier validation system
	result := v.ValidateWithLevel(vcon.ValidationComplete)
	if !result.Valid {
		log.Fatalf("vCon validation failed: %v", result.Errors)
	}
	fmt.Printf("✅ Validation passed at level: %s\n", result.Level.String())

	// Display the conversation
	fmt.Println("SMS Conversation vCon:")
	fmt.Println("===================")

	jsonData, err := v.ToJSONIndent()
	if err != nil {
		log.Fatalf("Failed to marshal vCon: %v", err)
	}

	fmt.Println(string(jsonData))

	// Show conversation summary
	fmt.Println("\n--- Conversation Summary ---")
	fmt.Printf("Participants: %d\n", len(v.Parties))
	for i, party := range v.Parties {
		fmt.Printf("  [%d] %s (%s)\n", i, *party.Name, *party.Tel)
	}

	fmt.Printf("\nMessages: %d\n", len(v.Dialog))
	for i, dialog := range v.Dialog {
		parties, ok := dialog.Parties.GetArray()
		if ok && len(parties) > 0 {
			senderName := *v.Parties[parties[0]].Name
			timestamp := dialog.Start.Format("15:04:05")
			fmt.Printf("  [%s] %s: %s\n", timestamp, senderName, dialog.Body)
		}

		// Show first few characters to avoid long output
		if i >= 5 {
			fmt.Printf("  ... (%d more messages)\n", len(v.Dialog)-6)
			break
		}
	}

	// Demonstrate filtering and analysis
	fmt.Println("\n--- Message Analysis ---")

	// Count messages per participant
	messageCounts := make(map[int]int)
	for _, dialog := range v.Dialog {
		parties, ok := dialog.Parties.GetArray()
		if ok && len(parties) > 0 {
			messageCounts[parties[0]]++
		}
	}

	for partyIdx, count := range messageCounts {
		partyName := *v.Parties[partyIdx].Name
		fmt.Printf("%s sent %d messages\n", partyName, count)
	}

	// Calculate conversation duration
	if len(v.Dialog) > 1 {
		duration := v.Dialog[len(v.Dialog)-1].Start.Sub(v.Dialog[0].Start)
		fmt.Printf("Conversation lasted: %s\n", duration.Round(time.Minute))
	}
}

func stringPtr(s string) *string {
	return &s
}
