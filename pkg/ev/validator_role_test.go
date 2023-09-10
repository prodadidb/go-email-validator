package ev_test

import (
	"reflect"
	"testing"

	"github.com/prodadidb/go-email-validator/pkg/ev"
	"github.com/prodadidb/go-email-validator/pkg/ev/contains"
	"github.com/prodadidb/go-email-validator/pkg/ev/evmail"
	"github.com/prodadidb/go-email-validator/pkg/ev/utils"
)

func Test_roleValidator_Validate(t *testing.T) {
	type fields struct {
		r contains.InSet
	}
	type args struct {
		email evmail.Address
		in1   []ev.ValidationResult
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   ev.ValidationResult
	}{
		{
			name: "email has not role",
			fields: fields{
				r: mockContains{t: t, want: validEmail.Username(), ret: false},
			},
			args: args{
				email: validEmail,
			},
			want: ev.NewResult(true, nil, nil, ev.RoleValidatorName),
		},
		{
			name: "email has role",
			fields: fields{
				r: mockContains{t: t, want: validEmail.Username(), ret: true},
			},
			args: args{
				email: validEmail,
			},
			want: ev.NewResult(false, utils.Errs(ev.RoleError{}), nil, ev.RoleValidatorName),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := ev.NewRoleValidator(tt.fields.r)
			if got := r.Validate(ev.NewInput(tt.args.email), tt.args.in1...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoleError_Error(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			want: ev.RoleErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ro := ev.RoleError{}
			if got := ro.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
