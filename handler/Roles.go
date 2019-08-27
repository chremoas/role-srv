package handler

import (
	"errors"
	"fmt"
	discord "github.com/chremoas/discord-gateway/proto"
	rolesrv "github.com/chremoas/role-srv/proto"
	common "github.com/chremoas/services-common/command"
	"github.com/chremoas/services-common/config"
	redis "github.com/chremoas/services-common/redis"
	"github.com/chremoas/services-common/sets"
	"github.com/fatih/structs"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/client"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type rolesHandler struct {
	Client client.Client
	Redis  *redis.Client
	Logger *zap.Logger
}

type clientList struct {
	discord discord.DiscordGatewayService
}

type syncData struct {
	ChannelId   string
	UserId      string
	SendMessage bool
}

var syncControl chan syncData
var clients clientList
var botRole string
var ignoredRoles []string
var roleKeys = []string{"Name", "Color", "Hoist", "Position", "Permissions", "Managed", "Mentionable", "Sync"}
var roleTypes = []string{"internal", "discord"}

func NewRolesHandler(config *config.Configuration, service micro.Service, log *zap.Logger) rolesrv.RolesHandler {
	c := service.Client()

	clients = clientList{
		discord: discord.NewDiscordGatewayService(config.LookupService("gateway", "discord"), c),
	}

	botRole = config.Bot.BotRole
	ignoredRoles = config.Bot.IgnoredRoles

	addr := fmt.Sprintf("%s:%d", config.Redis.Host, config.Redis.Port)
	redisClient := redis.Init(addr, config.Redis.Password, 0, config.LookupService("srv", "perms"))

	_, err := redisClient.Client.Ping().Result()
	if err != nil {
		panic(err)
	}

	// Let's create the role_admins and sig_admins stuff if it doesn't exist yet
	roleAdmin := redisClient.KeyName("perms:description:role_admins")
	exists, err := redisClient.Client.Exists(roleAdmin).Result()
	if err != nil {
		fmt.Println(err)
	} else {
		if exists == 0 {
			log.Info("role_admins doesn't exist. Creating it.")
			_, err = redisClient.Client.Set(roleAdmin, "Role Admins", 0).Result()
			if err != nil {
				log.Error(err.Error())
			}
		}
	}

	sigAdmin := redisClient.KeyName("perms:description:sig_admins")
	exists, err = redisClient.Client.Exists(sigAdmin).Result()
	if err != nil {
		fmt.Println(err)
	} else {
		if exists == 0 {
			log.Info("sig_admins doesn't exist. Creating it.")
			_, err = redisClient.Client.Set(sigAdmin, "SIG Admins", 0).Result()
			if err != nil {
				log.Error(err.Error())
			}
		}
	}

	rh := &rolesHandler{Redis: redisClient, Logger: log}

	// Check and update Redis schema as needed
	rh.updateSchema()

	// Start sync thread
	syncControl = make(chan syncData, 30)
	go rh.syncThread()

	return rh
}

func (h *rolesHandler) updateSchema() {
	sugar := h.Logger.Sugar()

	// Update Roles hash
	roles, err := h.getRoles()

	if err != nil {
		sugar.Errorf("Something went wrong getting the Roles: %s", err)
		return
	}

	for role := range roles {
		roleName := h.Redis.KeyName(fmt.Sprintf("role:%s", roles[role]))
		roleInfo, err := h.Redis.Client.HGetAll(roleName).Result()
		if err != nil {
			sugar.Errorf("Something went wrong getting the hash from Redis: %s", err)
			return
		}

		if roleInfo["Sync"] == "" {
			sugar.Infof("Updating role (%s) with default Sync value", roles[role])
			h.Redis.Client.HSet(roleName, "Sync", "1")
		}
	}
}

func (h *rolesHandler) GetRoleKeys(ctx context.Context, request *rolesrv.NilMessage, response *rolesrv.StringList) error {
	response.Value = roleKeys
	return nil
}

func (h *rolesHandler) GetRoleTypes(ctx context.Context, request *rolesrv.NilMessage, response *rolesrv.StringList) error {
	response.Value = roleTypes
	return nil
}

