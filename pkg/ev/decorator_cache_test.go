package ev_test

import (
	"context"
	"errors"
	"net/textproto"
	"reflect"
	"testing"
	"time"

	"github.com/allegro/bigcache"
	"github.com/prodadidb/go-email-validator/pkg/ev"
	"github.com/prodadidb/go-email-validator/pkg/ev/evcache"
	"github.com/prodadidb/go-email-validator/pkg/ev/evmail"
	"github.com/prodadidb/go-email-validator/pkg/ev/evsmtp"
	"github.com/prodadidb/gocache/marshaler"
	"github.com/prodadidb/gocache/store"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -destination=./mock_cache_test.go -package=ev_test -source=./evcache/evcache.go

func Test_cacheDecorator_Validate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	results := make([]ev.ValidationResult, 0)
	key := validEmail.String()

	type fields struct {
		validator ev.Validator
		cache     func() evcache.Interface
		getKey    ev.CacheKeyGetter
	}
	type args struct {
		email   evmail.Address
		results []ev.ValidationResult
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantResult ev.ValidationResult
	}{
		{
			name: "without cache, error and with set error",
			fields: fields{
				validator: inValidMockValidator,
				cache: func() evcache.Interface {
					cacheMock := NewMockInterface(ctrl)
					cacheMock.EXPECT().Get(ctx, key).Return(nil, nil).Times(1)
					cacheMock.EXPECT().Set(ctx, key, invalidResult).Return(errorSimple).Times(1)

					return cacheMock
				},
				getKey: ev.EmailCacheKeyGetter,
			},
			args: args{
				email:   validEmail,
				results: results,
			},
			wantResult: invalidResult,
		},
		{
			name: "without cache and with get error",
			fields: fields{
				validator: validMockValidator,
				cache: func() evcache.Interface {
					cacheMock := NewMockInterface(ctrl)
					cacheMock.EXPECT().Get(ctx, key).Return(nil, errorSimple).Times(1)
					cacheMock.EXPECT().Set(ctx, key, validResult).Return(nil).Times(1)

					return cacheMock
				},
				getKey: nil,
			},
			args: args{
				email:   validEmail,
				results: results,
			},
			wantResult: validResult,
		},
		{
			name: "with cache",
			fields: fields{
				validator: validMockValidator,
				cache: func() evcache.Interface {
					cacheMock := NewMockInterface(ctrl)
					cacheMock.EXPECT().Get(ctx, key).Return(&validResult, nil).Times(1)

					return cacheMock
				},
				getKey: ev.EmailCacheKeyGetter,
			},
			args: args{
				email:   validEmail,
				results: nil,
			},
			wantResult: validResult,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := ev.NewCacheDecorator(tt.fields.validator, tt.fields.cache(), tt.fields.getKey)
			if gotResult := c.Validate(ev.NewInput(tt.args.email), tt.args.results...); !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("Validate() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

func Test_cacheDecorator_GetDeps(t *testing.T) {
	deps := []ev.ValidatorName{ev.OtherValidator}

	type fields struct {
		validator ev.Validator
		cache     evcache.Interface
		getKey    ev.CacheKeyGetter
	}
	tests := []struct {
		name   string
		fields fields
		want   []ev.ValidatorName
	}{
		{
			name: "return deps",
			fields: fields{
				validator: mockValidator{deps: deps},
			},
			want: deps,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ev.CacheDecorator{
				Validator: tt.fields.validator,
				Cache:     tt.fields.cache,
				GetKey:    tt.fields.getKey,
			}
			if got := c.GetDeps(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDeps() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEmailCacheKeyGetter(t *testing.T) {
	type args struct {
		email   evmail.Address
		results []ev.ValidationResult
	}
	tests := []struct {
		name string
		args args
		want interface{}
	}{
		{
			name: "success",
			args: args{
				email:   validEmail,
				results: nil,
			},
			want: validEmail.String(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ev.EmailCacheKeyGetter(ev.NewInput(tt.args.email), tt.args.results...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EmailCacheKeyGetter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDomainCacheKeyGetter(t *testing.T) {
	type args struct {
		email evmail.Address
	}
	tests := []struct {
		name string
		args args
		want interface{}
	}{
		{
			name: "success",
			args: args{
				email: validEmail,
			},
			want: validEmail.Domain(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ev.DomainCacheKeyGetter(ev.NewInput(tt.args.email)); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DomainCacheKeyGetter() = %v, want %v", got, tt.want)
			}
		})
	}
}

var cacheErrs = []error{
	//error(&customErr{}), TODO find way to marshal and unmarshal all interfaces
	ev.NewDepsError(),
	evsmtp.NewError(1, &textproto.Error{Code: 505, Msg: "msg1"}),
	evsmtp.NewError(1, errors.New("msg2")),
}
var validatorResult = ev.NewResult(true, cacheErrs, cacheErrs, ev.OtherValidator)

func Test_Cache(t *testing.T) {
	bigCacheClient, err := bigcache.NewBigCache(bigcache.DefaultConfig(1 * time.Second))
	require.Nil(t, err)
	bigCacheStore := store.NewBigcache(bigCacheClient)

	ctx := context.Background()

	marshal := marshaler.New(bigCacheStore)

	cache := evcache.NewCacheMarshaller(marshal, func() interface{} {
		return new(ev.ValidationResult)
	})

	key := "key"

	err = cache.Set(ctx, key, validatorResult)
	require.Nil(t, err)

	gotInterface, err := cache.Get(ctx, key)
	require.Nil(t, err)
	got := *gotInterface.(*ev.ValidationResult)
	require.Equal(t, validatorResult, got)
}
