package evsmtp

import (
	"time"

	"github.com/prodadidb/go-email-validator/pkg/ev/evmail"
)

const (
	// DefaultTimeoutConnection is timeout for connection
	DefaultTimeoutConnection = 5 * time.Second
	// DefaultTimeoutResponse is timeout for communication with smtp server
	DefaultTimeoutResponse = 5 * time.Second
)

// Input describes data for Checker
type Input interface {
	Email() evmail.Address
	Options
}

// Options describes smtp options
type Options interface {
	EmailFrom() evmail.Address
	HelloName() string
	Proxy() string
	TimeoutConnection() time.Duration
	TimeoutResponse() time.Duration
	Port() int
}

// NewInput instantiates Input
func NewInput(email evmail.Address, opts Options) Input {
	if opts == nil {
		opts = EmptyOptions()
	}

	return &InputStruct{
		EmailAddress: email,
		Options:      opts,
	}
}

type InputStruct struct {
	EmailAddress evmail.Address
	Options
}

func (i *InputStruct) Email() evmail.Address {
	return i.EmailAddress
}

// OptionsDTO is dto for NewOptions
type OptionsDTO struct {
	EmailFrom   evmail.Address
	HelloName   string
	Proxy       string
	TimeoutCon  time.Duration
	TimeoutResp time.Duration
	Port        int
}

var defaultOptions = NewOptions(OptionsDTO{
	TimeoutCon:  DefaultTimeoutConnection,
	TimeoutResp: DefaultTimeoutResponse,
})

// DefaultOptions returns options with default values
func DefaultOptions() Options {
	return defaultOptions
}

var emptyOptions = NewOptions(OptionsDTO{})

// EmptyOptions returns empty options to avoid rewriting of default values
func EmptyOptions() Options {
	return emptyOptions
}

// NewOptions instantiates Options
func NewOptions(dto OptionsDTO) Options {
	return &OptionsStruct{
		EmailFromOption:   dto.EmailFrom,
		HelloNameOption:   dto.HelloName,
		ProxyOption:       dto.Proxy,
		TimeoutConOption:  dto.TimeoutCon,
		TimeoutRespOption: dto.TimeoutResp,
		PortOption:        dto.Port,
	}
}

type OptionsStruct struct {
	EmailFromOption   evmail.Address
	HelloNameOption   string
	ProxyOption       string
	TimeoutConOption  time.Duration
	TimeoutRespOption time.Duration
	PortOption        int
}

func (i *OptionsStruct) EmailFrom() evmail.Address {
	return i.EmailFromOption
}
func (i *OptionsStruct) HelloName() string {
	return i.HelloNameOption
}
func (i *OptionsStruct) Proxy() string {
	return i.ProxyOption
}
func (i *OptionsStruct) TimeoutConnection() time.Duration {
	return i.TimeoutConOption
}
func (i *OptionsStruct) TimeoutResponse() time.Duration {
	return i.TimeoutRespOption
}
func (i *OptionsStruct) Port() int {
	return i.PortOption
}
