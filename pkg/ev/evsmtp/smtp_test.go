package evsmtp_test

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"net/textproto"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/allegro/bigcache"
	"github.com/prodadidb/go-email-validator/pkg/ev/evcache"
	"github.com/prodadidb/go-email-validator/pkg/ev/evmail"
	"github.com/prodadidb/go-email-validator/pkg/ev/evsmtp"
	"github.com/prodadidb/go-email-validator/pkg/ev/evsmtp/smtpclient"
	"github.com/prodadidb/go-email-validator/pkg/ev/evtests"
	"github.com/prodadidb/go-email-validator/pkg/ev/utils"
	"github.com/prodadidb/gocache/marshaler"
	"github.com/prodadidb/gocache/store"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -destination=./mock_smtp_test.go -package=evsmtp_test -source=smtp.go
//go:generate mockgen -destination=./mock_cache_test.go -package=evsmtp_test -source=../evcache/evcache.go

func TestMain(m *testing.M) {
	evtests.TestMain(m)
}

func dialFunc(t *testing.T, client smtpclient.SMTPClient, err error, wantCtx context.Context, wantAddr, wantProxy string, sleep time.Duration) evsmtp.DialFunc {
	return func(ctx context.Context, addr, proxy string) (smtpclient.SMTPClient, error) {
		require.Equal(t, utils.StructName(wantCtx), utils.StructName(ctx))
		require.Equal(t, addr, wantAddr)
		require.Equal(t, wantProxy, proxy)

		time.Sleep(sleep)

		return client, err
	}
}

var (
	localhost      = "127.0.0.1"
	smtpLocalhost  = localhost + ":25"
	errorSimple    = errors.New("errorSimple")
	errorRandom    = errors.New("errorRandom")
	mxs            = evsmtp.MXs{&net.MX{Host: localhost}}
	emptyLocalName = ""
	simpleClient   = &smtp.Client{}
	emailFromStr   = "email@from.com"
	emailFrom      = evmail.FromString(emailFromStr)
	emailToStr     = "email@to.com"
	emailTo        = evmail.FromString(emailToStr)
	randomAddress  = getRandomAddress(emailTo)
	validEmail     = getValidTestEmail()
	getMockKey     = func(t *testing.T, wantEmail evmail.Address, ret interface{}) func(email evmail.Address) interface{} {
		return func(email evmail.Address) interface{} {
			require.Equal(t, wantEmail, email)
			return ret
		}
	}
)

func getRandomAddress(email evmail.Address) evmail.Address {
	return evmail.FromString("random.which.did.not.exist@" + email.Domain())
}

func mockRandomEmail(t *testing.T, email evmail.Address, err error) evsmtp.RandomEmail {
	return func(domain string) (evmail.Address, error) {
		if domain != email.Domain() {
			t.Errorf("domain of random email is not equal")
		}

		return email, err
	}
}

func localIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func Test_checker_Validate(t *testing.T) {
	type fields struct {
		sendMailFactory evsmtp.SendMailDialerFactory
		randomEmail     evsmtp.RandomEmail
		options         evsmtp.Options
	}
	type args struct {
		mx    evsmtp.MXs
		email evmail.Address
	}

	successDialFunc := dialFunc(t, simpleClient, nil, context.Background(), smtpLocalhost, "", 0)

	ctxTimeout, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()

	errConnection := evsmtp.NewError(evsmtp.ConnectionStage, errors.New(evsmtp.ErrConnectionMsg))

	tests := []struct {
		name     string
		fields   fields
		args     args
		wantErrs []error
	}{
		{
			name:     "empty mx",
			args:     args{},
			wantErrs: utils.Errs(errConnection),
		},
		{
			name: "cannot connection to mx",
			fields: fields{
				sendMailFactory: evsmtp.NewSendMailFactory(dialFunc(t, nil, errorSimple, context.Background(), smtpLocalhost, "", 0), nil),
				options:         &evsmtp.OptionsStruct{},
			},
			args: args{
				mx: mxs,
			},
			wantErrs: utils.Errs(errConnection),
		},
		{
			name: "Bad hello with helloName",
			fields: fields{
				sendMailFactory: evsmtp.NewSendMailCustom(successDialFunc, nil,
					func(client smtpclient.SMTPClient, tlsConfig *tls.Config) evsmtp.SendMail {
						return &mockSendMail{
							t:    t,
							want: failWant(&sendMailWant{stage: smHello, message: smHello + helloName, ret: errorSimple}, true),
						}
					}),
				options: &evsmtp.OptionsStruct{
					HelloNameOption: helloName,
				},
			},
			args: args{
				mx: mxs,
			},
			wantErrs: utils.Errs(evsmtp.NewError(evsmtp.HelloStage, errorSimple)),
		},
		{
			name: "Bad auth",
			fields: fields{
				sendMailFactory: evsmtp.NewSendMailCustom(successDialFunc, nil,
					func(client smtpclient.SMTPClient, tlsConfig *tls.Config) evsmtp.SendMail {
						return &mockSendMail{
							t: t,
							want: failWant(&sendMailWant{
								stage:   smAuth,
								message: smAuth,
								ret:     []interface{}{nil, errorSimple},
							}, true),
						}
					}),
				options: &evsmtp.OptionsStruct{},
			},
			args: args{
				mx: mxs,
			},
			wantErrs: utils.Errs(evsmtp.NewError(evsmtp.AuthStage, errorSimple)),
		},
		{
			name: "Bad Mail stage",
			fields: fields{
				sendMailFactory: evsmtp.NewSendMailCustom(successDialFunc, nil,
					func(client smtpclient.SMTPClient, tlsConfig *tls.Config) evsmtp.SendMail {
						return &mockSendMail{
							t: t,
							want: failWant(&sendMailWant{
								stage:   smMail,
								message: smMail + emailFrom.String(),
								ret:     errorSimple,
							}, true),
						}
					}),
				options: &evsmtp.OptionsStruct{
					EmailFromOption: emailFrom,
				},
			},
			args: args{
				mx: mxs,
			},
			wantErrs: utils.Errs(evsmtp.NewError(evsmtp.MailStage, errorSimple)),
		},
		{
			name: "Problem with generation Random email",
			fields: fields{
				sendMailFactory: evsmtp.NewSendMailCustom(successDialFunc, nil,
					func(client smtpclient.SMTPClient, tlsConfig *tls.Config) evsmtp.SendMail {
						return &mockSendMail{
							t: t,
							want: append(failWant(&sendMailWant{
								stage:   smMail,
								message: smMail + emailFrom.String(),
								ret:     nil,
							}, false),
								sendMailWant{
									stage:   smRCPTs,
									message: smRCPTs + emailTo.String(),
									ret:     errorSimple,
								},
								quitStageWant,
								closeStageWant,
							),
						}
					}),
				randomEmail: mockRandomEmail(t, randomAddress, errorRandom),
				options: &evsmtp.OptionsStruct{
					EmailFromOption: emailFrom,
				},
			},
			args: args{
				mx:    mxs,
				email: emailTo,
			},
			wantErrs: utils.Errs(
				evsmtp.NewError(evsmtp.RandomRCPTStage, errorRandom),
				evsmtp.NewError(evsmtp.RCPTsStage, errorSimple),
			),
		},
		{
			name: "Problem with RCPTs Random email",
			fields: fields{
				sendMailFactory: evsmtp.NewSendMailCustom(successDialFunc, nil,
					func(client smtpclient.SMTPClient, tlsConfig *tls.Config) evsmtp.SendMail {
						return &mockSendMail{
							t: t,
							want: append(failWant(&sendMailWant{
								stage:   smRCPTs,
								message: smRCPTs + randomAddress.String(),
								ret:     errorSimple,
							}, false),
								sendMailWant{
									stage:   smRCPTs,
									message: smRCPTs + emailTo.String(),
									ret:     errorSimple,
								},
								quitStageWant,
								closeStageWant,
							),
						}
					}),
				randomEmail: mockRandomEmail(t, randomAddress, nil),
				options: &evsmtp.OptionsStruct{
					EmailFromOption: emailFrom,
				},
			},
			args: args{
				mx:    mxs,
				email: emailTo,
			},
			wantErrs: utils.Errs(
				evsmtp.NewError(evsmtp.RandomRCPTStage, errorSimple),
				evsmtp.NewError(evsmtp.RCPTsStage, errorSimple),
			),
		},
		{
			name: "Quit problem",
			fields: fields{
				sendMailFactory: evsmtp.NewSendMailCustom(successDialFunc, nil,
					func(client smtpclient.SMTPClient, tlsConfig *tls.Config) evsmtp.SendMail {
						return &mockSendMail{
							t: t,
							want: failWant(&sendMailWant{
								stage:   smQuit,
								message: smQuit,
								ret:     errorSimple,
							}, true),
						}
					}),
				randomEmail: mockRandomEmail(t, randomAddress, nil),
				options: &evsmtp.OptionsStruct{
					EmailFromOption: emailFrom,
				},
			},
			args: args{
				mx:    mxs,
				email: emailTo,
			},
			wantErrs: utils.Errs(evsmtp.NewError(evsmtp.QuitStage, errorSimple)),
		},
		{
			name: "Success",
			fields: fields{
				sendMailFactory: evsmtp.NewSendMailCustom(successDialFunc, nil,
					func(client smtpclient.SMTPClient, tlsConfig *tls.Config) evsmtp.SendMail {
						return &mockSendMail{
							t:    t,
							want: failWant(nil, true),
						}
					}),
				randomEmail: mockRandomEmail(t, randomAddress, nil),
				options: &evsmtp.OptionsStruct{
					EmailFromOption: emailFrom,
				},
			},
			args: args{
				mx:    mxs,
				email: emailTo,
			},
			wantErrs: []error{},
		},
		{
			name: "with timeout success",
			fields: fields{
				sendMailFactory: evsmtp.NewSendMailCustom(
					dialFunc(t, simpleClient, nil, ctxTimeout, smtpLocalhost, "", 0),
					nil,
					func(client smtpclient.SMTPClient, tlsConfig *tls.Config) evsmtp.SendMail {
						return &mockSendMail{
							t:    t,
							want: failWant(nil, true),
						}
					}),
				randomEmail: mockRandomEmail(t, randomAddress, nil),
				options: &evsmtp.OptionsStruct{
					EmailFromOption:   emailFrom,
					TimeoutConOption:  5 * time.Second,
					TimeoutRespOption: 5 * time.Second,
				},
			},
			args: args{
				mx:    mxs,
				email: emailTo,
			},
			wantErrs: []error{},
		},
		{
			name: "with expired connection timeout",
			fields: fields{
				sendMailFactory: evsmtp.NewSendMailCustom(
					dialFunc(t, simpleClient, nil, ctxTimeout, smtpLocalhost, "", 2*time.Millisecond),
					nil,
					func(client smtpclient.SMTPClient, tlsConfig *tls.Config) evsmtp.SendMail {
						return &mockSendMail{
							t:    t,
							want: failWant(nil, true),
						}
					}),
				randomEmail: mockRandomEmail(t, randomAddress, nil),
				options: &evsmtp.OptionsStruct{
					EmailFromOption:  emailFrom,
					TimeoutConOption: 1,
				},
			},
			args: args{
				mx:    mxs,
				email: emailTo,
			},
			wantErrs: utils.Errs(errConnection),
		},
		{
			name: "with expired response timeout",
			fields: fields{
				sendMailFactory: evsmtp.NewSendMailCustom(successDialFunc, nil,
					func(client smtpclient.SMTPClient, tlsConfig *tls.Config) evsmtp.SendMail {
						return &mockSendMail{
							t: t,
							want: []sendMailWant{
								{
									sleep:   2 * time.Millisecond,
									stage:   smHello,
									message: smHelloLocalhost,
									ret:     context.DeadlineExceeded,
								},
								closeStageWant,
							},
						}
					}),
				randomEmail: mockRandomEmail(t, randomAddress, nil),
				options: &evsmtp.OptionsStruct{
					EmailFromOption:   emailFrom,
					TimeoutRespOption: 1 * time.Millisecond,
				},
			},
			args: args{
				mx:    mxs,
				email: emailTo,
			},
			wantErrs: utils.Errs(evsmtp.NewError(evsmtp.HelloStage, context.DeadlineExceeded)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := evsmtp.NewChecker(evsmtp.CheckerDTO{
				SendMailFactory: tt.fields.sendMailFactory,
				RandomEmail:     tt.fields.randomEmail,
				Options:         tt.fields.options,
			})
			gotErrs := c.Validate(tt.args.mx, evsmtp.NewInput(tt.args.email, nil))
			if !reflect.DeepEqual(gotErrs, tt.wantErrs) {
				t.Errorf("Validate() = %v, want %v", gotErrs, tt.wantErrs)
			}
		})
	}
}

