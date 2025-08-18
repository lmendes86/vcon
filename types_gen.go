package vcon

import (
	"time"

	"github.com/google/uuid"
)

// NOTE: Property handling constants and allowed properties have been moved to property_maps.go

// CivicAddress represents a structured postal address following the PIDF-LO civic address format.
// It contains hierarchical location information from country level down to specific building details.
type CivicAddress struct {
	Country *string `json:"country,omitempty" validate:"omitempty,len=2"` // ISO 3166-1 alpha-2 country code
	A1      *string `json:"a1,omitempty"`                                 // Administrative area 1 (state/province)
	A2      *string `json:"a2,omitempty"`                                 // Administrative area 2 (county/municipality)
	A3      *string `json:"a3,omitempty"`                                 // Administrative area 3 (city/town)
	A4      *string `json:"a4,omitempty"`                                 // Administrative area 4 (district)
	A5      *string `json:"a5,omitempty"`                                 // Administrative area 5 (postal code)
	A6      *string `json:"a6,omitempty"`                                 // Administrative area 6 (building/floor)
	Prd     *string `json:"prd,omitempty"`                                // Department or suite number
	Pod     *string `json:"pod,omitempty"`                                // PO Box identifier
	Sts     *string `json:"sts,omitempty"`                                // Street name
	Hno     *string `json:"hno,omitempty"`                                // House number
	Hns     *string `json:"hns,omitempty"`                                // House name
	Lmk     *string `json:"lmk,omitempty"`                                // Landmark name
	Loc     *string `json:"loc,omitempty"`                                // Location name
	Flr     *string `json:"flr,omitempty"`                                // Floor
	Nam     *string `json:"nam,omitempty"`                                // Name of the location
	Pc      *string `json:"pc,omitempty"`                                 // Postal code
}

// Party represents a participant in the conversation.
// Per IETF spec, all party fields are optional, including identifiers.
// Parties can include contact information, location data, and arbitrary metadata.
type Party struct {
	Tel          *string       `json:"tel,omitempty" validate:"omitempty,tel_uri"`        // Telephone number (supports both tel: URI and plain E.164 format)
	Stir         *string       `json:"stir,omitempty"`                                    // STIR identifier
	Mailto       *string       `json:"mailto,omitempty" validate:"omitempty,mailto_uri"`  // Email address (supports both mailto: URI and plain email format)
	Name         *string       `json:"name,omitempty" validate:"omitempty,min=1,max=255"` // Display name
	Validation   *string       `json:"validation,omitempty"`                              // Validation info
	Gmlpos       *string       `json:"gmlpos,omitempty"`                                  // GML position
	CivicAddress *CivicAddress `json:"civicaddress,omitempty"`                            // Structured postal address
	UUID         *uuid.UUID    `json:"uuid,omitempty"`                                    // Participant UUID
	Role         *string       `json:"role,omitempty" validate:"omitempty,min=1,max=100"` // Participant role
	ContactList  *string       `json:"contact_list,omitempty" validate:"omitempty,min=1"` // Contact list identifier
	// Note: SIP, DID, JCard, and Timezone fields removed for IETF draft-03 compliance
	// SIP and DID are not in the spec; JCard and Timezone are marked as TODO
	Meta                 map[string]any `json:"meta,omitempty"` // Arbitrary metadata
	AdditionalProperties map[string]any `json:"-"`              // Extra fields
}

// PartyHistory records an event for a party in the dialog.
// It tracks when parties join, drop, or have other state changes during the conversation.
type PartyHistory struct {
	Party int       `json:"party" validate:"min=0"`                                            // Party index
	Event string    `json:"event" validate:"required,oneof=join drop hold unhold mute unmute"` // e.g. "join", "drop", "hold", "unhold", "mute", "unmute" per IETF spec
	Time  time.Time `json:"time" validate:"required"`                                          // Timestamp of event
}

