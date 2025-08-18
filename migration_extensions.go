package vcon

import (
	"fmt"
	"reflect"
	"strings"
)

// getActualValue gets the actual value from a reflect.Value, dereferencing pointers.
func getActualValue(v reflect.Value) any {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		return v.Elem().Interface()
	}
	return v.Interface()
}

// ExtensionMigrationResult represents the result of extension field migration.
type ExtensionMigrationResult struct {
	MigratedFields  []string       // List of extension fields that were migrated
	MigrationsCount int            // Total number of fields migrated
	Warnings        []string       // Any warnings during migration
	MetaData        map[string]any // The migrated extension data
}

// MigrateExtensionsToMeta migrates all extension fields to the meta field for IETF strict compliance.
// This creates a new VCon with extension fields moved to appropriate meta sections.
func (v *VCon) MigrateExtensionsToMeta() (*VCon, *ExtensionMigrationResult, error) {
	result := &ExtensionMigrationResult{
		MigratedFields: []string{},
		Warnings:       []string{},
		MetaData:       make(map[string]any),
	}

	// Create a copy of the VCon to avoid modifying the original
	newVCon := *v

	// Migrate top-level extensions
	if err := migrateTopLevelExtensions(&newVCon, result); err != nil {
		return nil, nil, fmt.Errorf("failed to migrate top-level extensions: %w", err)
	}

	// Migrate party extensions
	if len(newVCon.Parties) > 0 {
		newParties := make([]Party, len(newVCon.Parties))
		for i, party := range newVCon.Parties {
			newParty := party
			if err := migratePartyExtensions(&newParty, result, i); err != nil {
				return nil, nil, fmt.Errorf("failed to migrate party[%d] extensions: %w", i, err)
			}
			newParties[i] = newParty
		}
		newVCon.Parties = newParties
	}

	// Migrate dialog extensions
	if len(newVCon.Dialog) > 0 {
		newDialogs := make([]Dialog, len(newVCon.Dialog))
		for i, dialog := range newVCon.Dialog {
			newDialog := dialog
			if err := migrateDialogExtensions(&newDialog, result, i); err != nil {
				return nil, nil, fmt.Errorf("failed to migrate dialog[%d] extensions: %w", i, err)
			}
			newDialogs[i] = newDialog
		}
		newVCon.Dialog = newDialogs
	}

	// Migrate analysis extensions
	if len(newVCon.Analysis) > 0 {
		newAnalysis := make([]Analysis, len(newVCon.Analysis))
		for i, analysis := range newVCon.Analysis {
			newAnalysisItem := analysis
			if err := migrateAnalysisExtensions(&newAnalysisItem, result, i); err != nil {
				return nil, nil, fmt.Errorf("failed to migrate analysis[%d] extensions: %w", i, err)
			}
			newAnalysis[i] = newAnalysisItem
		}
		newVCon.Analysis = newAnalysis
	}

	// Migrate attachment extensions
	if len(newVCon.Attachments) > 0 {
		newAttachments := make([]Attachment, len(newVCon.Attachments))
		for i, attachment := range newVCon.Attachments {
			newAttachment := attachment
			if err := migrateAttachmentExtensions(&newAttachment, result, i); err != nil {
				return nil, nil, fmt.Errorf("failed to migrate attachment[%d] extensions: %w", i, err)
			}
			newAttachments[i] = newAttachment
		}
		newVCon.Attachments = newAttachments
	}

	result.MigrationsCount = len(result.MigratedFields)

	return &newVCon, result, nil
}

// migrateTopLevelExtensions migrates deprecated top-level extension fields.
// Note: Extensions and MustSupport fields have been removed from VCon struct.
// This function now only handles other potential extension fields.
func migrateTopLevelExtensions(_ *VCon, _ *ExtensionMigrationResult) error {
	// Extensions and MustSupport fields have been removed - no migration needed
	// This function is kept for potential future extension field migrations

	return nil
}

