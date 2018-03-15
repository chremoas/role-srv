package handler

import (
	"errors"
	"fmt"
	discord "github.com/chremoas/discord-gateway/proto"
	rolesrv "github.com/chremoas/role-srv/proto"
	"github.com/chremoas/services-common/config"
	redis "github.com/chremoas/services-common/redis"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/client"
	"golang.org/x/net/context"
	//"regexp"
	"github.com/fatih/structs"
	"strings"
	"strconv"
)

type rolesHandler struct {
	Client client.Client
	Redis  *redis.Client
}

type clientList struct {
	discord discord.DiscordGatewayClient
}

var clients clientList

func NewRolesHandler(config *config.Configuration, service micro.Service) rolesrv.RolesHandler {
	c := service.Client()

	clients = clientList{
		discord: discord.NewDiscordGatewayClient(config.LookupService("gateway", "discord"), c),
	}

	addr := fmt.Sprintf("%s:%d", config.Redis.Host, config.Redis.Port)
	redisClient := redis.Init(addr, config.Redis.Password, config.Redis.Database, config.LookupService("srv", "perms"))

	_, err := redisClient.Client.Ping().Result()
	if err != nil {
		panic(err)
	}

	return &rolesHandler{Redis: redisClient}
}

func (h *rolesHandler) AddRole(ctx context.Context, request *rolesrv.Role, response *rolesrv.NilMessage) error {
	roleName := h.Redis.KeyName(fmt.Sprintf("role:%s", request.ShortName))

	// Type and Name are required so let's check for those
	if len(request.Type) == 0 {
		return errors.New("Type is required.")
	}

	if len(request.ShortName) == 0 {
		return errors.New("ShortName is required.")
	}

	exists, err := h.Redis.Client.Exists(roleName).Result()

	if err != nil {
		return err
	}

	if exists == 1 {
		return fmt.Errorf("Role `%s` already exists.", request.Name)
	}

	_, err = h.Redis.Client.HMSet(roleName, structs.Map(request)).Result()

	if err != nil {
		return err
	}

	response = &rolesrv.NilMessage{}

	return nil
}

func (h *rolesHandler) UpdateRole(ctx context.Context, request *rolesrv.UpdateInfo, response *rolesrv.NilMessage) error {
	roleName := h.Redis.KeyName(fmt.Sprintf("role:%s", request.Name))

	exists, err := h.Redis.Client.Exists(roleName).Result()

	if err != nil {
		return err
	}

	if exists == 0 {
		return fmt.Errorf("Role `%s` doesn't exists.", request.Name)
	}

	if !validRoleKey(request.Key) {
		return fmt.Errorf("`%s` isn't a valid Role Key.", request.Key)
	}

	h.Redis.Client.HSet(roleName, request.Key, request.Value)

	return nil
}

func validRoleKey(a string) bool {
	list := []string{"Name", "Color", "Hoist", "Position", "Permissions", "Managed", "Mentionable"}
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func (h *rolesHandler) RemoveRole(ctx context.Context, request *rolesrv.Role, response *rolesrv.NilMessage) error {
	roleName := h.Redis.KeyName(fmt.Sprintf("role:%s", request.Name))

	exists, err := h.Redis.Client.Exists(roleName).Result()

	if err != nil {
		return err
	}

	if exists == 0 {
		return fmt.Errorf("Role `%s` doesn't exists.", request.Name)
	}

	_, err = h.Redis.Client.Del(roleName).Result()

	if err != nil {
		return err
	}

	response = &rolesrv.NilMessage{}
	return nil
}

func (h *rolesHandler) GetRoles(ctx context.Context, request *rolesrv.NilMessage, response *rolesrv.GetRolesResponse) error {
	roles, err := h.Redis.Client.Keys(h.Redis.KeyName("role:*")).Result()

	if err != nil {
		return err
	}

	for role := range roles {
		roleName := strings.Split(roles[role], ":")
		response.Roles = append(response.Roles, &rolesrv.Role{Name: roleName[len(roleName)-1]})
	}
	return nil
}

func (h *rolesHandler) GetRole(ctx context.Context, request *rolesrv.Role, response *rolesrv.Role) error {
	roleName := h.Redis.KeyName(fmt.Sprintf("role:%s", request.ShortName))

	role, err := h.Redis.Client.HGetAll(roleName).Result()

	if err != nil {
		return err
	}

	color, err := strconv.ParseInt(role["Color"], 10, 32)
	position, err := strconv.ParseInt(role["Position"], 10, 32)
	permissions, err := strconv.ParseInt(role["Permissions"], 10, 32)
	hoist, err := strconv.ParseBool(role["Hoist"])
	managed, err := strconv.ParseBool(role["Managed"])
	mentionable, err := strconv.ParseBool(role["Mentionable"])

	response.ShortName = request.ShortName
	response.Type = role["Type"]
	response.Name = role["Name"]
	response.Color = int32(color)
	response.Hoist = hoist
	response.Position = int32(position)
	response.Permissions = int32(permissions)
	response.Managed = managed
	response.Mentionable = mentionable

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
