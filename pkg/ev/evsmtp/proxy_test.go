package evsmtp_test

import (
	"context"
	"errors"
	"net"
	"net/smtp"
	"os"
	"reflect"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/prodadidb/go-email-validator/pkg/ev/evsmtp"
	"github.com/prodadidb/go-email-validator/pkg/ev/evsmtp/smtpclient"
	"github.com/prodadidb/go-email-validator/pkg/ev/evtests"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"h12.io/socks"
)

//go:generate mockgen -mock_names=Conn=MockConn -destination=./mock_net_conn_test.go -package=evsmtp_test net Conn

const (
	proxyURL = "socks5://username:password@127.0.0.1:1080"
)

var (
	errMissingPort    = &net.AddrError{Err: "missing port in address", Addr: localhost}
	ctxBackground     = context.Background()
	ctxBackgroundFunc = func() context.Context { return ctxBackground }
)

func TestDirectDial(t *testing.T) {
	type fields struct {
		server []string
	}
	type args struct {
		ctx      func() context.Context
		addr     string
		proxyURL string
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantClient bool
		wantErr    error
	}{
		{
			name: "success",
			fields: fields{
				server: []string{
					"220 hello world",
				},
			},
			args: args{
				ctx:      ctxBackgroundFunc,
				proxyURL: "",
			},
			wantClient: true,
			wantErr:    nil,
		},
		{
			name: "fail port",
			args: args{
				ctx:      ctxBackgroundFunc,
				addr:     localhost,
				proxyURL: "",
			},
			wantClient: false,
			wantErr:    &net.OpError{Op: "dial", Net: "tcp", Err: errMissingPort},
		},
		{
			name: "fail",
			args: args{
				ctx:      ctxBackgroundFunc,
				addr:     localhost + ":25",
				proxyURL: "",
			},
			wantClient: false,
			wantErr: &net.OpError{
				Op:  "dial",
				Net: "tcp",
				Addr: &net.TCPAddr{
					IP:   net.IPv4(127, 0, 0, 1),
					Port: 25,
					Zone: "",
				},
				Err: &os.SyscallError{
					Syscall: "connect",
					Err:     syscall.ECONNREFUSED,
				},
			},
		},
		{
			name: "expired timeout",
			args: args{
				ctx: func() context.Context {
					return ctxBackground
				},
				addr: localhost + ":25",
			},
			wantClient: false,
			wantErr: &net.OpError{
				Op:  "dial",
				Net: "tcp",
				Addr: &net.TCPAddr{
					IP:   net.IPv4(127, 0, 0, 1),
					Port: 25,
					Zone: "",
				},
				Err: errors.New("i/o timeout"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var done chan string
			addr := tt.args.addr
			if len(tt.fields.server) > 0 {
				addr, done = Server(t, tt.fields.server, time.Second, "", false)
			}

			gotClient, err := evsmtp.DirectDial(tt.args.ctx(), addr, tt.args.proxyURL)
			if len(tt.fields.server) > 0 {
				<-done
				if gotClient != nil {
					_ = gotClient.Quit()
				}
			}

			var errStr string
			if errOp, ok := err.(*net.OpError); ok && errOp.Err != nil {
				errStr = errOp.Err.Error()
				errOp.Err = nil
			}
			var wantErrStr string
			wantErrOp, ok := tt.wantErr.(*net.OpError)
			if ok && wantErrOp.Err != nil {
				wantErrStr = wantErrOp.Err.Error()
				wantErrOp.Err = nil
			}
			if !reflect.DeepEqual(err, tt.wantErr) && errStr != wantErrStr {
				t.Errorf("DirectDial() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (gotClient == nil) == tt.wantClient {
				t.Errorf("DirectDial() got = %v, want %v", gotClient, tt.wantClient)
			}
		})
	}
}

func TestH12IODial(t *testing.T) {
	evtests.FunctionalSkip(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	defer func() {
		evsmtp.SmtpNewClientVar = smtp.NewClient
		evsmtp.H12ioDialVar = socks.Dial
	}()

	var cancel context.CancelFunc
	var wg sync.WaitGroup

	type fields struct {
		server        []string
		dial          func(proxyURI string) func(string, string) (net.Conn, error)
		smtpNewClient func(conn net.Conn, host string) (*smtp.Client, error)
	}

	type args struct {
		ctx      func() context.Context
		addr     string
		proxyURL string
	}

	tests := []struct {
		name       string
		fields     fields
		args       args
		wantClient bool
		wantErr    error
	}{
		{
			name: "success",
			fields: fields{
				server: []string{
					"220 hello world",
				},
				smtpNewClient: smtp.NewClient,
			},
			args: args{
				ctx:      ctxBackgroundFunc,
				proxyURL: proxyURL,
			},
			wantClient: true,
			wantErr:    nil,
		},
		{
			name: "faild proxy connection",
			fields: fields{
				smtpNewClient: smtp.NewClient,
			},
			args: args{
				ctx:      ctxBackgroundFunc,
				proxyURL: "asd",
			},
			wantClient: false,
			wantErr:    errors.New("unknown SOCKS protocol "),
		},
		{
			name: "expired timeout in connection",
			fields: fields{
				smtpNewClient: smtp.NewClient,
			},
			args: args{
				ctx: func() context.Context {
					return ctxBackground
				},
				proxyURL: "asd",
			},
			wantClient: false,
			wantErr:    context.DeadlineExceeded,
		},
		{
			name: "expired timeout smtp connection",
			fields: fields{
				dial: func(proxyURI string) func(string, string) (net.Conn, error) {
					wg.Add(1)

					return func(s string, s2 string) (net.Conn, error) {
						mock := NewMockConn(ctrl)
						mock.EXPECT().Close().Do(func() {
							wg.Done()
						}).Times(1)

						return mock, nil
					}
				},
				smtpNewClient: func(conn net.Conn, host string) (*smtp.Client, error) {
					cancel()
					time.Sleep(1 * time.Millisecond)
					return &smtp.Client{}, nil
				},
			},
			args: args{
				ctx: func() context.Context {
					var ctx context.Context
					ctx, cancel = context.WithTimeout(ctxBackground, 1*time.Second)
					return ctx
				},
				addr:     localhost + ":25",
				proxyURL: proxyURL,
			},
			wantClient: false,
			wantErr:    context.Canceled,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evsmtp.SmtpNewClientVar = tt.fields.smtpNewClient
			if tt.fields.dial != nil {
				evsmtp.H12ioDialVar = tt.fields.dial
			}

			var done chan string
			addr := tt.args.addr
			if len(tt.fields.server) > 0 {
				addr, done = Server(t, tt.fields.server, 1*time.Second, "", false)
				addr = localIP() + addr[4:]
			}

			gotClient, err := evsmtp.H12IODial(tt.args.ctx(), addr, tt.args.proxyURL)
			if len(tt.fields.server) > 0 {
				<-done
				if gotClient != nil {
					_ = gotClient.Quit()
				}
			}

			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("H12IODial() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if (gotClient == nil) == tt.wantClient {
				t.Errorf("H12IODial() got = %v, want %v", gotClient, tt.wantClient)
			}

			wg.Wait()
		})
	}
}

func TestH12IODial_Direct(t *testing.T) {
	wantAddr := localhost
	wantProxyURL := ""
	var wantErr error = nil
	wantCtx := context.Background()
	evsmtp.DirectDialVar = func(ctx context.Context, addr, proxyURL string) (smtpclient.SMTPClient, error) {
		require.Equal(t, wantCtx, ctx)
		require.Equal(t, wantAddr, addr)
		require.Equal(t, wantProxyURL, proxyURL)

		return nil, nil
	}
	got, err := evsmtp.H12IODial(wantCtx, wantAddr, wantProxyURL)
	evsmtp.DirectDialVar = evsmtp.DirectDial

	if !reflect.DeepEqual(err, wantErr) {
		t.Errorf("H12IODial() error = %v, wantErr %v", err, wantErr)
		return
	}

	if got != nil {
		t.Errorf("H12IODial() should not be null")
	}
}
