package evsmtp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"sync"

	"github.com/modern-go/reflect2"
	"github.com/prodadidb/go-email-validator/pkg/ev/evmail"
	"github.com/prodadidb/go-email-validator/pkg/ev/utils"
	"github.com/prodadidb/go-email-validator/pkg/log"
	"github.com/sethvargo/go-password/password"
	"github.com/tevino/abool"
	"go.uber.org/zap"
)

//go:generate mockgen -destination=./mock_smtp_test.go -package=evsmtp_test -source=smtp.go

// Configuration constants
const (
	ErrPrefix        = "evsmtp: "
	ErrConnectionMsg = ErrPrefix + "connection was not created"
	DefaultEmail     = "user@example.org"
	DefaultSMTPPort  = 25
	DefaultHelloName = "localhost"
)

// MXs is short alias for []*net.MX
type MXs = []*net.MX

// Constants of stages
const (
	RandomRCPTStage = CloseStage + 1
	ConnectionStage = RandomRCPTStage + 1
)

var (
	// ErrConnection is error of connection
	ErrConnection = NewError(ConnectionStage, errors.New(ErrConnectionMsg))
	// DefaultFromEmail is address, used as default From email
	DefaultFromEmail = evmail.FromString(DefaultEmail)
)

// Checker is SMTP validation
type Checker interface {
	Validate(mxs MXs, input Input) []error
}

// CheckerWithRandomRCPT is used for caching of RandomRCPT
type CheckerWithRandomRCPT interface {
	Checker
	RandomRCPT
}

// RandomRCPTFunc is function for checking of Catching All
type RandomRCPTFunc func(sm SendMail, email evmail.Address) (errs []error)

// RandomRCPT Need to realize of is-a relation (inheritance)
type RandomRCPT interface {
	Call(sm SendMail, email evmail.Address) []error
	Set(fn RandomRCPTFunc)
	Get() RandomRCPTFunc
}

// RandomEmail is function type to generate random email for checking of Catching All emails by RCPTs
type RandomEmail func(domain string) (evmail.Address, error)

func randomEmail(domain string) (evmail.Address, error) {
	gen, _ := password.NewGenerator(&password.GeneratorInput{
		LowerLetters: password.LowerLetters + password.Digits,
	})
	username, err := gen.Generate(15, 0, 0, true, true)
	return evmail.NewEmailAddress(username, domain), err
}

// CheckerDTO is DTO for NewChecker
type CheckerDTO struct {
	SendMailFactory SendMailDialerFactory
	RandomEmail     RandomEmail
	Options         Options
}

// NewChecker instantiates Checker
func NewChecker(dto CheckerDTO) Checker {
	if dto.SendMailFactory == nil {
		dto.SendMailFactory = NewSendMailFactory(DirectDial, nil)
	}

	if dto.RandomEmail == nil {
		dto.RandomEmail = randomEmail
	}

	if dto.Options == nil {
		dto.Options = DefaultOptions()
	}

	opts := OptionsDTO{
		EmailFrom:   evmail.EmptyEmail(dto.Options.EmailFrom(), DefaultFromEmail),
		HelloName:   utils.DefaultString(dto.Options.HelloName(), DefaultHelloName),
		Proxy:       dto.Options.Proxy(),
		TimeoutCon:  dto.Options.TimeoutConnection(),
		TimeoutResp: dto.Options.TimeoutResponse(),
		Port:        utils.DefaultInt(dto.Options.Port(), DefaultSMTPPort),
	}

	c := CheckerStruct{
		SendMailFactory: dto.SendMailFactory,
		Auth:            nil,
		RandomEmail:     dto.RandomEmail,
		Options:         NewOptions(opts),
	}
	c.RandomRCPT = &ARandomRCPT{fn: c.randomRCPT}

	return c
}

// CheckerStruct some SMTP server send additional message and we should read it 2.1.5 for OK message
type CheckerStruct struct {
	RandomRCPT
	SendMailFactory SendMailDialerFactory
	Auth            smtp.Auth
	RandomEmail     RandomEmail
	Options         Options
}

type sendMailRWMutex struct {
	m        sync.RWMutex
	sendMail SendMail
}

func (s *sendMailRWMutex) Set(sendMail SendMail) *sendMailRWMutex {
	s.m.Lock()
	defer s.m.Unlock()

	s.sendMail = sendMail

	return s
}

func (s *sendMailRWMutex) Get() SendMail {
	s.m.RLock()
	defer s.m.RUnlock()
	return s.sendMail
}