func TestChecker_Validate_WithProxy_Local(t *testing.T) {
	evtests.FunctionalSkip(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	successWantSMTP := []string{
		"EHLO helloName",
		"HELO helloName",
		"MAIL FROM:<user@example.org>",
		"RCPT TO:<random.which.did.not.exist@tradepro.net>",
		"RCPT TO:<asd@tradepro.net>",
		"QUIT",
		"",
	}

	type fields struct {
		SendMailFactory evsmtp.SendMailDialerFactory
		Auth            smtp.Auth
		RandomEmail     evsmtp.RandomEmail
		Server          []string
		OptionsDTO      evsmtp.OptionsDTO
	}
	type args struct {
		mxs   evsmtp.MXs
		email evmail.Address
	}

	emailString := "asd@tradepro.net"

	emailFrom := evmail.FromString(evsmtp.DefaultEmail)
	emailTest := evmail.FromString(emailString)

	tests := []struct {
		name     string
		fields   fields
		args     args
		wantErrs []error
		wantSMTP []string
	}{
		{
			name: "without proxy",
			fields: fields{
				SendMailFactory: evsmtp.NewSendMailFactory(evsmtp.DirectDial, nil),
				Auth:            nil,
				RandomEmail:     mockRandomEmail(t, getRandomAddress(emailTest), nil),
				Server:          SuccessServer,
				OptionsDTO: evsmtp.OptionsDTO{
					EmailFrom: emailFrom,
					HelloName: helloName,
				},
			},
			args: args{
				mxs:   mxs,
				email: emailTest,
			},
			wantErrs: []error{evsmtp.NewError(evsmtp.RandomRCPTStage, &textproto.Error{
				Code: 550,
				Msg:  "address does not exist",
			})},
			wantSMTP: successWantSMTP,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, done := Server(t, tt.fields.Server, time.Second, "", false)

			if tt.fields.OptionsDTO.Port == 0 {
				u, _ := url.Parse("http://" + addr)
				tt.fields.OptionsDTO.Port, _ = strconv.Atoi(u.Port())
			}

			c := evsmtp.CheckerStruct{
				SendMailFactory: tt.fields.SendMailFactory,
				Auth:            tt.fields.Auth,
				RandomEmail:     tt.fields.RandomEmail,
				Options:         evsmtp.NewOptions(tt.fields.OptionsDTO),
			}
			mockRandomRCPT := NewMockRandomRCPT(ctrl)
			mockRandomRCPT.EXPECT().Call(gomock.Any(), gomock.Any()).DoAndReturn(c.RandomRCPT).Times(1)
			//c.RandomRCPT = mockRandomRCPT
			//c.RandomRCPT = &ARandomRCPT{fn: c.randomRCPT}

			gotErrs := c.Validate(tt.args.mxs, evsmtp.NewInput(tt.args.email, nil))
			actualClient := <-done

			wantSMTP := strings.Join(tt.wantSMTP, Separator)
			if wantSMTP != actualClient {
				t.Errorf("Got:\n%s\nExpected:\n%s", actualClient, wantSMTP)
			}

			if !reflect.DeepEqual(gotErrs, tt.wantErrs) {
				t.Errorf("Validate() = %v, want %v", gotErrs, tt.wantErrs)
			}
		})
	}
}

