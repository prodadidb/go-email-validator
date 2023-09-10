package evsmtp

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net/smtp"
	"sync"

	"github.com/prodadidb/go-email-validator/pkg/ev/evsmtp/smtpclient"
)

// SendMailStage is stage type of SendMail
type SendMailStage uint8

// SafeSendMailStage is thread safe SendMailStage
type SafeSendMailStage struct {
	SendMailStage
	mu sync.RWMutex
}

// Set sets stage
func (s *SafeSendMailStage) Set(val SendMailStage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SendMailStage = val
}

// Get returns stage
func (s *SafeSendMailStage) Get() SendMailStage {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.SendMailStage
}

// Constants of stages
const (
	ClientStage SendMailStage = iota + 1
	HelloStage
	AuthStage
	MailStage
	RCPTsStage
	QuitStage
	CloseStage
)

// SendMail is interface of custom realization as smtp.SendMail
type SendMail interface {
	Client() smtpclient.SMTPClient
	Hello(helloName string) error
	Auth(a smtp.Auth) error
	Mail(from string) error
	RCPTs(addrs []string) map[string]error
	Data() (io.WriteCloser, error)
	Write(w io.WriteCloser, msg []byte) error
	Quit() error
	Close() error
}

var testHookStartTLS func(*tls.Config)

// SendMailDialerFactory is factory for SendMail with dialing
type SendMailDialerFactory func(ctx context.Context, host string, opts Options) (SendMail, error)

// NewSendMailFactory creates SendMailDialerFactory
func NewSendMailFactory(dialFunc DialFunc, tlsConfig *tls.Config) SendMailDialerFactory {
	return NewSendMailCustom(dialFunc, tlsConfig, NewSendMail)
}

// SendMailFactory is factory for SendMail
type SendMailFactory func(client smtpclient.SMTPClient, tlsConfig *tls.Config) SendMail

// NewSendMailCustom creates SendMailFactory with dialing and customization calling of SendMailFactory
func NewSendMailCustom(dialFunc DialFunc, tlsConfig *tls.Config, factory SendMailFactory) SendMailDialerFactory {
	return func(ctx context.Context, host string, opts Options) (SendMail, error) {
		conn, err := dialFunc(ctx, host, opts.Proxy())
		if err != nil {
			return nil, err
		}

		return factory(conn, tlsConfig), nil
	}
}

// NewSendMail instantiates SendMail
func NewSendMail(client smtpclient.SMTPClient, tlsConfig *tls.Config) SendMail {
	return &SendMailStruct{
		SMTPClient: client,
		TLSConfig:  tlsConfig,
	}
}

type SendMailStruct struct {
	SMTPClient smtpclient.SMTPClient
	TLSConfig  *tls.Config
}

func (s *SendMailStruct) Client() smtpclient.SMTPClient {
	return s.SMTPClient
}

func (s *SendMailStruct) Hello(helloName string) error {
	return s.SMTPClient.Hello(helloName)
}

func (s *SendMailStruct) Auth(a smtp.Auth) error {
	if ok, _ := s.SMTPClient.Extension("STARTTLS"); ok && s.TLSConfig != nil {
		if testHookStartTLS != nil {
			testHookStartTLS(s.TLSConfig)
		}
		if err := s.SMTPClient.StartTLS(s.TLSConfig); err != nil {
			return err
		}
	}

	if a != nil {
		if ok, _ := s.SMTPClient.Extension("AUTH"); !ok {
			return errors.New("smtp_checker: server doesn't support AUTH")
		}
		if err := s.SMTPClient.Auth(a); err != nil {
			return err
		}
	}
	return nil
}

func (s *SendMailStruct) Mail(from string) error {
	return s.SMTPClient.Mail(from)
}

func (s *SendMailStruct) RCPTs(addrs []string) map[string]error {
	errs := make(map[string]error)

	for _, addr := range addrs {
		if err := s.SMTPClient.Rcpt(addr); err != nil {
			errs[addr] = err
		}
	}

	return errs
}

func (s *SendMailStruct) Data() (io.WriteCloser, error) {
	return s.SMTPClient.Data()
}

func (s *SendMailStruct) Write(w io.WriteCloser, msg []byte) error {
	if _, err := w.Write(msg); err != nil {
		return err
	}

	return w.Close()
}

func (s *SendMailStruct) Quit() error {
	return s.SMTPClient.Quit()
}

func (s *SendMailStruct) Close() error {
	return s.SMTPClient.Close()
}
