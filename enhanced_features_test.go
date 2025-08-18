package vcon

import (
	"testing"
	"time"
)

// Test Extension Management - REMOVED for IETF draft-03 compliance.
// Extensions and MustSupport fields have been removed from VCon struct.
// Extension functionality should now be handled through proper extension patterns.

// Test MustSupport Management - REMOVED for IETF draft-03 compliance.
// MustSupport field has been removed from VCon struct.
// MustSupport functionality should now be handled through proper extension patterns.

// Test Tag Management.
func TestTagManagement(t *testing.T) {
	vcon := NewWithDefaults()

	// Test getting non-existent tag
	tag := vcon.GetTag("category")
	if tag != nil {
		t.Errorf("Expected nil tag, got %s", *tag)
	}

	// Test adding tag
	vcon.AddTag("category", "meeting")
	tag = vcon.GetTag("category")
	if tag == nil || *tag != "meeting" {
		t.Errorf("Expected 'meeting', got %v", tag)
	}

	// Test adding another tag
	vcon.AddTag("priority", "high")
	priority := vcon.GetTag("priority")
	if priority == nil || *priority != "high" {
		t.Errorf("Expected 'high', got %v", priority)
	}

	// Verify both tags exist
	category := vcon.GetTag("category")
	if category == nil || *category != "meeting" {
		t.Errorf("Expected 'meeting' for category, got %v", category)
	}

	// Test getting tags attachment
	tagsAttachment := vcon.GetTags()
	if tagsAttachment == nil {
		t.Error("Expected tags attachment to exist")
	} else if tagsAttachment.Type == nil || *tagsAttachment.Type != "tags" {
		t.Errorf("Expected tags type, got %v", tagsAttachment.Type)
	}
}

// Test Search Methods.
func TestFindAttachmentByType(t *testing.T) {
	vcon := NewWithDefaults()

	// Test finding non-existent attachment
	attachment := vcon.FindAttachmentByType("metadata")
	if attachment != nil {
		t.Error("Expected nil attachment")
	}

	// Add attachment
	metadataAttachment := Attachment{
		Type:     StringPtr("metadata"),
		Body:     map[string]interface{}{"version": "1.0"},
		Encoding: StringPtr("json"),
		Start:    func() *time.Time { t := time.Now().UTC(); return &t }(),
		Party:    IntPtr(0),
	}
	vcon.AddAttachment(metadataAttachment)

	// Test finding existing attachment
	found := vcon.FindAttachmentByType("metadata")
	if found == nil {
		t.Error("Expected to find metadata attachment")
	} else if found.Type == nil || *found.Type != "metadata" {
		t.Errorf("Expected metadata type, got %v", found.Type)
	}
}

func TestFindPartyIndex(t *testing.T) {
	vcon := NewWithDefaults()

	// Add parties
	party1 := Party{Name: StringPtr("John Doe"), Tel: StringPtr("+1234567890")}
	party2 := Party{Name: StringPtr("Jane Smith"), Mailto: StringPtr("jane@example.com")}
	vcon.AddParty(party1)
	vcon.AddParty(party2)

	// Test finding by name
	index := vcon.FindPartyIndex("name", "John Doe")
	if index == nil || *index != 0 {
		t.Errorf("Expected index 0, got %v", index)
	}

	// Test finding by tel
	index = vcon.FindPartyIndex("tel", "+1234567890")
	if index == nil || *index != 0 {
		t.Errorf("Expected index 0, got %v", index)
	}

	// Test finding by mailto
	index = vcon.FindPartyIndex("mailto", "jane@example.com")
	if index == nil || *index != 1 {
		t.Errorf("Expected index 1, got %v", index)
	}

	// Test finding non-existent
	index = vcon.FindPartyIndex("name", "Bob")
	if index != nil {
		t.Errorf("Expected nil, got %v", index)
	}
}

