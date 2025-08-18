package vcon

import (
	"fmt"
	"strings"
	"time"
)

// ValidationError represents a validation error with details about what failed.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error in %s: %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors.
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return "no validation errors"
	}
	msgs := make([]string, 0, len(e))
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// Validate performs validation on the VCon structure using struct tags and business rules.
func (v *VCon) Validate() error {
	// First run the struct tag validation
	if err := ValidateStruct(v); err != nil {
		return err
	}

	// Then run business logic validation
	var errors ValidationErrors

	// Validate party indices in dialogs
	for i, dialog := range v.Dialog {
		if dialog.Parties != nil {
			allIndices := dialog.Parties.GetAllPartyIndices()
			for _, partyIdx := range allIndices {
				if partyIdx < 0 || partyIdx >= len(v.Parties) {
					errors = append(errors, ValidationError{
						Field:   fmt.Sprintf("dialog[%d].parties", i),
						Message: fmt.Sprintf("invalid party index %d", partyIdx),
					})
				}
			}
		}

		// Check originator index
		if dialog.Originator != nil && (*dialog.Originator < 0 || *dialog.Originator >= len(v.Parties)) {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("dialog[%d].originator", i),
				Message: fmt.Sprintf("invalid originator index %d", *dialog.Originator),
			})
		}

		// Check transfer-related indices
		if dialog.Transferee != nil && (*dialog.Transferee < 0 || *dialog.Transferee >= len(v.Parties)) {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("dialog[%d].transferee", i),
				Message: fmt.Sprintf("invalid transferee index %d", *dialog.Transferee),
			})
		}

		if dialog.Transferor != nil && (*dialog.Transferor < 0 || *dialog.Transferor >= len(v.Parties)) {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("dialog[%d].transferor", i),
				Message: fmt.Sprintf("invalid transferor index %d", *dialog.Transferor),
			})
		}

		if dialog.TransferTarget != nil && (*dialog.TransferTarget < 0 || *dialog.TransferTarget >= len(v.Parties)) {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("dialog[%d].transfer-target", i),
				Message: fmt.Sprintf("invalid transfer target index %d", *dialog.TransferTarget),
			})
		}
	}

	// NOTE: IETF spec allows parties without identifiers (e.g., redacted parties)
	// The requirement for at least one identifier is moved to ValidateStrict() as a business rule

	if len(errors) > 0 {
		return errors
	}
	return nil
}

// IsValid is a convenience method that returns true if the VCon is valid.
func (v *VCon) IsValid() (bool, []string) {
	if err := v.Validate(); err != nil {
		if validationErrors, ok := err.(ValidationErrors); ok {
			messages := make([]string, len(validationErrors))
			for i, e := range validationErrors {
				messages[i] = e.Message
			}
			return false, messages
		}
		return false, []string{err.Error()}
	}
	return true, []string{}
}

// IsValidBool is a convenience method that returns only a boolean for simple validation checks.
func (v *VCon) IsValidBool() bool {
	valid, _ := v.IsValid()
	return valid
}

// ValidateStrict performs strict validation including additional business rules.
// This validation level enforces practical business requirements that are stricter
// than the IETF specification allows. Use this for production systems where data
// quality is critical. For pure IETF spec compliance, use ValidateIETF() instead.
//
// Business Rules Enforced (beyond IETF spec):
// - Party identifiers required (IETF allows parties without identifiers)
// - Chronological dialog order enforced
// - Duplicate party UUID detection.
func (v *VCon) ValidateStrict() error {
	// First perform basic validation
	if err := v.Validate(); err != nil {
		return err
	}

	var errors ValidationErrors

	// Check for duplicate party UUIDs
	partyUUIDs := make(map[string]int)
	for i, party := range v.Parties {
		if party.UUID != nil {
			key := party.UUID.String()
			if prevIdx, exists := partyUUIDs[key]; exists {
				errors = append(errors, ValidationError{
					Field:   fmt.Sprintf("parties[%d].uuid", i),
					Message: fmt.Sprintf("duplicate UUID with party[%d]", prevIdx),
				})
			}
			partyUUIDs[key] = i
		}
	}

	// Business rule: Check that each Party has at least one identifier (stricter than IETF spec)
	for i, party := range v.Parties {
		hasIdentifier := false
		if (party.Tel != nil && *party.Tel != "") ||
			(party.Mailto != nil && *party.Mailto != "") ||
			party.UUID != nil ||
			(party.Name != nil && *party.Name != "") {
			hasIdentifier = true
		}
		if !hasIdentifier {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("parties[%d]", i),
				Message: "party should have at least one identifier (tel, mailto, uuid, or name) - business rule",
			})
		}
	}

	// Check dialog chronological order
	var lastTime time.Time
	for i, dialog := range v.Dialog {
		if !lastTime.IsZero() && dialog.Start.Before(lastTime) {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("dialog[%d].start", i),
				Message: "dialogs should be in chronological order",
			})
		}
		lastTime = dialog.Start
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}

// Advanced validation methods to match Python library

// ValidateAdvanced performs advanced validation with both IETF compliance and business logic validation.
// This method combines IETF specification compliance with additional business rules and Python library compatibility.
func (v *VCon) ValidateAdvanced() error {
	// First perform basic validation
	if err := v.Validate(); err != nil {
		return err
	}

	// Perform IETF compliance validation (includes most advanced validation now)
	if err := v.ValidateIETF(); err != nil {
		return err
	}

	// Additional business logic validation can be added here in the future
	// Currently, IETF validation covers all the advanced validation requirements

	return nil
}

// validateTopLevelObjectMutualExclusion validates that only one of Redacted, Appended, or Group objects is present.
// Per IETF spec, these objects are mutually exclusive.
func (v *VCon) validateTopLevelObjectMutualExclusion() []ValidationError {
	var errors []ValidationError

	// Count which top-level objects are present
	objectsPresent := 0
	var presentObjects []string

	if v.Redacted != nil {
		objectsPresent++
		presentObjects = append(presentObjects, "redacted")
	}

	if v.Appended != nil {
		objectsPresent++
		presentObjects = append(presentObjects, "appended")
	}

	if len(v.Group) > 0 {
		objectsPresent++
		presentObjects = append(presentObjects, "group")
	}

	// IETF spec requires mutual exclusion - only one type should be present
	if objectsPresent > 1 {
		errors = append(errors, ValidationError{
			Field:   "top_level_objects",
			Message: fmt.Sprintf("only one of redacted, appended, or group objects is allowed per IETF spec, found: %s", strings.Join(presentObjects, ", ")),
		})
	}

	return errors
}
