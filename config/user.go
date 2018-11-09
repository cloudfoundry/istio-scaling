package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

type User struct {
	userName string
	password string
}

func NewUser(cfg Config) *User {
	return &User{
		userName: cfg.GetAdminUser(),
		password: cfg.GetAdminPassword(),
	}
}

func NewAdmin(cfg Config) *User {
	return &User{
		userName: cfg.GetAdminUser(),
		password: cfg.GetAdminPassword(),
	}
}

func (u *User) Create() {
	Expect(cf.Cf("create-user", u.userName, u.password).Wait(4 * defaultTimeout)).To(Exit(0))
}

func (u *User) Username() string { return u.userName }

func (u *User) Password() string { return u.password }

func (u *User) ShouldRemain() bool {
	return true
}

func (u *User) Destroy() {}

func generatePassword() string {
	const randomBytesLength = 16
	encoding := base64.RawURLEncoding

	randomBytes := make([]byte, encoding.DecodedLen(randomBytesLength))
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(fmt.Errorf("Could not generate random password: %s", err.Error()))
	}

	return "A0a!" + encoding.EncodeToString(randomBytes)
}
