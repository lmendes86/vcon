package vcon

import "fmt"

// ValidationLevel represents different levels of validation strictness.
type ValidationLevel int

const (
	// ValidationBasic performs core business logic validation.
	ValidationBasic ValidationLevel = iota

	// ValidationStrict performs business logic validation with additional strict rules.
	ValidationStrict

	// ValidationIETF performs IETF specification compliance validation.
	ValidationIETF

	// ValidationIETFDraft03 performs strict IETF draft-03 specification compliance validation.
	ValidationIETFDraft03

	// ValidationIETFStrict performs strict IETF compliance with extension field detection.
	ValidationIETFStrict

	// ValidationComplete performs comprehensive validation (business + IETF + strict).
	ValidationComplete
)

// ValidationResult contains the results of validation operations.
type ValidationResult struct {
	Level    ValidationLevel
	Valid    bool
	Errors   []ValidationError
	Warnings []ValidationWarning
}

// ValidationWarning represents a validation warning (non-blocking).
type ValidationWarning struct {
	Field   string
	Message string
	Level   string
}

func (w ValidationWarning) String() string {
	return fmt.Sprintf("validation warning in %s (%s): %s", w.Field, w.Level, w.Message)
}

// ValidateWithLevel performs validation at the specified level.
func (v *VCon) ValidateWithLevel(level ValidationLevel) *ValidationResult {
	result := &ValidationResult{
		Level:  level,
		Valid:  true,
		Errors: []ValidationError{},
	}

	var err error

	switch level {
	case ValidationBasic:
		err = v.Validate()
	case ValidationStrict:
		err = v.ValidateStrict()
	case ValidationIETF:
		err = v.ValidateIETF()
	case ValidationIETFDraft03:
		err = v.ValidateIETFDraft03()
	case ValidationIETFStrict:
		err = v.ValidateIETFStrict()
	case ValidationComplete:
		err = v.ValidateCompleteStrict()
	default:
		err = fmt.Errorf("unknown validation level: %d", level)
	}

	if err != nil {
		result.Valid = false
		if validationErrors, ok := err.(ValidationErrors); ok {
			result.Errors = append(result.Errors, validationErrors...)
		} else {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "general",
				Message: err.Error(),
			})
		}
	}

	return result
}

// String returns a string representation of the validation level.
func (vl ValidationLevel) String() string {
	switch vl {
	case ValidationBasic:
		return "Basic"
	case ValidationStrict:
		return "Strict"
	case ValidationIETF:
		return "IETF"
	case ValidationIETFDraft03:
		return "IETF-Draft-03"
	case ValidationIETFStrict:
		return "IETF-Strict"
	case ValidationComplete:
		return "Complete"
	default:
		return "Unknown"
	}
}
