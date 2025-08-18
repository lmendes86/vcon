package vcon

import (
	"fmt"
	"reflect"
	"strings"
)

// IETFStrictValidationError represents a strict IETF compliance error for extension fields.
type IETFStrictValidationError struct {
	Field     string
	Message   string
	Extension string
	Category  string
}

func (e IETFStrictValidationError) Error() string {
	return fmt.Sprintf("IETF strict validation error in %s: %s (extension: %s, category: %s)", e.Field, e.Message, e.Extension, e.Category)
}

// IETFStrictValidationErrors represents multiple IETF strict validation errors.
type IETFStrictValidationErrors []IETFStrictValidationError

func (e IETFStrictValidationErrors) Error() string {
	if len(e) == 0 {
		return "no IETF strict validation errors"
	}
	msgs := make([]string, 0, len(e))
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}
	return fmt.Sprintf("IETF strict validation failed with %d extension field violations: %s", len(msgs), msgs[0])
}

// ValidateIETFStrict performs strict IETF specification compliance validation.
// This mode flags any non-standard extension fields and provides migration recommendations.
func (v *VCon) ValidateIETFStrict() error {
	// First perform regular IETF validation
	if err := v.ValidateIETF(); err != nil {
		return err
	}

	var errors IETFStrictValidationErrors

	// Check for deprecated top-level fields
	errors = append(errors, v.validateStrictTopLevelFields()...)

	// Check for extension fields in all structs
	errors = append(errors, v.validateStrictPartyFields()...)
	errors = append(errors, v.validateStrictDialogFields()...)
	errors = append(errors, v.validateStrictAnalysisFields()...)
	errors = append(errors, v.validateStrictAttachmentFields()...)

	if len(errors) > 0 {
		return errors
	}
	return nil
}

// ValidateCompleteStrict performs comprehensive strict validation including both business rules
// and IETF extension field compliance. This combines ValidateStrict() and ValidateIETFStrict().
func (v *VCon) ValidateCompleteStrict() error {
	// First perform business rule strict validation
	if err := v.ValidateStrict(); err != nil {
		return err
	}

	// Then perform IETF extension field validation
	if err := v.ValidateIETFStrict(); err != nil {
		return err
	}

	return nil
}

// validateStrictTopLevelFields checks for deprecated top-level extension fields.
func (v *VCon) validateStrictTopLevelFields() []IETFStrictValidationError {
	var errors []IETFStrictValidationError

	// Extensions and MustSupport fields have been removed from VCon struct
	// No validation needed for these deprecated fields

	return errors
}

// validateStrictPartyFields checks for extension fields in Party objects.
func (v *VCon) validateStrictPartyFields() []IETFStrictValidationError {
	var errors []IETFStrictValidationError

	// IETF standard fields per specification
	ietfPartyFields := map[string]bool{
		"tel": true, "mailto": true, "name": true, "uuid": true,
		"gmlpos": true, "civicaddress": true, "role": true,
		"validation": true, // Actually extension but commonly used
		"meta":       true, // Standard IETF extension pattern
	}

	for i, party := range v.Parties {
		// Check for extension fields using reflection
		partyValue := reflect.ValueOf(party)
		partyType := partyValue.Type()

		for j := 0; j < partyValue.NumField(); j++ {
			field := partyType.Field(j)
			jsonTag := field.Tag.Get("json")
			if jsonTag == "" || jsonTag == "-" {
				continue
			}

			// Extract field name from JSON tag
			fieldName := jsonTag
			if comma := strings.IndexByte(jsonTag, ','); comma != -1 {
				fieldName = jsonTag[:comma]
			}

			// Skip if it's a standard IETF field
			if ietfPartyFields[fieldName] {
				continue
			}

			// Check if field has non-zero value
			fieldValue := partyValue.Field(j)
			if !isZeroValue(fieldValue) {
				category := categorizePartyExtension(fieldName)
				errors = append(errors, IETFStrictValidationError{
					Field:     fmt.Sprintf("parties[%d].%s", i, fieldName),
					Message:   fmt.Sprintf("Extension field '%s' not in IETF specification", fieldName),
					Extension: fieldName,
					Category:  category,
				})
			}
		}
	}

	return errors
}

// validateStrictDialogFields checks for extension fields in Dialog objects.
func (v *VCon) validateStrictDialogFields() []IETFStrictValidationError {
	var errors []IETFStrictValidationError

	// IETF standard fields per specification
	ietfDialogFields := map[string]bool{
		"type": true, "start": true, "parties": true, "originator": true,
		"mediatype": true, "filename": true, "body": true, "encoding": true,
		"url": true, "content_hash": true, "disposition": true, "party_history": true,
		"transferee": true, "transferor": true, "transfer-target": true,
		"original": true, "consultation": true, "target-dialog": true,
		"duration": true, "meta": true, // meta is standard extension pattern
	}

	for i, dialog := range v.Dialog {
		// Check for extension fields using reflection
		dialogValue := reflect.ValueOf(dialog)
		dialogType := dialogValue.Type()

		for j := 0; j < dialogValue.NumField(); j++ {
			field := dialogType.Field(j)
			jsonTag := field.Tag.Get("json")
			if jsonTag == "" || jsonTag == "-" {
				continue
			}

			// Extract field name from JSON tag
			fieldName := jsonTag
			if comma := strings.IndexByte(jsonTag, ','); comma != -1 {
				fieldName = jsonTag[:comma]
			}

			// Skip if it's a standard IETF field
			if ietfDialogFields[fieldName] {
				continue
			}

			// Check if field has non-zero value
			fieldValue := dialogValue.Field(j)
			if !isZeroValue(fieldValue) {
				category := categorizeDialogExtension(fieldName)
				errors = append(errors, IETFStrictValidationError{
					Field:     fmt.Sprintf("dialog[%d].%s", i, fieldName),
					Message:   fmt.Sprintf("Extension field '%s' not in IETF specification", fieldName),
					Extension: fieldName,
					Category:  category,
				})
			}
		}
	}

	return errors
}

