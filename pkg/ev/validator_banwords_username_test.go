package ev_test

import (
	"reflect"
	"testing"

	"github.com/prodadidb/go-email-validator/pkg/ev"
	"github.com/prodadidb/go-email-validator/pkg/ev/contains"
	"github.com/prodadidb/go-email-validator/pkg/ev/evmail"
	"github.com/prodadidb/go-email-validator/pkg/ev/utils"
)

func Test_banWordsUsernameValidator_Validate(t *testing.T) {
	type fields struct {
		d contains.InStrings
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
			name: "username is valid",
			fields: fields{
				d: mockInString{
					t:    t,
					want: validEmail.Username(),
					ret:  false,
				},
			},
			args: args{
				email: validEmail,
			},
			want: ev.NewResult(true, nil, nil, ev.BanWordsUsernameValidatorName),
		},
		{
			name: "username is banned",
			fields: fields{
				d: mockInString{
					t:    t,
					want: validEmail.Username(),
					ret:  true,
				},
			},
			args: args{
				email: validEmail,
			},
			want: ev.NewResult(false, utils.Errs(ev.BanWordsUsernameError{}), nil, ev.BanWordsUsernameValidatorName),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := ev.NewBanWordsUsername(tt.fields.d)
			if got := w.Validate(ev.NewInput(tt.args.email)); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBanWordsUsernameError_Error(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "success",
			want: ev.BanWordsUsernameErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ba := ev.BanWordsUsernameError{}
			if got := ba.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
