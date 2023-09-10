package ev

import (
	"github.com/prodadidb/go-email-validator/pkg/ev/contains"
	"github.com/prodadidb/go-email-validator/pkg/ev/free"
	"github.com/prodadidb/go-email-validator/pkg/ev/utils"
)

// FreeValidatorName is name of free validator
const FreeValidatorName ValidatorName = "FreeValidator"

// FreeErr is text for FreeError.Error
const FreeErr = "FreeError"

// FreeError is FreeValidatorName error
type FreeError struct{}

func (FreeError) Error() string {
	return FreeErr
}

// FreeDefaultValidator instantiates default FreeValidatorName based on free.NewWillWhiteSetFree()
func FreeDefaultValidator() Validator {
	return NewFreeValidator(free.NewWillWhiteSetFree())
}

// NewFreeValidator instantiates FreeValidatorName
func NewFreeValidator(f contains.InSet) Validator {
	return freeValidator{f: f}
}

type freeValidator struct {
	AValidatorWithoutDeps
	f contains.InSet
}

func (r freeValidator) Validate(input Input, _ ...ValidationResult) ValidationResult {
	var err error
	var isFree = r.f.Contains(input.Email().Domain())
	if isFree {
		err = FreeError{}
	}

	return NewResult(!isFree, utils.Errs(err), nil, FreeValidatorName)
}