func TestFindDialogsByType(t *testing.T) {
	vcon := NewWithDefaults()

	// Add different dialog types
	textDialog := Dialog{
		Type:    "text",
		Start:   time.Now(),
		Parties: NewDialogPartiesArrayPtr([]int{0}),
		Body:    "Hello",
	}

	recordingDialog := Dialog{
		Type:      "recording",
		Start:     time.Now().Add(time.Minute),
		Parties:   NewDialogPartiesArrayPtr([]int{0, 1}),
		Mediatype: StringPtr("audio/wav"),
	}

	transferDialog := Dialog{
		Type:    "transfer",
		Start:   time.Now().Add(2 * time.Minute),
		Parties: NewDialogPartiesArrayPtr([]int{0, 1}),
	}

	vcon.AddDialog(textDialog)
	vcon.AddDialog(recordingDialog)
	vcon.AddDialog(transferDialog)

	// Test finding text dialogs
	textDialogs := vcon.FindDialogsByType("text")
	if len(textDialogs) != 1 {
		t.Errorf("Expected 1 text dialog, got %d", len(textDialogs))
	}

	// Test finding transfer dialogs
	transferDialogs := vcon.FindDialogsByType("transfer")
	if len(transferDialogs) != 1 {
		t.Errorf("Expected 1 transfer dialog, got %d", len(transferDialogs))
	}

	// Test finding non-existent type
	videoDialogs := vcon.FindDialogsByType("video")
	if len(videoDialogs) != 0 {
		t.Errorf("Expected 0 video dialogs, got %d", len(videoDialogs))
	}
}

// Test Specialized Dialog Creators.
func TestAddTransferDialog(t *testing.T) {
	vcon := NewWithDefaults()

	transferData := map[string]any{
		"reason": "Call forwarded",
		"from":   "+1234567890",
		"to":     "+1987654321",
	}

	metadata := map[string]any{
		"system": "PBX",
	}

	index := vcon.AddTransferDialog(time.Now(), transferData, []int{0, 1}, metadata)
	if index != 0 {
		t.Errorf("Expected index 0, got %d", index)
	}

	// Verify dialog was added correctly
	dialog := vcon.GetDialog(0)
	if dialog == nil {
		t.Error("Expected dialog to be added")
		return
	}
	if dialog.Type != "transfer" {
		t.Errorf("Expected transfer type, got %s", dialog.Type)
	}
	if dialog.Transfer == nil {
		t.Error("Expected transfer data")
	}
	if dialog.Meta == nil {
		t.Error("Expected meta")
	}
}

func TestAddIncompleteDialog(t *testing.T) {
	vcon := NewWithDefaults()

	details := map[string]any{
		"ringDuration": 45000,
		"reason":       "No answer",
	}

	metadata := map[string]any{
		"attempts": 1,
	}

	index := vcon.AddIncompleteDialog(time.Now(), "NO_ANSWER", details, []int{0}, metadata)
	if index != 0 {
		t.Errorf("Expected index 0, got %d", index)
	}

	// Verify dialog was added correctly
	dialog := vcon.GetDialog(0)
	if dialog == nil {
		t.Error("Expected dialog to be added")
		return
	}
	if dialog.Type != "incomplete" {
		t.Errorf("Expected incomplete type, got %s", dialog.Type)
	}
	if dialog.Disposition == nil || *dialog.Disposition != "NO_ANSWER" {
		t.Errorf("Expected NO_ANSWER disposition, got %v", dialog.Disposition)
	}
	if dialog.Body == nil {
		t.Error("Expected body with details")
	}
}

// Test Analysis Management.
func TestAddAnalysis(t *testing.T) {
	vcon := NewWithDefaults()

	schemaStr := "version: 1.0"
	productStr := "sentiment-analyzer"
	meta := map[string]any{"confidence": "high"}

	vcon.AddAnalysis("sentiment", []int{0}, "acme", map[string]any{"score": 0.8}, "json", &productStr, &schemaStr, meta)

	if len(vcon.Analysis) != 1 {
		t.Errorf("Expected 1 analysis entry, got %d", len(vcon.Analysis))
	}

	analysis := vcon.Analysis[0]
	if analysis.Type != "sentiment" {
		t.Errorf("Expected sentiment type, got %v", analysis.Type)
	}
	if analysis.Vendor != "acme" {
		t.Errorf("Expected acme vendor, got %v", analysis.Vendor)
	}
	if analysis.Encoding == nil || *analysis.Encoding != "json" {
		encoding := "nil"
		if analysis.Encoding != nil {
			encoding = *analysis.Encoding
		}
		t.Errorf("Expected json encoding, got %v", encoding)
	}

	// Test finding analysis by type
	found := vcon.FindAnalysisByType("sentiment")
	if found == nil {
		t.Error("Expected to find sentiment analysis")
	} else if found.Type != "sentiment" {
		t.Errorf("Expected sentiment type, got %v", found.Type)
	}
}

// Helper functions.
func StringPtr(s string) *string {
	return &s
}
