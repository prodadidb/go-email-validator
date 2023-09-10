package ev_test

import (
	"testing"

	"github.com/emirpasic/gods/sets/hashset"
	"github.com/prodadidb/go-email-validator/pkg/ev"
	"github.com/prodadidb/go-email-validator/pkg/ev/evmail"
	"github.com/prodadidb/go-email-validator/pkg/ev/evsmtp"
	"github.com/stretchr/testify/require"
)

// test monicaramirezrestrepo@hotmail.com.
func newSMTPValidator() ev.Validator {
	return ev.NewSMTPValidator(evsmtp.NewChecker(evsmtp.CheckerDTO{
		Options: evsmtp.NewOptions(evsmtp.OptionsDTO{
			EmailFrom: evmail.FromString(evsmtp.DefaultEmail),
		}),
	}))
}

func getSMTPValidatorValidate() ev.Validator {
	return ev.NewDepValidator(
		map[ev.ValidatorName]ev.Validator{
			ev.SyntaxValidatorName: ev.NewSyntaxValidator(),
			ev.MXValidatorName:     ev.DefaultNewMXValidator(),
			ev.SMTPValidatorName: ev.NewWarningsDecorator(
				newSMTPValidator(),
				ev.NewIsWarning(hashset.New(evsmtp.RandomRCPTStage), func(warningMap ev.WarningSet) ev.IsWarning {
					return func(err error) bool {
						return warningMap.Contains(err.(evsmtp.Error).Stage())
					}
				}),
			),
		},
	)
}

func BenchmarkSMTPValidator_Validate(b *testing.B) {
	email := evmail.FromString(ValidEmailString)
	validator := getSMTPValidatorValidate()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(ev.NewInput(email))
	}
}

func TestSMTPValidator_Validate_WithoutMock(t *testing.T) {
	email := evmail.FromString(ValidEmailString)
	validator := getSMTPValidatorValidate()
	v := validator.Validate(ev.NewInput(email))
	require.True(t, v.IsValid())
}
