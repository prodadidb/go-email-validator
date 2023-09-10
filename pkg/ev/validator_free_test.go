package ev_test

import (
	"reflect"
	"testing"

	"github.com/prodadidb/go-email-validator/pkg/ev"
	"github.com/prodadidb/go-email-validator/pkg/ev/contains"
	"github.com/prodadidb/go-email-validator/pkg/ev/evmail"
	"github.com/prodadidb/go-email-validator/pkg/ev/free"
	"github.com/prodadidb/go-email-validator/pkg/ev/utils"
)

func Test_freeValidator_Validate(t *testing.T) {
	type fields struct {
		f contains.InSet
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
			name: "email is not free",
			fields: fields{
				f: mockContains{t: t, want: validEmail.Domain(), ret: false},
			},
			args: args{
				email: validEmail,
			},
			want: ev.NewResult(true, nil, nil, ev.FreeValidatorName),
		},
		{
			name: "email is free",
			fields: fields{
				f: mockContains{t: t, want: validEmail.Domain(), ret: true},
			},
			args: args{
				email: validEmail,
			},
			want: ev.NewResult(false, utils.Errs(ev.FreeError{}), nil, ev.FreeValidatorName),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := ev.NewFreeValidator(tt.fields.f)
			if got := r.Validate(ev.NewInput(tt.args.email)); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFreeDefaultValidator(t *testing.T) {
	tests := []struct {
		name string
		want ev.Validator
	}{
		{
			name: "success",
			want: ev.NewFreeValidator(free.NewWillWhiteSetFree()),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ev.FreeDefaultValidator(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FreeDefaultValidator() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFreeError_Error(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			want: ev.FreeErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fr := ev.FreeError{}
			if got := fr.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
