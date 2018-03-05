package main

import (
	"fmt"
	//uauthsvc "github.com/chremoas/auth-srv/proto"
	//discord "github.com/chremoas/discord-gateway/proto"
	"github.com/chremoas/role-srv/handler"
	"github.com/chremoas/role-srv/proto"
	"github.com/chremoas/services-common/config"
	"github.com/micro/go-micro"
	//"github.com/micro/go-micro/client"
)

var Version = "1.0.0"
var service micro.Service
var name = "role"
var conf *config.Configuration

func main() {
	service := config.NewService(Version, "srv", name, initialize)

	if err := service.Run(); err != nil {
		fmt.Println(err)
	}
}

func initialize(configuration *config.Configuration) error {
	chremoas_role.RegisterPermissionsHandler(service.Server(), handler.NewPermissionsHandler(configuration))
	chremoas_role.RegisterRolesHandler(service.Server(), handler.NewRolesHandler(configuration))
	return nil
}

//var ClientFactory = clientFactory{
//	name:   config.LookupService("srv", "auth"),
//	client: service.Client()}

//type clientFactory struct {
//	name   string
//	client client.Client
//}
//
//func (c clientFactory) NewEntityQueryClient() uauthsvc.EntityQueryClient {
//	return uauthsvc.NewEntityQueryClient(c.name, c.client)
//}
//
//func (c clientFactory) NewEntityAdminClient() uauthsvc.EntityAdminClient {
//	return uauthsvc.NewEntityAdminClient(c.name, c.client)
//}
//
//func (c clientFactory) NewDiscordGatewayClient() discord.DiscordGatewayClient {
//	return discord.NewDiscordGatewayClient(c.name, c.client)
//}
