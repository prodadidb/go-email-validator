package converter

import (
	"strings"

	"github.com/prodadidb/go-email-validator/pkg/ev/evmail"
)

// EmailsForTests returns emails for tests
func EmailsForTests() []string {
	return []string{
		"",
		"asdas.da2da",
		"asdas.dada",
		"asd@asd@asd",
		"zxczxczxc@joycasinoru",
		"sewag33689@gmail.com",
		"user99.doesnot.exist@gmail.com",
		"amazedfuckporno@gmail.com",
		"sewag33689@itymail.com",
		"derduzikne@nedoz.com",
		"tvzamhkdc@emlhub.com",
		"theofanis.giot2is@12pm.gr",
		"theofanisgiotis@12pm.gr",
		"asdasd@tradepro.net",
		"credit@mail.ru",
		"salestrade86@hotmail.com",
		"some.user.99@gmail.com",
		"monicaramirezrestrepo@hotmail.com",
		"admin@gmail.com",
		"name@yandex.ru",
		"admin@huntgear.ru",
		"go.email.validator@gmail.com",
		"radmal1982@yandex-team.ru",
		"pr@yandex-team.ru",
		"y-numata@senko.ed.jp",
		"alexey@life-in-travels.ru",
	}
}

// NewEmailAddress is a evmail.Address constructor
func NewEmailAddress(username, domain string, at *string) evmail.Address {
	return emailAddress{
		username: strings.ToLower(username),
		at:       at,
		domain:   strings.ToLower(domain),
	}
}

type emailAddress struct {
	username string
	at       *string
	domain   string
}

func (e emailAddress) Username() string {
	return e.username
}

func (e emailAddress) Domain() string {
	return e.domain
}

func (e emailAddress) String() string {
	if e.at == nil {
		return e.Username() + evmail.AT + e.Domain()
	}

	return e.Username() + *e.at + e.Domain()
}
