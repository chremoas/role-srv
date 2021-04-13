package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/chremoas/services-common/config"
	"github.com/micro/go-micro"
	"go.uber.org/zap"

	chremoasPrometheus "github.com/chremoas/services-common/prometheus"

	"github.com/chremoas/role-srv/handler"
	rolesrv "github.com/chremoas/role-srv/proto"
)

var (
	Version = "SET ME YOU KNOB"
	service micro.Service
	logger  *zap.Logger
	name    = "role"
)

func main() {
	var err error

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	// TODO pick stuff up from the config
	logger, err = zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	logger.Info("Initialized logger")

	go chremoasPrometheus.PrometheusExporter(logger)

	service = config.NewService(Version, "srv", name, initialize)

	if err := service.Run(); err != nil {
		fmt.Println(err)
	}
}

func initialize(config *config.Configuration) error {
	h, err := handler.NewRolesHandler(config, service, logger)
	if err != nil {
		return err
	}

	rolesrv.RegisterRolesHandler(service.Server(), h)
	return nil
}
