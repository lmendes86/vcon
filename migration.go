package vcon

import (
	"errors"
	"fmt"
	"time"
)

// MigrationMode defines how to handle migration between formats.
type MigrationMode int

const (
	// MigrationModeStrict fails if data cannot be migrated perfectly.
	MigrationModeStrict MigrationMode = iota
	// MigrationModeLenient attempts to migrate as much as possible.
	MigrationModeLenient
	// MigrationModePreserve keeps unknown fields in AdditionalProperties.
	MigrationModePreserve
)

// ErrNoMigrationNeeded is returned when no migration is required.
var ErrNoMigrationNeeded = errors.New("no migration needed")

// AnalysisMapToStruct converts an old-style Analysis map to the new struct format.
func AnalysisMapToStruct(analysisMap map[string]any, mode MigrationMode) (*Analysis, error) {
	analysis := &Analysis{
		AdditionalProperties: make(map[string]any),
	}

	// Handle required fields
	if err := migrateAnalysisRequiredFields(analysisMap, analysis, mode); err != nil {
		return nil, err
	}

	// Handle optional fields
	migrateAnalysisOptionalFields(analysisMap, analysis)

	// Handle content hash migration
	migrateAnalysisContentHash(analysisMap, analysis)

	// Handle additional properties
	if mode == MigrationModePreserve {
		migrateAnalysisAdditionalProperties(analysisMap, analysis)
	}

	return analysis, nil
}

// migrateAnalysisRequiredFields handles migration of required Analysis fields.
func migrateAnalysisRequiredFields(analysisMap map[string]any, analysis *Analysis, mode MigrationMode) error {
	if typeVal, ok := analysisMap["type"].(string); ok {
		analysis.Type = typeVal
	} else if mode == MigrationModeStrict {
		return fmt.Errorf("missing or invalid 'type' field")
	}

	if vendorVal, ok := analysisMap["vendor"].(string); ok {
		analysis.Vendor = vendorVal
	} else if mode == MigrationModeStrict {
		return fmt.Errorf("missing or invalid 'vendor' field")
	}

	if bodyVal, ok := analysisMap["body"]; ok {
		analysis.Body = bodyVal
	} else if mode == MigrationModeStrict {
		return fmt.Errorf("missing 'body' field")
	}

	if encodingVal, ok := analysisMap["encoding"].(string); ok {
		analysis.Encoding = &encodingVal
	} else if mode == MigrationModeStrict {
		return fmt.Errorf("missing or invalid 'encoding' field")
	}

	return nil
}

// migrateAnalysisOptionalFields handles migration of optional Analysis fields.
func migrateAnalysisOptionalFields(analysisMap map[string]any, analysis *Analysis) {
	if dialogVal, ok := analysisMap["dialog"]; ok {
		analysis.Dialog = dialogVal
	}

	if productVal, ok := analysisMap["product"].(string); ok && productVal != "" {
		analysis.Product = &productVal
	}

	if schemaVal, ok := analysisMap["schema"].(string); ok && schemaVal != "" {
		analysis.Schema = &schemaVal
	}

	if urlVal, ok := analysisMap["url"].(string); ok && urlVal != "" {
		analysis.URL = &urlVal
	}

	if filenameVal, ok := analysisMap["filename"].(string); ok && filenameVal != "" {
		analysis.Filename = &filenameVal
	}

	if mediatypeVal, ok := analysisMap["mediatype"].(string); ok && mediatypeVal != "" {
		analysis.Mediatype = &mediatypeVal
	}

	if metaVal, ok := analysisMap["meta"].(map[string]any); ok {
		analysis.Meta = metaVal
	}
}

// migrateAnalysisContentHash handles content hash migration from legacy alg+signature.
func migrateAnalysisContentHash(analysisMap map[string]any, analysis *Analysis) {
	if contentHashVal, ok := analysisMap["content_hash"].(string); ok && contentHashVal != "" {
		analysis.ContentHash = NewContentHashSingle(contentHashVal)
		return
	}

	// Convert legacy alg+signature to content_hash format
	if algVal, ok := analysisMap["alg"].(string); ok && algVal != "" {
		if sigVal, ok := analysisMap["signature"].(string); ok && sigVal != "" {
			contentHashStr := MigrateAlgSignatureToContentHash(&algVal, &sigVal)
			if contentHashStr != nil {
				analysis.ContentHash = NewContentHashSingle(*contentHashStr)
			}
		}
	}
}

