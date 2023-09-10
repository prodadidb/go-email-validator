package ev

import (
	"context"
	"fmt"

	"github.com/prodadidb/go-email-validator/pkg/ev/evcache"
	"github.com/prodadidb/go-email-validator/pkg/ev/utils"
	"github.com/prodadidb/go-email-validator/pkg/log"
	"go.uber.org/zap"
)

// CacheKeyGetter is type for key generators
// To use complex keys you can use https://github.com/vmihailenco/msgpack/
type CacheKeyGetter func(input Input, results ...ValidationResult) interface{}

// EmailCacheKeyGetter generates key as full email
func EmailCacheKeyGetter(input Input, _ ...ValidationResult) interface{} {
	return input.Email().String()
}

// DomainCacheKeyGetter generates key as domain
func DomainCacheKeyGetter(input Input, _ ...ValidationResult) interface{} {
	return input.Email().Domain()
}

// NewCacheDecorator instantiates cache decorator
func NewCacheDecorator(validator Validator, cache evcache.Interface, getKey CacheKeyGetter) Validator {
	if getKey == nil {
		getKey = EmailCacheKeyGetter
	}

	return &CacheDecorator{
		Validator: validator,
		Cache:     cache,
		GetKey:    getKey,
	}
}

type CacheDecorator struct {
	Validator Validator
	Cache     evcache.Interface
	GetKey    CacheKeyGetter
}

func (c *CacheDecorator) GetDeps() []ValidatorName {
	return c.Validator.GetDeps()
}

func (c *CacheDecorator) Validate(input Input, results ...ValidationResult) (result ValidationResult) {
	key := c.GetKey(input, results...)
	ctx := context.Background()
	resultInterface, err := c.Cache.Get(ctx, key)
	if err == nil && resultInterface != nil {
		result = *resultInterface.(*ValidationResult)
	} else {
		result = c.Validator.Validate(input, results...)
		if err := c.Cache.Set(ctx, key, result); err != nil {
			log.Logger().Error(fmt.Sprintf("cache decorator %v", err),
				zap.String("validator", utils.StructName(c.Validator)),
				zap.String("key", fmt.Sprint(key)),
				zap.String("email", input.Email().String()),
				zap.String("results", fmt.Sprint(results)),
			)
		}
	}

	return result
}
