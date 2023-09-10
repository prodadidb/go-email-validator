package evsmtp_test

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"net/smtp"
	"net/textproto"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/prodadidb/go-email-validator/pkg/ev/evsmtp/smtpclient"
	"github.com/prodadidb/go-email-validator/pkg/ev/evtests"
	"github.com/stretchr/testify/require"
)

// Separator separate mock message of Server
const Separator = "\r\n"

// SuccessServer contents success smtp server response
var SuccessServer = []string{
	"220 hello world",
	"502 EH?",
	"250 mx.google.com at your service",
	"250 Sender ok",
	"550 address does not exist",
	"250 Receiver ok",
	"221 Goodbye",
}

// Server to testing SMTP
// Partial copy of TestSendMail  from smtp.TestSendMail
func Server(
	t testing.TB,
	server []string,
	timeout time.Duration,
	addr string,
	infinite bool,
) (string, chan string) {
	var cmdbuf bytes.Buffer
	bcmdbuf := bufio.NewWriter(&cmdbuf)

	if addr == "" {
		addr = "0.0.0.0:0"
	}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatalf("Unable to create listener: %v", err)
	}

	var done = make(chan string)
	closedMu := &sync.Mutex{}
	closed := false
	closeServer := func() {
		closedMu.Lock()
		if !closed {
			closed = true
			_ = bcmdbuf.Flush()
			done <- cmdbuf.String()
			close(done)
			_ = l.Close()
		}
		closedMu.Unlock()
	}
	go func(data []string) {
		defer closeServer()

		if len(data) == 0 {
			return
		}
		for {
			func() {
				conn, err := l.Accept()
				if err != nil {
					t.Errorf("Accept error: %v", err)
					return
				}
				defer func() {
					_ = conn.Close()
				}()

				tc := textproto.NewConn(conn)

				for i := 0; i < len(data) && data[i] != ""; i++ {
					_ = tc.PrintfLine(data[i])
					for len(data[i]) >= 4 && data[i][3] == '-' {
						i++
						_ = tc.PrintfLine(data[i])
					}
					if data[i] == "221 Goodbye" {
						return
					}
					read := false
					for !read || data[i] == "354 Go ahead" {
						msg, err := tc.ReadLine()
						_, _ = bcmdbuf.Write([]byte(msg + Separator))
						read = true
						if err != nil {
							t.Errorf("Read error: %v", err)
							return
						}
						if data[i] == "354 Go ahead" && msg == "." {
							break
						}
					}
				}
			}()

			if !infinite {
				break
			}
		}
	}(server)

	go func() {
		if timeout > 0 {
			time.Sleep(timeout)
			closeServer()
		}
	}()

	return l.Addr().String(), done
}

const (
	helloName = "helloName"

	smClient         = "Client"
	smHello          = "Hello "
	smHelloLocalhost = "Hello localhost"
	smAuth           = "Auth"
	smMail           = "Mail "
	smRCPTs          = "Rcpts "
	smData           = "Data"
	smWrite          = "Write"
	smQuit           = "Quit"
	smClose          = "Close"
	smWriteMessage   = "Write message"
	smWCloseWriter   = "Close writer"
)

var (
	testUser           = "testUser"
	testPwd            = "testPwd"
	testHost           = "testHost"
	testAuth           = smtp.PlainAuth("", testUser, testPwd, testHost)
	testMsg            = "msg"
	mockWriterInstance = &mockWriter{}
)

func stringsJoin(strs []string) string {
	return strings.Join(strs, ",")
}

// TODO create mock by gomock
type sendMailWant struct {
	stage   string
	message string
	sleep   time.Duration
	ret     interface{}
}

type mockSendMail struct {
	mu   sync.Mutex
	t    *testing.T
	i    int
	want []sendMailWant
}

func (s *mockSendMail) Client() smtpclient.SMTPClient {
	return s.do(smClient).(smtpclient.SMTPClient)
}

func (s *mockSendMail) Hello(localName string) error {
	return evtests.ToError(s.do(smHello + localName))
}

func (s *mockSendMail) Auth(a smtp.Auth) error {
	ret := s.do(smAuth).([]interface{})
	if !reflect.DeepEqual(a, ret[0]) {
		s.t.Errorf("Invalid auth, got %#v, want %#v", a, testAuth)
	}
	return evtests.ToError(ret[1])
}

func (s *mockSendMail) Mail(from string) error {
	return evtests.ToError(s.do(smMail + from))
}

func (s *mockSendMail) RCPTs(addr []string) map[string]error {
	err := s.do(smRCPTs + stringsJoin(addr))

	if err == nil {
		return nil
	}

	return map[string]error{
		addr[0]: evtests.ToError(err),
	}
}

func (s *mockSendMail) Data() (io.WriteCloser, error) {
	return &mockWriter{s: s, want: testMsg}, evtests.ToError(s.do(smData))
}

func (s *mockSendMail) Write(w io.WriteCloser, msg []byte) error {
	_, _ = w.Write(msg)
	_ = w.Close()

	return evtests.ToError(s.do(smWrite))
}

func (s *mockSendMail) Quit() error {
	return evtests.ToError(s.do(smQuit))
}

func (s *mockSendMail) Close() error {
	return evtests.ToError(s.do(smClose))
}

func (s *mockSendMail) do(cmd string) interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.i >= len(s.want) {
		s.t.Fatalf("Invalid command %q", cmd)
	}

	want := s.want[s.i]
	if cmd != want.message {
		s.t.Fatalf("Invalid command, got %q, want %q", cmd, want.message)
	}
	s.i++

	time.Sleep(want.sleep)

	return want.ret
}

type mockWriter struct {
	want string
	s    *mockSendMail
	buf  bytes.Buffer
}

func (w *mockWriter) Write(p []byte) (int, error) {
	if w.buf.Len() == 0 {
		w.s.do(smWriteMessage)
	}
	w.buf.Write(p)
	return len(p), nil
}

func (w *mockWriter) Close() error {
	require.Equal(w.s.t, w.buf.String(), w.want)
	w.s.do(smWCloseWriter)
	return nil
}

var defaultWantMap = map[string]sendMailWant{
	smHello: {
		message: smHelloLocalhost,
	},
	smAuth: {
		message: smAuth,
		ret:     []interface{}{nil, nil},
	},
	smMail: {
		message: smMail + emailFrom.String(),
		ret:     nil,
	},
	smRCPTs: {
		message: smRCPTs + randomAddress.String(),
		ret:     nil,
	},
}

var quitStageWant = sendMailWant{
	message: smQuit,
	ret:     nil,
}

var closeStageWant = sendMailWant{
	message: smClose,
	ret:     nil,
}

var wantSuccessList = []string{
	smHello,
	smAuth,
	smMail,
	smRCPTs, // only random email call rcpt
	smQuit,
}

func failWant(failStage *sendMailWant, withClose bool) []sendMailWant {
	wants := make([]sendMailWant, 0)
	for _, stage := range wantSuccessList {

		var want sendMailWant
		want, ok := defaultWantMap[stage]
		if !ok {
			want = sendMailWant{message: stage}
		}

		if failStage != nil && stage == failStage.stage {
			wants = append(wants, *failStage)
			break
		}
		wants = append(wants, want)
	}

	if withClose {
		wants = append(wants, closeStageWant)
	}

	return wants
}