// Dialog represents a segment in the conversation (recording, text, transfer, etc.).
// Each dialog has a type that determines which fields are relevant and required.
// Supported types include "recording", "text", "transfer", and "incomplete" per IETF spec.
type Dialog struct {
	Type           string            `json:"type" validate:"required,oneof=recording text transfer incomplete"` // Dialog type per IETF spec
	Start          time.Time         `json:"start" validate:"required"`                                         // ISO 8601 timestamp
	Parties        *DialogParties    `json:"parties,omitempty"`                                                 // Flexible party representation (int, []int, or [][]int per IETF spec)
	Originator     *int              `json:"originator,omitempty" validate:"omitempty,min=0"`                   // Who initiated
	Mediatype      *string           `json:"mediatype,omitempty" validate:"omitempty,contains=/"`               // Media MIME type per IETF spec
	Filename       *string           `json:"filename,omitempty" validate:"omitempty,min=1"`                     // Filename if applicable
	Body           any               `json:"body,omitempty"`                                                    // Text body or structured details
	Encoding       *string           `json:"encoding,omitempty" validate:"omitempty,oneof=base64url json none"` // e.g. "base64url"
	URL            *string           `json:"url,omitempty" validate:"omitempty,url"`                            // External reference URL
	ContentHash    *ContentHashValue `json:"content_hash,omitempty" validate:"omitempty"`                       // Hash of dialog content for integrity (supports single hash or array)
	Disposition    *string           `json:"disposition,omitempty" validate:"omitempty,min=1"`                  // For "incomplete"
	PartyHistory   []PartyHistory    `json:"party_history,omitempty" validate:"omitempty,dive"`                 // Join/drop events
	Transferee     *int              `json:"transferee,omitempty" validate:"omitempty,min=0"`                   // For transfer
	Transferor     *int              `json:"transferor,omitempty" validate:"omitempty,min=0"`                   // For transfer
	TransferTarget *int              `json:"transfer-target,omitempty" validate:"omitempty,min=0"`              // For transfer
	Original       *int              `json:"original,omitempty" validate:"omitempty,min=0"`                     // For consultation
	Consultation   *int              `json:"consultation,omitempty" validate:"omitempty,min=0"`                 // For consultation
	TargetDialog   *int              `json:"target-dialog,omitempty" validate:"omitempty,min=0"`                // For transfer
	Campaign       *string           `json:"campaign,omitempty" validate:"omitempty,min=1,max=255"`             // Campaign tag
	Interaction    *string           `json:"interaction,omitempty" validate:"omitempty,min=1,max=255"`          // Interaction tag
	Skill          *string           `json:"skill,omitempty" validate:"omitempty,min=1,max=255"`                // Skill or queue
	Duration       *float64          `json:"duration,omitempty" validate:"omitempty,gt=0"`                      // Recording duration in seconds
	Meta           map[string]any    `json:"meta,omitempty"`                                                    // Legacy metadata
	Metadata       map[string]any    `json:"metadata,omitempty"`                                                // Structured metadata
	Transfer       map[string]any    `json:"transfer,omitempty"`                                                // Transfer-specific info
	Signaling      map[string]any    `json:"signaling,omitempty"`                                               // Signaling info
	Resolution     *string           `json:"resolution,omitempty" validate:"omitempty,min=1"`                   // e.g. "1920x1080"
	FrameRate      *float64          `json:"frame_rate,omitempty" validate:"omitempty,gt=0"`                    // fps
	Codec          *string           `json:"codec,omitempty" validate:"omitempty,min=1"`                        // Video codec
	Bitrate        *int              `json:"bitrate,omitempty" validate:"omitempty,gt=0"`                       // kbps
	Thumbnail      *string           `json:"thumbnail,omitempty" validate:"omitempty,min=1"`                    // Base64 image data
	SessionID      *string           `json:"session_id,omitempty" validate:"omitempty,min=1,max=255"`           // Session identifier for dialog
	// Note: Application and MessageID fields removed for IETF draft-03 compliance
	// These were non-standard fields not present in the IETF spec
	AdditionalProperties map[string]any `json:"-"` // Extra fields
}

