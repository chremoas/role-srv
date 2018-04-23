package client

import (
	"bytes"
	"fmt"
	rolesrv "github.com/chremoas/role-srv/proto"
	permsrv "github.com/chremoas/perms-srv/proto"
	common "github.com/chremoas/services-common/command"
	"context"
)

type Roles struct {
	RoleClient rolesrv.RolesClient
	PermsClient permsrv.PermissionsClient
	Permissions common.Permissions
}

func (r Roles) ListRoles(ctx context.Context, sig bool) string {
	var buffer bytes.Buffer
	var roleList = make(map[string]string)
	roles, err := r.RoleClient.GetRoles(ctx, &rolesrv.NilMessage{})

	if err != nil {
		return common.SendFatal(err.Error())
	}

	for role := range roles.Roles {
		if roles.Roles[role].Sig == sig {
			roleList[roles.Roles[role].ShortName] = roles.Roles[role].Name
		}
	}

	if len(roleList) == 0 {
		return common.SendError("No SIGs\n")
	}

	buffer.WriteString("SIGs:\n")
	for role := range roleList {
		buffer.WriteString(fmt.Sprintf("\t%s: %s\n", role, roleList[role]))
	}

	return fmt.Sprintf("```%s```", buffer.String())
}

func (r Roles) AddRole(ctx context.Context, sender, shortName, roleType, filterA, filterB, roleName string, sig bool) string {

	if len(roleName) > 0 && roleName[0] == '"' {
		roleName = roleName[1:]
	}

	if len(roleName) > 0 && roleName[len(roleName)-1] == '"' {
		roleName = roleName[:len(roleName)-1]
	}

	canPerform, err := r.Permissions.CanPerform(ctx, sender, []string{"sig_admins"})
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
		})

	if err != nil {
		return common.SendFatal(err.Error())
	}

	return common.SendSuccess(fmt.Sprintf("Added: %s\n", shortName))
}