// migrateAnalysisAdditionalProperties handles migration of additional properties.
func migrateAnalysisAdditionalProperties(analysisMap map[string]any, analysis *Analysis) {
	knownFields := map[string]bool{
		"type": true, "vendor": true, "body": true, "encoding": true,
		"dialog": true, "product": true, "schema": true, "meta": true,
		"url": true, "alg": true, "signature": true, "filename": true, "mediatype": true,
		"content_hash": true,
	}
	for k, v := range analysisMap {
		if !knownFields[k] {
			analysis.AdditionalProperties[k] = v
		}
	}
}

// AnalysisStructToMap converts a new Analysis struct to the old map format.
func AnalysisStructToMap(analysis *Analysis) map[string]any {
	result := make(map[string]any)

	if analysis.Type != "" {
		result["type"] = analysis.Type
	}
	result["vendor"] = analysis.Vendor
	result["body"] = analysis.Body
	if analysis.Encoding != nil {
		result["encoding"] = *analysis.Encoding
	}

	if analysis.Dialog != nil {
		result["dialog"] = analysis.Dialog
	}

	if analysis.Product != nil && *analysis.Product != "" {
		result["product"] = *analysis.Product
	}

	if analysis.Schema != nil && *analysis.Schema != "" {
		result["schema"] = *analysis.Schema
	}

	// Handle URL and content hash fields
	if analysis.URL != nil && *analysis.URL != "" {
		result["url"] = *analysis.URL
	}

	// Convert content_hash back to legacy alg+signature format for backward compatibility
	if analysis.ContentHash != nil && !analysis.ContentHash.IsEmpty() {
		contentHashStr := analysis.ContentHash.GetSingle()
		alg, signature, err := ConvertContentHashToAlgSignature(&contentHashStr)
		if err == nil && alg != nil && signature != nil {
			result["alg"] = *alg
			result["signature"] = *signature
		} else {
			// If conversion fails, store as content_hash (forward compatibility)
			result["content_hash"] = contentHashStr
		}
	}

	// Handle filename field
	if analysis.Filename != nil && *analysis.Filename != "" {
		result["filename"] = *analysis.Filename
	}

	// Handle mediatype field
	if analysis.Mediatype != nil && *analysis.Mediatype != "" {
		result["mediatype"] = *analysis.Mediatype
	}

	if analysis.Meta != nil {
		result["meta"] = analysis.Meta
	}

	// Include additional properties
	for k, v := range analysis.AdditionalProperties {
		if _, exists := result[k]; !exists {
			result[k] = v
		}
	}

	return result
}

// MigrateVConToIETF migrates a VCon to be IETF-compliant.
func MigrateVConToIETF(v *VCon, mode MigrationMode) error {
	// Update version to IETF spec
	if v.Vcon != IETFVersion {
		if mode == MigrationModeStrict {
			return fmt.Errorf("vcon version is %s, expected %s", v.Vcon, IETFVersion)
		}
		v.Vcon = IETFVersion
	}

	// Ensure parties array exists (required per IETF)
	if v.Parties == nil {
		v.Parties = []Party{}
	}

	// Validate party identifiers
	for i, party := range v.Parties {
		if !hasPartyIdentifier(party) {
			if mode == MigrationModeStrict {
				return fmt.Errorf("party at index %d has no identifier", i)
			}
			// In lenient mode, add a generic name if no identifier exists
			if mode == MigrationModeLenient {
				if party.Name == nil {
					name := fmt.Sprintf("Party %d", i)
					v.Parties[i].Name = &name
				}
			}
		}
	}

	// Ensure dialog disposition for incomplete type
	for i, dialog := range v.Dialog {
		if dialog.Type == "incomplete" && (dialog.Disposition == nil || *dialog.Disposition == "") {
			if mode == MigrationModeStrict {
				return fmt.Errorf("incomplete dialog at index %d missing disposition", i)
			}
			if mode == MigrationModeLenient {
				disposition := "unknown"
				v.Dialog[i].Disposition = &disposition
			}
		}
	}

	// Analysis is already in the new struct format after type changes

	return nil
}

