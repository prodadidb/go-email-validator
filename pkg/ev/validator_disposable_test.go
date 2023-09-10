package ev_test

import (
	"reflect"
	"testing"

	"github.com/prodadidb/go-email-validator/pkg/ev"
	"github.com/prodadidb/go-email-validator/pkg/ev/contains"
	"github.com/prodadidb/go-email-validator/pkg/ev/evmail"
)

func TestDisposableValidator_Validate(t *testing.T) {
	type fields struct {
		d contains.InSet
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
			name: "valid",
			fields: fields{
				d: mockContains{t: t, want: validEmail.Domain(), ret: false},
			},
			args: args{email: validEmail},
			want: ev.NewResult(true, nil, nil, ev.DisposableValidatorName),
		},
		{
			name: "invalid",
			fields: fields{
				d: mockContains{t: t, want: validEmail.Domain(), ret: true},
			},
			args: args{email: validEmail},
			want: ev.NewResult(false, []error{ev.DisposableError{}}, nil, ev.DisposableValidatorName),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := ev.NewDisposableValidator(tt.fields.d)
			if got := d.Validate(ev.NewInput(tt.args.email), tt.args.in1...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDisposableError_Error(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			want: ev.DisposableErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			di := ev.DisposableError{}
			if got := di.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