// validateStrictAnalysisFields checks for extension fields in Analysis objects.
func (v *VCon) validateStrictAnalysisFields() []IETFStrictValidationError {
	var errors []IETFStrictValidationError

	// IETF standard fields per specification
	ietfAnalysisFields := map[string]bool{
		"type": true, "dialog": true, "vendor": true, "body": true,
		"encoding": true, "url": true, "content_hash": true, "meta": true,
	}

	for i, analysis := range v.Analysis {
		// Check for extension fields using reflection
		analysisValue := reflect.ValueOf(analysis)
		analysisType := analysisValue.Type()

		for j := 0; j < analysisValue.NumField(); j++ {
			field := analysisType.Field(j)
			jsonTag := field.Tag.Get("json")
			if jsonTag == "" || jsonTag == "-" {
				continue
			}

			// Extract field name from JSON tag
			fieldName := jsonTag
			if comma := strings.IndexByte(jsonTag, ','); comma != -1 {
				fieldName = jsonTag[:comma]
			}

			// Skip if it's a standard IETF field
			if ietfAnalysisFields[fieldName] {
				continue
			}

			// Check if field has non-zero value
			fieldValue := analysisValue.Field(j)
			if !isZeroValue(fieldValue) {
				category := categorizeAnalysisExtension(fieldName)
				errors = append(errors, IETFStrictValidationError{
					Field:     fmt.Sprintf("analysis[%d].%s", i, fieldName),
					Message:   fmt.Sprintf("Extension field '%s' not in IETF specification", fieldName),
					Extension: fieldName,
					Category:  category,
				})
			}
		}
	}

	return errors
}

// validateStrictAttachmentFields checks for extension fields in Attachment objects.
func (v *VCon) validateStrictAttachmentFields() []IETFStrictValidationError {
	var errors []IETFStrictValidationError

	// IETF standard fields per specification
	ietfAttachmentFields := map[string]bool{
		"type": true, "start": true, "party": true, "body": true,
		"encoding": true, "url": true, "content_hash": true,
		"meta": true, // Standard IETF extension pattern for attachments
	}

	for i, attachment := range v.Attachments {
		// Check for extension fields using reflection
		attachmentValue := reflect.ValueOf(attachment)
		attachmentType := attachmentValue.Type()

		for j := 0; j < attachmentValue.NumField(); j++ {
			field := attachmentType.Field(j)
			jsonTag := field.Tag.Get("json")
			if jsonTag == "" || jsonTag == "-" {
				continue
			}

			// Extract field name from JSON tag
			fieldName := jsonTag
			if comma := strings.IndexByte(jsonTag, ','); comma != -1 {
				fieldName = jsonTag[:comma]
			}

			// Skip if it's a standard IETF field
			if ietfAttachmentFields[fieldName] {
				continue
			}

			// Check if field has non-zero value
			fieldValue := attachmentValue.Field(j)
			if !isZeroValue(fieldValue) {
				category := categorizeAttachmentExtension(fieldName)
				errors = append(errors, IETFStrictValidationError{
					Field:     fmt.Sprintf("attachments[%d].%s", i, fieldName),
					Message:   fmt.Sprintf("Extension field '%s' not in IETF specification", fieldName),
					Extension: fieldName,
					Category:  category,
				})
			}
		}
	}

	return errors
}

// isZeroValue checks if a reflect.Value is the zero value for its type.
func isZeroValue(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}

	switch v.Kind() {
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return v.IsNil()
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if !isZeroValue(v.Index(i)) {
				return false
			}
		}
		return true
	case reflect.String:
		return v.String() == ""
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if !isZeroValue(v.Field(i)) {
				return false
			}
		}
		return true
	}
	return false
}

// Extension field categorization functions

func categorizePartyExtension(fieldName string) string {
	switch fieldName {
	case "stir", "sip", "validation":
		return "telephony"
	case "did":
		return "identity"
	case "contact_list", "jCard":
		return "contact"
	case "timezone":
		return "temporal"
	default:
		return "other"
	}
}

func categorizeDialogExtension(fieldName string) string {
	switch fieldName {
	case "resolution", "frame_rate", "codec", "bitrate", "thumbnail":
		return "video_media"
	case "campaign", "interaction", "skill":
		return "contact_center"
	case "session_id", "signaling":
		// Note: "application" and "message_id" removed for IETF draft-03 compliance
		return "technical"
	case "metadata", "transfer":
		return "metadata"
	default:
		return "other"
	}
}

func categorizeAnalysisExtension(fieldName string) string {
	switch fieldName {
	case "product", "schema":
		return "vendor"
	case "mediatype", "filename":
		return "content"
	default:
		return "other"
	}
}

func categorizeAttachmentExtension(fieldName string) string {
	switch fieldName {
	case "dialog":
		return "reference"
	case "mediatype", "filename":
		return "content"
	case "metadata":
		return "metadata"
	default:
		return "other"
	}
}
