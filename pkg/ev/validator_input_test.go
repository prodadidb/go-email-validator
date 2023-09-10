package ev_test

import (
	"reflect"
	"testing"

	"github.com/prodadidb/go-email-validator/pkg/ev"
	"github.com/prodadidb/go-email-validator/pkg/ev/evmail"
)

func TestNewInput(t *testing.T) {
	type args struct {
		email     evmail.Address
		kvOptions []ev.KVOption
	}

	kvOptions := []ev.KVOption{
		{Name: ev.OtherValidator, Option: 1},
		{Name: ev.OtherValidator, Option: 3},
		{Name: ev.SMTPValidatorName, Option: 2},
	}

	tests := []struct {
		name string
		args args
		want ev.Input
	}{
		{
			name: "success",
			args: args{
				email:     GetValidTestEmail(),
				kvOptions: kvOptions,
			},
			want: &ev.InputStruct{
				EmailAddress: GetValidTestEmail(),
				Options: map[ev.ValidatorName]interface{}{
					ev.OtherValidator:    3,
					ev.SMTPValidatorName: 2,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := make([]ev.KVOption, 0)

			for _, opt := range tt.args.kvOptions {
				opts = append(opts, ev.NewKVOption(opt.Name, opt.Option))
			}

			if got := ev.NewInput(tt.args.email, opts...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewInput() = %v, want %v", got, tt.want)
			}
		})
	}
}
