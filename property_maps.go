package vcon

// Property handling modes for non-standard properties.
const (
	PropertyHandlingDefault = "default" // Keep non-standard properties
	PropertyHandlingStrict  = "strict"  // Remove non-standard properties
	PropertyHandlingMeta    = "meta"    // Move non-standard properties to meta
)

// Allowed properties for each object type (for property handling).
var (
	// AllowedVConProperties defines valid properties for the top-level vCon object.
	// This includes standard IETF fields and common extension fields.
	AllowedVConProperties = map[string]bool{
		"uuid": true, "vcon": true, "created_at": true, "updated_at": true, "redacted": true,
		"group": true, "parties": true, "dialog": true, "attachments": true, "analysis": true,
		"signatures": true, "payload": true, "meta": true, "subject": true, "appended": true,
		"extensions": true, "must_support": true,
	}

	// AllowedPartyProperties defines valid properties for Party objects.
	// Includes standard IETF identifiers and extension fields for telephony and identity.
	AllowedPartyProperties = map[string]bool{
		"type": true, "name": true, "contact": true, "meta": true, "external_id": true, "party_id": true,
		"tel": true, "stir": true, "mailto": true, "validation": true, "gmlpos": true, "civicaddress": true,
		"uuid": true, "role": true, "contact_list": true, "sip": true, "did": true, "jCard": true, "timezone": true,
	}

	// AllowedDialogProperties defines valid properties for Dialog objects.
	// Includes standard IETF fields and extension fields for video, contact center, and technical metadata.
	AllowedDialogProperties = map[string]bool{
		"type": true, "start": true, "parties": true, "duration": true, "mediatype": true, "filename": true,
		"body": true, "encoding": true, "url": true, "content_hash": true, "disposition": true,
		"party_history": true, "transferee": true, "transferor": true, "transfer-target": true,
		"original": true, "consultation": true, "target-dialog": true, "campaign": true,
		"interaction": true, "skill": true, "meta": true, "metadata": true, "transfer": true,
		"signaling": true, "originator": true, "resolution": true, "frame_rate": true,
		"codec": true, "bitrate": true, "thumbnail": true, "streaming": true, "video": true,
		"session_id": true,
		// Note: "application" and "message_id" removed for IETF draft-03 compliance
	}

	// AllowedAttachmentProperties defines valid properties for Attachment objects.
	// Includes standard IETF fields and extension fields for enhanced metadata.
	AllowedAttachmentProperties = map[string]bool{
		"type": true, "start": true, "party": true, "dialog": true, "mediatype": true, "filename": true,
		"body": true, "encoding": true, "url": true, "content_hash": true, "meta": true,
	}

	// AllowedAnalysisProperties defines valid properties for Analysis objects.
	// Includes standard IETF fields and extension fields for vendor-specific analysis.
	AllowedAnalysisProperties = map[string]bool{
		"type": true, "dialog": true, "vendor": true, "product": true, "schema": true,
		"mediatype": true, "filename": true, "body": true, "encoding": true, "url": true, "content_hash": true, "meta": true,
	}
)

// IsPropertyAllowed checks if a property is allowed for the given object type.
func IsPropertyAllowed(objectType, property string) bool {
	switch objectType {
	case "vcon":
		return AllowedVConProperties[property]
	case "party":
		return AllowedPartyProperties[property]
	case "dialog":
		return AllowedDialogProperties[property]
	case "attachment":
		return AllowedAttachmentProperties[property]
	case "analysis":
		return AllowedAnalysisProperties[property]
	default:
		return false
	}
}

// GetAllowedProperties returns the property map for the given object type.
func GetAllowedProperties(objectType string) map[string]bool {
	switch objectType {
	case "vcon":
		return AllowedVConProperties
	case "party":
		return AllowedPartyProperties
	case "dialog":
		return AllowedDialogProperties
	case "attachment":
		return AllowedAttachmentProperties
	case "analysis":
		return AllowedAnalysisProperties
	default:
		return make(map[string]bool)
	}
}
