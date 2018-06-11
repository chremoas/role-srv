package client

import (
	"context"
	rolesrv "github.com/chremoas/role-srv/proto"
	common "github.com/chremoas/services-common/command"
	"fmt"
	"strings"
)

func (r Roles) AddSIG(ctx context.Context, sender, sig string) string {
	return r.sigAction(ctx, sender, sig, true, false)
}

func (r Roles) RemoveSIG(ctx context.Context, sender, sig string) string {
	return r.sigAction(ctx, sender, sig, false, false)
}

func (r Roles) JoinSIG(ctx context.Context, sender, sig string) string {
	return r.sigAction(ctx, sender, sig, true, true)
}

func (r Roles) LeaveSIG(ctx context.Context, sender, sig string) string {
	return r.sigAction(ctx, sender, sig, false, false)
}

func (r Roles) sigAction(ctx context.Context, sender, sig string, join, joinable bool) string {
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

	// Is this a joinable role? Only check on Join/Leave not Add/Remove
	if joinable {
		if !role.Joinable {
			return common.SendError(fmt.Sprintf("'%s' is not a joinable SIG, talk to an admin", sig))
		}
	}

	// add member to role
	if join {
		_, err = r.RoleClient.AddMembers(ctx, &rolesrv.Members{Name: []string{s[1]}, Filter: role.FilterB})
	} else {
		_, err = r.RoleClient.RemoveMembers(ctx, &rolesrv.Members{Name: []string{s[1]}, Filter: role.FilterB})
	}
	if err != nil {
		return common.SendError(err.Error())
	}

	_, err = r.RoleClient.SyncToChatService(ctx, r.GetSyncRequest(sender, false))
	if err != nil {
		return common.SendError(err.Error())
	}

	_, outputName, err := r.MapName(ctx, []string{s[1]})

	if join {
		return common.SendSuccess(fmt.Sprintf("Added %s to %s", outputName[0], sig))
	} else {
		return common.SendSuccess(fmt.Sprintf("Removed %s from %s", outputName[0], sig))
	}
}
