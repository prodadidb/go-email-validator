package ev

import (
	"github.com/emirpasic/gods/sets/hashset"
	"github.com/prodadidb/go-email-validator/pkg/ev/contains"
	"github.com/prodadidb/go-email-validator/pkg/ev/disposable"
	"github.com/prodadidb/go-email-validator/pkg/ev/evsmtp"
	"github.com/prodadidb/go-email-validator/pkg/ev/role"
)

// GetDefaultSMTPValidator instantiates default SMTPValidatorName
func GetDefaultSMTPValidator(dto evsmtp.CheckerDTO) Validator {
	return NewWarningsDecorator(
		NewSMTPValidator(evsmtp.NewChecker(dto)),
		NewIsWarning(hashset.New(evsmtp.RandomRCPTStage), func(warningMap WarningSet) IsWarning {
			return func(err error) bool {
				errSMTP, ok := err.(evsmtp.Error)
				if !ok {
					return false
				}
				return warningMap.Contains(errSMTP.Stage())
			}
		}),
	)
}

// GetDefaultFactories returns default list ValidatorMap
func GetDefaultFactories() ValidatorMap {
	return ValidatorMap{
		RoleValidatorName:       NewRoleValidator(role.NewRBEASetRole()),
		DisposableValidatorName: NewDisposableValidator(contains.NewFunc(disposable.MailChecker)),
		SyntaxValidatorName:     NewSyntaxValidator(),
		MXValidatorName:         DefaultNewMXValidator(),
		SMTPValidatorName:       GetDefaultSMTPValidator(evsmtp.CheckerDTO{}),
	}
}

// NewDepBuilder instantiates Validator with ValidatorMap or GetDefaultFactories validators
func NewDepBuilder(validators ValidatorMap) *DepBuilder {
	if validators == nil {
		validators = GetDefaultFactories()
	}

	return &DepBuilder{Validators: validators}
}

// DepBuilder is used to form Validator
type DepBuilder struct {
	Validators ValidatorMap
}

// Set sets validator by ValidatorName
func (d *DepBuilder) Set(name ValidatorName, validator Validator) *DepBuilder {
	d.Validators[name] = validator

	return d
}

// Get returns validator by ValidatorName
func (d *DepBuilder) Get(name ValidatorName) Validator {
	if d.Has(name) {
		return d.Validators[name]
	}

	return nil
}

// Has checks for existing validators by ValidatorName...
func (d *DepBuilder) Has(names ...ValidatorName) bool {
	for _, name := range names {
		if _, has := d.Validators[name]; !has {
			return false
		}
	}

	return true
}

// Delete deletes validators by ValidatorName...
func (d *DepBuilder) Delete(names ...ValidatorName) *DepBuilder {
	for _, name := range names {
		if d.Has(name) {
			delete(d.Validators, name)
		}
	}

	return d
}

// Build builds Validator based on configuration
func (d *DepBuilder) Build() Validator {
	return NewDepValidator(d.Validators)
}
