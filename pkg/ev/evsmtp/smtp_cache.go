package evsmtp

import (
	"context"
	"fmt"

	"github.com/prodadidb/go-email-validator/pkg/ev/evcache"
	"github.com/prodadidb/go-email-validator/pkg/ev/evmail"
	"github.com/prodadidb/go-email-validator/pkg/log"
	"go.uber.org/zap"
)

// ARandomRCPT is abstract realization of RandomRCPT
type ARandomRCPT struct {
	fn RandomRCPTFunc
}

// Call is calling of RandomRCPTFunc
func (a *ARandomRCPT) Call(sm SendMail, email evmail.Address) []error {
	return a.fn(sm, email)
}

func (a *ARandomRCPT) Set(fn RandomRCPTFunc) {
	a.fn = fn
}

func (a *ARandomRCPT) Get() RandomRCPTFunc {
	return a.fn
}

// RandomCacheKeyGetter is type of function to Get cache key
type RandomCacheKeyGetter func(email evmail.Address) interface{}

// DefaultRandomCacheKeyGetter generates of cache key for RandomRCPT
func DefaultRandomCacheKeyGetter(email evmail.Address) interface{} {
	return email.Domain()
}

// NewCheckerCacheRandomRCPT creates Checker with caching of RandomRCPT calling
func NewCheckerCacheRandomRCPT(checker CheckerWithRandomRCPT, cache evcache.Interface, getKey RandomCacheKeyGetter) Checker {
	if getKey == nil {
		getKey = DefaultRandomCacheKeyGetter
	}

	c := &CheckerCacheRandomRCPTStruct{
		CheckerWithRandomRCPT: checker,
		RandomRCPTOpt:         &ARandomRCPT{fn: checker.Get()},
		Cache:                 cache,
		GetKey:                getKey,
	}

	c.CheckerWithRandomRCPT.Set(c.RandomRCPT)

	return c
}

type CheckerCacheRandomRCPTStruct struct {
	CheckerWithRandomRCPT
	RandomRCPTOpt RandomRCPT
	Cache         evcache.Interface
	GetKey        RandomCacheKeyGetter
}

func (c CheckerCacheRandomRCPTStruct) RandomRCPT(sm SendMail, email evmail.Address) (errs []error) {
	key := c.GetKey(email)
	ctx := context.Background()
	resultInterface, err := c.Cache.Get(ctx, key)
	if err == nil && resultInterface != nil {
		errs = *resultInterface.(*[]error)
	} else {
		errs = c.RandomRCPTOpt.Call(sm, email)
		if err = c.Cache.Set(ctx, key, ErrorsToEVSMTPErrors(errs)); err != nil {
			log.Logger().Error(fmt.Sprintf("cache RandomRCPT: %s", err),
				zap.String("email", email.String()),
				zap.String("key", fmt.Sprint(key)),
			)
		}
	}

	return errs
}