// ConvertFromLegacyAnalysis converts legacy analysis format (map) to new struct format.
func ConvertFromLegacyAnalysis(legacy []map[string]any, mode MigrationMode) ([]Analysis, error) {
	result := make([]Analysis, 0, len(legacy))

	for i, analysisMap := range legacy {
		analysis, err := AnalysisMapToStruct(analysisMap, mode)
		if err != nil {
			if mode == MigrationModeStrict {
				return nil, fmt.Errorf("failed to convert analysis at index %d: %w", i, err)
			}
			// In lenient mode, skip failed conversions
			continue
		}
		result = append(result, *analysis)
	}

	return result, nil
}

// ConvertToLegacyAnalysis converts new Analysis struct format to legacy map format.
func ConvertToLegacyAnalysis(analyses []Analysis) []map[string]any {
	result := make([]map[string]any, len(analyses))

	for i, analysis := range analyses {
		result[i] = AnalysisStructToMap(&analysis)
	}

	return result
}

// TimePtr is a helper function to create a pointer to a time.Time.
func TimePtr(t time.Time) *time.Time {
	return &t
}

// ConvertLegacyRedactedObject converts legacy redacted format (map[string]any) to RedactedObject.
func ConvertLegacyRedactedObject(legacy map[string]any, mode MigrationMode) (*RedactedObject, error) {
	if legacy == nil {
		return nil, ErrNoMigrationNeeded
	}

	redacted := &RedactedObject{}

	// UUID is required
	if uuidVal, ok := legacy["uuid"].(string); ok && uuidVal != "" {
		redacted.UUID = uuidVal
	} else {
		if mode == MigrationModeStrict {
			return nil, fmt.Errorf("missing or invalid 'uuid' field in redacted object")
		}
		// In lenient mode, generate a placeholder
		redacted.UUID = "legacy-redacted-" + fmt.Sprintf("%d", len(legacy))
	}

	// Optional fields
	if typeVal, ok := legacy["type"].(string); ok && typeVal != "" {
		redacted.Type = &typeVal
	}

	if bodyVal, ok := legacy["body"].(string); ok && bodyVal != "" {
		redacted.Body = &bodyVal
	}

	if encodingVal, ok := legacy["encoding"].(string); ok && encodingVal != "" {
		redacted.Encoding = &encodingVal
	}

	if urlVal, ok := legacy["url"].(string); ok && urlVal != "" {
		redacted.URL = &urlVal
	}

	if hashVal, ok := legacy["content_hash"].(string); ok && hashVal != "" {
		redacted.ContentHash = NewContentHashSingle(hashVal)
	}

	return redacted, nil
}

