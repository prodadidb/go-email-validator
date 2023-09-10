package ev_test

import (
	"reflect"
	"testing"

	"github.com/prodadidb/go-email-validator/pkg/ev"
)

func TestDepBuilder_Build(t *testing.T) {
	type fields struct {
		validators ev.ValidatorMap
	}
	tests := []struct {
		name   string
		fields fields
		want   ev.Validator
	}{
		{
			name: "nil",
			fields: fields{
				validators: nil,
			},
			want: ev.NewDepValidator(ev.GetDefaultFactories()),
		},
		{
			name: "empty map",
			fields: fields{
				validators: ev.ValidatorMap{},
			},
			want: ev.NewDepValidator(ev.ValidatorMap{}),
		},
		{
			name: "map",
			fields: fields{
				validators: ev.ValidatorMap{
					mockValidatorName: newMockValidator(true),
				},
			},
			want: ev.NewDepValidator(ev.ValidatorMap{
				mockValidatorName: newMockValidator(true),
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := ev.NewDepBuilder(tt.fields.validators)
			if got := d.Build(); !reflect.DeepEqual(got, tt.want) {
				/*
					TODO find right way to compare struct with function.
					1. Use pointer for function
					2. Use InterfaceData()
				*/
				if tt.name != "nil" || len(got.(ev.DepValidator).Deps) != len(tt.want.(ev.DepValidator).Deps) {
					t.Errorf("Build() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestDepBuilder_Delete(t *testing.T) {
	type fields struct {
		validators ev.ValidatorMap
	}
	type args struct {
		names []ev.ValidatorName
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *ev.DepBuilder
	}{
		{
			name: "delete not exist element",
			fields: fields{
				validators: ev.ValidatorMap{},
			},
			args: args{
				names: []ev.ValidatorName{mockValidatorName, ev.SyntaxValidatorName},
			},
			want: &ev.DepBuilder{
				Validators: ev.ValidatorMap{},
			},
		},
		{
			name: "delete exist element",
			fields: fields{
				validators: ev.ValidatorMap{
					mockValidatorName:  newMockValidator(false),
					ev.MXValidatorName: newMockValidator(false)},
			},
			args: args{
				names: []ev.ValidatorName{mockValidatorName, ev.SyntaxValidatorName},
			},
			want: &ev.DepBuilder{
				Validators: ev.ValidatorMap{ev.MXValidatorName: newMockValidator(false)},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := ev.NewDepBuilder(tt.fields.validators)
			if got := d.Delete(tt.args.names...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Delete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDepBuilder_Set(t *testing.T) {
	type fields struct {
		validators ev.ValidatorMap
	}
	type args struct {
		name      ev.ValidatorName
		validator ev.Validator
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *ev.DepBuilder
	}{
		{
			name: "set",
			fields: fields{
				validators: ev.ValidatorMap{},
			},
			args: args{
				name:      mockValidatorName,
				validator: newMockValidator(false),
			},
			want: &ev.DepBuilder{
				Validators: ev.ValidatorMap{mockValidatorName: newMockValidator(false)},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &ev.DepBuilder{
				Validators: tt.fields.validators,
			}
			if got := d.Set(tt.args.name, tt.args.validator); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Set() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDepBuilder_Get(t *testing.T) {
	type fields struct {
		validators ev.ValidatorMap
	}
	type args struct {
		name ev.ValidatorName
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   ev.Validator
	}{
		{
			name: "has",
			fields: fields{
				validators: ev.ValidatorMap{mockValidatorName: newMockValidator(false)},
			},
			args: args{
				name: mockValidatorName,
			},
			want: newMockValidator(false),
		},
		{
			name: "has not",
			fields: fields{
				validators: ev.ValidatorMap{},
			},
			args: args{
				name: mockValidatorName,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &ev.DepBuilder{
				Validators: tt.fields.validators,
			}
			if got := d.Get(tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}
