package handler

import (
	"errors"
	"fmt"
	uauthsvc "github.com/chremoas/auth-srv/proto"
	discord "github.com/chremoas/discord-gateway/proto"
	"github.com/chremoas/role-srv/proto"
	"github.com/chremoas/services-common/config"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/client"
	"golang.org/x/net/context"
	"regexp"
)

var service micro.Service

type ClientFactory interface {
	NewEntityQueryClient() uauthsvc.EntityQueryClient
	NewEntityAdminClient() uauthsvc.EntityAdminClient
	NewDiscordGatewayClient() discord.DiscordGatewayClient
}

//var clientFactory ClientFactory

type clients struct {
	chremoasQuery uauthsvc.EntityQueryClient
	chremoasAdmin uauthsvc.EntityAdminClient
	discord       discord.DiscordGatewayClient
}

func initClientConnections() (clients, error) {
	client := service.Client()

	clients := clients{
		chremoasQuery: uauthsvc.NewEntityQueryClient(config.LookupService("srv", "auth"), client),
		chremoasAdmin: uauthsvc.NewEntityAdminClient(config.LookupService("srv", "auth"), client),
		discord:       discord.NewDiscordGatewayClient(config.LookupService("gateway", "discord"), client),
	}
	// TODO: Catch errors or something
	return clients, nil
}

type rolesHandler struct {
	Client client.Client
}

func NewRolesHandler() chremoas_role.RolesHandler {
	return &rolesHandler{}
}

func (h *rolesHandler) AddRole(ctx context.Context, request *chremoas_role.AddRoleRequest, response *chremoas_role.AddRoleResponse) error {
	clients, err := initClientConnections()

	roleName := request.Role.Name
	chatServiceGroup := request.Role.RoleNick

	_, err = clients.chremoasAdmin.RoleUpdate(ctx, &uauthsvc.RoleAdminRequest{
		Role:      &uauthsvc.Role{RoleName: roleName, ChatServiceGroup: chatServiceGroup},
		Operation: uauthsvc.EntityOperation_ADD_OR_UPDATE,
	})

	if err != nil {
		return err
	}

	_, err = clients.discord.CreateRole(ctx, &discord.CreateRoleRequest{Name: chatServiceGroup})

	if err != nil {
		return err
	}

	return nil
}

func (h *rolesHandler) RemoveRole(ctx context.Context, request *chremoas_role.RemoveRoleRequest, response *chremoas_role.RemoveRoleResponse) error {
	var dRoleName string
	clients, err := initClientConnections()
	roleName := request.Name

	chremoasRoles, err := clients.chremoasQuery.GetRoles(ctx, &uauthsvc.EntityQueryRequest{})

	for cr := range chremoasRoles.List {
		if chremoasRoles.List[cr].RoleName == roleName {
			dRoleName = chremoasRoles.List[cr].ChatServiceGroup
		}
	}

	_, err = clients.chremoasAdmin.RoleUpdate(ctx, &uauthsvc.RoleAdminRequest{
		Role:      &uauthsvc.Role{RoleName: roleName, ChatServiceGroup: "Doesn't matter"},
		Operation: uauthsvc.EntityOperation_REMOVE,
	})

	if err != nil {
		return err
	}

	_, err = clients.discord.DeleteRole(ctx, &discord.DeleteRoleRequest{Name: dRoleName})

	if err != nil {
		return err
	}

	return nil
}

func (h *rolesHandler) GetRoles(ctx context.Context, request *chremoas_role.GetRolesRequest, response *chremoas_role.GetRolesResponse) error {
	clients, err := initClientConnections()
	output, err := clients.chremoasQuery.GetRoles(ctx, &uauthsvc.EntityQueryRequest{})

	if err != nil {
		return err
	}

	if output.String() == "" {
		return errors.New("There are no roles defined")
	}

	for role := range output.List {
		response.Roles = append(response.Roles, &chremoas_role.DiscordRole{Name: output.List[role].RoleName, RoleNick: output.List[role].ChatServiceGroup})
	}

	return nil
}

func (h *rolesHandler) SyncRoles(ctx context.Context, request *chremoas_role.SyncRolesRequest, response *chremoas_role.SyncRolesResponse) error {
	var matchSpace = regexp.MustCompile(`\s`)
	var matchDBError = regexp.MustCompile(`^Error 1062:`)
	var matchDiscordError = regexp.MustCompile(`^The role '.*' already exists$`)

	clients, err := initClientConnections()

	//listDRoles(ctx, req)
	discordRoles, err := clients.discord.GetAllRoles(ctx, &discord.GuildObjectRequest{})

	if err != nil {
		return err
	}

	//listRoles(ctx, req)
	chremoasRoles, err := clients.chremoasQuery.GetRoles(ctx, &uauthsvc.EntityQueryRequest{})

	if err != nil {
		return err
	}

	for dr := range discordRoles.Roles {
		_, err := clients.chremoasAdmin.RoleUpdate(ctx, &uauthsvc.RoleAdminRequest{
			Role:      &uauthsvc.Role{ChatServiceGroup: discordRoles.Roles[dr].Name, RoleName: matchSpace.ReplaceAllString(discordRoles.Roles[dr].Name, "_")},
			Operation: uauthsvc.EntityOperation_ADD_OR_UPDATE,
		})

		if err != nil {
			if !matchDBError.MatchString(err.Error()) {
				return err
			}
		}
		response.Roles = append(response.Roles, &chremoas_role.SyncRoles{Source: "Discord", Destination: "Chremoas", Name: discordRoles.Roles[dr].Name})
	}

	for cr := range chremoasRoles.List {
		_, err := clients.discord.CreateRole(ctx, &discord.CreateRoleRequest{Name: chremoasRoles.List[cr].ChatServiceGroup})

		if err != nil {
			if !matchDiscordError.MatchString(err.Error()) {
				return err
			}
		}
		response.Roles = append(response.Roles, &chremoas_role.SyncRoles{Source: "Chremoas", Destination: "Discord", Name: chremoasRoles.List[cr].ChatServiceGroup})
	}

	// Let's see what this looks like after it's run
	fmt.Printf("response.Roles = %d\n", len(response.Roles))

	//if buffer.Len() == 0 {
	//	buffer.WriteString("No roles needed to be synced")
	//}

	return nil
}
