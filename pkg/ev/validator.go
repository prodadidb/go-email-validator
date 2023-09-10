package ev

import (
	"github.com/prodadidb/go-email-validator/pkg/ev/evsmtp"
	"github.com/prodadidb/go-email-validator/pkg/ev/utils"
	"github.com/vmihailenco/msgpack"
)

func init() {
	msgpack.RegisterExt(evsmtp.ExtID(), new(DepsError))
	msgpack.RegisterExt(evsmtp.ExtID(), new(AValidationResult))
}

// OtherValidator is ValidatorName for unknown Validator
const OtherValidator ValidatorName = "other"

// Validator is interface for validators
type Validator interface {
	GetDeps() []ValidatorName
	Validate(input Input, results ...ValidationResult) ValidationResult
}

// ChangeableValidationResult is ValidationResult with changeable errors and warnings
type ChangeableValidationResult interface {
	SetErrors([]error)
	SetWarnings([]error)
}

// ValidationResult is interface to represent result of validation
type ValidationResult interface {
	// IsValid is status of validation
	IsValid() bool
	// Errors of result after validation
	Errors() []error
	// HasErrors checks for the presence of the Errors
	HasErrors() bool
	// Warnings of result after validation
	Warnings() []error
	// HasWarnings checks for the presence of the Warnings
	HasWarnings() bool
	// ValidatorName returns name of validator
	ValidatorName() ValidatorName
}

// AValidationResult is abstract class for extending of validation
type AValidationResult struct {
	IsValidVal  bool
	ErrorsVal   []error
	WarningsVal []error
	NameVal     ValidatorName
}

// IsValid is status of validation
func (a *AValidationResult) IsValid() bool {
	return a.IsValidVal
}

// SetErrors sets errors
func (a *AValidationResult) SetErrors(errors []error) {
	a.ErrorsVal = errors
	a.IsValidVal = !a.HasErrors()
}

// Errors of result after validation
func (a *AValidationResult) Errors() []error {
	return a.ErrorsVal
}

// HasErrors checks for the presence of the Errors
func (a *AValidationResult) HasErrors() bool {
	return utils.RangeLen(a.Errors()) > 0
}

// SetWarnings sets warnings
func (a *AValidationResult) SetWarnings(warnings []error) {
	a.WarningsVal = warnings
}

// Warnings of result after validation
func (a *AValidationResult) Warnings() []error {
	return a.WarningsVal
}

// HasWarnings checks for the presence of the Warnings
func (a *AValidationResult) HasWarnings() bool {
	return utils.RangeLen(a.Warnings()) > 0
}

// ValidatorName returns name of validator
func (a *AValidationResult) ValidatorName() ValidatorName {
	return a.NameVal
}

// EncodeMsgpack is used to fix this problem https://github.com/vmihailenco/msgpack/issues/294
func (a *AValidationResult) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.EncodeMulti(
		a.IsValidVal,
		evsmtp.ErrorsToEVSMTPErrors(a.ErrorsVal),
		evsmtp.ErrorsToEVSMTPErrors(a.WarningsVal),
		a.NameVal,
	)
}

// DecodeMsgpack is used to fix this problem https://github.com/vmihailenco/msgpack/issues/294
func (a *AValidationResult) DecodeMsgpack(dec *msgpack.Decoder) error {
	return dec.DecodeMulti(&a.IsValidVal, &a.ErrorsVal, &a.WarningsVal, &a.NameVal)
}

type ValidationResultStruct = AValidationResult

// NewValidResult returns valid validation result
func NewValidResult(name ValidatorName) ValidationResult {
	return NewResult(true, nil, nil, name)
}

// NewResult returns result of validation by parameters
func NewResult(isValid bool, errors []error, warnings []error, name ValidatorName) ValidationResult {
	if name == "" {
		name = OtherValidator
	}

	return &ValidationResultStruct{isValid, errors, warnings, name}
}

var EmptyDeps = make([]ValidatorName, 0)

// AValidatorWithoutDeps is an abstract structure for validator without dependencies
type AValidatorWithoutDeps struct{}

// GetDeps returns dependencies of Validator
func (a AValidatorWithoutDeps) GetDeps() []ValidatorName {
	return EmptyDeps
}
