package promptemailverificationapi

import (
	"github.com/prodadidb/go-email-validator/pkg/ev"
	"github.com/prodadidb/go-email-validator/pkg/ev/evmail"
	"github.com/prodadidb/go-email-validator/pkg/presentation/converter"
)

//go:generate go run cmd/dep_test_generator/gen.go

const (
	// Name is a name of converter
	Name converter.Name = "PromptEmailVerificationApi"

	// ErrorSyntaxInvalid is a syntax error
	ErrorSyntaxInvalid = "Invalid email syntax."
)

type mx struct {
	AcceptsMail bool     `json:"accepts_mail"`
	Records     []string `json:"records"`
}

// DepPresentation is representation of https://promptapi.com/marketplace/description/email_verification-api
type DepPresentation struct {
	Email          string `json:"email"`
	SyntaxValid    bool   `json:"syntax_valid"`
	IsDisposable   bool   `json:"is_disposable"`
	IsRoleAccount  bool   `json:"is_role_account"`
	IsCatchAll     bool   `json:"is_catch_all"`
	IsDeliverable  bool   `json:"is_deliverable"`
	CanConnectSMTP bool   `json:"can_connect_smtp"`
	IsInboxFull    bool   `json:"is_inbox_full"`
	IsDisabled     bool   `json:"is_disabled"`
	MxRecords      mx     `json:"mx_records"`
	Message        string `json:"message"`
}

// NewDepConverter creates DepConverter
func NewDepConverter() DepConverter {
	return DepConverter{}
}

// DepConverter is the converter for https://promptapi.com/marketplace/description/email_verification-api
type DepConverter struct{}

// Can be used for ev.ValidationResult
func (DepConverter) Can(_ evmail.Address, result ev.ValidationResult, _ converter.Options) bool {
	return result.ValidatorName() == ev.DepValidatorName
}

// Convert converts ev.ValidationResult in DepPresentation
func (d DepConverter) Convert(email evmail.Address, resultInterface ev.ValidationResult, _ converter.Options) (result interface{}) {
	var message string
	depResult := resultInterface.(ev.DepValidationResult)
	validationResults := depResult.GetResults()
	mxResult := validationResults[ev.MXValidatorName].(ev.MXValidationResult)

	smtpPresentation := converter.NewSMTPConverter().Convert(email, validationResults[ev.SMTPValidatorName], nil).(converter.SMTPPresentation)

	Email := email.String()
	isSyntaxValid := validationResults[ev.SyntaxValidatorName].IsValid()
	if !isSyntaxValid && len(Email) == 0 {
		message = ErrorSyntaxInvalid
	}

	depPresentation := DepPresentation{
		Email:          Email,
		SyntaxValid:    isSyntaxValid,
		IsDisposable:   !validationResults[ev.DisposableValidatorName].IsValid(),
		IsRoleAccount:  !validationResults[ev.RoleValidatorName].IsValid(),
		IsCatchAll:     smtpPresentation.IsCatchAll,
		IsDeliverable:  smtpPresentation.IsDeliverable,
		CanConnectSMTP: smtpPresentation.CanConnectSMTP,
		IsInboxFull:    smtpPresentation.HasFullInbox,
		IsDisabled:     smtpPresentation.IsDisabled,
		MxRecords: mx{
			AcceptsMail: mxResult.IsValid(),
			Records:     converter.MX2String(mxResult.MX()),
		},
		Message: message,
	}

	return depPresentation
}

// NewDepValidator is a common dep validator object
func NewDepValidator(smtpValidator ev.Validator) ev.Validator {
	builder := ev.NewDepBuilder(nil)
	if smtpValidator == nil {
		smtpValidator = builder.Get(ev.SMTPValidatorName)
	}

	return ev.NewDepBuilder(nil).Set(
		ev.SyntaxValidatorName,
		ev.NewSyntaxRegexValidator(nil),
	).Set(ev.SMTPValidatorName, smtpValidator).Build()
}