func (h *rolesHandler) AddRole(ctx context.Context, request *rolesrv.Role, response *rolesrv.NilMessage) error {
	roleName := h.Redis.KeyName(fmt.Sprintf("role:%s", request.ShortName))
	filterA := h.Redis.KeyName(fmt.Sprintf("filter_description:%s", request.FilterA))
	filterB := h.Redis.KeyName(fmt.Sprintf("filter_description:%s", request.FilterB))

	// Type, Name and the filters are required so let's check for those
	if len(request.Type) == 0 {
		return errors.New("Type is required.")
	}

	if len(request.ShortName) == 0 {
		return errors.New("ShortName is required.")
	}

	if len(request.Name) == 0 {
		return errors.New("Name is required.")
	}

	if len(request.FilterA) == 0 {
		return errors.New("FilterA is required.")
	}

	if len(request.FilterB) == 0 {
		return errors.New("FilterB is required.")
	}

	if !validListItem(request.Type, roleTypes) {
		return fmt.Errorf("`%s` isn't a valid Role Type.", request.Type)
	}

	exists, err := h.Redis.Client.Exists(roleName).Result()

	if err != nil {
		return err
	}

	if exists == 1 {
		return fmt.Errorf("Role `%s` already exists.", request.Name)
	}

	// Check if filter A exists
	exists, err = h.Redis.Client.Exists(filterA).Result()

	if err != nil {
		return err
	}

	if exists == 0 && request.FilterA != "wildcard" {
		return fmt.Errorf("FilterA `%s` doesn't exists.", request.FilterA)
	}

	// Check if filter B exists
	exists, err = h.Redis.Client.Exists(filterB).Result()

	if err != nil {
		return err
	}

	if exists == 0 && request.FilterB != "wildcard" {
		return fmt.Errorf("FilterB `%s` doesn't exists.", request.FilterB)
	}

	_, err = h.Redis.Client.HMSet(roleName, structs.Map(request)).Result()

	if err != nil {
		return err
	}

	response = &rolesrv.NilMessage{}

	return nil
}

func (h *rolesHandler) UpdateRole(ctx context.Context, request *rolesrv.UpdateInfo, response *rolesrv.NilMessage) error {
	// Does this actually work? -brian
	roleName := h.Redis.KeyName(fmt.Sprintf("role:%s", request.Name))

	exists, err := h.Redis.Client.Exists(roleName).Result()

	if err != nil {
		return err
	}

	if exists == 0 {
		return fmt.Errorf("Role `%s` doesn't exists.", request.Name)
	}

	if !validListItem(request.Key, roleKeys) {
		return fmt.Errorf("`%s` isn't a valid Role Key.", request.Key)
	}

	h.Redis.Client.HSet(roleName, request.Key, request.Value)

	return nil
}

