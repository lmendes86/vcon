package vcon

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// SchemaValidator is the global validator instance.
var SchemaValidator *validator.Validate

func init() {
	// Initialize validator with required struct validation enabled.
	SchemaValidator = validator.New(validator.WithRequiredStructEnabled())

	// Register custom validators for IETF URI format compliance
	if err := SchemaValidator.RegisterValidation("tel_uri", validateTelURI); err != nil {
		panic("failed to register tel_uri validator: " + err.Error())
	}
	if err := SchemaValidator.RegisterValidation("mailto_uri", validateMailtoURI); err != nil {
		panic("failed to register mailto_uri validator: " + err.Error())
	}
}

// JSONValidationErrors wraps validator.ValidationErrors to make it JSON-serializable.
type JSONValidationErrors validator.ValidationErrors

// Error returns the error as a string.
func (jve JSONValidationErrors) Error() string {
	vve := validator.ValidationErrors(jve)
	return vve.Error()
}

// MarshalJSON returns validation errors as JSON.
func (jve JSONValidationErrors) MarshalJSON() ([]byte, error) {
	errorMap := make(map[string]JSONValidationErrorDetails)

	for _, fieldError := range jve {
		errorMap[strings.ToLower(fieldError.Field())] = JSONValidationErrorDetails{
			Validator: fieldError.ActualTag(),
			Value:     fieldError.Value(),
			Message:   fmt.Sprintf("Field validation for '%s' failed on the '%s' tag", fieldError.Field(), fieldError.ActualTag()),
		}
	}

	return json.Marshal(errorMap)
}

// JSONValidationErrorDetails contains the actual body of the JSON error used by JSONValidationErrors.
type JSONValidationErrorDetails struct {
	Validator string      `json:"validator"`
	Value     interface{} `json:"value"`
	Message   string      `json:"message"`
}

// ToJSONValidationErrors converts ValidationErrors into JSONValidationErrors for JSON serialization.
func ToJSONValidationErrors(err error) error {
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		return JSONValidationErrors(validationErrs)
	}
	return err
}

// ValidateStruct validates a struct using the global validator instance.
// Returns JSONValidationErrors if validation fails, nil if successful.
func ValidateStruct(s interface{}) error {
	if err := SchemaValidator.Struct(s); err != nil {
		return ToJSONValidationErrors(err)
	}
	return nil
}

// ValidateStructJSON validates a struct and returns JSON-formatted validation errors.
// This is useful for API responses where you want structured error information.
func ValidateStructJSON(s interface{}) ([]byte, error) {
	if err := ValidateStruct(s); err != nil {
		if jsonErr, ok := err.(JSONValidationErrors); ok {
			return json.Marshal(jsonErr)
		}
		// Fallback for other error types
		return json.Marshal(map[string]string{"error": err.Error()})
	}
	return nil, nil
}

// validateTelURI validates telephone numbers in both URI format (tel:+1234567890) and plain format (+1234567890).
// Per IETF vCon spec, tel fields should accept URI format as well as plain E.164 format.
func validateTelURI(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // Allow empty values (omitempty handles this)
	}

	// Check if it's a tel: URI format
	if strings.HasPrefix(value, "tel:") {
		// Extract the phone number part after "tel:"
		phoneNumber := strings.TrimPrefix(value, "tel:")
		return isValidE164(phoneNumber)
	}

	// Otherwise validate as plain E.164 format
	return isValidE164(value)
}

// validateMailtoURI validates email addresses in both URI format (mailto:user@domain.com) and plain format (user@domain.com).
// Per IETF vCon spec, mailto fields should accept URI format as well as plain email format.
func validateMailtoURI(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // Allow empty values (omitempty handles this)
	}

	// Check if it's a mailto: URI format
	if strings.HasPrefix(value, "mailto:") {
		// Extract the email address part after "mailto:"
		emailAddr := strings.TrimPrefix(value, "mailto:")
		return isValidEmail(emailAddr)
	}

	// Otherwise validate as plain email format
	return isValidEmail(value)
}

// isValidE164 validates a phone number in E.164 format (+1234567890).
func isValidE164(phone string) bool {
	if phone == "" {
		return false
	}

	// E.164 format: starts with +, followed by up to 15 digits
	e164Regex := regexp.MustCompile(`^\+[1-9]\d{0,14}$`)
	return e164Regex.MatchString(phone)
}

// isValidEmail validates an email address using a basic regex pattern.
func isValidEmail(email string) bool {
	if email == "" {
		return false
	}

	// Basic email validation regex (simplified but sufficient for most cases)
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
