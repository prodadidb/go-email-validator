package ev_test

import (
	"reflect"
	"testing"

	"github.com/prodadidb/go-email-validator/pkg/ev"
	"github.com/prodadidb/go-email-validator/pkg/ev/contains"
	"github.com/prodadidb/go-email-validator/pkg/ev/evmail"
	"github.com/prodadidb/go-email-validator/pkg/ev/utils"
)

func Test_blackListEmailsValidator_Validate(t *testing.T) {
	type fields struct {
		d contains.InSet
	}
	type args struct {
		email evmail.Address
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   ev.ValidationResult
	}{
		{
			name: "email is valid",
			fields: fields{
				d: mockContains{t: t, want: validEmail.String(), ret: false},
			},
			args: args{
				email: validEmail,
			},
			want: ev.NewResult(true, nil, nil, ev.BlackListEmailsValidatorName),
		},
		{
			name: "email is invalid",
			fields: fields{
				d: mockContains{t: t, want: validEmail.String(), ret: true},
			},
			args: args{
				email: validEmail,
			},
			want: ev.NewResult(false, utils.Errs(ev.BlackListEmailsError{}), nil, ev.BlackListEmailsValidatorName),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := ev.NewBlackListEmailsValidator(tt.fields.d)
			if got := w.Validate(ev.NewInput(tt.args.email)); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlackListEmailsError_Error(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "success",
			want: ev.BlackListEmailsErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bl := ev.BlackListEmailsError{}
			if got := bl.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