// migratePartyExtensions migrates party extension fields to meta.
func migratePartyExtensions(party *Party, result *ExtensionMigrationResult, index int) error {
	// IETF standard fields
	ietfFields := map[string]bool{
		"tel": true, "mailto": true, "name": true, "uuid": true,
		"gmlpos": true, "civicaddress": true, "role": true,
		"validation": true, // Common extension but keeping for now
	}

	extensions := make(map[string]any)
	partyValue := reflect.ValueOf(party).Elem()
	partyType := partyValue.Type()

	for i := 0; i < partyValue.NumField(); i++ {
		field := partyType.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" || jsonTag == "meta,omitempty" {
			continue
		}

		// Extract field name from JSON tag
		fieldName := jsonTag
		if comma := strings.IndexByte(jsonTag, ','); comma != -1 {
			fieldName = jsonTag[:comma]
		}

		// Skip if it's a standard IETF field
		if ietfFields[fieldName] {
			continue
		}

		// Check if field has non-zero value
		fieldValue := partyValue.Field(i)
		if !isZeroValue(fieldValue) {
			// Get the actual value, dereferencing pointers
			value := getActualValue(fieldValue)
			extensions[fieldName] = value
			result.MigratedFields = append(result.MigratedFields, fmt.Sprintf("parties[%d].%s", index, fieldName))

			// Clear the original field
			zeroValue := reflect.Zero(fieldValue.Type())
			fieldValue.Set(zeroValue)
		}
	}

	// Add extensions to party meta if we have any
	if len(extensions) > 0 {
		if party.Meta == nil {
			party.Meta = make(map[string]any)
		}
		party.Meta["extensions"] = extensions
		result.MetaData[fmt.Sprintf("party_%d", index)] = extensions
	}

	return nil
}

// migrateDialogExtensions migrates dialog extension fields to meta.
func migrateDialogExtensions(dialog *Dialog, result *ExtensionMigrationResult, index int) error {
	// IETF standard fields
	ietfFields := map[string]bool{
		"type": true, "start": true, "parties": true, "originator": true,
		"mediatype": true, "filename": true, "body": true, "encoding": true,
		"url": true, "content_hash": true, "disposition": true, "party_history": true,
		"transferee": true, "transferor": true, "transfer-target": true,
		"original": true, "consultation": true, "target-dialog": true,
		"duration": true, "meta": true, // meta is standard extension pattern
	}

	extensions := make(map[string]any)
	dialogValue := reflect.ValueOf(dialog).Elem()
	dialogType := dialogValue.Type()

	for i := 0; i < dialogValue.NumField(); i++ {
		field := dialogType.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" || jsonTag == "meta,omitempty" {
			continue
		}

		// Extract field name from JSON tag
		fieldName := jsonTag
		if comma := strings.IndexByte(jsonTag, ','); comma != -1 {
			fieldName = jsonTag[:comma]
		}

		// Skip if it's a standard IETF field
		if ietfFields[fieldName] {
			continue
		}

		// Check if field has non-zero value
		fieldValue := dialogValue.Field(i)
		if !isZeroValue(fieldValue) {
			// Get the actual value, dereferencing pointers
			value := getActualValue(fieldValue)
			extensions[fieldName] = value
			result.MigratedFields = append(result.MigratedFields, fmt.Sprintf("dialog[%d].%s", index, fieldName))

			// Clear the original field
			zeroValue := reflect.Zero(fieldValue.Type())
			fieldValue.Set(zeroValue)
		}
	}

	// Add extensions to dialog meta if we have any
	if len(extensions) > 0 {
		if dialog.Meta == nil {
			dialog.Meta = make(map[string]any)
		}
		dialog.Meta["extensions"] = extensions
		result.MetaData[fmt.Sprintf("dialog_%d", index)] = extensions
	}

	return nil
}

// migrateAnalysisExtensions migrates analysis extension fields to meta.
func migrateAnalysisExtensions(analysis *Analysis, result *ExtensionMigrationResult, index int) error {
	// IETF standard fields
	ietfFields := map[string]bool{
		"type": true, "dialog": true, "vendor": true, "body": true,
		"encoding": true, "url": true, "content_hash": true, "meta": true,
	}

	extensions := make(map[string]any)
	analysisValue := reflect.ValueOf(analysis).Elem()
	analysisType := analysisValue.Type()

	for i := 0; i < analysisValue.NumField(); i++ {
		field := analysisType.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" || jsonTag == "meta,omitempty" {
			continue
		}

		// Extract field name from JSON tag
		fieldName := jsonTag
		if comma := strings.IndexByte(jsonTag, ','); comma != -1 {
			fieldName = jsonTag[:comma]
		}

		// Skip if it's a standard IETF field
		if ietfFields[fieldName] {
			continue
		}

		// Check if field has non-zero value
		fieldValue := analysisValue.Field(i)
		if !isZeroValue(fieldValue) {
			// Get the actual value, dereferencing pointers
			value := getActualValue(fieldValue)
			extensions[fieldName] = value
			result.MigratedFields = append(result.MigratedFields, fmt.Sprintf("analysis[%d].%s", index, fieldName))

			// Clear the original field
			zeroValue := reflect.Zero(fieldValue.Type())
			fieldValue.Set(zeroValue)
		}
	}

	// Add extensions to analysis meta if we have any
	if len(extensions) > 0 {
		if analysis.Meta == nil {
			analysis.Meta = make(map[string]any)
		}
		analysis.Meta["extensions"] = extensions
		result.MetaData[fmt.Sprintf("analysis_%d", index)] = extensions
	}

	return nil
}

