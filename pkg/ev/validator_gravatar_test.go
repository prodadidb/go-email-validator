package ev_test

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"testing"

	"github.com/prodadidb/go-email-validator/pkg/ev"
	"github.com/prodadidb/go-email-validator/pkg/ev/evmail"
	"github.com/prodadidb/go-email-validator/pkg/ev/evtests"
)

const GravatarExistEmail = "beau@dentedreality.com.au"
const GravatarExistEmailURL = "https://www.gravatar.com/avatar/205e460b479e2e5b48aec07710c08d50?d=404"

// TODO mocking Gravatar
func Test_gravatarValidator_Validate(t *testing.T) {
	evtests.FunctionalSkip(t)

	type args struct {
		email   evmail.Address
		options []ev.KVOption
		results []ev.ValidationResult
	}

	tests := []struct {
		name string
		args args
		want ev.ValidationResult
	}{
		{
			name: "valid",
			args: args{
				email:   evmail.FromString(GravatarExistEmail),
				results: []ev.ValidationResult{ev.NewValidResult(ev.SyntaxValidatorName)},
			},
			want: ev.NewGravatarValidationResult(
				GravatarExistEmailURL,
				ev.NewValidResult(ev.GravatarValidatorName).(*ev.AValidationResult),
			),
		},
		{
			name: "invalid syntax",
			args: args{
				email:   evmail.FromString(""),
				results: []ev.ValidationResult{ev.SyntaxGetError()},
			},
			want: ev.GravatarGetError(ev.NewDepsError()),
		},
		{
			name: "invalid in gravatar",
			args: args{
				email:   evmail.FromString("some.none.exist@with.non.exist.domain"),
				results: []ev.ValidationResult{ev.NewValidResult(ev.SyntaxValidatorName)},
			},
			want: ev.GravatarGetError(ev.GravatarError{}),
		},
		{
			name: "expired timeout",
			args: args{
				email:   evmail.FromString("some.none.exist@with.non.exist.domain"),
				results: []ev.ValidationResult{ev.NewValidResult(ev.SyntaxValidatorName)},
				options: []ev.KVOption{ev.NewKVOption(
					ev.GravatarValidatorName,
					ev.NewGravatarOptions(ev.GravatarOptionsDTO{Timeout: 1}),
				)},
			},
			want: ev.GravatarGetError(&url.Error{
				Op:  "Head",
				URL: "https://www.gravatar.com/avatar/77996abfe12fc2141488a60b29aa4844?d=404",
				Err: errors.New("context deadline exceeded (Client.Timeout exceeded while awaiting headers)"),
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var errStr string
			var wantErrStr string

			w := ev.NewGravatarValidator()
			gotInterface := w.Validate(ev.NewInput(tt.args.email, tt.args.options...), tt.args.results...)

			got := gotInterface.(ev.GravatarValidationResult)
			want := tt.want.(ev.GravatarValidationResult)

			if len(got.Errors()) > 0 && len(want.Errors()) > 0 {
				if errOp, ok := got.Errors()[0].(*url.Error); ok && errOp.Err != nil {
					errStr = errOp.Err.Error()
					errOp.Err = nil
				}

				wantErrOp, ok := want.Errors()[0].(*url.Error)
				if ok && wantErrOp.Err != nil {
					wantErrStr = wantErrOp.Err.Error()
					wantErrOp.Err = nil
				}
			}

			if !reflect.DeepEqual(got, want) || got.URL() != want.URL() || errStr != wantErrStr {
				t.Errorf("Validate() = %v, want %v", gotInterface, tt.want)
			}
		})
	}
}

func Test_gravatarValidator_race_parallel(t *testing.T) {
	evtests.FunctionalSkip(t)

	w := ev.NewGravatarValidator()
	wantStr := "context deadline exceeded (Client.Timeout exceeded while awaiting headers)"
	for i := 0; i < 100; i++ {
		email := evmail.NewEmailAddress(fmt.Sprintf("someNoneExistUserName%d", i), "someNonExists")
		t.Run(email.String(), func(t *testing.T) {
			t.Parallel()

			gotInterface := w.Validate(
				ev.NewInput(email, ev.NewKVOption(
					ev.GravatarValidatorName,
					ev.NewGravatarOptions(ev.GravatarOptionsDTO{Timeout: 1}),
				)),
				ev.NewValidResult(ev.SyntaxValidatorName))

			got := gotInterface.(ev.GravatarValidationResult)
			gotStr := got.Errors()[0].(*url.Error).Err.Error()

			if gotStr != wantStr {
				t.Errorf("Validate() = %v, wantStr %v", gotStr, wantStr)
			}
		})
	}
}

func Test_gravatarValidator_GetDeps(t *testing.T) {
	tests := []struct {
		name string
		want []ev.ValidatorName
	}{
		{
			name: "success",
			want: []ev.ValidatorName{ev.SyntaxValidatorName},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := ev.NewGravatarValidator()
			if got := g.GetDeps(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDeps() = %v, want %v", got, tt.want)
			}
		})
	}
}
