package ev_test

import (
	"reflect"
	"testing"

	"github.com/prodadidb/go-email-validator/pkg/ev"
	"github.com/prodadidb/go-email-validator/pkg/ev/contains"
	"github.com/prodadidb/go-email-validator/pkg/ev/evmail"
	"github.com/prodadidb/go-email-validator/pkg/ev/utils"
)

func Test_blackListValidator_Validate(t *testing.T) {
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
			name: "domain is valid",
			fields: fields{
				d: mockContains{t: t, want: validEmail.Domain(), ret: false},
			},
			args: args{
				email: validEmail,
			},
			want: ev.NewResult(true, nil, nil, ev.BlackListDomainsValidatorName),
		},
		{
			name: "domain is invalid",
			fields: fields{
				d: mockContains{t: t, want: validEmail.Domain(), ret: true},
			},
			args: args{
				email: validEmail,
			},
			want: ev.NewResult(false, utils.Errs(ev.BlackListDomainsError{}), nil, ev.BlackListDomainsValidatorName),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := ev.NewBlackListValidator(tt.fields.d)
			if got := w.Validate(ev.NewInput(tt.args.email)); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}
