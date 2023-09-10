package converter

import (
	"errors"
	"net/textproto"
	"strings"
	"sync"

	"github.com/prodadidb/go-email-validator/pkg/ev"
	"github.com/prodadidb/go-email-validator/pkg/ev/evmail"
	"github.com/prodadidb/go-email-validator/pkg/ev/evsmtp"
)

// SMTPPresentation is a presentation of smtp from ev.ValidationResult
type SMTPPresentation struct {
	CanConnectSMTP bool `json:"can_connect_smtp"`
	HasFullInbox   bool `json:"has_full_inbox"`
	IsCatchAll     bool `json:"is_catch_all"`
	IsDeliverable  bool `json:"is_deliverable"`
	IsDisabled     bool `json:"is_disabled"`
	IsGreyListed   bool `json:"is_grey_listed"`
}

// Default results
var (
	WithoutErrsSMTPPresentation = SMTPPresentation{
		CanConnectSMTP: true,
		HasFullInbox:   false,
		IsCatchAll:     true,
		IsDeliverable:  true,
		IsDisabled:     false,
		IsGreyListed:   false,
	}
	FalseSMTPPresentation = SMTPPresentation{
		CanConnectSMTP: false,
		HasFullInbox:   false,
		IsCatchAll:     false,
		IsDeliverable:  false,
		IsDisabled:     false,
		IsGreyListed:   false,
	}
)

var smtpConverter *SMTPConverter

var muNewSMTPConverter sync.Mutex

// NewSMTPConverter creates SMTPConverter
func NewSMTPConverter() *SMTPConverter {
	muNewSMTPConverter.Lock()
	defer muNewSMTPConverter.Unlock()

	if smtpConverter == nil {
		smtpConverter = &SMTPConverter{}
	}

	return smtpConverter
}

// SMTPConverter converts ev.ValidationResult in SMTPConverter
type SMTPConverter struct{}

// Can ev.ValidationResult be converted in SMTPConverter
func (SMTPConverter) Can(_ evmail.Address, result ev.ValidationResult, _ Options) bool {
	return result.ValidatorName() == ev.SMTPValidatorName
}

// Convert ev.ValidationResult in SMTPConverter
func (SMTPConverter) Convert(_ evmail.Address, result ev.ValidationResult, _ Options) interface{} {
	var presentation = WithoutErrsSMTPPresentation
	var errString string
	var errCode int
	var smtpError evsmtp.Error
	var depError *ev.DepsError

	errs := result.Errors()
	errs = append(errs, result.Warnings()...)

	for _, err := range errs {
		if !errors.As(err, &smtpError) {
			if errors.As(err, &depError) {
				return FalseSMTPPresentation
			}
			continue
		}

		sourceErr := errors.Unwrap(smtpError)
		errString = strings.ToLower(sourceErr.Error())

		errCode = 0
		switch v := sourceErr.(type) {
		case *textproto.Error:
			errCode = v.Code
		}
		if strings.Contains(errString, "greylist") {
			presentation.IsGreyListed = true
		}

		switch smtpError.Stage() {
		case evsmtp.ConnectionStage:
			presentation = FalseSMTPPresentation
		case evsmtp.HelloStage,
			evsmtp.AuthStage,
			evsmtp.MailStage:
			presentation.IsDeliverable = false
		case evsmtp.RandomRCPTStage:
			presentation.IsCatchAll = false
		case evsmtp.RCPTsStage:
			presentation.IsDeliverable = false
			switch {
			case strings.Contains(errString, "disabled") ||
				strings.Contains(errString, "discontinued"):
				presentation.IsDisabled = true
			case errCode == 452 && (strings.Contains(errString, "full") ||
				strings.Contains(errString, "insufficient") ||
				strings.Contains(errString, "over quota") ||
				strings.Contains(errString, "space") ||
				strings.Contains(errString, "too many messages")):

				presentation.HasFullInbox = true
			case strings.Contains(errString, "the user you are trying to contact is receiving mail at a rate that"):
				presentation.IsDeliverable = true
			}
		}
	}

	return presentation
}
