package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/emirpasic/gods/sets/hashset"
)

// Default configuration constants
const (
	RoleURL      = "https://raw.githubusercontent.com/mixmaxhq/role-based-email-addresses/master/index.js"
	RBEARolePath = "pkg/ev/role/rbea_roles.go"
)

var excludes = hashset.New(
	"asd",
	"asdasd",
	"asdf",
)

func rbeaRolesUpdate(url, path string) {
	rolesResp, err := http.Get(url)
	errPanic(err)
	defer func() {
		_ = rolesResp.Body.Close()
	}()

	var roles = make([]string, 0)
	rolesBytes, err := io.ReadAll(rolesResp.Body)
	errPanic(err)

	rolesBytes = bytes.ReplaceAll(rolesBytes[17:len(rolesBytes)-2], []byte{'\''}, []byte{'"'})
	err = json.Unmarshal(rolesBytes, &roles)
	errPanic(err)

	f, err := os.Create(path)
	errPanic(err)
	defer func() {
		_ = f.Close()
	}()

	_, _ = f.WriteString(generateRoleCode(roles))
}

func generateRoleCode(roles []string) string {
	strBuilder := strings.Builder{}
	strBuilder.WriteString(
		`package role

import (
	"github.com/emirpasic/gods/sets/hashset"
	"github.com/prodadidb/go-email-validator/pkg/ev/contains"
	"strings"
)

// RBEARoles returns the list of roles
func RBEARoles() []string {
	return rbeaRoles
}

// NewRBEASetRole forms contains.InSet from roles (https://github.com/mixmaxhq/role-based-email-addresses)
func NewRBEASetRole() contains.InSet {
	RBEARoles := RBEARoles()
	roles := make([]interface{}, len(RBEARoles))
	for i, role := range RBEARoles {
		roles[i] = strings.ToLower(role)
	}

	return contains.NewSet(hashset.New(roles...))
}
`)

	strBuilder.WriteString(
		`
var rbeaRoles = []string{
`)
	for _, role := range roles {
		if excludes.Contains(role) {
			continue
		}

		strBuilder.WriteString("\t\"")
		strBuilder.WriteString(role)
		strBuilder.WriteString("\",\n")
	}
	strBuilder.WriteString("}\n")

	return strBuilder.String()
}
