package converter

import (
	"github.com/prodadidb/go-email-validator/pkg/ev/evsmtp"
)

// MX2String converts ms records to string array
func MX2String(MXs evsmtp.MXs) []string {
	var result = make([]string, len(MXs))
	for i, mx := range MXs {
		result[i] = mx.Host
	}

	return result
}
