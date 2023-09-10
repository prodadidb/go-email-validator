package mailboxvalidator

import (
	"strings"

	"github.com/prodadidb/go-email-validator/pkg/ev/evmail"
	"github.com/prodadidb/go-email-validator/pkg/presentation/converter"
)

var emptyString = ""

// EmailFromString creates evmail.Address from string
func EmailFromString(email string) evmail.Address {
	pos := strings.LastIndexByte(email, '@')

	if pos == -1 || len(email) < 3 {
		return converter.NewEmailAddress("", email, &emptyString)
	}

	return converter.NewEmailAddress(email[:pos], email[pos+1:], nil)
}