func validListItem(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func (h *rolesHandler) RemoveRole(ctx context.Context, request *rolesrv.Role, response *rolesrv.NilMessage) error {
	roleName := h.Redis.KeyName(fmt.Sprintf("role:%s", request.ShortName))

	exists, err := h.Redis.Client.Exists(roleName).Result()

	if err != nil {
		return err
	}

	if exists == 0 {
		return fmt.Errorf("Role `%s` doesn't exists.", request.ShortName)
	}

	_, err = h.Redis.Client.Del(roleName).Result()

	if err != nil {
		return err
	}

	response = &rolesrv.NilMessage{}
	return nil
}

func (h *rolesHandler) GetRoles(ctx context.Context, request *rolesrv.NilMessage, response *rolesrv.GetRolesResponse) error {
	var sigValue, joinableValue, syncValue bool
	roles, err := h.getRoles()

	if err != nil {
		return err
	}

	for role := range roles {
		roleInfo, err := h.Redis.Client.HGetAll(h.Redis.KeyName(fmt.Sprintf("role:%s", roles[role]))).Result()
		if err != nil {
			return err
		}

		if roleInfo["Sig"] == "0" {
			sigValue = false
		} else {
			sigValue = true
		}

		if roleInfo["Joinable"] == "0" {
			joinableValue = false
		} else {
			joinableValue = true
		}

		if roleInfo["Sync"] == "0" {
			syncValue = false
		} else {
			syncValue = true
		}

		response.Roles = append(response.Roles, &rolesrv.Role{
			ShortName: roles[role],
			Name:      roleInfo["Name"],
			Sig:       sigValue,
			Joinable:  joinableValue,
			Sync:      syncValue,
		})
	}

	return nil
}

func (h *rolesHandler) getRoles() ([]string, error) {
	var roleList []string
	roles, err := h.Redis.Client.Keys(h.Redis.KeyName("role:*")).Result()

	if err != nil {
		return nil, err
	}

	for role := range roles {
		roleName := strings.Split(roles[role], ":")
		roleList = append(roleList, roleName[len(roleName)-1])
	}

	return roleList, nil
}

func (h *rolesHandler) getRole(name string) (role map[string]string, err error) {
	roleName := h.Redis.KeyName(fmt.Sprintf("role:%s", name))

	exists, err := h.Redis.Client.Exists(roleName).Result()
	if err != nil {
		return nil, err
	}

	if exists == 0 {
		return nil, fmt.Errorf("role doesn't exist: %s", name)
	}

	r, err := h.Redis.Client.HGetAll(roleName).Result()
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (h *rolesHandler) mapRoleToProtobufRole(role map[string]string) *rolesrv.Role {
	color, _ := strconv.ParseInt(role["Color"], 10, 32)
	position, _ := strconv.ParseInt(role["Position"], 10, 32)
	permissions, _ := strconv.ParseInt(role["Permissions"], 10, 32)
	hoist, _ := strconv.ParseBool(role["Hoist"])
	managed, _ := strconv.ParseBool(role["Managed"])
	mentionable, _ := strconv.ParseBool(role["Mentionable"])
	sig, _ := strconv.ParseBool(role["Sig"])
	joinable, _ := strconv.ParseBool(role["Joinable"])
	sync, _ := strconv.ParseBool(role["Sync"])

	return &rolesrv.Role{
		ShortName:   role["ShortName"],
		Type:        role["Type"],
		FilterA:     role["FilterA"],
		FilterB:     role["FilterB"],
		Name:        role["Name"],
		Color:       int32(color),
		Hoist:       hoist,
		Position:    int32(position),
		Permissions: int32(permissions),
		Managed:     managed,
		Mentionable: mentionable,
		Sig:         sig,
		Joinable:    joinable,
		Sync:        sync,
	}
}

func (h *rolesHandler) GetRole(ctx context.Context, request *rolesrv.Role, response *rolesrv.Role) error {
	role, err := h.getRole(request.ShortName)
	if err != nil {
		return err
	}

	*response = *h.mapRoleToProtobufRole(role)
	return nil
}

func (h *rolesHandler) sendMessage(ctx context.Context, channelId, message string, sendMessage bool) {
	sugar := h.Logger.Sugar()

	if sendMessage {
		_, err := clients.discord.SendMessage(ctx, &discord.SendMessageRequest{ChannelId: channelId, Message: message})
		if err != nil {
			msg := fmt.Sprintf("sendMessage: %s", err.Error())
			sugar.Error(msg)
		}
	}
}

func (h *rolesHandler) syncMembers(channelId, userId string, sendMessage bool) error {
	sugar := h.Logger.Sugar()
	var roleNameMap = make(map[string]string)
	var discordMemberships = make(map[string]*sets.StringSet)
	var chremoasMemberships = make(map[string]*sets.StringSet)
	var updateMembers = make(map[string]*sets.StringSet)

	// Discord limit is 1000, should probably make this a config option. -brian
	var numberPerPage int32 = 1000
	var memberCount = 1
	var memberId = ""

	t := time.Now()

	// Need to pre-populate the membership sets with all the users so we can pick up users with no roles.
	for memberCount > 0 {
		//longCtx, _ := context.WithTimeout(context.Background(), time.Second * 20)

		members, err := clients.discord.GetAllMembers(context.Background(), &discord.GetAllMembersRequest{NumberPerPage: numberPerPage, After: memberId})
		if err != nil {
			msg := fmt.Sprintf("syncMembers: GetAllMembers: %s", err.Error())
			h.sendMessage(context.Background(), channelId, common.SendFatal(msg), true)
			sugar.Error(msg)
			return err
		}

		for m := range members.Members {
			userId := members.Members[m].User.Id
			if _, ok := discordMemberships[userId]; !ok {
				discordMemberships[userId] = sets.NewStringSet()
			}

			for r := range members.Members[m].Roles {
				discordMemberships[userId].Add(members.Members[m].Roles[r].Name)
			}

			if _, ok := chremoasMemberships[userId]; !ok {
				chremoasMemberships[userId] = sets.NewStringSet()
			}

			oldNum, _ := strconv.Atoi(members.Members[m].User.Id)
			newNum, _ := strconv.Atoi(memberId)

			if oldNum > newNum {
				memberId = members.Members[m].User.Id
			}
		}

		memberCount = len(members.Members)
	}

	h.sendDualMessage(
		fmt.Sprintf("Got all Discord members [%s]", time.Since(t)),
		channelId,
		sendMessage,
	)

	t = time.Now()

	// Get all the Roles from discord and create a map of their name to their Id
	discordRoles, err := clients.discord.GetAllRoles(context.Background(), &discord.GuildObjectRequest{})
	if err != nil {
		msg := fmt.Sprintf("syncMembers: GetAllRoles: %s", err.Error())
		h.sendMessage(context.Background(), channelId, common.SendFatal(msg), true)
		sugar.Error(msg)
		return err
	}

	for d := range discordRoles.Roles {
		roleNameMap[discordRoles.Roles[d].Name] = discordRoles.Roles[d].Id
	}

	h.sendDualMessage(
		fmt.Sprintf("Got all Discord roles [%s]", time.Since(t)),
		channelId,
		sendMessage,
	)

	t = time.Now()

	// Get all the Chremoas roles and build membership Sets
	chremoasRoles, err := h.getRoles()
	if err != nil {
		msg := fmt.Sprintf("syncMembers: getRoles: %s", err.Error())
		h.sendMessage(context.Background(), channelId, common.SendFatal(msg), true)
		sugar.Error(msg)
		return err
	}

	h.sendDualMessage(
		fmt.Sprintf("Got all Chremoas roles [%s]", time.Since(t)),
		channelId,
		sendMessage,
	)

	t = time.Now()

	for r := range chremoasRoles {
		sugar.Infof("Checking role: %s", chremoasRoles[r])
		role, err := h.getRole(chremoasRoles[r])
		if err != nil {
			msg := fmt.Sprintf("syncMembers: getRole: %s: %s", chremoasRoles[r], err.Error())
			h.sendMessage(context.Background(), channelId, common.SendFatal(msg), true)
			sugar.Error(msg)
			return err
		}

		if role["Sync"] == "0" || role["Sync"] == "false" {
			continue
		}

		membership, err := h.getRoleMembership(chremoasRoles[r])
		if err != nil {
			msg := fmt.Sprintf("syncMembers: getRoleMembership: %s", err.Error())
			h.sendMessage(context.Background(), channelId, common.SendFatal(msg), true)
			sugar.Error(msg)
			return err
		}

		roleName, err := h.getRole(chremoasRoles[r])
		if err != nil {
			msg := fmt.Sprintf("syncMembers: getRole: %s", err.Error())
			h.sendMessage(context.Background(), channelId, common.SendFatal(msg), true)
			sugar.Error(msg)
			return err
		}

		//roleId := roleNameMap[roleName["Name"]]

		for m := range membership.Set {
			sugar.Debugf("Key is: %s", m)
			if len(m) != 0 {
				sugar.Debugf("Set is %v", chremoasMemberships[m])
				if chremoasMemberships[m] == nil {
					chremoasMemberships[m] = sets.NewStringSet()
				}
				chremoasMemberships[m].Add(roleName["Name"])
			}
		}
	}

	h.sendDualMessage(
		fmt.Sprintf("Got all role Memberships [%s]", time.Since(t)),
		channelId,
		sendMessage,
	)

	t = time.Now()

	for m := range chremoasMemberships {
		if discordMemberships[m] == nil {
			sugar.Debugf("not in discord: %v", m)
			continue
		}

		diff := chremoasMemberships[m].Difference(discordMemberships[m])
		if diff.Len() != 0 {
			sugar.Infof("diff1: %v", diff)
			updateMembers[m] = sets.NewStringSet()
			for r := range chremoasMemberships[m].Set {
				//for i := range ignoredRoles {
				//	sugar.Infof("Checking %s == %s", roleNameMap[r], ignoredRoles[i])
				//	if roleNameMap[r] == ignoredRoles[i] {
				//		continue
				//	}
				//}

				updateMembers[m].Add(roleNameMap[r])
			}
		}

		// TODO: Figure out if we really need this?
		//diff = discordMemberships[m].Difference(chremoasMemberships[m])
		//if diff.Len() != 0 {
		//	sugar.Infof("diff2: %v", diff)
		//	updateMembers[m] = sets.NewStringSet()
		//	for r := range chremoasMemberships[m].Set {
		//		for i := range ignoredRoles {
		//			sugar.Infof("Checking %s == %s", roleNameMap[r], ignoredRoles[i])
		//			//if roleNameMap[r] == ignoredRoles[i] {
		//			//	continue
		//			//}
		//		}
		//
		//		updateMembers[m].Add(roleNameMap[r])
		//	}
		//}

		if updateMembers[m] != nil {
			sugar.Infof("m: %v updateMembers: %v", m, updateMembers[m])
		}
	}

	// Apply the membership sets to discord overwriting anything that's there.
	h.sendDualMessage(
		fmt.Sprintf("Updating %d discord users", len(updateMembers)),
		channelId,
		sendMessage,
	)

	noSyncList := h.Redis.KeyName("members:no_sync")
	sugar.Infof("noSyncList: %v", noSyncList)
	for m := range updateMembers {
		// Don't sync people who we don't want to mess with. Always put the Discord Server Owner here
		// because we literally can't sync them no matter what.
		noSync, _ := h.Redis.Client.SIsMember(noSyncList, m).Result()
		if noSync {
			sugar.Infof("Skipping noSync user: %s", m)
			continue
		}

		ctx, _ := context.WithTimeout(context.Background(), time.Second*20)
		_, err = clients.discord.UpdateMember(ctx, &discord.UpdateMemberRequest{
			Operation: discord.MemberUpdateOperation_ADD_OR_UPDATE_ROLES,
			UserId:    m,
			RoleIds:   updateMembers[m].ToSlice(),
		})
		if err != nil {
			msg := fmt.Sprintf("syncMembers: UpdateMember: %s", err.Error())
			h.sendMessage(context.Background(), channelId, common.SendFatal(msg), true)
			sugar.Error(msg)
		}
		sugar.Infof("Updating Discord User: %s", m)
	}

	h.sendDualMessage(
		fmt.Sprintf("Updated Discord Roles [%s]", time.Since(t)),
		channelId,
		sendMessage,
	)

	return nil
}

func (h *rolesHandler) GetRoleMembership(ctx context.Context, request *rolesrv.RoleMembershipRequest, response *rolesrv.RoleMembershipResponse) error {
	members, err := h.getRoleMembership(request.Name)
	if err != nil {
		return err
	}

	for m := range members.Set {
		response.Members = append(response.Members, m)
	}

	return nil
}

func (h *rolesHandler) getRoleMembership(role string) (members *sets.StringSet, err error) {
	var filterASet = sets.NewStringSet()
	var filterBSet = sets.NewStringSet()

	roleName := h.Redis.KeyName(fmt.Sprintf("role:%s", role))

	r, err := h.Redis.Client.HGetAll(roleName).Result()
	if err != nil {
		return filterASet, err
	}

	filterADesc := h.Redis.KeyName(fmt.Sprintf("filter_description:%s", r["FilterA"]))
	filterBDesc := h.Redis.KeyName(fmt.Sprintf("filter_description:%s", r["FilterB"]))

	filterAMembers := h.Redis.KeyName(fmt.Sprintf("filter_members:%s", r["FilterA"]))
	filterBMembers := h.Redis.KeyName(fmt.Sprintf("filter_members:%s", r["FilterB"]))

	if r["FilterB"] == "wildcard" {
		exists, err := h.Redis.Client.Exists(filterADesc).Result()
		if err != nil {
			return filterASet, err
		}

		if exists == 0 {
			return filterASet, fmt.Errorf("Filter `%s` doesn't exists.", r["FilterA"])
		}

		filterA, err := h.Redis.Client.SMembers(filterAMembers).Result()
		if err != nil {
			return filterASet, err
		}

		filterASet.FromSlice(filterA)
		return filterASet, nil
	}

	if r["FilterA"] == "wildcard" {
		exists, err := h.Redis.Client.Exists(filterBDesc).Result()
		if err != nil {
			return filterASet, err
		}

		if exists == 0 {
			return filterASet, fmt.Errorf("Filter `%s` doesn't exists.", r["FilterB"])
		}

		filterB, err := h.Redis.Client.SMembers(filterBMembers).Result()
		if err != nil {
			return filterASet, err
		}

		filterBSet.FromSlice(filterB)
		return filterBSet, nil
	}

	filterInter, err := h.Redis.Client.SInter(filterAMembers, filterBMembers).Result()
	if err != nil {
		return filterASet, err
	}

	filterASet.FromSlice(filterInter)
	return filterASet, nil
}

func (h *rolesHandler) syncRoles(channelId, userId string, sendMessage bool) error {
	ctx := context.Background()
	var matchDiscordError = regexp.MustCompile(`^The role '.*' already exists$`)
	chremoasRoleSet := sets.NewStringSet()
	discordRoleSet := sets.NewStringSet()
	sugar := h.Logger.Sugar()
	var chremoasRoleData = make(map[string]map[string]string)

	chremoasRoles, err := h.getRoles()
	if err != nil {
		msg := fmt.Sprintf("syncRoles: h.getRoles(): %s", err.Error())
		h.sendMessage(ctx, channelId, common.SendFatal(msg), true)
		sugar.Error(msg)
		return err
	}

	for role := range chremoasRoles {
		roleName := h.Redis.KeyName(fmt.Sprintf("role:%s", chremoasRoles[role]))
		c, err := h.Redis.Client.HGetAll(roleName).Result()

		if err != nil {
			msg := fmt.Sprintf("syncRoles: HGetAll(): %s", err.Error())
			h.sendMessage(ctx, channelId, common.SendFatal(msg), true)
			sugar.Error(msg)
			return err
		}

		sugar.Infof("Checking %s: %s", c["Name"], c["Sync"])
		if c["Sync"] == "1" || c["Sync"] == "true" {
			chremoasRoleSet.Add(c["Name"])

			if _, ok := chremoasRoleData[c["Name"]]; !ok {
				chremoasRoleData[c["Name"]] = make(map[string]string)
			}
			chremoasRoleData[c["Name"]] = c
		}
	}

	discordRoles, err := clients.discord.GetAllRoles(ctx, &discord.GuildObjectRequest{})
	if err != nil {
		msg := fmt.Sprintf("syncRoles: GetAllRoles: %s", err.Error())
		h.sendMessage(ctx, channelId, common.SendFatal(msg), true)
		sugar.Error(msg)
		return err
	}

	ignoreSet := sets.NewStringSet()
	ignoreSet.Add(botRole)
	ignoreSet.Add("@everyone")
	for i := range ignoredRoles {
		ignoreSet.Add(ignoredRoles[i])
	}

	for role := range discordRoles.Roles {
		if !ignoreSet.Contains(discordRoles.Roles[role].Name) {
			discordRoleSet.Add(discordRoles.Roles[role].Name)
		}
	}

	toAdd := chremoasRoleSet.Difference(discordRoleSet)
	toDelete := discordRoleSet.Difference(chremoasRoleSet)
	toUpdate := discordRoleSet.Intersection(chremoasRoleSet)

	sugar.Debugf("toAdd: %v", toAdd)
	sugar.Debugf("toDelete: %v", toDelete)
	sugar.Debugf("toUpdate: %v", toUpdate)

	for r := range toAdd.Set {
		_, err := clients.discord.CreateRole(ctx, &discord.CreateRoleRequest{Name: r})

		if err != nil {
			if matchDiscordError.MatchString(err.Error()) {
				// The role list was cached most likely so we'll pretend we didn't try
				// to create it just now. -brian
				sugar.Debugf("syncRoles added: %s", r)
				continue
			} else {
				msg := fmt.Sprintf("syncRoles: CreateRole(): %s", err.Error())
				h.sendMessage(ctx, channelId, common.SendFatal(msg), true)
				sugar.Error(msg)
				return err
			}
		}

		sugar.Debugf("syncRoles added: %s", r)
	}

	for r := range toDelete.Set {
		_, err := clients.discord.DeleteRole(ctx, &discord.DeleteRoleRequest{Name: r})

		if err != nil {
			msg := fmt.Sprintf("syncRoles: DeleteRole(): %s", err.Error())
			h.sendMessage(ctx, channelId, common.SendFatal(msg), true)
			sugar.Error(msg)
			return err
		}

		sugar.Debugf("syncRoles removed: %s", r)
	}

	for r := range toUpdate.Set {
		color, _ := strconv.ParseInt(chremoasRoleData[r]["Color"], 10, 64)
		perm, _ := strconv.ParseInt(chremoasRoleData[r]["Permissions"], 10, 64)
		position, _ := strconv.ParseInt(chremoasRoleData[r]["Position"], 10, 64)
		hoist, _ := strconv.ParseBool(chremoasRoleData[r]["Hoist"])
		mention, _ := strconv.ParseBool(chremoasRoleData[r]["Mentionable"])
		managed, _ := strconv.ParseBool(chremoasRoleData[r]["Managed"])

		editRequest := &discord.EditRoleRequest{
			Name:     chremoasRoleData[r]["Name"],
			Color:    color,
			Perm:     perm,
			Position: position,
			Hoist:    hoist,
			Mention:  mention,
			Managed:  managed,
		}

		longCtx, _ := context.WithTimeout(ctx, time.Minute*5)
		_, err := clients.discord.EditRole(longCtx, editRequest)
		if err != nil {
			msg := fmt.Sprintf("syncRoles: EditRole(): %s", err.Error())
			h.sendMessage(ctx, channelId, common.SendFatal(msg), true)
			sugar.Error(msg)
			return err
		}

		sugar.Debugf("syncRoles updated: %s", r)
	}

	return nil
}

//
// Filter related stuff
//

func (h *rolesHandler) GetFilters(ctx context.Context, request *rolesrv.NilMessage, response *rolesrv.FilterList) error {
	filters, err := h.Redis.Client.Keys(h.Redis.KeyName("filter_description:*")).Result()

	if err != nil {
		return err
	}

	for filter := range filters {
		filterDescription, err := h.Redis.Client.Get(filters[filter]).Result()

		if err != nil {
			return err
		}

		filterName := strings.Split(filters[filter], ":")

		response.FilterList = append(response.FilterList,
			&rolesrv.Filter{Name: filterName[len(filterName)-1], Description: filterDescription})
	}

	return nil
}

func (h *rolesHandler) AddFilter(ctx context.Context, request *rolesrv.Filter, response *rolesrv.NilMessage) error {
	filterName := h.Redis.KeyName(fmt.Sprintf("filter_description:%s", request.Name))

	// Type and Name are required so let's check for those
	if len(request.Name) == 0 {
		return errors.New("Name is required.")
	}

	if len(request.Description) == 0 {
		return errors.New("Description is required.")
	}

	exists, err := h.Redis.Client.Exists(filterName).Result()

	if err != nil {
		return err
	}

	if exists == 1 {
		return fmt.Errorf("Filter `%s` already exists.", request.Name)
	}

	_, err = h.Redis.Client.Set(filterName, request.Description, 0).Result()

	if err != nil {
		return err
	}

	response = &rolesrv.NilMessage{}

	return nil
}

func (h *rolesHandler) RemoveFilter(ctx context.Context, request *rolesrv.Filter, response *rolesrv.NilMessage) error {
	filterName := h.Redis.KeyName(fmt.Sprintf("filter_description:%s", request.Name))
	filterMembers := h.Redis.KeyName(fmt.Sprintf("filter_members:%s", request.Name))

	exists, err := h.Redis.Client.Exists(filterName).Result()

	if err != nil {
		return err
	}

	if exists == 0 {
		return fmt.Errorf("Filter `%s` doesn't exists.", request.Name)
	}

	members, err := h.Redis.Client.SMembers(filterMembers).Result()

	if len(members) > 0 {
		return fmt.Errorf("Filter `%s` not empty.", request.Name)
	}

	_, err = h.Redis.Client.Del(filterName).Result()

	if err != nil {
		return err
	}

	response = &rolesrv.NilMessage{}
	return nil
}

func (h *rolesHandler) GetMembers(ctx context.Context, request *rolesrv.Filter, response *rolesrv.MemberList) error {
	var memberlist []string
	filterName := h.Redis.KeyName(fmt.Sprintf("filter_members:%s", request.Name))

	filters, err := h.Redis.Client.SMembers(filterName).Result()

	if err != nil {
		return err
	}

	for filter := range filters {
		if len(filters[filter]) > 0 {
			memberlist = append(memberlist, filters[filter])
		}
	}

	response.Members = memberlist
	return nil
}

func (h *rolesHandler) AddMembers(ctx context.Context, request *rolesrv.Members, response *rolesrv.NilMessage) error {
	filterName := h.Redis.KeyName(fmt.Sprintf("filter_members:%s", request.Filter))
	filterDesc := h.Redis.KeyName(fmt.Sprintf("filter_description:%s", request.Filter))

	exists, err := h.Redis.Client.Exists(filterDesc).Result()

	if err != nil {
		return err
	}

	if exists == 0 {
		return fmt.Errorf("Filter `%s` doesn't exists.", request.Filter)
	}

	for member := range request.Name {
		_, err = h.Redis.Client.SAdd(filterName, request.Name[member]).Result()
	}

	if err != nil {
		return err
	}

	response = &rolesrv.NilMessage{}
	return nil
}

func (h *rolesHandler) RemoveMembers(ctx context.Context, request *rolesrv.Members, response *rolesrv.NilMessage) error {
	filterName := h.Redis.KeyName(fmt.Sprintf("filter_members:%s", request.Filter))
	filterDesc := h.Redis.KeyName(fmt.Sprintf("filter_description:%s", request.Filter))

	exists, err := h.Redis.Client.Exists(filterDesc).Result()

	if err != nil {
		return err
	}

	if exists == 0 {
		return fmt.Errorf("Filter `%s` doesn't exists.", request.Filter)
	}

	for member := range request.Name {
		_, err = h.Redis.Client.SRem(filterName, request.Name[member]).Result()
	}

	if err != nil {
		return err
	}

	response = &rolesrv.NilMessage{}
	return nil
}

func (h *rolesHandler) GetDiscordUser(ctx context.Context, request *rolesrv.GetDiscordUserRequest, response *rolesrv.GetDiscordUserResponse) error {
	members, err := clients.discord.GetAllMembersAsSlice(ctx, &discord.GetAllMembersRequest{})
	if err != nil {
		return err
	}

	for m := range members.Members {
		if request.UserId == members.Members[m].User.Id {
			response.Username = members.Members[m].User.Username
			response.Avatar = members.Members[m].User.Avatar
			response.Bot = members.Members[m].User.Bot
			response.Discriminator = members.Members[m].User.Discriminator
			response.Email = members.Members[m].User.Email
			response.MfaEnabled = members.Members[m].User.MFAEnabled
			response.Verified = members.Members[m].User.Verified

			return nil
		}
	}

	return errors.New("User not found")
}

func (h *rolesHandler) GetDiscordUserList(ctx context.Context, request *rolesrv.NilMessage, response *rolesrv.GetDiscordUserListResponse) error {
	members, err := clients.discord.GetAllMembersAsSlice(ctx, &discord.GetAllMembersRequest{})
	if err != nil {
		return err
	}

	for m := range members.Members {
		response.Users = append(response.Users, &rolesrv.GetDiscordUserResponse{
			Nick:          members.Members[m].Nick,
			Id:            members.Members[m].User.Id,
			Username:      members.Members[m].User.Username,
			Avatar:        members.Members[m].User.Avatar,
			Bot:           members.Members[m].User.Bot,
			Discriminator: members.Members[m].User.Discriminator,
			Email:         members.Members[m].User.Email,
			MfaEnabled:    members.Members[m].User.MFAEnabled,
			Verified:      members.Members[m].User.Verified,
		})
	}

	return nil
}

func (h *rolesHandler) SyncToChatService(ctx context.Context, request *rolesrv.SyncRequest, response *rolesrv.NilMessage) error {
	syncControl <- syncData{ChannelId: request.ChannelId, UserId: request.UserId, SendMessage: request.SendMessage}
	return nil
}

func (h *rolesHandler) sendDualMessage(msg, channelId string, sendMessage bool) {
	ctx := context.Background()
	sugar := h.Logger.Sugar()

	sugar.Info(msg)
	h.sendMessage(ctx, channelId, common.SendSuccess(msg), sendMessage)
}

func (h *rolesHandler) syncThread() {
	for {
		request := <-syncControl

		t1 := time.Now()

		h.sendDualMessage("Starting Role Sync", request.ChannelId, request.SendMessage)

		h.syncRoles(request.ChannelId, request.UserId, request.SendMessage)

		msg := fmt.Sprintf("Completed Role Sync [%s]", time.Since(t1))
		h.sendDualMessage(msg, request.ChannelId, request.SendMessage)

		t2 := time.Now()
		h.sendDualMessage("Starting Member Sync", request.ChannelId, request.SendMessage)

		h.syncMembers(request.ChannelId, request.UserId, request.SendMessage)

		msg = fmt.Sprintf("Completed Member Sync [%s]", time.Since(t2))
		h.sendDualMessage(msg, request.ChannelId, request.SendMessage)

		msg = fmt.Sprintf("Completed All Syncing [%s]", time.Since(t1))
		h.sendDualMessage(msg, request.ChannelId, request.SendMessage)
	}
}

func (h *rolesHandler) ListUserRoles(ctx context.Context, request *rolesrv.ListUserRolesRequest, response *rolesrv.ListUserRolesResponse) error {
	roles, err := h.getRoles()
	if err != nil {
		return err
	}

	for role := range roles {
		r, err := h.getRoleMembership(roles[role])
		if err != nil {
			return err
		}

		if r.Contains(request.UserId) {
			rInfo, err := h.getRole(roles[role])
			if err != nil {
				return err
			}

			response.Roles = append(response.Roles, h.mapRoleToProtobufRole(rInfo))
		}
	}

	return nil
}
