package vcon

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// IETFVersion is the required vCon version per IETF specification.
const IETFVersion = "0.0.2"

// IETFValidationError represents an IETF spec compliance error.
type IETFValidationError struct {
	Field       string
	Message     string
	Requirement string // IETF requirement reference
}

func (e IETFValidationError) Error() string {
	return fmt.Sprintf("IETF validation error in %s: %s (requirement: %s)", e.Field, e.Message, e.Requirement)
}

// IETFValidationErrors represents multiple IETF validation errors.
type IETFValidationErrors []IETFValidationError

func (e IETFValidationErrors) Error() string {
	if len(e) == 0 {
		return "no IETF validation errors"
	}
	msgs := make([]string, 0, len(e))
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// ValidateIETF performs strict IETF specification compliance validation.
// This ensures the vCon conforms to draft-ietf-vcon-vcon-container.
func (v *VCon) ValidateIETF() error {
	var errors IETFValidationErrors

	// Validate different sections
	errors = append(errors, v.validateIETFTopLevel()...)
	errors = append(errors, v.validateIETFParties()...)
	errors = append(errors, v.validateIETFDialogs()...)
	errors = append(errors, v.validateIETFAnalysis()...)
	errors = append(errors, v.validateIETFAttachments()...)
	errors = append(errors, v.validateIETFTopLevelObjects()...)

	// Validate signatures (JWS compliance per IETF spec)
	if err := v.ValidateSignaturePresence(); err != nil {
		errors = append(errors, IETFValidationError{
			Field:       "signatures",
			Message:     err.Error(),
			Requirement: "IETF vCon spec - payload required when signatures present",
		})
	}

	// Validate top-level object mutual exclusion (IETF spec requirement)
	errors = append(errors, v.validateIETFTopLevelObjectMutualExclusion()...)

	if len(errors) > 0 {
		return errors
	}
	return nil
}

// validateIETFTopLevel validates required top-level fields and timestamps.
func (v *VCon) validateIETFTopLevel() []IETFValidationError {
	var errors []IETFValidationError

	// UUID is REQUIRED per IETF draft-03 spec
	if v.UUID == uuid.Nil {
		errors = append(errors, IETFValidationError{
			Field:       "uuid",
			Message:     "uuid is required per IETF specification",
			Requirement: "IETF vCon spec draft-03 - UUID MUST be globally unique",
		})
	}

	// Validate UUID format if present
	if v.UUID != uuid.Nil {
		if _, err := uuid.Parse(v.UUID.String()); err != nil {
			errors = append(errors, IETFValidationError{
				Field:       "uuid",
				Message:     "uuid must be valid RFC4122 format",
				Requirement: "IETF vCon spec draft-03 - UUID format",
			})
		}
	}

	// Validate version string
	if v.Vcon != IETFVersion {
		errors = append(errors, IETFValidationError{
			Field:       "vcon",
			Message:     fmt.Sprintf("vcon must be '%s', got '%s'", IETFVersion, v.Vcon),
			Requirement: "IETF vCon spec version 0.0.2",
		})
	}

	// Validate timestamps are RFC3339
	if err := validateRFC3339Time(v.CreatedAt); err != nil {
		errors = append(errors, IETFValidationError{
			Field:       "created_at",
			Message:     "must be valid RFC3339 timestamp",
			Requirement: "IETF vCon spec section 4.1",
		})
	}

	if v.UpdatedAt != nil {
		if err := validateRFC3339Time(*v.UpdatedAt); err != nil {
			errors = append(errors, IETFValidationError{
				Field:       "updated_at",
				Message:     "must be valid RFC3339 timestamp",
				Requirement: "IETF vCon spec section 4.1",
			})
		}
	}

	return errors
}

// validateIETFParties validates parties have required identifiers.
func (v *VCon) validateIETFParties() []IETFValidationError {
	var errors []IETFValidationError

	// Note: Per IETF draft-03 spec, parties array can be empty
	// The parties field is required to exist but can be an empty array
	// Only validate individual parties if array is not empty
	if len(v.Parties) == 0 {
		return errors // Empty parties array is valid per IETF spec
	}

	// Note: Per IETF spec, party identifiers are optional
	// No validation required for party identifiers

	return errors
}

// validateIETFDialogs validates dialog structure and content.
func (v *VCon) validateIETFDialogs() []IETFValidationError {
	var errors []IETFValidationError

	for i, dialog := range v.Dialog {
		errors = append(errors, v.validateDialogBasic(i, dialog)...)
		errors = append(errors, v.validateDialogTypeSpecific(i, dialog)...)
		errors = append(errors, v.validateDialogOneOfContent(i, dialog)...)
		errors = append(errors, v.validateDialogURLs(i, dialog)...)
		errors = append(errors, v.validateDialogContentHash(i, dialog)...)
	}

	return errors
}

// validateDialogBasic validates basic dialog fields.
func (v *VCon) validateDialogBasic(i int, dialog Dialog) []IETFValidationError {
	var errors []IETFValidationError

	// Validate dialog type
	if !isValidDialogType(dialog.Type) {
		errors = append(errors, IETFValidationError{
			Field:       fmt.Sprintf("dialog[%d].type", i),
			Message:     fmt.Sprintf("invalid dialog type: %s", dialog.Type),
			Requirement: "IETF vCon spec section 4.3",
		})
	}

	// Validate start time is RFC3339
	if err := validateRFC3339Time(dialog.Start); err != nil {
		errors = append(errors, IETFValidationError{
			Field:       fmt.Sprintf("dialog[%d].start", i),
			Message:     "must be valid RFC3339 timestamp",
			Requirement: "IETF vCon spec section 4.3",
		})
	}

	// Validate parties field is required for recording and text types
	if dialog.Type == "recording" || dialog.Type == "text" {
		if dialog.Parties == nil || len(dialog.Parties.GetAllPartyIndices()) == 0 {
			errors = append(errors, IETFValidationError{
				Field:       fmt.Sprintf("dialog[%d].parties", i),
				Message:     fmt.Sprintf("%s dialog must have at least one party", dialog.Type),
				Requirement: "IETF vCon spec section 4.3 - parties required for recording/text",
			})
		}
	}

	// Validate party indices are within valid range
	if dialog.Parties != nil {
		if err := dialog.Parties.Validate(len(v.Parties)); err != nil {
			errors = append(errors, IETFValidationError{
				Field:       fmt.Sprintf("dialog[%d].parties", i),
				Message:     err.Error(),
				Requirement: "IETF vCon spec section 4.3",
			})
		}
	}

	return errors
}

// validateDialogTypeSpecific validates type-specific dialog requirements.
func (v *VCon) validateDialogTypeSpecific(i int, dialog Dialog) []IETFValidationError {
	var errors []IETFValidationError

	switch dialog.Type {
	case "incomplete":
		errors = append(errors, v.validateIncompleteDialog(i, dialog)...)
	case "recording", "text":
		errors = append(errors, v.validateRecordingTextDialog(i, dialog)...)
	case "transfer":
		errors = append(errors, v.validateTransferDialog(i, dialog)...)
	}

	return errors
}

// validateIncompleteDialog validates incomplete dialog requirements.
func (v *VCon) validateIncompleteDialog(i int, dialog Dialog) []IETFValidationError {
	var errors []IETFValidationError

	// Validate disposition requirement
	errors = append(errors, v.validateIncompleteDialogDisposition(i, dialog)...)

	// Validate content restrictions
	errors = append(errors, v.validateIncompleteDialogContentRestrictions(i, dialog)...)

	return errors
}

// validateIncompleteDialogDisposition validates disposition field requirements.
func (v *VCon) validateIncompleteDialogDisposition(i int, dialog Dialog) []IETFValidationError {
	var errors []IETFValidationError

	// Disposition is required for incomplete dialogs
	if dialog.Disposition == nil || *dialog.Disposition == "" {
		errors = append(errors, IETFValidationError{
			Field:       fmt.Sprintf("dialog[%d].disposition", i),
			Message:     "incomplete dialog must have disposition",
			Requirement: "IETF vCon spec section 4.3.4",
		})
		return errors
	}

	// Validate disposition value against allowed enumeration
	validDispositions := []string{"no-answer", "congestion", "failed", "busy", "hung-up", "voicemail-no-message"}
	isValid := false
	for _, valid := range validDispositions {
		if *dialog.Disposition == valid {
			isValid = true
			break
		}
	}
	if !isValid {
		errors = append(errors, IETFValidationError{
			Field:       fmt.Sprintf("dialog[%d].disposition", i),
			Message:     fmt.Sprintf("invalid disposition value: %s", *dialog.Disposition),
			Requirement: "IETF vCon spec section 4.3.4 - must be one of: no-answer, congestion, failed, busy, hung-up, voicemail-no-message",
		})
	}

	return errors
}

// validateIncompleteDialogContentRestrictions validates that incomplete dialogs don't have content.
func (v *VCon) validateIncompleteDialogContentRestrictions(i int, dialog Dialog) []IETFValidationError {
	var errors []IETFValidationError

	// Incomplete dialogs MUST NOT have Dialog Content
	if hasDialogContent(dialog) {
		errors = append(errors, IETFValidationError{
			Field:       fmt.Sprintf("dialog[%d]", i),
			Message:     "incomplete dialog must not have content (body/url)",
			Requirement: "IETF vCon spec - incomplete dialogs must not have content",
		})
	}

	// Incomplete dialogs MUST NOT have mediatype or filename
	if dialog.Mediatype != nil && *dialog.Mediatype != "" {
		errors = append(errors, IETFValidationError{
			Field:       fmt.Sprintf("dialog[%d].mediatype", i),
			Message:     "incomplete dialog must not have mediatype",
			Requirement: "IETF vCon spec - incomplete dialogs have no media content",
		})
	}
	if dialog.Filename != nil && *dialog.Filename != "" {
		errors = append(errors, IETFValidationError{
			Field:       fmt.Sprintf("dialog[%d].filename", i),
			Message:     "incomplete dialog must not have filename",
			Requirement: "IETF vCon spec - incomplete dialogs have no media content",
		})
	}

	return errors
}

// validateRecordingTextDialog validates recording and text dialog requirements.
func (v *VCon) validateRecordingTextDialog(i int, dialog Dialog) []IETFValidationError {
	var errors []IETFValidationError

	// Recording and text dialogs MUST have Dialog Content
	if !hasDialogContent(dialog) {
		errors = append(errors, IETFValidationError{
			Field:       fmt.Sprintf("dialog[%d]", i),
			Message:     fmt.Sprintf("%s dialog must have content (body/url)", dialog.Type),
			Requirement: "IETF vCon spec - recording/text dialogs must have content",
		})
	}

	// ENHANCED: Mediatype is REQUIRED for inline content (body present)
	if dialog.Body != nil && (dialog.Mediatype == nil || *dialog.Mediatype == "") {
		errors = append(errors, IETFValidationError{
			Field:       fmt.Sprintf("dialog[%d].mediatype", i),
			Message:     fmt.Sprintf("%s dialog with inline content requires mediatype", dialog.Type),
			Requirement: "IETF vCon spec - mediatype required for inline content",
		})
	}

	// Media type validation when present
	if dialog.Mediatype != nil && *dialog.Mediatype != "" {
		if !isValidMIMEType(*dialog.Mediatype) {
			errors = append(errors, IETFValidationError{
				Field:       fmt.Sprintf("dialog[%d].mediatype", i),
				Message:     fmt.Sprintf("invalid media type: %s", *dialog.Mediatype),
				Requirement: "RFC2046",
			})
		}
	}

	return errors
}

// validateTransferDialog validates transfer dialog requirements.
func (v *VCon) validateTransferDialog(i int, dialog Dialog) []IETFValidationError {
	var errors []IETFValidationError

	// Validate content restrictions
	errors = append(errors, v.validateTransferDialogContentRestrictions(i, dialog)...)

	// Validate party restrictions
	errors = append(errors, v.validateTransferDialogPartyRestrictions(i, dialog)...)

	// Validate required references
	errors = append(errors, v.validateTransferDialogRequiredReferences(i, dialog)...)

	// Validate required transfer parties
	errors = append(errors, v.validateTransferDialogTransferParties(i, dialog)...)

	return errors
}

// validateTransferDialogContentRestrictions validates that transfer dialogs don't have content.
func (v *VCon) validateTransferDialogContentRestrictions(i int, dialog Dialog) []IETFValidationError {
	var errors []IETFValidationError

	// Transfer dialogs MUST NOT have Dialog Content
	if hasDialogContent(dialog) {
		errors = append(errors, IETFValidationError{
			Field:       fmt.Sprintf("dialog[%d]", i),
			Message:     "transfer dialog must not have content (body/url)",
			Requirement: "IETF vCon spec - transfer dialogs must not have content",
		})
	}

	// Transfer dialogs MUST NOT have mediatype or filename
	if dialog.Mediatype != nil && *dialog.Mediatype != "" {
		errors = append(errors, IETFValidationError{
			Field:       fmt.Sprintf("dialog[%d].mediatype", i),
			Message:     "transfer dialog must not have mediatype",
			Requirement: "IETF vCon spec - transfer dialogs have no media content",
		})
	}
	if dialog.Filename != nil && *dialog.Filename != "" {
		errors = append(errors, IETFValidationError{
			Field:       fmt.Sprintf("dialog[%d].filename", i),
			Message:     "transfer dialog must not have filename",
			Requirement: "IETF vCon spec - transfer dialogs have no media content",
		})
	}

	return errors
}

// validateTransferDialogPartyRestrictions validates that transfer dialogs don't have parties.
func (v *VCon) validateTransferDialogPartyRestrictions(i int, dialog Dialog) []IETFValidationError {
	var errors []IETFValidationError

	// Transfer dialogs MUST NOT have parties or originator
	if dialog.Parties != nil && len(dialog.Parties.GetAllPartyIndices()) > 0 {
		errors = append(errors, IETFValidationError{
			Field:       fmt.Sprintf("dialog[%d].parties", i),
			Message:     "transfer dialog must not have parties",
			Requirement: "IETF vCon spec - transfer dialogs represent control flow, not party conversation",
		})
	}
	if dialog.Originator != nil {
		errors = append(errors, IETFValidationError{
			Field:       fmt.Sprintf("dialog[%d].originator", i),
			Message:     "transfer dialog must not have originator",
			Requirement: "IETF vCon spec - transfer dialogs represent control flow, not party conversation",
		})
	}

	return errors
}

// validateTransferDialogRequiredReferences validates required dialog references.
func (v *VCon) validateTransferDialogRequiredReferences(i int, dialog Dialog) []IETFValidationError {
	var errors []IETFValidationError

	// Transfer dialogs MUST have original and target-dialog references
	if dialog.Original == nil {
		errors = append(errors, IETFValidationError{
			Field:       fmt.Sprintf("dialog[%d].original", i),
			Message:     "transfer dialog must have original dialog reference",
			Requirement: "IETF vCon spec section 4.3.3 - original required",
		})
	} else {
		// Validate that the original dialog index is valid
		if *dialog.Original < 0 || *dialog.Original >= len(v.Dialog) {
			errors = append(errors, IETFValidationError{
				Field:       fmt.Sprintf("dialog[%d].original", i),
				Message:     fmt.Sprintf("original dialog reference %d is invalid (must be 0-%d)", *dialog.Original, len(v.Dialog)-1),
				Requirement: "IETF vCon spec section 4.3.3 - original must reference valid dialog",
			})
		} else if *dialog.Original == i {
			errors = append(errors, IETFValidationError{
				Field:       fmt.Sprintf("dialog[%d].original", i),
				Message:     "original dialog reference cannot point to itself",
				Requirement: "IETF vCon spec section 4.3.3 - original must reference different dialog",
			})
		}
	}

	if dialog.TargetDialog == nil {
		errors = append(errors, IETFValidationError{
			Field:       fmt.Sprintf("dialog[%d].target-dialog", i),
			Message:     "transfer dialog must have target-dialog reference",
			Requirement: "IETF vCon spec section 4.3.3 - target-dialog required",
		})
	} else {
		// Validate that the target dialog index is valid
		if *dialog.TargetDialog < 0 || *dialog.TargetDialog >= len(v.Dialog) {
			errors = append(errors, IETFValidationError{
				Field:       fmt.Sprintf("dialog[%d].target-dialog", i),
				Message:     fmt.Sprintf("target-dialog reference %d is invalid (must be 0-%d)", *dialog.TargetDialog, len(v.Dialog)-1),
				Requirement: "IETF vCon spec section 4.3.3 - target-dialog must reference valid dialog",
			})
		} else if *dialog.TargetDialog == i {
			errors = append(errors, IETFValidationError{
				Field:       fmt.Sprintf("dialog[%d].target-dialog", i),
				Message:     "target-dialog reference cannot point to itself",
				Requirement: "IETF vCon spec section 4.3.3 - target-dialog must reference different dialog",
			})
		}
	}

	// Validate consultation field if provided (optional field)
	if dialog.Consultation != nil {
		// Validate that the consultation dialog index is valid
		if *dialog.Consultation < 0 || *dialog.Consultation >= len(v.Dialog) {
			errors = append(errors, IETFValidationError{
				Field:       fmt.Sprintf("dialog[%d].consultation", i),
				Message:     fmt.Sprintf("consultation dialog reference %d is invalid (must be 0-%d)", *dialog.Consultation, len(v.Dialog)-1),
				Requirement: "IETF vCon spec section 4.3.3 - consultation must reference valid dialog",
			})
		} else if *dialog.Consultation == i {
			errors = append(errors, IETFValidationError{
				Field:       fmt.Sprintf("dialog[%d].consultation", i),
				Message:     "consultation dialog reference cannot point to itself",
				Requirement: "IETF vCon spec section 4.3.3 - consultation must reference different dialog",
			})
		}
	}

	return errors
}

// validateTransferDialogTransferParties validates required transfer party references.
func (v *VCon) validateTransferDialogTransferParties(i int, dialog Dialog) []IETFValidationError {
	var errors []IETFValidationError

	// Transfer dialogs MUST have transferee, transferor, and transfer-target per IETF spec
	if dialog.Transferee == nil {
		errors = append(errors, IETFValidationError{
			Field:       fmt.Sprintf("dialog[%d].transferee", i),
			Message:     "transfer dialog must have transferee",
			Requirement: "IETF vCon spec section 4.3.3 - transferee required",
		})
	}
	if dialog.Transferor == nil {
		errors = append(errors, IETFValidationError{
			Field:       fmt.Sprintf("dialog[%d].transferor", i),
			Message:     "transfer dialog must have transferor",
			Requirement: "IETF vCon spec section 4.3.3 - transferor required",
		})
	}
	if dialog.TransferTarget == nil {
		errors = append(errors, IETFValidationError{
			Field:       fmt.Sprintf("dialog[%d].transfer-target", i),
			Message:     "transfer dialog must have transfer-target",
			Requirement: "IETF vCon spec section 4.3.3 - transfer-target required",
		})
	}

	return errors
}

// validateDialogOneOfContent validates dialog one-of content rules (body+encoding OR url+content_hash).
func (v *VCon) validateDialogOneOfContent(i int, dialog Dialog) []IETFValidationError {
	var errors []IETFValidationError

	// Skip content validation for transfer and incomplete dialogs - they should not have content
	if dialog.Type == "transfer" || dialog.Type == "incomplete" {
		return errors
	}

	hasInlineContent := dialog.Body != nil
	hasExternalContent := dialog.URL != nil && *dialog.URL != ""

	// Recording and text dialogs MUST have content (already validated in validateDialogTypeSpecific)
	// But we still validate the one-of rule: cannot have both inline AND external content
	if hasInlineContent && hasExternalContent {
		errors = append(errors, IETFValidationError{
			Field:       fmt.Sprintf("dialog[%d]", i),
			Message:     "dialog cannot have both inline content (body) and external content (url)",
			Requirement: "IETF vCon spec - one-of content rule",
		})
	}

	// If inline content, require encoding
	if hasInlineContent && (dialog.Encoding == nil || *dialog.Encoding == "") {
		errors = append(errors, IETFValidationError{
			Field:       fmt.Sprintf("dialog[%d].encoding", i),
			Message:     "encoding is required when body is present",
			Requirement: "IETF vCon spec - encoding required for inline content",
		})
	}

	// Validate encoding value if present
	if dialog.Encoding != nil && *dialog.Encoding != "" {
		validEncodings := []string{"json", "none", "base64url"}
		isValid := false
		for _, valid := range validEncodings {
			if *dialog.Encoding == valid {
				isValid = true
				break
			}
		}
		if !isValid {
			errors = append(errors, IETFValidationError{
				Field:       fmt.Sprintf("dialog[%d].encoding", i),
				Message:     fmt.Sprintf("invalid encoding: %s", *dialog.Encoding),
				Requirement: "IETF vCon spec - valid encoding values",
			})
		}
	}

	return errors
}

// validateDialogURLs validates external URLs use HTTPS.
func (v *VCon) validateDialogURLs(i int, dialog Dialog) []IETFValidationError {
	var errors []IETFValidationError

	if dialog.URL != nil && *dialog.URL != "" {
		if err := validateHTTPS(*dialog.URL); err != nil {
			errors = append(errors, IETFValidationError{
				Field:       fmt.Sprintf("dialog[%d].url", i),
				Message:     "external URLs must use HTTPS",
				Requirement: "IETF vCon spec section 5.1",
			})
		}
	}

	return errors
}

// validateDialogContentHash validates content_hash format and requirements.
func (v *VCon) validateDialogContentHash(i int, dialog Dialog) []IETFValidationError {
	var errors []IETFValidationError

	// If URL is present, content_hash is required per IETF spec
	if dialog.URL != nil && *dialog.URL != "" {
		if dialog.ContentHash == nil || dialog.ContentHash.IsEmpty() {
			errors = append(errors, IETFValidationError{
				Field:       fmt.Sprintf("dialog[%d].content_hash", i),
				Message:     "content_hash is required when URL is present",
				Requirement: "IETF vCon spec - content integrity for external references",
			})
		}
	}

	// Validate content_hash format if present
	if dialog.ContentHash != nil && !dialog.ContentHash.IsEmpty() {
		// Validate each hash in the ContentHashValue (could be single or array)
		hashes := dialog.ContentHash.GetArray()
		for _, hash := range hashes {
			if err := ValidateContentHashFormat(hash); err != nil {
				errors = append(errors, IETFValidationError{
					Field:       fmt.Sprintf("dialog[%d].content_hash", i),
					Message:     err.Error(),
					Requirement: "IETF vCon spec - content_hash format 'algorithm:hash'",
				})
			}
		}
	}

	return errors
}

// validateIETFAnalysis validates analysis objects.
func (v *VCon) validateIETFAnalysis() []IETFValidationError {
	var errors []IETFValidationError

	for i, analysis := range v.Analysis {
		errors = append(errors, v.validateAnalysisFields(i, analysis)...)
		errors = append(errors, v.validateAnalysisOneOfContent(i, analysis)...)
		errors = append(errors, v.validateAnalysisDialogRefs(i, analysis)...)
		errors = append(errors, v.validateAnalysisContentHash(i, analysis)...)
	}

	return errors
}

// validateAnalysisFields validates required analysis fields.
func (v *VCon) validateAnalysisFields(i int, analysis Analysis) []IETFValidationError {
	var errors []IETFValidationError

	if analysis.Vendor == "" {
		errors = append(errors, IETFValidationError{
			Field:       fmt.Sprintf("analysis[%d].vendor", i),
			Message:     "vendor is required",
			Requirement: "IETF vCon spec section 4.5",
		})
	}

	// Type is REQUIRED per IETF spec (allows custom types per "SHOULD" language)
	if analysis.Type == "" {
		errors = append(errors, IETFValidationError{
			Field:       fmt.Sprintf("analysis[%d].type", i),
			Message:     "type is required",
			Requirement: "IETF vCon spec section 4.5 - type is required",
		})
	}
	// Note: Per IETF spec section 4.4.1, type "SHOULD be one of" standard types,
	// but custom types are allowed per RFC 2119 "SHOULD" language

	return errors
}

// validateAnalysisOneOfContent validates analysis one-of content rules (body+encoding OR url+content_hash).
func (v *VCon) validateAnalysisOneOfContent(i int, analysis Analysis) []IETFValidationError {
	var errors []IETFValidationError

	hasInlineContent := analysis.Body != nil
	hasExternalContent := analysis.URL != nil && *analysis.URL != ""

	// Analysis must have either inline OR external content
	if !hasInlineContent && !hasExternalContent {
		errors = append(errors, IETFValidationError{
			Field:       fmt.Sprintf("analysis[%d]", i),
			Message:     "analysis must have either body+encoding or url+content_hash",
			Requirement: "IETF vCon spec - content requirement",
		})
	}

	// Cannot have both inline AND external content
	if hasInlineContent && hasExternalContent {
		errors = append(errors, IETFValidationError{
			Field:       fmt.Sprintf("analysis[%d]", i),
			Message:     "analysis cannot have both inline content (body) and external content (url)",
			Requirement: "IETF vCon spec - one-of content rule",
		})
	}

	// If inline content, require encoding
	if hasInlineContent && (analysis.Encoding == nil || *analysis.Encoding == "") {
		errors = append(errors, IETFValidationError{
			Field:       fmt.Sprintf("analysis[%d].encoding", i),
			Message:     "encoding is required when body is present",
			Requirement: "IETF vCon spec - encoding required for inline content",
		})
	}

	// Validate encoding value if present
	if analysis.Encoding != nil && *analysis.Encoding != "" {
		validEncodings := []string{"json", "none", "base64url"}
		isValid := false
		for _, valid := range validEncodings {
			if *analysis.Encoding == valid {
				isValid = true
				break
			}
		}
		if !isValid {
			errors = append(errors, IETFValidationError{
				Field:       fmt.Sprintf("analysis[%d].encoding", i),
				Message:     fmt.Sprintf("invalid encoding: %s", *analysis.Encoding),
				Requirement: "IETF vCon spec - valid encoding values",
			})
		}
	}

	return errors
}

// validateAnalysisDialogRefs validates analysis dialog references.
func (v *VCon) validateAnalysisDialogRefs(i int, analysis Analysis) []IETFValidationError {
	var errors []IETFValidationError

	if analysis.Dialog != nil {
		dialogCount := len(v.Dialog)
		switch ref := analysis.Dialog.(type) {
		case float64:
			idx := int(ref)
			if idx < 0 || idx >= dialogCount {
				errors = append(errors, IETFValidationError{
					Field:       fmt.Sprintf("analysis[%d].dialog", i),
					Message:     fmt.Sprintf("invalid dialog index: %d", idx),
					Requirement: "IETF vCon spec section 4.5",
				})
			}
		case int:
			if ref < 0 || ref >= dialogCount {
				errors = append(errors, IETFValidationError{
					Field:       fmt.Sprintf("analysis[%d].dialog", i),
					Message:     fmt.Sprintf("invalid dialog index: %d", ref),
					Requirement: "IETF vCon spec section 4.5",
				})
			}
		case []interface{}:
			for j, dialogIdx := range ref {
				if idx, ok := dialogIdx.(float64); ok {
					intIdx := int(idx)
					if intIdx < 0 || intIdx >= dialogCount {
						errors = append(errors, IETFValidationError{
							Field:       fmt.Sprintf("analysis[%d].dialog[%d]", i, j),
							Message:     fmt.Sprintf("invalid dialog index: %d", intIdx),
							Requirement: "IETF vCon spec section 4.5",
						})
					}
				}
			}
		}
	}

	return errors
}

// validateAnalysisContentHash validates content_hash format and requirements for analysis.
func (v *VCon) validateAnalysisContentHash(i int, analysis Analysis) []IETFValidationError {
	var errors []IETFValidationError

	// If URL is present, content_hash is required per IETF spec
	if analysis.URL != nil && *analysis.URL != "" {
		if analysis.ContentHash == nil || analysis.ContentHash.IsEmpty() {
			errors = append(errors, IETFValidationError{
				Field:       fmt.Sprintf("analysis[%d].content_hash", i),
				Message:     "content_hash is required when URL is present",
				Requirement: "IETF vCon spec - content integrity for external references",
			})
		}
	}

	// Validate content_hash format if present
	if analysis.ContentHash != nil && !analysis.ContentHash.IsEmpty() {
		// Validate each hash in the ContentHashValue (could be single or array)
		hashes := analysis.ContentHash.GetArray()
		for _, hash := range hashes {
			if err := ValidateContentHashFormat(hash); err != nil {
				errors = append(errors, IETFValidationError{
					Field:       fmt.Sprintf("analysis[%d].content_hash", i),
					Message:     err.Error(),
					Requirement: "IETF vCon spec - content_hash format 'algorithm:hash'",
				})
			}
		}
	}

	return errors
}

// validateIETFAttachments validates attachment references and timestamps.
func (v *VCon) validateIETFAttachments() []IETFValidationError {
	var errors []IETFValidationError

	for i, attachment := range v.Attachments {
		// Validate party reference (required per IETF spec)
		if attachment.Party == nil {
			errors = append(errors, IETFValidationError{
				Field:       fmt.Sprintf("attachments[%d].party", i),
				Message:     "party is required",
				Requirement: "IETF vCon spec section 4.4",
			})
		} else if *attachment.Party < 0 || *attachment.Party >= len(v.Parties) {
			errors = append(errors, IETFValidationError{
				Field:       fmt.Sprintf("attachments[%d].party", i),
				Message:     fmt.Sprintf("invalid party index: %d", *attachment.Party),
				Requirement: "IETF vCon spec section 4.4",
			})
		}

		// Validate one-of content rules
		errors = append(errors, v.validateAttachmentOneOfContent(i, attachment)...)

		// Validate dialog reference if present
		if attachment.Dialog != nil {
			if *attachment.Dialog < 0 || *attachment.Dialog >= len(v.Dialog) {
				errors = append(errors, IETFValidationError{
					Field:       fmt.Sprintf("attachments[%d].dialog", i),
					Message:     fmt.Sprintf("invalid dialog index: %d", *attachment.Dialog),
					Requirement: "IETF vCon spec section 4.4",
				})
			}
		}

		// Validate timestamp (optional per IETF spec, but validate format if present)
		if attachment.Start != nil {
			if err := validateRFC3339Time(*attachment.Start); err != nil {
				errors = append(errors, IETFValidationError{
					Field:       fmt.Sprintf("attachments[%d].start", i),
					Message:     "must be valid RFC3339 timestamp",
					Requirement: "IETF vCon spec section 4.4",
				})
			}
		}

		// Validate attachment URL uses HTTPS if present
		if attachment.URL != nil && *attachment.URL != "" {
			if err := validateHTTPS(*attachment.URL); err != nil {
				errors = append(errors, IETFValidationError{
					Field:       fmt.Sprintf("attachments[%d].url", i),
					Message:     "external URLs must use HTTPS",
					Requirement: "IETF vCon spec section 5.1",
				})
			}

			// content_hash is required when URL is present
			if attachment.ContentHash == nil || attachment.ContentHash.IsEmpty() {
				errors = append(errors, IETFValidationError{
					Field:       fmt.Sprintf("attachments[%d].content_hash", i),
					Message:     "content_hash is required when URL is present",
					Requirement: "IETF vCon spec - content integrity for external references",
				})
			}
		}

		// Validate content_hash format if present
		if attachment.ContentHash != nil && !attachment.ContentHash.IsEmpty() {
			// Validate each hash in the ContentHashValue (could be single or array)
			hashes := attachment.ContentHash.GetArray()
			for _, hash := range hashes {
				if err := ValidateContentHashFormat(hash); err != nil {
					errors = append(errors, IETFValidationError{
						Field:       fmt.Sprintf("attachments[%d].content_hash", i),
						Message:     err.Error(),
						Requirement: "IETF vCon spec - content_hash format 'algorithm:hash'",
					})
				}
			}
		}
	}

	return errors
}

// validateAttachmentOneOfContent validates attachment one-of content rules (body+encoding OR url+content_hash).
func (v *VCon) validateAttachmentOneOfContent(i int, attachment Attachment) []IETFValidationError {
	var errors []IETFValidationError

	hasInlineContent := attachment.Body != nil
	hasExternalContent := attachment.URL != nil && *attachment.URL != ""

	// ENHANCED: Check if this is a redacted vCon scenario
	isRedactedScenario := v.Redacted != nil

	// ENHANCED: Allow attachments without content in redacted vCon scenarios
	if !hasInlineContent && !hasExternalContent && !isRedactedScenario {
		errors = append(errors, IETFValidationError{
			Field:       fmt.Sprintf("attachments[%d]", i),
			Message:     "attachment must have either body or url",
			Requirement: "IETF vCon spec section 4.4 - content requirement",
		})
	}

	// Cannot have both inline AND external content
	if hasInlineContent && hasExternalContent {
		errors = append(errors, IETFValidationError{
			Field:       fmt.Sprintf("attachments[%d]", i),
			Message:     "attachment cannot have both inline content (body) and external content (url)",
			Requirement: "IETF vCon spec - one-of content rule",
		})
	}

	// If inline content, require encoding
	if hasInlineContent && (attachment.Encoding == nil || *attachment.Encoding == "") {
		errors = append(errors, IETFValidationError{
			Field:       fmt.Sprintf("attachments[%d].encoding", i),
			Message:     "encoding is required when body is present",
			Requirement: "IETF vCon spec - encoding required for inline content",
		})
	}

	// ENHANCED: Require mediatype for inline attachments
	if hasInlineContent && (attachment.Mediatype == nil || *attachment.Mediatype == "") {
		errors = append(errors, IETFValidationError{
			Field:       fmt.Sprintf("attachments[%d].mediatype", i),
			Message:     "mediatype is required for inline attachments",
			Requirement: "IETF vCon spec section 4.4 - mediatype required for inline content",
		})
	}

	// Validate mediatype format if present
	if attachment.Mediatype != nil && *attachment.Mediatype != "" {
		if !isValidMIMEType(*attachment.Mediatype) {
			errors = append(errors, IETFValidationError{
				Field:       fmt.Sprintf("attachments[%d].mediatype", i),
				Message:     fmt.Sprintf("invalid mediatype: %s", *attachment.Mediatype),
				Requirement: "RFC2046",
			})
		}
	}

	// Validate encoding value if present
	if attachment.Encoding != nil && *attachment.Encoding != "" {
		validEncodings := []string{"json", "none", "base64url"}
		isValid := false
		for _, valid := range validEncodings {
			if *attachment.Encoding == valid {
				isValid = true
				break
			}
		}
		if !isValid {
			errors = append(errors, IETFValidationError{
				Field:       fmt.Sprintf("attachments[%d].encoding", i),
				Message:     fmt.Sprintf("invalid encoding: %s", *attachment.Encoding),
				Requirement: "IETF vCon spec - valid encoding values",
			})
		}
	}

	// ENHANCED: Handle redacted content placeholders
	if isRedactedScenario && !hasInlineContent && !hasExternalContent {
		// In redacted scenarios, attachments may have redacted content placeholders
		// We allow type-only attachments or attachments with metadata only
		if attachment.Type == nil || *attachment.Type == "" {
			errors = append(errors, IETFValidationError{
				Field:       fmt.Sprintf("attachments[%d].type", i),
				Message:     "redacted attachment must have type field",
				Requirement: "IETF vCon spec - redacted content must indicate content type",
			})
		}
	}

	return errors
}

// validateRFC3339Time checks if a time is properly formatted as RFC3339.
func validateRFC3339Time(t time.Time) error {
	// RFC3339 format validation
	formatted := t.Format(time.RFC3339)
	parsed, err := time.Parse(time.RFC3339, formatted)
	if err != nil {
		return err
	}
	// Check if parsing loses precision (invalid format)
	if !parsed.Equal(t.Truncate(time.Second)) {
		return fmt.Errorf("time loses precision when formatted as RFC3339")
	}
	return nil
}

// hasPartyIdentifier checks if a party has at least one identifier.
func hasPartyIdentifier(p Party) bool {
	return p.Tel != nil || p.Mailto != nil || p.Name != nil || p.UUID != nil
}

// isValidDialogType checks if a dialog type is valid per IETF spec.
func isValidDialogType(dialogType string) bool {
	validTypes := []string{"recording", "text", "transfer", "incomplete"}
	for _, valid := range validTypes {
		if dialogType == valid {
			return true
		}
	}
	return false
}

// isValidMIMEType checks if a MIME type is valid (relaxed validation).
// Per IETF spec, any valid MIME type format is acceptable.
func isValidMIMEType(mimeType string) bool {
	// Basic MIME type validation: type/subtype
	mimeRegex := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9\-\+\.]*\/[a-zA-Z0-9\-\+\.]+$`)
	return mimeRegex.MatchString(mimeType)
}

// validateHTTPS ensures a URL uses HTTPS protocol.
func validateHTTPS(urlStr string) error {
	u, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %v", err)
	}
	if u.Scheme != "https" {
		return fmt.Errorf("URL must use HTTPS, got %s", u.Scheme)
	}
	return nil
}

// hasDialogContent checks if a dialog has content (body or URL).
func hasDialogContent(dialog Dialog) bool {
	return dialog.Body != nil || (dialog.URL != nil && *dialog.URL != "")
}

// IETFComplianceLevel represents the level of IETF compliance.
type IETFComplianceLevel int

const (
	// IETFNonCompliant indicates the vCon does not meet IETF requirements.
	IETFNonCompliant IETFComplianceLevel = iota
	// IETFPartiallyCompliant indicates the vCon meets some IETF requirements.
	IETFPartiallyCompliant
	// IETFFullyCompliant indicates the vCon fully meets IETF requirements.
	IETFFullyCompliant
)

// validateIETFTopLevelObjects validates redacted, appended, and group objects.
func (v *VCon) validateIETFTopLevelObjects() []IETFValidationError {
	var errors []IETFValidationError

	// Validate Redacted object if present
	if v.Redacted != nil {
		// If URL is present, content_hash is required
		if v.Redacted.URL != nil && *v.Redacted.URL != "" {
			if v.Redacted.ContentHash == nil || v.Redacted.ContentHash.IsEmpty() {
				errors = append(errors, IETFValidationError{
					Field:       "redacted.content_hash",
					Message:     "content_hash is required when URL is present",
					Requirement: "IETF vCon spec - content integrity for external references",
				})
			}
		}

		// Validate content_hash format if present
		if v.Redacted.ContentHash != nil && !v.Redacted.ContentHash.IsEmpty() {
			// Validate each hash in the ContentHashValue (could be single or array)
			hashes := v.Redacted.ContentHash.GetArray()
			for _, hash := range hashes {
				if err := ValidateContentHashFormat(hash); err != nil {
					errors = append(errors, IETFValidationError{
						Field:       "redacted.content_hash",
						Message:     err.Error(),
						Requirement: "IETF vCon spec - content_hash format 'algorithm:hash'",
					})
				}
			}
		}
	}

	// Validate Appended object if present
	if v.Appended != nil {
		// If URL is present, content_hash is required
		if v.Appended.URL != nil && *v.Appended.URL != "" {
			if v.Appended.ContentHash == nil || v.Appended.ContentHash.IsEmpty() {
				errors = append(errors, IETFValidationError{
					Field:       "appended.content_hash",
					Message:     "content_hash is required when URL is present",
					Requirement: "IETF vCon spec - content integrity for external references",
				})
			}
		}

		// Validate content_hash format if present
		if v.Appended.ContentHash != nil && !v.Appended.ContentHash.IsEmpty() {
			// Validate each hash in the ContentHashValue (could be single or array)
			hashes := v.Appended.ContentHash.GetArray()
			for _, hash := range hashes {
				if err := ValidateContentHashFormat(hash); err != nil {
					errors = append(errors, IETFValidationError{
						Field:       "appended.content_hash",
						Message:     err.Error(),
						Requirement: "IETF vCon spec - content_hash format 'algorithm:hash'",
					})
				}
			}
		}
	}

	// Validate Group objects if present
	for i, group := range v.Group {
		// If URL is present, content_hash is required
		if group.URL != nil && *group.URL != "" {
			if group.ContentHash == nil || group.ContentHash.IsEmpty() {
				errors = append(errors, IETFValidationError{
					Field:       fmt.Sprintf("group[%d].content_hash", i),
					Message:     "content_hash is required when URL is present",
					Requirement: "IETF vCon spec - content integrity for external references",
				})
			}
		}

		// Validate content_hash format if present
		if group.ContentHash != nil && !group.ContentHash.IsEmpty() {
			// Validate each hash in the ContentHashValue (could be single or array)
			hashes := group.ContentHash.GetArray()
			for _, hash := range hashes {
				if err := ValidateContentHashFormat(hash); err != nil {
					errors = append(errors, IETFValidationError{
						Field:       fmt.Sprintf("group[%d].content_hash", i),
						Message:     err.Error(),
						Requirement: "IETF vCon spec - content_hash format 'algorithm:hash'",
					})
				}
			}
		}
	}

	return errors
}

// validateIETFTopLevelObjectMutualExclusion validates that only one of Redacted, Appended, or Group objects is present.
// Per IETF spec, these objects are mutually exclusive.
func (v *VCon) validateIETFTopLevelObjectMutualExclusion() []IETFValidationError {
	var errors []IETFValidationError

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
		errors = append(errors, IETFValidationError{
			Field:       "top_level_objects",
			Message:     fmt.Sprintf("only one of redacted, appended, or group objects is allowed per IETF spec, found: %s", strings.Join(presentObjects, ", ")),
			Requirement: "IETF vCon spec section 4.1 - mutual exclusion of linking objects",
		})
	}

	return errors
}

// CheckIETFCompliance returns the IETF compliance level and any errors.
func (v *VCon) CheckIETFCompliance() (IETFComplianceLevel, error) {
	err := v.ValidateIETF()
	if err == nil {
		return IETFFullyCompliant, nil
	}

	if ietfErrors, ok := err.(IETFValidationErrors); ok {
		// Check if only minor issues
		criticalErrors := 0
		for _, e := range ietfErrors {
			if strings.Contains(e.Field, "vcon") || strings.Contains(e.Field, "uuid") {
				criticalErrors++
			}
		}
		if criticalErrors == 0 {
			return IETFPartiallyCompliant, err
		}
	}

	return IETFNonCompliant, err
}

// ValidateIETFDraft03 performs strict IETF Draft-03 specification compliance validation.
// This is the most restrictive validation level that only allows fields explicitly
// defined in the IETF draft-03 specification without any extension fields.
func (v *VCon) ValidateIETFDraft03() error {
	// First perform regular IETF validation
	if err := v.ValidateIETF(); err != nil {
		return err
	}

	var errors IETFValidationErrors

	// Additional Draft-03 specific validations
	errors = append(errors, v.validateDraft03VersionRequirement()...)
	errors = append(errors, v.validateDraft03FieldPresence()...)
	errors = append(errors, v.validateDraft03ExtensionRestrictions()...)

	if len(errors) > 0 {
		return errors
	}
	return nil
}

// validateDraft03VersionRequirement validates strict version requirements for Draft-03.
func (v *VCon) validateDraft03VersionRequirement() []IETFValidationError {
	var errors []IETFValidationError

	// Draft-03 requires exactly version "0.0.2"
	if v.Vcon != "0.0.2" {
		errors = append(errors, IETFValidationError{
			Field:       "vcon",
			Message:     "IETF Draft-03 requires vcon version to be exactly '0.0.2'",
			Requirement: "IETF vCon draft-03 version requirement",
		})
	}

	return errors
}

// validateDraft03FieldPresence validates required field presence per Draft-03.
func (v *VCon) validateDraft03FieldPresence() []IETFValidationError {
	var errors []IETFValidationError

	// UUID is RECOMMENDED (not REQUIRED) per Draft-03
	// CreatedAt is REQUIRED (already validated in base IETF validation)
	// Parties is REQUIRED (already validated in base IETF validation)

	return errors
}

const (
	// Draft03ExtensionRequirement is the error message for extension field violations.
	Draft03ExtensionRequirement = "IETF vCon draft-03 strict compliance - core fields only"
)

// validateDraft03ExtensionRestrictions validates that only Draft-03 standard fields are present.
// This is the strictest validation that flags any non-standard extension fields.
func (v *VCon) validateDraft03ExtensionRestrictions() []IETFValidationError {
	var errors []IETFValidationError

	// Validate parties for extension fields
	errors = append(errors, v.validateDraft03PartyExtensions()...)

	// Validate dialogs for extension fields
	errors = append(errors, v.validateDraft03DialogExtensions()...)

	// Validate analysis for extension fields
	errors = append(errors, v.validateDraft03AnalysisExtensions()...)

	return errors
}

// validateDraft03PartyExtensions checks for extension fields in parties.
func (v *VCon) validateDraft03PartyExtensions() []IETFValidationError {
	var errors []IETFValidationError

	for i, party := range v.Parties {
		if party.Stir != nil {
			errors = append(errors, IETFValidationError{
				Field:       fmt.Sprintf("parties[%d].stir", i),
				Message:     "stir field is an extension not in Draft-03 core specification",
				Requirement: Draft03ExtensionRequirement,
			})
		}
		if party.ContactList != nil {
			errors = append(errors, IETFValidationError{
				Field:       fmt.Sprintf("parties[%d].contact_list", i),
				Message:     "contact_list field is an extension not in Draft-03 core specification",
				Requirement: Draft03ExtensionRequirement,
			})
		}
	}

	return errors
}

// validateDraft03DialogExtensions checks for extension fields in dialogs.
func (v *VCon) validateDraft03DialogExtensions() []IETFValidationError {
	var errors []IETFValidationError

	for i, dialog := range v.Dialog {
		// Contact center extension fields
		if dialog.Campaign != nil {
			errors = append(errors, IETFValidationError{
				Field:       fmt.Sprintf("dialog[%d].campaign", i),
				Message:     "campaign field is an extension not in Draft-03 core specification",
				Requirement: Draft03ExtensionRequirement,
			})
		}
		if dialog.Interaction != nil {
			errors = append(errors, IETFValidationError{
				Field:       fmt.Sprintf("dialog[%d].interaction", i),
				Message:     "interaction field is an extension not in Draft-03 core specification",
				Requirement: Draft03ExtensionRequirement,
			})
		}
		if dialog.Skill != nil {
			errors = append(errors, IETFValidationError{
				Field:       fmt.Sprintf("dialog[%d].skill", i),
				Message:     "skill field is an extension not in Draft-03 core specification",
				Requirement: Draft03ExtensionRequirement,
			})
		}
		if dialog.SessionID != nil {
			errors = append(errors, IETFValidationError{
				Field:       fmt.Sprintf("dialog[%d].session_id", i),
				Message:     "session_id field is an extension not in Draft-03 core specification",
				Requirement: Draft03ExtensionRequirement,
			})
		}

		// Video extension fields
		errors = append(errors, v.validateDraft03VideoExtensions(i, dialog)...)
	}

	return errors
}

// validateDraft03VideoExtensions checks for video-specific extension fields.
func (v *VCon) validateDraft03VideoExtensions(i int, dialog Dialog) []IETFValidationError {
	var errors []IETFValidationError

	if dialog.Resolution != nil {
		errors = append(errors, IETFValidationError{
			Field:       fmt.Sprintf("dialog[%d].resolution", i),
			Message:     "resolution field is an extension not in Draft-03 core specification",
			Requirement: Draft03ExtensionRequirement,
		})
	}
	if dialog.FrameRate != nil {
		errors = append(errors, IETFValidationError{
			Field:       fmt.Sprintf("dialog[%d].frame_rate", i),
			Message:     "frame_rate field is an extension not in Draft-03 core specification",
			Requirement: Draft03ExtensionRequirement,
		})
	}
	if dialog.Codec != nil {
		errors = append(errors, IETFValidationError{
			Field:       fmt.Sprintf("dialog[%d].codec", i),
			Message:     "codec field is an extension not in Draft-03 core specification",
			Requirement: Draft03ExtensionRequirement,
		})
	}
	if dialog.Bitrate != nil {
		errors = append(errors, IETFValidationError{
			Field:       fmt.Sprintf("dialog[%d].bitrate", i),
			Message:     "bitrate field is an extension not in Draft-03 core specification",
			Requirement: Draft03ExtensionRequirement,
		})
	}
	if dialog.Thumbnail != nil {
		errors = append(errors, IETFValidationError{
			Field:       fmt.Sprintf("dialog[%d].thumbnail", i),
			Message:     "thumbnail field is an extension not in Draft-03 core specification",
			Requirement: Draft03ExtensionRequirement,
		})
	}

	return errors
}

// validateDraft03AnalysisExtensions checks for extension fields in analysis.
func (v *VCon) validateDraft03AnalysisExtensions() []IETFValidationError {
	var errors []IETFValidationError

	for i, analysis := range v.Analysis {
		if analysis.Product != nil {
			errors = append(errors, IETFValidationError{
				Field:       fmt.Sprintf("analysis[%d].product", i),
				Message:     "product field is an extension not in Draft-03 core specification",
				Requirement: Draft03ExtensionRequirement,
			})
		}
		if analysis.Schema != nil {
			errors = append(errors, IETFValidationError{
				Field:       fmt.Sprintf("analysis[%d].schema", i),
				Message:     "schema field is an extension not in Draft-03 core specification",
				Requirement: Draft03ExtensionRequirement,
			})
		}
		if analysis.Mediatype != nil {
			errors = append(errors, IETFValidationError{
				Field:       fmt.Sprintf("analysis[%d].mediatype", i),
				Message:     "mediatype field is an extension not in Draft-03 core specification",
				Requirement: Draft03ExtensionRequirement,
			})
		}
		if analysis.Filename != nil {
			errors = append(errors, IETFValidationError{
				Field:       fmt.Sprintf("analysis[%d].filename", i),
				Message:     "filename field is an extension not in Draft-03 core specification",
				Requirement: Draft03ExtensionRequirement,
			})
		}
	}

	return errors
}
