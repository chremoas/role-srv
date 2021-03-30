package handler

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	discord "github.com/chremoas/discord-gateway/proto"
)

var matchDiscordError = regexp.MustCompile(`^The role '.*' already exists$`)

func (h *rolesHandler) addDiscordRole(ctx context.Context, name string) error {
	var sugar = h.Sugar()

	_, err := h.clients.discord.CreateRole(ctx, &discord.CreateRoleRequest{Name: name})
	if err != nil {
		if matchDiscordError.MatchString(err.Error()) {
			// The role list was cached most likely so we'll pretend we didn't try
			// to create it just now. -brian
		} else {
			msg := fmt.Sprintf("addDiscordRole '%s': %s", name, err.Error())
			sugar.Error(msg)
			return err
		}
	}

	return nil
}

func (h *rolesHandler) removeDiscordRole(ctx context.Context, name string) error {
	var sugar = h.Sugar()

	_, err := h.clients.discord.DeleteRole(ctx, &discord.DeleteRoleRequest{Name: name})
	if err != nil {
		msg := fmt.Sprintf("removeDiscordRole '%s': %s", name, err.Error())
		sugar.Error(msg)
		return err
	}

	return nil
}

func (h *rolesHandler) updateDiscordRole(ctx context.Context, key, value string) error {
	var (
		sugar       = h.Sugar()
		editRequest = &discord.EditRoleRequest{}
	)

	switch key {
	case "Color":
		editRequest.Color, _ = strconv.ParseInt(value, 10, 64)
	case "Permissions":
		editRequest.Perm, _ = strconv.ParseInt(value, 10, 64)
	case "Position":
		editRequest.Position, _ = strconv.ParseInt(value, 10, 64)
	case "Hoist":
		editRequest.Hoist, _ = strconv.ParseBool(value)
	case "mentionable":
		editRequest.Mention, _ = strconv.ParseBool(value)
	case "Managed":
		editRequest.Managed, _ = strconv.ParseBool(value)
	case "Name":
		editRequest.Name = value
	}

	longCtx, cancel := context.WithTimeout(ctx, time.Minute*5)
	defer cancel()

	_, err := h.clients.discord.EditRole(longCtx, editRequest)
	if err != nil {
		msg := fmt.Sprintf("updateDiscordRole: %s", err.Error())
		sugar.Error(msg)
		return err
	}

	return nil
}

func (h *rolesHandler) updateUserRoles(ctx context.Context, userID string, roleID []string) error {
	var sugar = h.Sugar()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	_, err := h.clients.discord.UpdateMember(ctx, &discord.UpdateMemberRequest{
		Operation: discord.MemberUpdateOperation_ADD_OR_UPDATE_ROLES,
		UserId:    userID,
		RoleIds:   roleID,
	})
	if err != nil {
		msg := fmt.Sprintf("updateUserRoles: %s", err.Error())
		sugar.Error(msg)
		return err
	}

	return nil
}
