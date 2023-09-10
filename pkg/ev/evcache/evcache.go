package evcache

import (
	"context"

	"github.com/prodadidb/gocache/cache"
	"github.com/prodadidb/gocache/store"
)

//go:generate mockgen -destination=./mock_evcache_test.go -package=evcache_test -source=evcache.go

// Interface is Cache interface
type Interface interface {
	Get(ctx context.Context, key interface{}) (interface{}, error)
	Set(ctx context.Context, key, object interface{}) error
}

// NewCache instantiates adapter for cache.CacheInterface
func NewCache(cache cache.CacheInterface[any], option ...store.Option) Interface {
	return &gocacheAdapter{
		cache:  cache,
		option: option,
	}
}

type gocacheAdapter struct {
	cache  cache.CacheInterface[any]
	option []store.Option
}

func (c *gocacheAdapter) Get(ctx context.Context, key interface{}) (interface{}, error) {
	return c.cache.Get(ctx, key)
}

func (c *gocacheAdapter) Set(ctx context.Context, key, object interface{}) error {
	return c.cache.Set(ctx, key, object, c.option...)
}

// Marshaler is interface for marshaler.Marshaler
type Marshaler interface {
	Get(ctx context.Context, key interface{}, returnObj interface{}) (interface{}, error)
	Set(ctx context.Context, key, object interface{}, options ...store.Option) error
	Delete(ctx context.Context, key interface{}) error
	Invalidate(ctx context.Context, options ...store.InvalidateOption) error
	Clear(ctx context.Context) error
}

// MarshallerReturnObj generates returned object to fill by marshaler.Marshaler
type MarshallerReturnObj func() interface{}

// NewCacheMarshaller generates adapter for marshaler.Marshaler
func NewCacheMarshaller(marshaller Marshaler, returnObj MarshallerReturnObj, option ...store.Option) Interface {
	return &gocacheMarshallerAdapter{
		marshaller: marshaller,
		returnObj:  returnObj,
		option:     option,
	}
}

type gocacheMarshallerAdapter struct {
	marshaller Marshaler
	returnObj  MarshallerReturnObj
	option     []store.Option
}

func (c *gocacheMarshallerAdapter) Get(ctx context.Context, key interface{}) (interface{}, error) {
	return c.marshaller.Get(ctx, key, c.returnObj())
}

func (c *gocacheMarshallerAdapter) Set(ctx context.Context, key, object interface{}) error {
	return c.marshaller.Set(ctx, key, object, c.option...)
}
