package main

import (
	"flag"
	"fmt"

	"github.com/supergiant/supergiant/pkg/core"
	"github.com/supergiant/supergiant/pkg/model"
	"github.com/supergiant/supergiant/pkg/util"
)

func main() {
	c := new(core.Core)

	flag.StringVar(&c.ConfigFilePath, "config-file", "", "")
	flag.Parse()

	if err := c.InitializeForeground(); err != nil {
		panic(err)
	}

	password := util.RandomString(16)

	user := &model.User{
		Username: "admin",
		Password: password,
		Role:     model.UserRoleAdmin,
	}
	if err := c.Users.Create(user); err != nil {
		panic(err)
	}

	msg := fmt.Sprintf(`
  ==============================
  |     ADMIN USER CREATED     |
  ==============================
  |                            |
  | Username: admin            |
  | Password: %s |
  |                            |
  ==============================
  |           ( ͡° ͜ʖ ͡°)         |
  ==============================
  `, password)
	fmt.Println(msg)
}
