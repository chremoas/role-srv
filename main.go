package main

import (
	"fmt"
	"github.com/chremoas/services-common/config"
	"github.com/chremoas/role-srv/handler"
	"github.com/chremoas/role-srv/proto"
)

var Version = "1.0.0"
var name = "role"

func main() {
	service := config.NewService(Version, "srv", name, config.NilInit)

	chremoas_role.RegisterPermissionsHandler(service.Server(), handler.NewPermissionsHandler())
	chremoas_role.RegisterRolesHandler(service.Server(), handler.NewRolesHandler())

	if err := service.Run(); err != nil {
		fmt.Println(err)
	}
}
