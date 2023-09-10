package ev_test

import (
	"errors"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/prodadidb/go-email-validator/pkg/ev"
	"github.com/prodadidb/go-email-validator/pkg/ev/evmail"
	"github.com/prodadidb/go-email-validator/pkg/ev/evtests"
	"github.com/prodadidb/go-email-validator/pkg/ev/utils"
	"github.com/stretchr/testify/require"
)

// Test constants
const (
	ValidUsername    = "go.email.validator"
	ValidDomain      = "gmail.com"
	ValidEmailString = ValidUsername + "@" + ValidDomain
)

// GetValidTestEmail returns valid email.Address
func GetValidTestEmail() evmail.Address {
	return evmail.NewEmailAddress(ValidUsername, ValidDomain)
}

var (
	validEmail                        = evmail.FromString(ValidEmailString)
	invalidEmail                      = evmail.FromString("some%..@invalid.%.email")
	validMockValidator   ev.Validator = mockValidator{result: true}
	inValidMockValidator ev.Validator = mockValidator{result: false}
	errorSimple                       = errors.New("errorSimple")
	errorSimple2                      = errors.New("errorSimple2")
	validResult                       = ev.NewResult(true, nil, nil, ev.OtherValidator)
	invalidResult                     = ev.NewResult(false, utils.Errs(newMockError()), nil, ev.OtherValidator)
)

func sortErrors(errs []error) []error {
	sort.Slice(errs, func(l, r int) bool {
		return strings.Compare(errs[l].Error(), errs[r].Error()) >= 0
	})

	return errs
}

type mockContains struct {
	t    *testing.T
	want interface{}
	ret  bool
}

func (m mockContains) Contains(value interface{}) bool {
	require.Equal(m.t, value, m.want)

	return m.ret
}

type mockInString struct {
	t    *testing.T
	want interface{}
	ret  bool
}

func (m mockInString) Contains(value string) bool {
	require.Equal(m.t, value, m.want)

	return m.ret
}

func newMockError() error {
	return mockError{}
}

type mockError struct{}

func (mockError) Error() string {
	return "mockError"
}

const mockValidatorName ev.ValidatorName = "mockValidatorName"

func newMockValidator(result bool) mockValidator {
	return mockValidator{result: result}
}

type mockValidator struct {
	result bool
	deps   []ev.ValidatorName
}

func (m mockValidator) Validate(_ ev.Input, _ ...ev.ValidationResult) ev.ValidationResult {
	var err error
	if !m.result {
		err = newMockError()
	}

	return ev.NewResult(m.result, utils.Errs(err), nil, ev.OtherValidator)
}

func (m mockValidator) GetDeps() []ev.ValidatorName {
	return m.deps
}

type mockValidationResult struct {
	errs  []error
	warns []error
	name  ev.ValidatorName
}

func (m mockValidationResult) IsValid() bool {
	return m.HasErrors()
}

func (m mockValidationResult) Errors() []error {
	return m.errs
}

func (m mockValidationResult) HasErrors() bool {
	return reflect.ValueOf(m.Errors()).Len() > 0
}

func (m mockValidationResult) Warnings() []error {
	return m.warns
}

func (m mockValidationResult) HasWarnings() bool {
	return reflect.ValueOf(m.Warnings()).Len() > 0
}

func (m mockValidationResult) ValidatorName() ev.ValidatorName {
	return m.name
}

func TestMain(m *testing.M) {
	evtests.TestMain(m)
}

func TestMockValidator(t *testing.T) {
	cases := []struct {
		validator mockValidator
		expected  ev.ValidationResult
	}{
		{
			validator: newMockValidator(true),
			expected:  ev.NewResult(true, nil, nil, ev.OtherValidator),
		},
		{
			validator: newMockValidator(false),
			expected:  ev.NewResult(false, utils.Errs(newMockError()), nil, ev.OtherValidator),
		},
	}

	var emptyEmail evmail.Address
	for _, c := range cases {
		actual := c.validator.Validate(ev.NewInput(emptyEmail))
		require.Equal(t, c.expected, actual)
	}
}

func TestAValidatorWithoutDeps(t *testing.T) {
	validator := ev.AValidatorWithoutDeps{}

	require.Equal(t, ev.EmptyDeps, validator.GetDeps())
}

func TestNewValidatorResult(t *testing.T) {
	type args struct {
		isValid  bool
		errors   []error
		warnings []error
		name     ev.ValidatorName
	}
	tests := []struct {
		name string
		args args
		want ev.ValidationResult
	}{
		{
			name: "empty name",
			args: args{
				isValid:  true,
				errors:   nil,
				warnings: nil,
				name:     "",
			},
			want: &ev.ValidationResultStruct{IsValidVal: true, NameVal: ev.OtherValidator},
		},
		{
			name: "invalid with errors and warnings",
			args: args{
				isValid:  false,
				errors:   []error{errorSimple},
				warnings: []error{errorSimple},
				name:     mockValidatorName,
			},
			want: &ev.ValidationResultStruct{ErrorsVal: []error{errorSimple}, WarningsVal: []error{errorSimple}, NameVal: mockValidatorName},
		},
		{
			name: "invalid",
			args: args{
				isValid:  false,
				errors:   nil,
				warnings: nil,
				name:     mockValidatorName,
			},
			want: &ev.ValidationResultStruct{NameVal: mockValidatorName},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ev.NewResult(tt.args.isValid, tt.args.errors, tt.args.warnings, tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewResult() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidatorName_String(t *testing.T) {
	tests := []struct {
		name string
		v    ev.ValidatorName
		want string
	}{
		{
			name: "success",
			v:    mockValidatorName,
			want: string(mockValidatorName),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAValidationResult_HasWarnings(t *testing.T) {
	type fields struct {
		isValid  bool
		errors   []error
		warnings []error
		name     ev.ValidatorName
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "true",
			fields: fields{
				warnings: utils.Errs(errorSimple),
			},
			want: true,
		},
		{
			name: "false empty",
			fields: fields{
				warnings: []error{},
			},
			want: false,
		},
		{
			name: "false nil",
			fields: fields{
				warnings: nil,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &ev.AValidationResult{
				IsValidVal:  tt.fields.isValid,
				ErrorsVal:   tt.fields.errors,
				WarningsVal: tt.fields.warnings,
				NameVal:     tt.fields.name,
			}
			if got := a.HasWarnings(); got != tt.want {
				t.Errorf("HasWarnings() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAValidationResult_ValidatorName(t *testing.T) {
	type fields struct {
		isValid  bool
		errors   []error
		warnings []error
		name     ev.ValidatorName
	}
	tests := []struct {
		name   string
		fields fields
		want   ev.ValidatorName
	}{
		{
			name: "success",
			fields: fields{
				name: ev.OtherValidator,
			},
			want: ev.OtherValidator,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &ev.AValidationResult{
				IsValidVal:  tt.fields.isValid,
				ErrorsVal:   tt.fields.errors,
				WarningsVal: tt.fields.warnings,
				NameVal:     tt.fields.name,
			}
			if got := a.ValidatorName(); got != tt.want {
				t.Errorf("ValidatorName() = %v, want %v", got, tt.want)
			}
		})
	}
}
