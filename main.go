package main

import (
	"fmt"
	"github.com/chremoas/role-srv/handler"
	"github.com/chremoas/role-srv/proto"
	"github.com/chremoas/services-common/config"
	uauthsvc "github.com/chremoas/auth-srv/proto"
	"github.com/micro/go-micro/client"
	discord "github.com/chremoas/discord-gateway/proto"
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

type clientFactory struct {
	name   string
	client client.Client
}

func (c clientFactory) NewEntityQueryClient() uauthsvc.EntityQueryClient {
	return uauthsvc.NewEntityQueryClient(c.name, c.client)
}

func (c clientFactory) NewEntityAdminClient() uauthsvc.EntityAdminClient {
	return uauthsvc.NewEntityAdminClient(c.name, c.client)
}

func (c clientFactory) NewDiscordGatewayClient() discord.DiscordGatewayClient {
	return discord.NewDiscordGatewayClient(c.name, c.client)
}