// ConvertLegacyAppendedObject converts legacy appended format (map[string]any or []any) to AppendedObject.
func ConvertLegacyAppendedObject(legacy any, mode MigrationMode) (*AppendedObject, error) {
	if legacy == nil {
		return nil, ErrNoMigrationNeeded
	}

	// Handle both single object and array (take first element)
	var appendedMap map[string]any

	switch v := legacy.(type) {
	case map[string]any:
		appendedMap = v
	case []any:
		if len(v) > 0 {
			if firstMap, ok := v[0].(map[string]any); ok {
				appendedMap = firstMap
			} else {
				if mode == MigrationModeStrict {
					return nil, fmt.Errorf("invalid appended object format in array")
				}
				return nil, ErrNoMigrationNeeded // Skip in lenient mode
			}
		} else {
			return nil, ErrNoMigrationNeeded // Empty array
		}
	case []map[string]any:
		if len(v) > 0 {
			appendedMap = v[0]
		} else {
			return nil, ErrNoMigrationNeeded // Empty array
		}
	default:
		if mode == MigrationModeStrict {
			return nil, fmt.Errorf("unsupported appended object format: %T", legacy)
		}
		return nil, ErrNoMigrationNeeded // Skip in lenient mode
	}

	if appendedMap == nil {
		return nil, ErrNoMigrationNeeded
	}

	appended := &AppendedObject{}

	// UUID is required
	if uuidVal, ok := appendedMap["uuid"].(string); ok && uuidVal != "" {
		appended.UUID = uuidVal
	} else {
		if mode == MigrationModeStrict {
			return nil, fmt.Errorf("missing or invalid 'uuid' field in appended object")
		}
		// In lenient mode, generate a placeholder
		appended.UUID = "legacy-appended-" + fmt.Sprintf("%d", len(appendedMap))
	}

	// Optional fields
	if typeVal, ok := appendedMap["type"].(string); ok && typeVal != "" {
		appended.Type = &typeVal
	}

	if bodyVal, ok := appendedMap["body"].(string); ok && bodyVal != "" {
		appended.Body = &bodyVal
	}

	if encodingVal, ok := appendedMap["encoding"].(string); ok && encodingVal != "" {
		appended.Encoding = &encodingVal
	}

	if urlVal, ok := appendedMap["url"].(string); ok && urlVal != "" {
		appended.URL = &urlVal
	}

	if hashVal, ok := appendedMap["content_hash"].(string); ok && hashVal != "" {
		appended.ContentHash = NewContentHashSingle(hashVal)
	}

	return appended, nil
}

// ConvertLegacyGroupObjects converts legacy group format ([]any) to []GroupObject.
func ConvertLegacyGroupObjects(legacy []any, mode MigrationMode) ([]GroupObject, error) {
	if legacy == nil {
		return nil, ErrNoMigrationNeeded
	}

	result := make([]GroupObject, 0, len(legacy))

	for i, item := range legacy {
		var groupMap map[string]any

		switch v := item.(type) {
		case map[string]any:
			groupMap = v
		case string:
			// Legacy format might just be UUIDs
			result = append(result, GroupObject{UUID: v})
			continue
		default:
			if mode == MigrationModeStrict {
				return nil, fmt.Errorf("invalid group object format at index %d: %T", i, item)
			}
			continue // Skip in lenient mode
		}

		group := GroupObject{}

		// UUID is required
		if uuidVal, ok := groupMap["uuid"].(string); ok && uuidVal != "" {
			group.UUID = uuidVal
		} else {
			if mode == MigrationModeStrict {
				return nil, fmt.Errorf("missing or invalid 'uuid' field in group object at index %d", i)
			}
			// In lenient mode, generate a placeholder
			group.UUID = "legacy-group-" + fmt.Sprintf("%d", i)
		}

		// Optional fields
		if typeVal, ok := groupMap["type"].(string); ok && typeVal != "" {
			group.Type = &typeVal
		}

		if bodyVal, ok := groupMap["body"].(string); ok && bodyVal != "" {
			group.Body = &bodyVal
		}

		if encodingVal, ok := groupMap["encoding"].(string); ok && encodingVal != "" {
			group.Encoding = &encodingVal
		}

		if urlVal, ok := groupMap["url"].(string); ok && urlVal != "" {
			group.URL = &urlVal
		}

		if hashVal, ok := groupMap["content_hash"].(string); ok && hashVal != "" {
			group.ContentHash = NewContentHashSingle(hashVal)
		}

		result = append(result, group)
	}

	return result, nil
}

// MigrateTopLevelObjects migrates legacy top-level object formats to new IETF-compliant structures.
func (v *VCon) MigrateTopLevelObjects(_ MigrationMode) error {
	// This function can be used to migrate vCons that were loaded with generic JSON unmarshaling
	// and contain map[string]any or []any in top-level object fields

	// Note: In normal usage, the new struct definitions should handle this automatically
	// This function is primarily for cases where raw JSON needs to be migrated

	return nil // No migration needed for properly typed structures
}
