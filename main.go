package main

import (
	"fmt"
	"github.com/chremoas/role-srv/handler"
	rolesrv "github.com/chremoas/role-srv/proto"
	"github.com/chremoas/services-common/config"
	"github.com/micro/go-micro"
)

var Version = "1.0.0"
var service micro.Service
var name = "role"

func main() {
	service = config.NewService(Version, "srv", name, initialize)

	if err := service.Run(); err != nil {
		fmt.Println(err)
	}
}

func initialize(config *config.Configuration) error {
	rolesrv.RegisterRolesHandler(service.Server(), handler.NewRolesHandler(config, service))
	return nil
}
