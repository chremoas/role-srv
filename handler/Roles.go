package handler

import (
	uauthsvc "github.com/chremoas/auth-srv/proto"
	discord "github.com/chremoas/discord-gateway/proto"
	"github.com/chremoas/role-srv/proto"
	"github.com/micro/go-micro/client"
	"golang.org/x/net/context"
	"errors"
	"fmt"
	"regexp"
)

type ClientFactory interface {
	NewClient() uauthsvc.UserAuthenticationClient
	NewAdminClient() uauthsvc.UserAuthenticationAdminClient
	NewEntityQueryClient() uauthsvc.EntityQueryClient
	NewEntityAdminClient() uauthsvc.EntityAdminClient
	NewDiscordGatewayClient() discord.DiscordGatewayClient
}

var clientFactory ClientFactory

type rolesHandler struct {
	Client client.Client
}

func NewRolesHandler() chremoas_role.RolesHandler {
	return &rolesHandler{}
}

func (h *rolesHandler) AddRole(ctx context.Context, request *chremoas_role.AddRoleRequest, response *chremoas_role.AddRoleResponse) error {
	chremoasClient := clientFactory.NewEntityAdminClient()
	discordClient := clientFactory.NewDiscordGatewayClient()

	roleName := request.Role.Name
	chatServiceGroup := request.Role.RoleNick

	_, err := chremoasClient.RoleUpdate(ctx, &uauthsvc.RoleAdminRequest{
		Role:      &uauthsvc.Role{RoleName: roleName, ChatServiceGroup: chatServiceGroup},
		Operation: uauthsvc.EntityOperation_ADD_OR_UPDATE,
	})

	if err != nil {
		return err
	}

	_, err = discordClient.CreateRole(ctx, &discord.CreateRoleRequest{Name: chatServiceGroup})

	if err != nil {
		return err
	}

	return nil
}

func (h *rolesHandler) RemoveRole(ctx context.Context, request *chremoas_role.RemoveRoleRequest, response *chremoas_role.RemoveRoleResponse) error {
	var dRoleName string
	chremoasClient := clientFactory.NewEntityAdminClient()
	discordClient := clientFactory.NewDiscordGatewayClient()
	roleName := request.Name

	chremoasQueryClient := clientFactory.NewEntityQueryClient()
	chremoasRoles, err := chremoasQueryClient.GetRoles(ctx, &uauthsvc.EntityQueryRequest{})

	for cr := range chremoasRoles.List {
		if chremoasRoles.List[cr].RoleName == roleName {
			dRoleName = chremoasRoles.List[cr].ChatServiceGroup
		}
	}

	_, err = chremoasClient.RoleUpdate(ctx, &uauthsvc.RoleAdminRequest{
		Role:      &uauthsvc.Role{RoleName: roleName, ChatServiceGroup: "Doesn't matter"},
		Operation: uauthsvc.EntityOperation_REMOVE,
	})

	if err != nil {
		return err
	}

	_, err = discordClient.DeleteRole(ctx, &discord.DeleteRoleRequest{Name: dRoleName})

	if err != nil {
		return err
	}

	return nil
}

func (h *rolesHandler) GetRoles(ctx context.Context, request *chremoas_role.GetRolesRequest, response *chremoas_role.GetRolesResponse) error {
	chremoasClient := clientFactory.NewEntityQueryClient()
	output, err := chremoasClient.GetRoles(ctx, &uauthsvc.EntityQueryRequest{})

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

	//listDRoles(ctx, req)
	discordClient := clientFactory.NewDiscordGatewayClient()
	discordRoles, err := discordClient.GetAllRoles(ctx, &discord.GuildObjectRequest{})

	if err != nil {
		return err
	}

	//listRoles(ctx, req)
	chremoasClient := clientFactory.NewEntityQueryClient()
	chremoasRoles, err := chremoasClient.GetRoles(ctx, &uauthsvc.EntityQueryRequest{})

	if err != nil {
		return err
	}

	for dr := range discordRoles.Roles {
		chremoasClient := clientFactory.NewEntityAdminClient()

		_, err := chremoasClient.RoleUpdate(ctx, &uauthsvc.RoleAdminRequest{
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
		discordClient := clientFactory.NewDiscordGatewayClient()
		_, err := discordClient.CreateRole(ctx, &discord.CreateRoleRequest{Name: chremoasRoles.List[cr].ChatServiceGroup})

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