func Test_checkerCacheRandomRCPT_RandomRCPT(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	type fields struct {
		checkerWithRandomRPCT func() evsmtp.CheckerWithRandomRCPT
		cache                 func() evcache.Interface
		getKey                evsmtp.RandomCacheKeyGetter
	}
	type args struct {
		email evmail.Address
	}

	errs := []error{errorSimple}
	errsAlias := []evsmtp.AliasError{errorSimple}
	emptyChecker := func() evsmtp.CheckerWithRandomRCPT {
		mock := NewMockCheckerWithRandomRCPT(ctrl)
		mock.EXPECT().Get().Return(nil).Times(1)
		mock.EXPECT().Set(gomock.Any()).Times(1)

		return mock
	}

	tests := []struct {
		name     string
		fields   fields
		args     args
		wantErrs []error
	}{
		{
			name: "with cache",
			fields: fields{
				checkerWithRandomRPCT: emptyChecker,
				cache: func() evcache.Interface {
					mock := NewMockInterface(ctrl)
					mock.EXPECT().Get(ctx, validEmail.Domain()).Return(&errs, nil).Times(1)

					return mock
				},
				getKey: getMockKey(t, validEmail, validEmail.Domain()),
			},
			args: args{
				email: validEmail,
			},
			wantErrs: errs,
		},
		{
			name: "without cache",
			fields: fields{
				checkerWithRandomRPCT: func() evsmtp.CheckerWithRandomRCPT {
					mock := NewMockCheckerWithRandomRCPT(ctrl)
					mock.EXPECT().Get().Return(mock.Call).Times(1)
					mock.EXPECT().Set(gomock.Any()).Times(1)
					mock.EXPECT().Call(gomock.Any(), validEmail).Return(errs).Times(1)

					return mock
				},
				cache: func() evcache.Interface {
					mock := NewMockInterface(ctrl)
					mock.EXPECT().Get(ctx, validEmail.Domain()).Return(nil, nil).Times(1)
					mock.EXPECT().Set(ctx, validEmail.Domain(), errsAlias).Return(nil).Times(1)

					return mock
				},
				getKey: getMockKey(t, validEmail, validEmail.Domain()),
			},
			args: args{
				email: validEmail,
			},
			wantErrs: errs,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := evsmtp.NewCheckerCacheRandomRCPT(tt.fields.checkerWithRandomRPCT(), tt.fields.cache(), tt.fields.getKey).(*evsmtp.CheckerCacheRandomRCPTStruct)
			if gotErrs := c.RandomRCPT(nil, tt.args.email); !reflect.DeepEqual(gotErrs, tt.wantErrs) {
				t.Errorf("RandomRCPT() = %v, want %v", gotErrs, tt.wantErrs)
			}
		})
	}
}