// Analysis represents analysis performed on the conversation.
// Each analysis has a type (e.g., "transcript", "summary", "sentiment") and vendor information.
// The Body field contains the analysis result, encoded according to the Encoding field.
type Analysis struct {
	Type                 string            `json:"type" validate:"required,min=1"`                                    // Analysis type (REQUIRED per IETF draft-03, allows custom types per "SHOULD" language)
	Dialog               any               `json:"dialog,omitempty"`                                                  // Dialog index or array of indices
	Vendor               string            `json:"vendor" validate:"required,min=1"`                                  // Vendor/product that created analysis (REQUIRED per IETF)
	Product              *string           `json:"product,omitempty" validate:"omitempty,min=1"`                      // Product name
	Schema               *string           `json:"schema,omitempty" validate:"omitempty,min=1"`                       // Schema for analysis format
	Mediatype            *string           `json:"mediatype,omitempty" validate:"omitempty,contains=/"`               // Media MIME type
	Filename             *string           `json:"filename,omitempty" validate:"omitempty,min=1"`                     // Optional filename
	Body                 any               `json:"body,omitempty"`                                                    // Analysis result (optional - can use URL)
	Encoding             *string           `json:"encoding,omitempty" validate:"omitempty,oneof=base64url json none"` // Encoding type
	URL                  *string           `json:"url,omitempty" validate:"omitempty,url"`                            // External reference URL
	ContentHash          *ContentHashValue `json:"content_hash,omitempty" validate:"omitempty"`                       // Hash of analysis content for integrity (supports single hash or array)
	Meta                 map[string]any    `json:"meta,omitempty"`                                                    // Additional metadata
	AdditionalProperties map[string]any    `json:"-"`                                                                 // Extra fields
}

// Attachment represents an attachment (inline or external) in the vCon.
// Attachments contain supplementary data such as documents, images, or other files.
// The Body field contains the actual data, encoded according to the Encoding field.
type Attachment struct {
	Type                 *string           `json:"type,omitempty" validate:"omitempty,min=1"`                         // Attachment type/purpose (optional per IETF)
	Start                *time.Time        `json:"start,omitempty"`                                                   // Optional timestamp (recommended per IETF spec)
	Party                *int              `json:"party" validate:"required,min=0"`                                   // Required party index per IETF spec
	Dialog               *int              `json:"dialog,omitempty" validate:"omitempty,min=0"`                       // Related dialog index
	Mediatype            *string           `json:"mediatype,omitempty" validate:"omitempty,contains=/"`               // Media MIME type per IETF spec
	Filename             *string           `json:"filename,omitempty" validate:"omitempty,min=1"`                     // Optional filename
	Body                 any               `json:"body,omitempty"`                                                    // Inline data (optional - can use URL instead)
	Encoding             *string           `json:"encoding,omitempty" validate:"omitempty,oneof=base64url json none"` // Encoding for body
	URL                  *string           `json:"url,omitempty" validate:"omitempty,url"`                            // External reference URL
	ContentHash          *ContentHashValue `json:"content_hash,omitempty" validate:"omitempty"`                       // Hash of attachment content for integrity (supports single hash or array)
	Meta                 map[string]any    `json:"meta,omitempty"`                                                    // Additional metadata per IETF extension pattern
	AdditionalProperties map[string]any    `json:"-"`                                                                 // Extra fields
}

// Signature represents a digital signature for vCon integrity.
type Signature struct {
	Protected string                 `json:"protected" validate:"required,min=1"` // Protected header (base64url-encoded)
	Header    map[string]interface{} `json:"header,omitempty"`                    // JWS unprotected header parameters
	Signature string                 `json:"signature" validate:"required,min=1"` // Signature value (base64url-encoded)
}

// RedactedObject represents a redacted vCon reference per IETF spec.
// This object indicates that the original vCon has been redacted or replaced.
type RedactedObject struct {
	UUID        string            `json:"uuid" validate:"required"`                                          // UUID of the redacted vCon
	Type        *string           `json:"type,omitempty" validate:"omitempty,min=1"`                         // Type of redaction
	Body        *string           `json:"body,omitempty"`                                                    // Redacted content (optional)
	Encoding    *string           `json:"encoding,omitempty" validate:"omitempty,oneof=base64url json none"` // Encoding for body
	URL         *string           `json:"url,omitempty" validate:"omitempty,url"`                            // External reference URL
	ContentHash *ContentHashValue `json:"content_hash,omitempty" validate:"omitempty"`                       // Hash of redacted content (supports single hash or array)
}

