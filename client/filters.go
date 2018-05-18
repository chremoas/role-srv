package client

import (
	"context"
	"fmt"
	rolesrv "github.com/chremoas/role-srv/proto"
	common "github.com/chremoas/services-common/command"
	"bytes"
	"time"
)

func (r Roles) AddFilter(ctx context.Context, sender, filterName, filterDescription string) string {
	if len(filterDescription) > 0 && filterDescription[0] == '"' {
		filterDescription = filterDescription[1:]
	}

	if len(filterDescription) > 0 && filterDescription[len(filterDescription)-1] == '"' {
		filterDescription = filterDescription[:len(filterDescription)-1]
	}

	canPerform, err := r.Permissions.CanPerform(ctx, sender)
	if err != nil {
		return common.SendFatal(err.Error())
	}

	if !canPerform {
		return common.SendError("User doesn't have permission to this command")
	}

	_, err = r.RoleClient.AddFilter(ctx, &rolesrv.Filter{Name: filterName, Description: filterDescription})
	if err != nil {
		return common.SendFatal(err.Error())
	}

	return common.SendSuccess(fmt.Sprintf("Added: %s\n", filterName))
}

func (r Roles) ListFilters(ctx context.Context) string {
	var buffer bytes.Buffer
	filters, err := r.RoleClient.GetFilters(ctx, &rolesrv.NilMessage{})

	if err != nil {
		return common.SendFatal(err.Error())
	}

	if len(filters.FilterList) == 0 {
		return common.SendError("No Filters\n")
	}

	buffer.WriteString("Filters:\n")
	for filter := range filters.FilterList {
		buffer.WriteString(fmt.Sprintf("\t%s: %s\n", filters.FilterList[filter].Name, filters.FilterList[filter].Description))
	}

	return fmt.Sprintf("```%s```", buffer.String())
}

func (r Roles) RemoveFilter(ctx context.Context, sender, name string) string {
	canPerform, err := r.Permissions.CanPerform(ctx, sender)
	if err != nil {
		return common.SendFatal(err.Error())
	}

	if !canPerform {
		return common.SendError("User doesn't have permission to this command")
	}

	_, err = r.RoleClient.RemoveFilter(ctx, &rolesrv.Filter{Name: name})
	if err != nil {
		return common.SendFatal(err.Error())
	}

	return common.SendSuccess(fmt.Sprintf("Removed: %s\n", name))
}

func (r Roles) ListMembers(ctx context.Context, name string) string {
	t := time.Now()

	members, err := r.RoleClient.GetMembers(ctx, &rolesrv.Filter{Name: name})
	if err != nil {
		return common.SendFatal(err.Error())
	}

	if len(members.Members) == 0 {
		return common.SendError("No members in filter")
	}

	buffer, _, err := r.MapName(ctx, members.Members)
	if err != nil {
		return common.SendFatal(err.Error())
	}

	if buffer.Len() == 0 {
		return "```Empty list```\n"
	}

	buffer.WriteString(fmt.Sprintf("\nListeMembers returned in %s", time.Since(t)))
	return fmt.Sprintf("```%s Members:\n%s```\n", name, buffer.String())
}

func (r Roles) RemoveAllMembers(ctx context.Context, name, sender string) error {
	members, err := r.RoleClient.GetMembers(ctx, &rolesrv.Filter{Name: name})
	if err != nil {
		return err
	}

	_, err = r.RoleClient.RemoveMembers(ctx, &rolesrv.Members{Name: members.Members, Filter: name})
	if err != nil {
		return err
	}

	_, err = r.RoleClient.SyncToChatService(ctx, r.GetSyncRequest(sender, false))
	if err != nil {
		return err
	}

	return nil
}

func (r Roles) AddMember(ctx context.Context, sender, user, filter string) string {
	canPerform, err := r.Permissions.CanPerform(ctx, sender)
	if err != nil {
		return common.SendFatal(err.Error())
	}

	if !canPerform {
		return common.SendError("User doesn't have permission to this command")
	}

	_, err = r.RoleClient.AddMembers(ctx,
		&rolesrv.Members{Name: []string{user}, Filter: filter})
	if err != nil {
		return common.SendFatal(err.Error())
	}

	_, err = r.RoleClient.SyncToChatService(ctx, r.GetSyncRequest(sender, false))
	if err != nil {
		return common.SendFatal(err.Error())
	}

	return common.SendSuccess(fmt.Sprintf("Added '%s' to '%s'\n", user, filter))
}

func (r Roles) RemoveMember(ctx context.Context, sender, user, filter string) string {
	canPerform, err := r.Permissions.CanPerform(ctx, sender)
	if err != nil {
		return common.SendFatal(err.Error())
	}

	if !canPerform {
		return common.SendError("User doesn't have permission to this command")
	}

	_, err = r.RoleClient.RemoveMembers(ctx,
		&rolesrv.Members{Name: []string{user}, Filter: filter})
	if err != nil {
		return common.SendFatal(err.Error())
	}

	_, err = r.RoleClient.SyncToChatService(ctx, r.GetSyncRequest(sender, false))
	if err != nil {
		return common.SendFatal(err.Error())
	}

	return common.SendSuccess(fmt.Sprintf("Removed '%s' from '%s'\n", user, filter))
}


func (r Roles) SyncMembers(ctx context.Context, sender string) string {
	//var buffer bytes.Buffer
	_, err := r.RoleClient.SyncToChatService(ctx, r.GetSyncRequest(sender, true))

	if err != nil {
		return common.SendFatal(err.Error())
	}

	return common.SendSuccess("Done")
}