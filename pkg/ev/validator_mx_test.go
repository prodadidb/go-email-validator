package ev_test

import (
	"net"
	"reflect"
	"testing"

	"github.com/prodadidb/go-email-validator/pkg/ev"
	"github.com/prodadidb/go-email-validator/pkg/ev/evmail"
	"github.com/prodadidb/go-email-validator/pkg/ev/evsmtp"
	"github.com/prodadidb/go-email-validator/pkg/ev/utils"
	"github.com/stretchr/testify/require"
)

func mockLookupMX(t *testing.T, domainExpected string, ret evsmtp.MXs, err error) evsmtp.FuncLookupMX {
	return func(domain string) ([]*net.MX, error) {
		require.Equal(t, domainExpected, domain)

		return ret, err
	}
}

func Test_mxValidator_Validate(t *testing.T) {
	type fields struct {
		lookupMX evsmtp.FuncLookupMX
	}
	type args struct {
		email evmail.Address
	}

	mxs := evsmtp.MXs{&net.MX{}}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   ev.ValidationResult
	}{
		{
			name: "existed domain",
			fields: fields{
				lookupMX: mockLookupMX(t, validEmail.Domain(), mxs, nil),
			},
			args: args{
				email: validEmail,
			},
			want: ev.NewMXValidationResult(
				mxs,
				ev.NewResult(true, nil, nil, ev.MXValidatorName).(*ev.AValidationResult),
			),
		},
		{
			name: "empty mx list",
			fields: fields{
				lookupMX: mockLookupMX(t, validEmail.Domain(), nil, nil),
			},
			args: args{
				email: validEmail,
			},
			want: ev.NewMXValidationResult(
				nil,
				ev.NewResult(false, utils.Errs(ev.EmptyMXsError{}), nil, ev.MXValidatorName).(*ev.AValidationResult),
			),
		},
		{
			name: "unexisted domain",
			fields: fields{
				lookupMX: mockLookupMX(t, validEmail.Domain(), nil, errorSimple),
			},
			args: args{
				email: validEmail,
			},
			want: ev.NewMXValidationResult(
				nil,
				ev.NewResult(false, utils.Errs(errorSimple), nil, ev.MXValidatorName).(*ev.AValidationResult),
			),
		},
		{
			name: "unexisted domain with mxs",
			fields: fields{
				lookupMX: mockLookupMX(t, validEmail.Domain(), mxs, errorSimple),
			},
			args: args{
				email: validEmail,
			},
			want: ev.NewMXValidationResult(
				mxs,
				ev.NewResult(false, utils.Errs(errorSimple), nil, ev.MXValidatorName).(*ev.AValidationResult),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := ev.NewMXValidator(tt.fields.lookupMX)
			if got := v.Validate(ev.NewInput(tt.args.email)); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkSMTPValidator_Validate_MX(b *testing.B) {
	email := evmail.FromString(ValidEmailString)

	depValidator := ev.NewDepValidator(
		map[ev.ValidatorName]ev.Validator{
			ev.SyntaxValidatorName: ev.DefaultNewMXValidator(),
			ev.MXValidatorName:     ev.NewSyntaxValidator(),
		},
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		depValidator.Validate(ev.NewInput(email))
	}
}
