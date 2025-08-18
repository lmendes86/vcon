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
	// Create a vCon for a call center interaction
	v := vcon.NewWithDefaults()

	subject := "Technical Support Call - Ticket #12345"
	v.Subject = &subject

	// Add customer
	customerPhone := "+1555123456"
	customerName := "Sarah Miller"
	customer := v.AddParty(vcon.Party{
		Tel:  &customerPhone,
		Name: &customerName,
		Role: stringPtr("customer"),
		Meta: map[string]any{
			"customer_id":    "CUST789123",
			"account_type":   "premium",
			"previous_calls": 2,
		},
	})

	// Add support agent
	agentEmail := "mike.support@techcorp.com"
	agentName := "Mike Anderson"
	agent := v.AddParty(vcon.Party{
		Mailto: &agentEmail,
		Name:   &agentName,
		Role:   stringPtr("support_agent"),
		Meta: map[string]any{
			"employee_id": "EMP456",
			"department":  "technical_support",
			"skill_level": "senior",
		},
	})

	// Add supervisor (joined later)
	supervisorEmail := "lisa.supervisor@techcorp.com"
	supervisorName := "Lisa Thompson"
	supervisor := v.AddParty(vcon.Party{
		Mailto: &supervisorEmail,
		Name:   &supervisorName,
		Role:   stringPtr("supervisor"),
		Meta: map[string]any{
			"employee_id": "EMP123",
			"department":  "technical_support",
		},
	})

	// Define call timeline
	callStart := time.Now().Add(-45 * time.Minute)

	// Initial call setup and greeting
	v.AddDialog(vcon.Dialog{
		Type:        "incomplete",
		Start:       callStart,
		Parties:     vcon.NewDialogPartiesArrayPtr([]int{customer, agent}),
		Originator:  &customer,
		Disposition: stringPtr("no-answer"), // IETF compliant disposition
		Meta: map[string]any{
			"call_queue":    "technical_support",
			"wait_time_sec": 45,
		},
	})

	// Add party history for call events
	greetingTime := callStart.Add(45 * time.Second)
	v.AddDialog(vcon.Dialog{
		Type:        "recording",
		Start:       greetingTime,
		Parties:     vcon.NewDialogPartiesArrayPtr([]int{customer, agent}),
		Originator:  &agent,
		Duration:    floatPtr(180), // 3 minutes greeting and problem description
		Mediatype:   stringPtr("audio/wav"),
		URL:         stringPtr("https://recordings.techcorp.com/calls/2024/01/15/call_12345_part1.wav"),
		ContentHash: vcon.NewContentHashSingle("sha-256:LCa0a2j_xo_5m0U8HTBBNBNCLXBkg7-g-YpeiGJm564"), // Required for external URLs
		PartyHistory: []vcon.PartyHistory{
			{
				Party: agent,
				Event: "join",
				Time:  greetingTime,
			},
		},
		Meta: map[string]any{
			"segment":       "greeting_and_problem_description",
			"quality_score": 4.5,
		},
	})

	// Troubleshooting segment
	troubleshootingStart := greetingTime.Add(3 * time.Minute)
	v.AddDialog(vcon.Dialog{
		Type:        "recording",
		Start:       troubleshootingStart,
		Parties:     vcon.NewDialogPartiesArrayPtr([]int{customer, agent}),
		Duration:    floatPtr(600), // 10 minutes of troubleshooting
		Mediatype:   stringPtr("audio/wav"),
		URL:         stringPtr("https://recordings.techcorp.com/calls/2024/01/15/call_12345_part2.wav"),
		ContentHash: vcon.NewContentHashSingle("sha-256:MJ7MSJwS1utMxA9QyQLytNDtd-5RGnx6m808qG1M2G-Y"), // Required for external URLs
		Meta: map[string]any{
			"segment":       "troubleshooting",
			"issue_type":    "software_configuration",
			"resolution":    "partial",
			"quality_score": 3.8,
		},
	})

	// Escalation to supervisor (transfer dialog)
	escalationTime := troubleshootingStart.Add(10 * time.Minute)
	targetDialogRef := len(v.Dialog) + 1
	v.AddDialog(vcon.Dialog{
		Type:           "transfer",
		Start:          escalationTime,
		Transferee:     &customer,
		Transferor:     &agent,
		Original:       &[]int{1}[0],     // Reference to original dialog (greeting)
		TargetDialog:   &targetDialogRef, // Reference to target dialog (will be consultation)
		TransferTarget: &supervisor,      // Transfer target party index
		PartyHistory: []vcon.PartyHistory{
			{
				Party: supervisor,
				Event: "join",
				Time:  escalationTime,
			},
		},
		Meta: map[string]any{
			"transfer_reason":  "escalation_required",
			"issue_complexity": "high",
		},
	})

	// Supervisor consultation (brief hold)
	consultationTime := escalationTime.Add(30 * time.Second)
	v.AddDialog(vcon.Dialog{
		Type:        "recording",
		Start:       consultationTime,
		Parties:     vcon.NewDialogPartiesArrayPtr([]int{agent, supervisor}),
		Duration:    floatPtr(120), // 2 minutes consultation
		Mediatype:   stringPtr("audio/wav"),
		URL:         stringPtr("https://recordings.techcorp.com/calls/2024/01/15/call_12345_consultation.wav"),
		ContentHash: vcon.NewContentHashSingle("sha-256:XpY8e9r3G7k1M4N5p6Q7R8s9T0u1V2w3X4y5Z6a7B8c9"), // Required for external URLs
		Meta: map[string]any{
			"segment": "supervisor_consultation",
			"private": true,
		},
	})

	// Resolution with all parties
	resolutionTime := consultationTime.Add(2 * time.Minute)
	v.AddDialog(vcon.Dialog{
		Type:        "recording",
		Start:       resolutionTime,
		Parties:     vcon.NewDialogPartiesArrayPtr([]int{customer, agent, supervisor}),
		Duration:    floatPtr(420), // 7 minutes for resolution
		Mediatype:   stringPtr("audio/wav"),
		URL:         stringPtr("https://recordings.techcorp.com/calls/2024/01/15/call_12345_resolution.wav"),
		ContentHash: vcon.NewContentHashSingle("sha-256:Z9a8B7c6D5e4F3g2H1i0J9k8L7m6N5o4P3q2R1s0T9u8"), // Required for external URLs
		PartyHistory: []vcon.PartyHistory{
			{
				Party: supervisor,
				Event: "drop",
				Time:  resolutionTime.Add(5 * time.Minute),
			},
		},
		Meta: map[string]any{
			"segment":       "resolution",
			"resolution":    "complete",
			"quality_score": 4.8,
			"satisfaction":  "high",
		},
	})

	// Call wrap-up
	wrapupTime := resolutionTime.Add(7 * time.Minute)
	v.AddDialog(vcon.Dialog{
		Type:        "recording",
		Start:       wrapupTime,
		Parties:     vcon.NewDialogPartiesArrayPtr([]int{customer, agent}),
		Duration:    floatPtr(90), // 1.5 minutes wrap-up
		Mediatype:   stringPtr("audio/wav"),
		URL:         stringPtr("https://recordings.techcorp.com/calls/2024/01/15/call_12345_wrapup.wav"),
		ContentHash: vcon.NewContentHashSingle("sha-256:U8v7W6x5Y4z3A2b1C0d9E8f7G6h5I4j3K2l1M0n9O8p7"), // Required for external URLs
		PartyHistory: []vcon.PartyHistory{
			{
				Party: agent,
				Event: "drop",
				Time:  wrapupTime.Add(90 * time.Second),
			},
			{
				Party: customer,
				Event: "drop",
				Time:  wrapupTime.Add(90 * time.Second),
			},
		},
		Meta: map[string]any{
			"segment":        "call_wrapup",
			"survey_offered": true,
		},
	})

	// Add call center metadata to appropriate analysis objects
	// Note: Top-level Meta field removed for IETF draft-03 compliance
	// Move metadata to analysis objects that will be added later

	// Add analysis attachment
	v.AddAttachment(vcon.Attachment{
		Type:      stringPtr("analysis"),
		Body:      generateCallAnalysis(),
		Encoding:  stringPtr("json"),
		Mediatype: stringPtr("application/json"),
		Start:     func() *time.Time { t := time.Now(); return &t }(),
		Party:     &customer,
		Meta: map[string]any{
			"analyzer":      "CallAnalytics_v2.1",
			"analysis_time": time.Now().Format(time.RFC3339),
			"confidence":    0.94,
			"language":      "en-US",
			// Call center metadata moved from top-level Meta field (removed for IETF compliance)
			"call_center": map[string]any{
				"location":     "Atlanta_Center_01",
				"campaign":     "technical_support",
				"ticket_id":    "TICK-12345",
				"category":     "software_support",
				"subcategory":  "configuration_issue",
				"resolution":   "resolved",
				"satisfaction": 4.7,
			},
			"quality": map[string]any{
				"overall_score":    4.4,
				"adherence_score":  4.2,
				"resolution_score": 4.8,
				"courtesy_score":   4.6,
				"reviewed_by":      "QA_ANALYST_001",
			},
			"analytics": map[string]any{
				"total_duration_sec":  1617, // About 27 minutes
				"talk_time_sec":       1485,
				"hold_time_sec":       132,
				"transfers":           1,
				"escalations":         1,
				"resolution_attempts": 2,
			},
		},
	})

	v.UpdateTimestamp()

	// Validate the vCon using the new multi-tier validation system
	result := v.ValidateWithLevel(vcon.ValidationIETF)
	if !result.Valid {
		log.Fatalf("vCon validation failed: %v", result.Errors)
	}
	fmt.Printf("✅ Validation passed at level: %s\n", result.Level.String())

	// Display results
	fmt.Println("Call Center vCon Generated Successfully!")
	fmt.Println("=====================================")

	// Show summary
	fmt.Printf("Call Subject: %s\n", *v.Subject)
	fmt.Printf("Participants: %d\n", len(v.Parties))
	fmt.Printf("Dialog Segments: %d\n", len(v.Dialog))
	fmt.Printf("Attachments: %d\n", len(v.Attachments))

	// Show call timeline
	fmt.Println("\nCall Timeline:")
	for i, dialog := range v.Dialog {
		duration := ""
		if dialog.Duration != nil {
			minutes := int(*dialog.Duration / 60)
			seconds := int(*dialog.Duration) % 60
			duration = fmt.Sprintf(" (%dm%ds)", minutes, seconds)
		}

		allPartyIndices := dialog.Parties.GetAllPartyIndices()
		parties := make([]string, len(allPartyIndices))
		for j, partyIdx := range allPartyIndices {
			parties[j] = *v.Parties[partyIdx].Name
		}

		fmt.Printf("  %d. %s - %s%s - %v\n",
			i+1,
			dialog.Start.Format("15:04:05"),
			dialog.Type,
			duration,
			parties,
		)
	}

	// Export to JSON file
	jsonData, err := v.ToJSONIndent()
	if err != nil {
		log.Fatalf("Failed to marshal vCon: %v", err)
	}

	fmt.Printf("\nvCon JSON size: %d bytes\n", len(jsonData))
	fmt.Println("\nFirst 500 characters of JSON:")
	if len(jsonData) > 500 {
		fmt.Printf("%s...\n", string(jsonData[:500]))
	} else {
		fmt.Println(string(jsonData))
	}
}

func generateCallAnalysis() map[string]any {
	return map[string]any{
		"sentiment": map[string]any{
			"customer": map[string]any{
				"overall":     "neutral_to_positive",
				"progression": []string{"frustrated", "neutral", "satisfied"},
				"final_score": 0.72,
			},
			"agent": map[string]any{
				"overall":     "professional",
				"empathy":     0.85,
				"knowledge":   0.78,
				"final_score": 0.82,
			},
		},
		"keywords": []string{
			"software configuration",
			"installation error",
			"system compatibility",
			"resolved",
			"satisfied",
		},
		"resolution": map[string]any{
			"time_to_resolution_sec": 1617,
			"first_call_resolution":  true,
			"escalation_required":    true,
			"resolution_quality":     "high",
		},
		"compliance": map[string]any{
			"greeting_adherence":      true,
			"privacy_statement":       true,
			"resolution_confirmation": true,
			"survey_offered":          true,
		},
	}
}

func stringPtr(s string) *string {
	return &s
}

func floatPtr(f float64) *float64 {
	return &f
}
