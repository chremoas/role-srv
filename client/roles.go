package client

import (
	"bytes"
	"context"
	"fmt"
	permsrv "github.com/chremoas/perms-srv/proto"
	rolesrv "github.com/chremoas/role-srv/proto"
	common "github.com/chremoas/services-common/command"
	"go.uber.org/zap"
	"github.com/chremoas/services-common/sets"
	"strconv"
)

type Roles struct {
	RoleClient  rolesrv.RolesService
	PermsClient permsrv.PermissionsService
	Permissions common.Permissions
	Logger      *zap.Logger
}

var clientType = map[bool]string{true: "SIG", false: "Role"}

func (r Roles) ListRoles(ctx context.Context, all, sig bool) string {
	var buffer bytes.Buffer
	var roleList = make(map[string]string)
	roles, err := r.RoleClient.GetRoles(ctx, &rolesrv.NilMessage{})

	if err != nil {
		return common.SendFatal(err.Error())
	}

	for role := range roles.Roles {
		if roles.Roles[role].Sig == sig {
			if roles.Roles[role].Sig && !roles.Roles[role].Joinable && !all {
				continue
			}
			roleList[roles.Roles[role].ShortName] = roles.Roles[role].Name
		}
	}

	if len(roleList) == 0 {
		return common.SendError(fmt.Sprintf("No %ss\n", clientType[sig]))
	}

	buffer.WriteString(fmt.Sprintf("%ss:\n", clientType[sig]))
	for role := range roleList {
		buffer.WriteString(fmt.Sprintf("\t%s: %s\n", role, roleList[role]))
	}

	return fmt.Sprintf("```%s```", buffer.String())
}

func (r Roles) AddRole(ctx context.Context, sender, shortName, roleType, filterA, filterB string, joinable bool, roleName string, sig bool) string {
	if len(roleName) > 0 && roleName[0] == '"' {
		roleName = roleName[1:]
	}

	if len(roleName) > 0 && roleName[len(roleName)-1] == '"' {
		roleName = roleName[:len(roleName)-1]
	}

	canPerform, err := r.Permissions.CanPerform(ctx, sender)
	if err != nil {
		return common.SendFatal(err.Error())
	}

	if !canPerform {
		return common.SendError("User doesn't have permission to this command")
	}

	_, err = r.RoleClient.AddRole(ctx,
		&rolesrv.Role{
			Sig:       sig,
			ShortName: shortName,
			Type:      roleType,
			Name:      roleName,
			FilterA:   filterA,
			FilterB:   filterB,
			Joinable:  joinable,
		})

	if err != nil {
		return common.SendFatal(err.Error())
	}

	_, err = r.RoleClient.SyncRoles(ctx, r.GetSyncRequest(sender, false))
	if err != nil {
		return common.SendFatal(err.Error())
	}

	return common.SendSuccess(fmt.Sprintf("Added: %s\n", shortName))
}

func (r Roles) RemoveRole(ctx context.Context, sender, shortName string, sig bool) string {
	canPerform, err := r.Permissions.CanPerform(ctx, sender)
	if err != nil {
		return common.SendFatal(err.Error())
	}

	if !canPerform {
		return common.SendError("User doesn't have permission to this command")
	}

	// Need to check if it's a sig or not
	role, err := r.RoleClient.GetRole(ctx, &rolesrv.Role{ShortName: shortName})
	if err != nil {
		return common.SendFatal(err.Error())
	}

	if role.Sig != sig {
		return common.SendError(fmt.Sprintf("'%s' doesn't exist", shortName))
	}

	_, err = r.RoleClient.RemoveRole(ctx, &rolesrv.Role{ShortName: shortName})
	if err != nil {
		return common.SendFatal(err.Error())
	}

	_, err = r.RoleClient.SyncRoles(ctx, r.GetSyncRequest(sender, false))
	if err != nil {
		return common.SendFatal(err.Error())
	}

	return common.SendSuccess(fmt.Sprintf("Removed: %s\n", shortName))
}

