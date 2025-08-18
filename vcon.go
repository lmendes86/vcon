// Package vcon provides Go types and utilities for working with vCon (virtual conversation) format.
// vCon is a standard for representing conversation data including recordings, text messages,
// transfers, and metadata in a structured JSON format.
package vcon

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// New creates a new VCon instance with required fields initialized.
// It generates a new UUID and sets the creation timestamp to the current time.
func New(version string) *VCon {
	id := uuid.New()
	return &VCon{
		UUID:      id,
		Vcon:      version,
		CreatedAt: time.Now().UTC(),
	}
}

// NewWithDefaults creates a new VCon instance with the IETF spec version "0.0.2".
func NewWithDefaults() *VCon {
	return New("0.0.2")
}

// AddParty adds a new party (participant) to the conversation.
// Returns the index of the added party.
func (v *VCon) AddParty(party Party) int {
	v.Parties = append(v.Parties, party)
	return len(v.Parties) - 1
}

// AddDialog adds a new dialog segment to the conversation.
// Returns the index of the added dialog.
func (v *VCon) AddDialog(dialog Dialog) int {
	v.Dialog = append(v.Dialog, dialog)
	return len(v.Dialog) - 1
}

// AddAttachment adds a new attachment to the conversation.
// Returns the index of the added attachment.
func (v *VCon) AddAttachment(attachment Attachment) int {
	v.Attachments = append(v.Attachments, attachment)
	return len(v.Attachments) - 1
}

// GetParty returns the party at the specified index.
// Returns nil if the index is out of bounds.
func (v *VCon) GetParty(index int) *Party {
	if index < 0 || index >= len(v.Parties) {
		return nil
	}
	return &v.Parties[index]
}

// GetDialog returns the dialog at the specified index.
// Returns nil if the index is out of bounds.
func (v *VCon) GetDialog(index int) *Dialog {
	if index < 0 || index >= len(v.Dialog) {
		return nil
	}
	return &v.Dialog[index]
}

// UpdateTimestamp sets the UpdatedAt field to the current time.
func (v *VCon) UpdateTimestamp() {
	now := time.Now().UTC()
	v.UpdatedAt = &now
}

// MarshalJSON implements the json.Marshaler interface.
// It handles custom JSON marshaling for the VCon type.
func (v *VCon) MarshalJSON() ([]byte, error) {
	// Create an alias to avoid infinite recursion
	type Alias VCon

	// Handle additional properties
	if len(v.AdditionalProperties) > 0 {
		// Marshal the main struct
		mainData, err := json.Marshal((*Alias)(v))
		if err != nil {
			return nil, err
		}

		// Unmarshal to a map to merge additional properties
		var result map[string]interface{}
		if err := json.Unmarshal(mainData, &result); err != nil {
			return nil, err
		}

		// Merge additional properties
		for k, v := range v.AdditionalProperties {
			if _, exists := result[k]; !exists {
				result[k] = v
			}
		}

		return json.Marshal(result)
	}

	return json.Marshal((*Alias)(v))
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// It handles custom JSON unmarshaling for the VCon type.
func (v *VCon) UnmarshalJSON(data []byte) error {
	// Create an alias to avoid infinite recursion
	type Alias VCon
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(v),
	}

	// First unmarshal to the struct
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// Then unmarshal to a map to capture additional properties
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Known fields that should not be in AdditionalProperties
	knownFields := map[string]bool{
		"uuid": true, "vcon": true, "created_at": true, "updated_at": true,
		"subject": true, "redacted": true, "group": true, "parties": true,
		"dialog": true, "attachments": true, "analysis": true, "signatures": true,
		"payload": true, "meta": true, "appended": true, "extensions": true, "must_support": true,
	}

	// Collect unknown fields
	v.AdditionalProperties = make(map[string]any)
	for key, value := range raw {
		if !knownFields[key] {
			var val any
			if err := json.Unmarshal(value, &val); err != nil {
				return fmt.Errorf("failed to unmarshal additional property %s: %w", key, err)
			}
			v.AdditionalProperties[key] = val
		}
	}

	// Clean up if no additional properties
	if len(v.AdditionalProperties) == 0 {
		v.AdditionalProperties = nil
	}

	return nil
}

// Tag Management Methods

// FindAttachmentByType finds the first attachment with the specified type.
func (v *VCon) FindAttachmentByType(attachmentType string) *Attachment {
	for i := range v.Attachments {
		if v.Attachments[i].Type != nil && *v.Attachments[i].Type == attachmentType {
			return &v.Attachments[i]
		}
	}
	return nil
}

// GetTags returns the tags attachment from the vCon.
func (v *VCon) GetTags() *Attachment {
	return v.FindAttachmentByType("tags")
}

