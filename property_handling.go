package vcon

import (
	"encoding/json"
)

// PropertyHandler manages how non-standard properties are handled during JSON processing.
type PropertyHandler struct {
	Mode string // default, strict, or meta
}

// NewPropertyHandler creates a new PropertyHandler with the specified mode.
func NewPropertyHandler(mode string) *PropertyHandler {
	if mode != PropertyHandlingDefault && mode != PropertyHandlingStrict && mode != PropertyHandlingMeta {
		mode = PropertyHandlingDefault
	}
	return &PropertyHandler{Mode: mode}
}

// ProcessProperties processes an object's properties based on the handler mode.
func (h *PropertyHandler) ProcessProperties(data map[string]interface{}, allowedProperties map[string]bool) map[string]interface{} {
	if data == nil {
		return nil
	}

	result := make(map[string]interface{})
	nonStandard := make(map[string]interface{})

	// Separate standard and non-standard properties
	for key, value := range data {
		if allowedProperties[key] {
			result[key] = value
		} else {
			nonStandard[key] = value
		}
	}

	// Handle non-standard properties based on mode
	switch h.Mode {
	case PropertyHandlingStrict:
		// Ignore non-standard properties - they're already excluded
	case PropertyHandlingMeta:
		// Move non-standard properties to meta
		if len(nonStandard) > 0 {
			if result["meta"] == nil {
				result["meta"] = make(map[string]interface{})
			}
			if meta, ok := result["meta"].(map[string]interface{}); ok {
				for k, v := range nonStandard {
					meta[k] = v
				}
			}
		}
	default: // PropertyHandlingDefault
		// Keep non-standard properties
		for k, v := range nonStandard {
			result[k] = v
		}
	}

	return result
}

// ProcessVCon processes a VCon object according to property handling rules.
func (h *PropertyHandler) ProcessVCon(data map[string]interface{}) map[string]interface{} {
	processed := h.ProcessProperties(data, AllowedVConProperties)

	// Process embedded objects
	if parties, ok := processed["parties"].([]interface{}); ok {
		processedParties := make([]interface{}, len(parties))
		for i, party := range parties {
			if partyMap, ok := party.(map[string]interface{}); ok {
				processedParties[i] = h.ProcessProperties(partyMap, AllowedPartyProperties)
			} else {
				processedParties[i] = party
			}
		}
		processed["parties"] = processedParties
	}

	if dialogs, ok := processed["dialog"].([]interface{}); ok {
		processedDialogs := make([]interface{}, len(dialogs))
		for i, dialog := range dialogs {
			if dialogMap, ok := dialog.(map[string]interface{}); ok {
				processedDialogs[i] = h.ProcessProperties(dialogMap, AllowedDialogProperties)
			} else {
				processedDialogs[i] = dialog
			}
		}
		processed["dialog"] = processedDialogs
	}

	if attachments, ok := processed["attachments"].([]interface{}); ok {
		processedAttachments := make([]interface{}, len(attachments))
		for i, attachment := range attachments {
			if attachmentMap, ok := attachment.(map[string]interface{}); ok {
				processedAttachments[i] = h.ProcessProperties(attachmentMap, AllowedAttachmentProperties)
			} else {
				processedAttachments[i] = attachment
			}
		}
		processed["attachments"] = processedAttachments
	}

	if analysis, ok := processed["analysis"].([]interface{}); ok {
		processedAnalysis := make([]interface{}, len(analysis))
		for i, item := range analysis {
			if analysisMap, ok := item.(map[string]interface{}); ok {
				processedAnalysis[i] = h.ProcessProperties(analysisMap, AllowedAnalysisProperties)
			} else {
				processedAnalysis[i] = item
			}
		}
		processed["analysis"] = processedAnalysis
	}

	return processed
}

// MarshalVConWithPropertyHandling marshals a VCon with property handling.
func (h *PropertyHandler) MarshalVConWithPropertyHandling(v *VCon) ([]byte, error) {
	// First marshal to get JSON representation
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	// Unmarshal to map for processing
	var vconMap map[string]interface{}
	if err := json.Unmarshal(data, &vconMap); err != nil {
		return nil, err
	}

	// Process properties
	processed := h.ProcessVCon(vconMap)

	// Marshal the processed result
	return json.Marshal(processed)
}

// UnmarshalVConWithPropertyHandling unmarshals JSON to VCon with property handling.
func (h *PropertyHandler) UnmarshalVConWithPropertyHandling(data []byte, v *VCon) error {
	// First unmarshal to map
	var vconMap map[string]interface{}
	if err := json.Unmarshal(data, &vconMap); err != nil {
		return err
	}

	// Process properties
	processed := h.ProcessVCon(vconMap)

	// Marshal processed data and unmarshal to struct
	processedData, err := json.Marshal(processed)
	if err != nil {
		return err
	}

	return json.Unmarshal(processedData, v)
}