func (c CheckerStruct) Validate(mxs MXs, input Input) (errs []error) {
	var smMutex = &sendMailRWMutex{}
	var err error
	errs = make([]error, 0)
	var host string

	email := input.Email()
	opts := NewOptions(OptionsDTO{
		EmailFrom:   evmail.EmptyEmail(input.EmailFrom(), c.Options.EmailFrom()),
		HelloName:   utils.DefaultString(input.HelloName(), c.Options.HelloName()),
		Proxy:       utils.DefaultString(input.Proxy(), c.Options.Proxy()),
		TimeoutCon:  utils.DefaultDuration(input.TimeoutConnection(), c.Options.TimeoutConnection()),
		TimeoutResp: utils.DefaultDuration(input.TimeoutResponse(), c.Options.TimeoutResponse()),
		Port:        utils.DefaultInt(input.Port(), c.Options.Port()),
	})

	for _, mx := range mxs {
		host = fmt.Sprintf("%v:%v", mx.Host, opts.Port())

		func() {
			var cancel context.CancelFunc
			var ctx context.Context
			ctx = context.Background()
			if opts.TimeoutConnection() > 0 {
				// TODO think about logging of timeout connection error
				ctx, cancel = context.WithTimeout(ctx, opts.TimeoutConnection())
				defer cancel()
			}

			done := make(chan struct{}, 1)
			go func() {
				defer close(done)
				var errSM error

				sendMail, errSM := c.SendMailFactory(ctx, host, input)
				if errSM == nil {
					smMutex.Set(sendMail)
				}
			}()

			select {
			case <-ctx.Done():
				return
			case <-done:
				return
			}
		}()

		if reflect2.IsNil(smMutex.Get()) {
			break
		}
	}

	stage := SafeSendMailStage{SendMailStage: ConnectionStage}
	sm := smMutex.Get()
	if reflect2.IsNil(sm) {
		return append(errs, ErrConnection)
	}

	needClose := abool.NewBool(true)
	defer func() {
		if needClose.IsNotSet() {
			return
		}
		if err := sm.Close(); err != nil {
			log.Logger().Error(fmt.Sprintf("SendMailStruct.Close %v", err),
				zap.String("email", email.String()),
				zap.String("mxs", fmt.Sprint(mxs)),
			)
		}
	}()

	done := make(chan struct{}, 1)
	isDone := abool.New()
	errAppend := func(elems ...error) bool {
		if isDone.IsNotSet() {
			errs = append(errs, elems...)
		}
		return isDone.IsSet()
	}

	timeoutResponse := utils.DefaultDuration(input.TimeoutResponse(), c.Options.TimeoutResponse())
	ctx := context.Background()
	if timeoutResponse > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeoutResponse)
		defer cancel()
	}

	go func() {
		defer close(done)

		stage.Set(HelloStage)
		if err = sm.Hello(opts.HelloName()); err != nil {
			errAppend(NewError(stage.Get(), err))
			return
		}

		stage.Set(AuthStage)
		if err = sm.Auth(c.Auth); err != nil {
			errAppend(NewError(stage.Get(), err))
			return
		}

		stage.Set(MailStage)
		if err = sm.Mail(opts.EmailFrom().String()); err != nil {
			errAppend(NewError(stage.Get(), err))
			return
		}

		stage.Set(RandomRCPTStage)
		if errsRandomRCPTs := c.RandomRCPT.Call(sm, email); len(errsRandomRCPTs) > 0 {
			if errAppend(errsRandomRCPTs...) {
				return
			}
			stage.Set(RCPTsStage)
			if errsRCPTs := sm.RCPTs([]string{email.String()}); len(errsRCPTs) > 0 {
				errAppend(NewError(stage.Get(), errsRCPTs[email.String()]))
			}
		}

		stage.Set(QuitStage)
		if err = sm.Quit(); err != nil {
			errAppend(NewError(stage.Get(), err))
		}
		needClose.UnSet()
	}()

	defer isDone.Set()
	select {
	case <-ctx.Done():
		errAppend(NewError(stage.Get(), ctx.Err()))
		return
	case <-done:
		return
	}
}

func (c CheckerStruct) randomRCPT(sm SendMail, email evmail.Address) (errs []error) {
	randomEmail, err := c.RandomEmail(email.Domain())
	if err != nil {
		randomEmailErr := NewError(RandomRCPTStage, err)
		log.Logger().Error(
			fmt.Sprintf("generate random email: %v", randomEmailErr),
			zap.String("email", email.String()),
		)
		return append(errs, randomEmailErr)
	}

	if errsRCPTs := sm.RCPTs([]string{randomEmail.String()}); len(errsRCPTs) > 0 {
		return append(errs, NewError(RandomRCPTStage, errsRCPTs[randomEmail.String()]))
	}

	return
}
