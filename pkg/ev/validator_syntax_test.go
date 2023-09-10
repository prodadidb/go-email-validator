package ev_test

import (
	"reflect"
	"regexp"
	"testing"

	"github.com/prodadidb/go-email-validator/pkg/ev"
	"github.com/prodadidb/go-email-validator/pkg/ev/evmail"
)

func Test_syntaxValidator_Validate(t *testing.T) {
	type fields struct{}
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
			name:   "success",
			fields: fields{},
			args: args{
				email: validEmail,
			},
			want: ev.NewValidResult(ev.SyntaxValidatorName),
		},
		{
			name:   "invalid",
			fields: fields{},
			args: args{
				email: invalidEmail,
			},
			want: ev.SyntaxGetError(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sy := ev.NewSyntaxValidator()
			if got := sy.Validate(ev.NewInput(tt.args.email)); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_syntaxRegexValidator_Validate(t *testing.T) {
	invalidRegExp := regexp.MustCompile("^@$")

	type fields struct {
		AValidatorWithoutDeps ev.AValidatorWithoutDeps
		emailRegex            *regexp.Regexp
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
			name: "success with default regex",
			fields: fields{
				emailRegex: nil,
			},
			args: args{
				email: validEmail,
			},
			want: ev.NewValidResult(ev.SyntaxValidatorName),
		},
		{
			name: "invalid with default regex",
			fields: fields{
				emailRegex: ev.DefaultEmailRegex,
			},
			args: args{
				email: invalidEmail,
			},
			want: ev.SyntaxGetError(),
		},
		{
			name: "invalid with custom regex",
			fields: fields{
				emailRegex: invalidRegExp,
			},
			args: args{
				email: validEmail,
			},
			want: ev.SyntaxGetError(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := ev.NewSyntaxRegexValidator(tt.fields.emailRegex)
			if got := s.Validate(ev.NewInput(tt.args.email)); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSyntaxError_Error(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			want: ev.SyntaxErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sy := ev.SyntaxError{}
			if got := sy.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
