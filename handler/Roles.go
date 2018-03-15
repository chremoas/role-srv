package handler

import (
	//"errors"
	//"fmt"
	discord "github.com/chremoas/discord-gateway/proto"
	rolesrv "github.com/chremoas/role-srv/proto"
	"github.com/chremoas/services-common/config"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/client"
	"golang.org/x/net/context"
	//"regexp"
)

type rolesHandler struct {
	Client client.Client
}

type clientList struct {
	discord       discord.DiscordGatewayClient
}

var clients clientList

func NewRolesHandler(conf *config.Configuration, service micro.Service) rolesrv.RolesHandler {
	c := service.Client()

	clients = clientList{
		discord:       discord.NewDiscordGatewayClient(conf.LookupService("gateway", "discord"), c),
	}
	return &rolesHandler{}
}

func (h *rolesHandler) AddRole(ctx context.Context, request *rolesrv.Role, response *rolesrv.NilMessage) error {
	return nil
}

func (h *rolesHandler) UpdateRole(ctx context.Context, request *rolesrv.Role, response *rolesrv.NilMessage) error {
	return nil
}

func (h *rolesHandler) RemoveRole(ctx context.Context, request *rolesrv.RemoveRoleRequest, response *rolesrv.NilMessage) error {
	return nil
}

func (h *rolesHandler) GetRoles(ctx context.Context, request *rolesrv.NilMessage, response *rolesrv.GetRolesResponse) error {
	return nil
}

func (h *rolesHandler) SyncRoles(ctx context.Context, request *rolesrv.NilMessage, response *rolesrv.SyncRolesResponse) error {
	//var matchSpace = regexp.MustCompile(`\s`)
	//var matchDBError = regexp.MustCompile(`^Error 1062:`)
	//var matchDiscordError = regexp.MustCompile(`^The role '.*' already exists$`)
	//
	////listDRoles(ctx, req)
	//discordRoles, err := clients.discord.GetAllRoles(ctx, &discord.GuildObjectRequest{})
	//
	//if err != nil {
	//	return err
	//}
	//
	////listRoles(ctx, req)
	//chremoasRoles, err := clients.chremoasQuery.GetRoles(ctx, &uauthsvc.EntityQueryRequest{})
	//
	//if err != nil {
	//	return err
	//}
	//
	//for dr := range discordRoles.Roles {
	//	_, err := clients.chremoasAdmin.RoleUpdate(ctx, &uauthsvc.RoleAdminRequest{
	//		Role:      &uauthsvc.Role{ChatServiceGroup: discordRoles.Roles[dr].Name, RoleName: matchSpace.ReplaceAllString(discordRoles.Roles[dr].Name, "_")},
	//		Operation: uauthsvc.EntityOperation_ADD_OR_UPDATE,
	//	})
	//
	//	if err != nil {
	//		if !matchDBError.MatchString(err.Error()) {
	//			return err
	//		}
	//		fmt.Printf("dr err: %+v\n", err)
	//	} else {
	//		response.Roles = append(response.Roles, &rolesrv.SyncRoles{Source: "Discord", Destination: "Chremoas", Name: discordRoles.Roles[dr].Name})
	//	}
	//}
	//
	//for cr := range chremoasRoles.List {
	//	_, err := clients.discord.CreateRole(ctx, &discord.CreateRoleRequest{Name: chremoasRoles.List[cr].ChatServiceGroup})
	//
	//	if err != nil {
	//		if !matchDiscordError.MatchString(err.Error()) {
	//			return err
	//		}
	//		fmt.Printf("cr err: %+v\n", err)
	//	} else {
	//		response.Roles = append(response.Roles, &rolesrv.SyncRoles{Source: "Chremoas", Destination: "Discord", Name: chremoasRoles.List[cr].ChatServiceGroup})
	//	}
	//}

	return nil
}
