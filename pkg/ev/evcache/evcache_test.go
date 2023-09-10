package evcache_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/prodadidb/go-email-validator/pkg/ev/evcache"
	"github.com/prodadidb/go-email-validator/pkg/ev/evtests"
	"github.com/prodadidb/gocache/cache"
	"github.com/prodadidb/gocache/store"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -mock_names=CacheInterface[any]=MockCI -destination=./mock_gocache_test.go -package=evcache_test github.com/prodadidb/gocache/cache CacheInterface[any]

const key = "key"

var (
	simpleValue  = map[int]string{0: "123"}
	errorSimple  = errors.New("errorSimple")
	emptyOptions = make([]store.Option, 0)
)

func TestMain(m *testing.M) {
	evtests.TestMain(m)
}

func Test_gocacheAdapter_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	type fields struct {
		cache func() cache.CacheInterface[any]
	}
	type args struct {
		key interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "success get",
			fields: fields{
				cache: func() cache.CacheInterface[any] {
					cacheMock := NewMockCI(ctrl)
					cacheMock.EXPECT().Get(ctx, key).Return(simpleValue, nil).Times(1)

					return cacheMock
				},
			},
			args: args{
				key: key,
			},
			want:    simpleValue,
			wantErr: false,
		},
		{
			name: "error get",
			fields: fields{
				cache: func() cache.CacheInterface[any] {
					cacheMock := NewMockCI(ctrl)
					cacheMock.EXPECT().Get(ctx, key).Return(nil, errorSimple).Times(1)

					return cacheMock
				},
			},
			args: args{
				key: key,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := evcache.NewCache(tt.fields.cache(), nil)
			got, err := c.Get(ctx, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_gocacheAdapter_Set(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	type fields struct {
		cache  func() cache.CacheInterface[any]
		option []store.Option
	}
	type args struct {
		key    interface{}
		object interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				cache: func() cache.CacheInterface[any] {
					cacheMock := NewMockCI(ctrl)
					cacheMock.EXPECT().Set(ctx, key, simpleValue, emptyOptions).Return(nil).Times(1)

					return cacheMock
				},
				option: emptyOptions,
			},
			args: args{
				key:    key,
				object: simpleValue,
			},
			wantErr: false,
		},
		{
			name: "error",
			fields: fields{
				cache: func() cache.CacheInterface[any] {
					cacheMock := NewMockCI(ctrl)
					cacheMock.EXPECT().Set(ctx, key, simpleValue, emptyOptions).Return(errorSimple).Times(1)

					return cacheMock
				},
				option: emptyOptions,
			},
			args: args{
				key:    key,
				object: simpleValue,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := evcache.NewCache(tt.fields.cache(), tt.fields.option...)
			if err := c.Set(ctx, tt.args.key, tt.args.object); (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

var returnObj = func() interface{} {
	return make(map[int]string)
}

func Test_gocacheMarshallerAdapter_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	type fields struct {
		marshaller func() evcache.Marshaler
		returnObj  evcache.MarshallerReturnObj
		option     []store.Option
	}
	type args struct {
		key interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "success get",
			fields: fields{
				marshaller: func() evcache.Marshaler {
					cacheMock := NewMockMarshaler(ctrl)
					cacheMock.EXPECT().Get(ctx, key, returnObj()).Return(simpleValue, nil).Times(1)

					return cacheMock
				},
				returnObj: returnObj,
			},
			args: args{
				key: key,
			},
			want:    simpleValue,
			wantErr: false,
		},
		{
			name: "error get",
			fields: fields{
				marshaller: func() evcache.Marshaler {
					cacheMock := NewMockMarshaler(ctrl)
					cacheMock.EXPECT().Get(ctx, key, returnObj()).Return(nil, errorSimple).Times(1)

					return cacheMock
				},
				returnObj: returnObj,
			},
			args: args{
				key: key,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := evcache.NewCacheMarshaller(tt.fields.marshaller(), tt.fields.returnObj, tt.fields.option...)
			got, err := c.Get(ctx, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_gocacheMarshallerAdapter_Set(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	type fields struct {
		marshaller func() evcache.Marshaler
		returnObj  evcache.MarshallerReturnObj
		option     []store.Option
	}
	type args struct {
		key    interface{}
		object interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				marshaller: func() evcache.Marshaler {
					cacheMock := NewMockMarshaler(ctrl)
					cacheMock.EXPECT().Set(ctx, key, simpleValue, emptyOptions).Return(nil).Times(1)

					return cacheMock
				},
				option: emptyOptions,
			},
			args: args{
				key:    key,
				object: simpleValue,
			},
			wantErr: false,
		},
		{
			name: "error",
			fields: fields{
				marshaller: func() evcache.Marshaler {
					cacheMock := NewMockMarshaler(ctrl)
					cacheMock.EXPECT().Set(ctx, key, simpleValue, emptyOptions).Return(errorSimple).Times(1)

					return cacheMock
				},
				option: emptyOptions,
			},
			args: args{
				key:    key,
				object: simpleValue,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := evcache.NewCacheMarshaller(tt.fields.marshaller(), tt.fields.returnObj, tt.fields.option...)
			if err := c.Set(ctx, tt.args.key, tt.args.object); (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