func (r Roles) RoleInfo(ctx context.Context, sender, shortName string, sig bool) string {
	var buffer bytes.Buffer

	canPerform, err := r.Permissions.CanPerform(ctx, sender)
	if err != nil {
		return common.SendFatal(err.Error())
	}

	if !canPerform {
		return common.SendError("User doesn't have permission to this command")
	}

	info, err := r.RoleClient.GetRole(ctx, &rolesrv.Role{ShortName: shortName})
	if err != nil {
		return common.SendFatal(err.Error())
	}

	buffer.WriteString(fmt.Sprintf("ShortName: %s\n", info.ShortName))
	buffer.WriteString(fmt.Sprintf("Type: %s\n", info.Type))
	buffer.WriteString(fmt.Sprintf("FilterA: %s\n", info.FilterA))
	buffer.WriteString(fmt.Sprintf("FilterB: %s\n", info.FilterB))
	buffer.WriteString(fmt.Sprintf("Name: %s\n", info.Name))
	buffer.WriteString(fmt.Sprintf("Color: %d\n", info.Color))
	buffer.WriteString(fmt.Sprintf("Hoist: %t\n", info.Hoist))
	buffer.WriteString(fmt.Sprintf("Position: %d\n", info.Position))
	buffer.WriteString(fmt.Sprintf("Permissions: %d\n", info.Permissions))
	buffer.WriteString(fmt.Sprintf("Manged: %t\n", info.Managed))
	buffer.WriteString(fmt.Sprintf("Mentionable: %t\n", info.Mentionable))
	if sig {
		buffer.WriteString(fmt.Sprintf("Joinable: %t\n", info.Joinable))
	}

	return fmt.Sprintf("```%s```", buffer.String())
}

func (r Roles) SyncRoles(ctx context.Context, sender string) string {
	r.Logger.Info("Calling SyncRoles()")

	_, err := r.RoleClient.SyncRoles(ctx, r.GetSyncRequest(sender, true))
	if err != nil {
		return common.SendFatal(err.Error())
	}

	return common.SendSuccess("Done")
}

func (r Roles) Set(ctx context.Context, sender, name, key, value string) string {
	var validKeys = sets.NewStringSet()
	validKeys.FromSlice([]string{"Color", "Hoist", "Position", "Permissions", "Managed", "Mentionable"})

	if !validKeys.Contains(key) {
		var buffer bytes.Buffer
		buffer.WriteString(fmt.Sprintf("Unknown key: %s\nValid Options are:\n", key))
		for k := range validKeys.Set {
			buffer.WriteString(fmt.Sprintf("\t%s\n", k))
		}
		return common.SendError(buffer.String())
	}

	if key == "Color" {
		if string(value[0]) == "#" {
			i, _ := strconv.ParseInt(value[1:], 16, 64)
			value = strconv.Itoa(int(i))
		}
	}

	_, err := r.RoleClient.UpdateRole(ctx, &rolesrv.UpdateInfo{Name: name, Key: key, Value: value})
	if err != nil {
		return common.SendFatal(err.Error())
	}

	_, err = r.RoleClient.SyncRoles(ctx, r.GetSyncRequest(sender, false))
	if err != nil {
		return common.SendFatal(err.Error())
	}

	return common.SendSuccess(fmt.Sprintf("Set '%s' to '%s' for '%s'", key, value, name))
}

func (r Roles) GetMembers(ctx context.Context, role string) string {
	var buffer bytes.Buffer

	members, err := r.RoleClient.GetRoleMembership(ctx, &rolesrv.RoleMembershipRequest{Name: role})
	if err != nil {
		return common.SendFatal(err.Error())
	}

	for m := range members.Members {
		if len(members.Members[m]) > 0 {
			user, err := r.RoleClient.GetDiscordUser(ctx, &rolesrv.GetDiscordUserRequest{UserId: members.Members[m]})
			if err != nil {
				return common.SendError(err.Error())
			}
			buffer.WriteString(fmt.Sprintf("\t%s\n", user.Username))
		}
	}

	if buffer.Len() == 0 {
		return "```Empty list```\n"
	}

	return fmt.Sprintf("```%s Members:\n%s```\n", role, buffer.String())
}