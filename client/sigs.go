package client

import (
	"context"
	rolesrv "github.com/chremoas/role-srv/proto"
	common "github.com/chremoas/services-common/command"
	"fmt"
	"strings"
)

func (r Roles) JoinSIG(ctx context.Context, sender, sig string) string {
	s := strings.Split(sender, ":")
	foo, err := r.RoleClient.GetRole(ctx, &rolesrv.Role{ShortName: sig})
	if err != nil {
		return common.SendError(err.Error())
	}

	if !foo.Sig {
		return common.SendError("Not a SIG")
	}

	// get the filter from from the role
	role, err := r.RoleClient.GetRole(ctx, &rolesrv.Role{ShortName: sig})
	if err != nil {
		return common.SendError(err.Error())
	}

	// Is this a joinable role?
	if !role.Joinable {
		return common.SendError(fmt.Sprintf("'%s' is not a joinable SIG, talk to an admin", sig))
	}

	// add member to role
	_, err = r.RoleClient.AddMembers(ctx, &rolesrv.Members{Name: []string{s[1]}, Filter: role.FilterB})
	if err != nil {
		return common.SendError(err.Error())
	}

	return common.SendSuccess(fmt.Sprintf("Added %s to %s", sender, sig))
}