// GetTag retrieves the value of a specific tag by name.
func (v *VCon) GetTag(tagName string) *string {
	tagsAttachment := v.FindAttachmentByType("tags")
	if tagsAttachment == nil {
		return nil
	}

	// Tags are stored as []string in the body
	if tagList, ok := tagsAttachment.Body.([]interface{}); ok {
		for _, tag := range tagList {
			if tagStr, ok := tag.(string); ok {
				if len(tagStr) > len(tagName)+1 && tagStr[:len(tagName)+1] == tagName+":" {
					value := tagStr[len(tagName)+1:]
					return &value
				}
			}
		}
	}
	return nil
}

// AddTag adds a tag to the vCon. Creates a tags attachment if one doesn't exist.
func (v *VCon) AddTag(tagName, tagValue string) {
	tagsAttachment := v.FindAttachmentByType("tags")
	if tagsAttachment == nil {
		// Create new tags attachment
		tagList := []interface{}{tagName + ":" + tagValue}
		tagsType := "tags"
		jsonEncoding := "json"
		attachment := Attachment{
			Type:     &tagsType,
			Body:     tagList,
			Encoding: &jsonEncoding,
		}
		v.AddAttachment(attachment)
	} else {
		// Add to existing tags
		if tagList, ok := tagsAttachment.Body.([]interface{}); ok {
			tagList = append(tagList, tagName+":"+tagValue)
			tagsAttachment.Body = tagList
		} else {
			// Initialize as slice if not already
			tagsAttachment.Body = []interface{}{tagName + ":" + tagValue}
		}
	}
	v.UpdateTimestamp()
}

// FindAnalysisByType finds the first analysis entry with the specified type.
func (v *VCon) FindAnalysisByType(analysisType string) *Analysis {
	for _, analysis := range v.Analysis {
		if analysis.Type == analysisType {
			return &analysis
		}
	}
	return nil
}

// AddAnalysis adds analysis data to the vCon.
func (v *VCon) AddAnalysis(analysisType string, dialog interface{}, vendor string, body interface{}, encoding string, product *string, schema *string, meta map[string]any) {
	analysis := Analysis{
		Type:     analysisType,
		Dialog:   dialog,
		Vendor:   vendor,
		Body:     body,
		Encoding: &encoding,
		Product:  product,
		Schema:   schema,
		Meta:     meta,
	}

	v.Analysis = append(v.Analysis, analysis)
	v.UpdateTimestamp()
}

// Search Methods

// FindPartyIndex finds the index of a party by searching for a field value.
func (v *VCon) FindPartyIndex(fieldName, value string) *int {
	for i, party := range v.Parties {
		switch fieldName {
		case "name":
			if party.Name != nil && *party.Name == value {
				return &i
			}
		case "tel":
			if party.Tel != nil && *party.Tel == value {
				return &i
			}
		case "mailto":
			if party.Mailto != nil && *party.Mailto == value {
				return &i
			}
		case "uuid":
			if party.UUID != nil && party.UUID.String() == value {
				return &i
			}
		case "role":
			if party.Role != nil && *party.Role == value {
				return &i
			}
		}
	}
	return nil
}

// FindDialogsByType finds all dialog entries with the specified type.
func (v *VCon) FindDialogsByType(dialogType string) []Dialog {
	var results []Dialog
	for _, dialog := range v.Dialog {
		if dialog.Type == dialogType {
			results = append(results, dialog)
		}
	}
	return results
}

// Specialized Dialog Creators

// AddTransferDialog creates and adds a transfer dialog entry to the vCon.
func (v *VCon) AddTransferDialog(start time.Time, transferData map[string]any, parties []int, metadata map[string]any) int {
	var dialogParties *DialogParties
	if len(parties) > 0 {
		dp := NewDialogPartiesArray(parties)
		dialogParties = &dp
	}
	dialog := Dialog{
		Type:     "transfer",
		Start:    start,
		Parties:  dialogParties,
		Transfer: transferData,
		Meta:     metadata,
	}
	return v.AddDialog(dialog)
}

// AddIncompleteDialog creates and adds an incomplete dialog entry to the vCon.
func (v *VCon) AddIncompleteDialog(start time.Time, disposition string, details map[string]any, parties []int, metadata map[string]any) int {
	var dialogParties *DialogParties
	if len(parties) > 0 {
		dp := NewDialogPartiesArray(parties)
		dialogParties = &dp
	}
	dialog := Dialog{
		Type:        "incomplete",
		Start:       start,
		Parties:     dialogParties,
		Disposition: &disposition,
		Body:        details,
		Meta:        metadata,
	}
	return v.AddDialog(dialog)
}
