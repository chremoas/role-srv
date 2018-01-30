package main

import (
	"fmt"
	uauthsvc "github.com/chremoas/auth-srv/proto"
	proto "github.com/chremoas/chremoas/proto"
	discord "github.com/chremoas/discord-gateway/proto"
	"github.com/chremoas/role-cmd/command"
	"github.com/chremoas/role-srv/handler"
	"github.com/chremoas/role-srv/proto"
	"github.com/chremoas/services-common/config"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/client"
)

var Version = "1.0.0"
var service micro.Service
var name = "role"

func main() {
	service := config.NewService(Version, "srv", name, initialize)

	chremoas_role.RegisterPermissionsHandler(service.Server(), handler.NewPermissionsHandler())
	chremoas_role.RegisterRolesHandler(service.Server(), handler.NewRolesHandler())

	if err := service.Run(); err != nil {
		fmt.Println(err)
	}
}

// This function is a callback from the config.NewService function.  Read those docs
func initialize(config *config.Configuration) error {
	clientFactory := clientFactory{
		authSrv:        config.LookupService("srv", "auth"),
		discordGateway: config.LookupService("gateway", "discord"),
		client:         service.Client()}

	proto.RegisterCommandHandler(service.Server(),
		command.NewCommand(name,
			&clientFactory,
		),
	)

	return nil
}

type clientFactory struct {
	authSrv        string
	discordGateway string
	client         client.Client
}

func (c clientFactory) NewClient() uauthsvc.UserAuthenticationClient {
	return uauthsvc.NewUserAuthenticationClient(c.authSrv, c.client)
}

func (c clientFactory) NewAdminClient() uauthsvc.UserAuthenticationAdminClient {
	return uauthsvc.NewUserAuthenticationAdminClient(c.authSrv, c.client)
}

func (c clientFactory) NewEntityQueryClient() uauthsvc.EntityQueryClient {
	return uauthsvc.NewEntityQueryClient(c.authSrv, c.client)
}

func (c clientFactory) NewEntityAdminClient() uauthsvc.EntityAdminClient {
	return uauthsvc.NewEntityAdminClient(c.authSrv, c.client)
}

func (c clientFactory) NewDiscordGatewayClient() discord.DiscordGatewayClient {
	return discord.NewDiscordGatewayClient(c.discordGateway, c.client)
}