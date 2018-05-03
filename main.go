package main

import (
	"fmt"
	"github.com/chremoas/role-srv/handler"
	rolesrv "github.com/chremoas/role-srv/proto"
	"github.com/chremoas/services-common/config"
	"github.com/micro/go-micro"
	"go.uber.org/zap"
)

var Version = "SET ME YOU KNOB"
var service micro.Service
var logger *zap.Logger
var name = "role"

func main() {
	service = config.NewService(Version, "srv", name, initialize)
	var err error

	// TODO pick stuff up from the config
	logger, err = zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	logger.Info("Initialized logger")

	if err := service.Run(); err != nil {
		fmt.Println(err)
	}
}

func initialize(config *config.Configuration) error {
	rolesrv.RegisterRolesHandler(service.Server(), handler.NewRolesHandler(config, service, logger))
	return nil
}