// migrateAttachmentExtensions migrates attachment extension fields to meta or appropriate locations.
func migrateAttachmentExtensions(attachment *Attachment, result *ExtensionMigrationResult, index int) error {
	// IETF standard fields
	ietfFields := map[string]bool{
		"type": true, "start": true, "party": true, "body": true,
		"encoding": true, "url": true, "content_hash": true,
	}

	extensions := make(map[string]any)
	attachmentValue := reflect.ValueOf(attachment).Elem()
	attachmentType := attachmentValue.Type()

	for i := 0; i < attachmentValue.NumField(); i++ {
		field := attachmentType.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" || jsonTag == "metadata,omitempty" {
			continue
		}

		// Extract field name from JSON tag
		fieldName := jsonTag
		if comma := strings.IndexByte(jsonTag, ','); comma != -1 {
			fieldName = jsonTag[:comma]
		}

		// Skip if it's a standard IETF field
		if ietfFields[fieldName] {
			continue
		}

		// Check if field has non-zero value
		fieldValue := attachmentValue.Field(i)
		if !isZeroValue(fieldValue) {
			// Get the actual value, dereferencing pointers
			value := getActualValue(fieldValue)

			// Special handling for some fields
			switch fieldName {
			case "dialog":
				// Dialog reference is useful but not in IETF spec
				extensions[fieldName] = value
				result.Warnings = append(result.Warnings, fmt.Sprintf("attachment[%d].dialog field migrated - consider alternative referencing", index))
			case "metadata":
				// Metadata field can stay as is - it's a recognized pattern
				// Don't migrate this one, it's acceptable
				continue
			default:
				extensions[fieldName] = value
			}

			result.MigratedFields = append(result.MigratedFields, fmt.Sprintf("attachments[%d].%s", index, fieldName))

			// Clear the original field
			zeroValue := reflect.Zero(fieldValue.Type())
			fieldValue.Set(zeroValue)
		}
	}

	// Add extensions to attachment meta if we have any
	if len(extensions) > 0 {
		if attachment.Meta == nil {
			attachment.Meta = make(map[string]any)
		}
		if attachment.Meta["extensions"] == nil {
			attachment.Meta["extensions"] = extensions
		} else {
			// Merge with existing meta
			if existing, ok := attachment.Meta["extensions"].(map[string]any); ok {
				for k, v := range extensions {
					existing[k] = v
				}
			}
		}
		result.MetaData[fmt.Sprintf("attachment_%d", index)] = extensions
	}

	return nil
}

// RestoreExtensionsFromMeta attempts to restore extension fields from meta back to their original locations.
// This is the reverse operation of MigrateExtensionsToMeta.
func (v *VCon) RestoreExtensionsFromMeta() (*VCon, *ExtensionMigrationResult, error) {
	result := &ExtensionMigrationResult{
		MigratedFields: []string{},
		Warnings:       []string{},
		MetaData:       make(map[string]any),
	}

	// Create a copy of the VCon to avoid modifying the original
	newVCon := *v

	// Restore top-level extensions
	if err := restoreTopLevelExtensions(&newVCon, result); err != nil {
		return nil, nil, fmt.Errorf("failed to restore top-level extensions: %w", err)
	}

	// Note: Restoration of struct-level extensions would require more complex reflection
	// to populate the original fields. This is left for future implementation if needed.
	result.Warnings = append(result.Warnings, "Struct-level extension restoration not yet implemented")

	result.MigrationsCount = len(result.MigratedFields)

	return &newVCon, result, nil
}

// restoreTopLevelExtensions restores deprecated top-level fields from meta.
// Note: Extensions, MustSupport, and top-level Meta fields have been removed from VCon struct for IETF draft-03 compliance.
// This function cannot restore these fields as they no longer exist.
func restoreTopLevelExtensions(_ *VCon, result *ExtensionMigrationResult) error {
	// Top-level Meta field has been removed for IETF draft-03 compliance
	// Extensions and MustSupport fields were previously removed
	// Cannot restore deprecated fields that no longer exist in the struct

	result.Warnings = append(result.Warnings, "Cannot restore deprecated top-level Meta field - field removed from VCon struct for IETF compliance")
	result.Warnings = append(result.Warnings, "Cannot restore deprecated Extensions field - field removed from VCon struct")
	result.Warnings = append(result.Warnings, "Cannot restore deprecated MustSupport field - field removed from VCon struct")

	return nil
}