func TestDefaultRandomCacheKeyGetter(t *testing.T) {
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
				email: getValidTestEmail(),
			},
			want: getValidTestEmail().Domain(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := evsmtp.DefaultRandomCacheKeyGetter(tt.args.email); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DefaultRandomCacheKeyGetter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewCheckerCacheRandomRCPT(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		checker func() evsmtp.CheckerWithRandomRCPT
		cache   evcache.Interface
		getKey  evsmtp.RandomCacheKeyGetter
	}
	tests := []struct {
		name string
		args args
		want evsmtp.Checker
	}{
		{
			name: "fill empty",
			args: args{
				checker: func() evsmtp.CheckerWithRandomRCPT {
					mock := NewMockCheckerWithRandomRCPT(ctrl)
					mock.EXPECT().Get().Return(nil).Times(1)
					mock.EXPECT().Set(gomock.Any()).Times(1)

					return mock
				},
				cache:  nil,
				getKey: nil,
			},
			want: &evsmtp.CheckerCacheRandomRCPTStruct{
				GetKey:        evsmtp.DefaultRandomCacheKeyGetter,
				RandomRCPTOpt: &evsmtp.ARandomRCPT{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := evsmtp.NewCheckerCacheRandomRCPT(tt.args.checker(), tt.args.cache, tt.args.getKey)

			gotChecker := got.(*evsmtp.CheckerCacheRandomRCPTStruct)
			gotGetKey := gotChecker.GetKey
			gotChecker.GetKey = nil
			gotChecker.CheckerWithRandomRCPT = nil
			want := tt.want.(*evsmtp.CheckerCacheRandomRCPTStruct)
			wantGetKey := want.GetKey
			want.GetKey = nil

			if !reflect.DeepEqual(got, tt.want) || fmt.Sprint(gotGetKey) != fmt.Sprint(wantGetKey) {
				t.Errorf(
					"NewCheckerCacheRandomRCPT() = %v, want %v\n gotGetKey = %v, wantGetKey %v",
					got, tt.want, gotGetKey, wantGetKey)
			}
		})
	}
}

var cacheErrs = []error{
	evsmtp.NewError(1, &textproto.Error{Code: 505, Msg: "msg1"}),
	evsmtp.NewError(1, errors.New("msg2")),
}

func Test_Cache(t *testing.T) {
	ctx := context.Background()
	bigCacheClient, err := bigcache.NewBigCache(bigcache.DefaultConfig(5 * time.Minute))
	require.Nil(t, err)
	bigCacheStore := store.NewBigcache(bigCacheClient)

	marshal := marshaler.New(bigCacheStore)

	cache := evcache.NewCacheMarshaller(marshal, func() interface{} {
		return new([]error)
	})

	key := "key"

	err = cache.Set(ctx, key, evsmtp.ErrorsToEVSMTPErrors(cacheErrs))
	require.Nil(t, err)

	got, err := cache.Get(ctx, key)
	require.Nil(t, err)
	require.Equal(t, cacheErrs, *got.(*[]error))
}

func Test_checkerCacheRandomRCPT_RandomRCPT_RealCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	type fields struct {
		CheckerWithRandomRCPT func() evsmtp.CheckerWithRandomRCPT
		// nolint:unused
		randomRCPT evsmtp.RandomRCPT
		cache      func() evcache.Interface
	}
	type args struct {
		email evmail.Address
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantErrs []error
	}{
		{
			name: "with cache",
			fields: fields{
				CheckerWithRandomRCPT: func() evsmtp.CheckerWithRandomRCPT {
					mock := NewMockCheckerWithRandomRCPT(ctrl)
					mock.EXPECT().Get().Return(mock.Call).Times(1)
					mock.EXPECT().Set(gomock.Any()).Times(1)

					return mock
				},
				cache: func() evcache.Interface {
					bigCacheClient, err := bigcache.NewBigCache(bigcache.DefaultConfig(5 * time.Minute))
					require.Nil(t, err)
					bigCacheStore := store.NewBigcache(bigCacheClient)

					marshal := marshaler.New(bigCacheStore)

					// Add value to cache
					key := evsmtp.DefaultRandomCacheKeyGetter(validEmail)
					err = marshal.Set(ctx, key, evsmtp.ErrorsToEVSMTPErrors(cacheErrs))
					require.Nil(t, err)

					return evcache.NewCacheMarshaller(marshal, func() interface{} {
						return new([]error)
					}, nil)
				},
			},
			args: args{
				email: validEmail,
			},
			wantErrs: cacheErrs,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := evsmtp.NewCheckerCacheRandomRCPT(tt.fields.CheckerWithRandomRCPT(), tt.fields.cache(), evsmtp.DefaultRandomCacheKeyGetter).(*evsmtp.CheckerCacheRandomRCPTStruct)
			if gotErrs := c.RandomRCPT(nil, tt.args.email); !reflect.DeepEqual(gotErrs, tt.wantErrs) {
				t.Errorf("RandomRCPT() = %v, want %v", gotErrs, tt.wantErrs)
			}
		})
	}
}

// Test constants
const (
	ValidUsername = "go.email.validator"
	ValidDomain   = "gmail.com"
)

// GetValidTestEmail returns valid email.Address
func getValidTestEmail() evmail.Address {
	return evmail.NewEmailAddress(ValidUsername, ValidDomain)
}
