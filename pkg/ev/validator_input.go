package ev

import (
	"github.com/prodadidb/go-email-validator/pkg/ev/evmail"
)

// ValidatorName is type to represent validator name
type ValidatorName string

func (v ValidatorName) String() string {
	return string(v)
}

// Input consists of input data for Validator.Validate
type Input interface {
	Email() evmail.Address
	Option(name ValidatorName) interface{}
}

// NewInput create Input from evmail.Address and KVOption list
func NewInput(email evmail.Address, kvOptions ...KVOption) Input {
	var options = make(map[ValidatorName]interface{})

	for _, kvOption := range kvOptions {
		options[kvOption.Name] = kvOption.Option
	}

	return NewInputFromMap(email, options)
}

// NewInputFromMap create Input from evmail.Address and options
func NewInputFromMap(email evmail.Address, options map[ValidatorName]interface{}) Input {
	return &InputStruct{
		EmailAddress: email,
		Options:      options,
	}
}

type InputStruct struct {
	EmailAddress evmail.Address
	Options      map[ValidatorName]interface{}
}

func (i *InputStruct) Email() evmail.Address {
	return i.EmailAddress
}

func (i *InputStruct) Option(name ValidatorName) interface{} {
	return i.Options[name]
}

// NewKVOption instantiates KVOption
func NewKVOption(name ValidatorName, option interface{}) KVOption {
	return KVOption{
		Name:   name,
		Option: option,
	}
}

// KVOption needs to form options in Input
type KVOption struct {
	Name   ValidatorName
	Option interface{}
}
