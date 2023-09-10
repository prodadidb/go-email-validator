package ev_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/prodadidb/go-email-validator/pkg/ev"
	"github.com/prodadidb/go-email-validator/pkg/ev/evmail"
	"github.com/prodadidb/go-email-validator/pkg/ev/evtests"
	"github.com/stretchr/testify/require"
)

type testSleep struct {
	sleep time.Duration
	mockValidator
	deps []ev.ValidatorName
}

func (t testSleep) GetDeps() []ev.ValidatorName {
	return t.deps
}

func (t testSleep) Validate(_ ev.Input, results ...ev.ValidationResult) ev.ValidationResult {
	time.Sleep(t.sleep)

	var isValid = true
	for _, result := range results {
		if !result.IsValid() {
			isValid = false
			break
		}
	}

	return ev.NewDepValidatorResult(isValid && t.result, nil)
}

func TestDepValidator_Validate_Independent(t *testing.T) {
	email := GetValidTestEmail()
	strings := ev.EmptyDeps

	depValidator := ev.NewDepValidator(
		map[ev.ValidatorName]ev.Validator{
			"test1": &testSleep{
				0,
				newMockValidator(true),
				strings,
			},
			"test2": &testSleep{
				0,
				newMockValidator(true),
				strings,
			},
			"test3": &testSleep{
				0,
				newMockValidator(false),
				strings,
			},
		},
	)

	v := depValidator.Validate(ev.NewInput(email))
	require.False(t, v.IsValid())
}

func TestDepValidator_Validate_Dependent(t *testing.T) {
	email := GetValidTestEmail()
	strings := ev.EmptyDeps

	depValidator := ev.NewDepValidator(map[ev.ValidatorName]ev.Validator{
		"test1": &testSleep{
			100 * time.Millisecond,
			newMockValidator(true),
			strings,
		},
		"test2": &testSleep{
			100 * time.Millisecond,
			newMockValidator(true),
			strings,
		},
		"test3": &testSleep{
			100 * time.Millisecond,
			newMockValidator(true),
			[]ev.ValidatorName{"test1", "test2"},
		},
	},
	)

	v := depValidator.Validate(ev.NewInput(email))
	require.True(t, v.IsValid())
}

func TestDepValidator_Validate_Full(t *testing.T) {
	evtests.FunctionalSkip(t)

	email := evmail.FromString(ValidEmailString)
	depValidator := ev.NewDepBuilder(nil).Build()

	v := depValidator.Validate(ev.NewInput(email))
	require.True(t, v.IsValid())
}

func Test_depValidationResult_Errors(t *testing.T) {
	type fields struct {
		isValid bool
		results ev.DepResult
	}
	tests := []struct {
		name   string
		fields fields
		want   []error
	}{
		{
			name: "with Errors",
			fields: fields{
				isValid: false,
				results: ev.DepResult{
					mockValidatorName:      mockValidationResult{errs: []error{errorSimple, errorSimple2}},
					ev.SyntaxValidatorName: mockValidationResult{errs: []error{errorSimple2, errorSimple}},
				},
			},
			want: []error{errorSimple, errorSimple2, errorSimple2, errorSimple},
		},
		{
			name: "without Errors",
			fields: fields{
				isValid: false,
				results: ev.DepResult{
					mockValidatorName:      mockValidationResult{},
					ev.SyntaxValidatorName: mockValidationResult{},
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := ev.NewDepValidatorResult(tt.fields.isValid, tt.fields.results)

			got := sortErrors(d.Errors())
			tt.want = sortErrors(tt.want)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Errors() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_depValidationResult_Warnings(t *testing.T) {
	type fields struct {
		isValid bool
		results ev.DepResult
	}
	tests := []struct {
		name         string
		fields       fields
		wantWarnings []error
	}{
		{
			name: "with Warnings",
			fields: fields{
				isValid: false,
				results: ev.DepResult{
					mockValidatorName:      mockValidationResult{warns: []error{errorSimple, errorSimple2}},
					ev.SyntaxValidatorName: mockValidationResult{warns: []error{errorSimple2, errorSimple}},
				},
			},
			wantWarnings: []error{errorSimple, errorSimple2, errorSimple2, errorSimple},
		},
		{
			name: "without Warnings",
			fields: fields{
				isValid: false,
				results: ev.DepResult{
					mockValidatorName:      mockValidationResult{},
					ev.SyntaxValidatorName: mockValidationResult{},
				},
			},
			wantWarnings: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := ev.NewDepValidatorResult(tt.fields.isValid, tt.fields.results)

			gotWarnings := sortErrors(d.Warnings())
			tt.wantWarnings = sortErrors(tt.wantWarnings)

			if !reflect.DeepEqual(gotWarnings, tt.wantWarnings) {
				t.Errorf("Warnings() = %v, want %v", gotWarnings, tt.wantWarnings)
			}
		})
	}
}