// AppendedObject represents an appended vCon reference per IETF spec.
// This object indicates additional data that has been appended to the vCon.
type AppendedObject struct {
	UUID        string            `json:"uuid" validate:"required"`                                          // UUID of the appended vCon
	Type        *string           `json:"type,omitempty" validate:"omitempty,min=1"`                         // Type of appended data
	Body        *string           `json:"body,omitempty"`                                                    // Appended content (optional)
	Encoding    *string           `json:"encoding,omitempty" validate:"omitempty,oneof=base64url json none"` // Encoding for body
	URL         *string           `json:"url,omitempty" validate:"omitempty,url"`                            // External reference URL
	ContentHash *ContentHashValue `json:"content_hash,omitempty" validate:"omitempty"`                       // Hash of appended content (supports single hash or array)
}

// GroupObject represents a grouped vCon reference per IETF spec.
// This object indicates that this vCon is part of a group or collection.
type GroupObject struct {
	UUID        string            `json:"uuid" validate:"required"`                                          // UUID of the grouped vCon
	Type        *string           `json:"type,omitempty" validate:"omitempty,min=1"`                         // Type of grouping relationship
	Body        *string           `json:"body,omitempty"`                                                    // Group metadata (optional)
	Encoding    *string           `json:"encoding,omitempty" validate:"omitempty,oneof=base64url json none"` // Encoding for body
	URL         *string           `json:"url,omitempty" validate:"omitempty,url"`                            // External reference URL
	ContentHash *ContentHashValue `json:"content_hash,omitempty" validate:"omitempty"`                       // Hash of group content (supports single hash or array)
}

// VCon is the top-level conversation container that holds all conversation data.
// It includes metadata, participants (parties), conversation segments (dialog),
// attachments, and analysis results. Each VCon has a unique UUID and version information.
type VCon struct {
	UUID                 uuid.UUID       `json:"uuid" validate:"required"`                                    // Conversation UUID (REQUIRED per IETF draft-03)
	Vcon                 string          `json:"vcon" validate:"required,eq=0.0.2"`                           // Version, must be "0.0.2" per IETF spec
	CreatedAt            time.Time       `json:"created_at" validate:"required"`                              // ISO timestamp
	UpdatedAt            *time.Time      `json:"updated_at,omitempty" validate:"omitempty,gtfield=CreatedAt"` // ISO timestamp
	Subject              *string         `json:"subject,omitempty" validate:"omitempty,min=1,max=255"`        // Optional subject line
	Redacted             *RedactedObject `json:"redacted,omitempty" validate:"omitempty"`                     // Redacted vCon reference (IETF spec)
	Group                []GroupObject   `json:"group,omitempty" validate:"omitempty,dive"`                   // Related vCon UUIDs (IETF spec)
	Parties              []Party         `json:"parties" validate:"required,dive"`                            // Participant list (REQUIRED per IETF spec)
	Dialog               []Dialog        `json:"dialog,omitempty" validate:"omitempty,dive"`                  // Conversation segments
	Attachments          []Attachment    `json:"attachments,omitempty" validate:"omitempty,dive"`             // Attachments
	Analysis             []Analysis      `json:"analysis,omitempty" validate:"omitempty,dive"`                // Analysis results (now properly typed)
	Signatures           []Signature     `json:"signatures,omitempty" validate:"omitempty,dive"`              // Digital signatures
	Payload              *string         `json:"payload,omitempty"`                                           // Signed payload for integrity
	Appended             *AppendedObject `json:"appended,omitempty" validate:"omitempty"`                     // Append-only data (IETF spec)
	AdditionalProperties map[string]any  `json:"-"`                                                           // Extra fields
}
