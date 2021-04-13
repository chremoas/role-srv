package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	discord "github.com/chremoas/discord-gateway/proto"
	common "github.com/chremoas/services-common/command"
	"github.com/chremoas/services-common/config"
	"github.com/micro/go-micro"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/net/context"

	sq "github.com/Masterminds/squirrel"

	rolesrv "github.com/chremoas/role-srv/proto"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type rolesHandler struct {
	db sq.StatementBuilderType
	*zap.Logger
	clients   clientList
	namespace string
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

//var ignoredRoles []string

// Role keys are database columns we're allowed up update
var roleKeys = []string{"Name", "Color", "Hoist", "Position", "Permissions", "Joinable", "Managed", "Mentionable", "Sync"}
var roleTypes = []string{"internal", "discord"}

func NewRolesHandler(config *config.Configuration, service micro.Service, log *zap.Logger) (rolesrv.RolesHandler, error) {
	var (
		sugar = log.Sugar()
		err   error
		c     = service.Client()
	)

	clients := clientList{
		discord: discord.NewDiscordGatewayService(config.LookupService("gateway", "discord"), c),
	}

	//ignoredRoles = viper.GetStringSlice("bot.ignoredRoles")

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
		viper.GetString("database.host"),
		viper.GetInt("database.port"),
		viper.GetString("database.username"),
		viper.GetString("database.password"),
		viper.GetString("database.roledb"),
	)

	ldb, err := sqlx.Connect(viper.GetString("database.driver"), dsn)
	if err != nil {
		sugar.Error(err)
		return nil, err
	}

	err = ldb.Ping()
	if err != nil {
		sugar.Error(err)
		return nil, err
	}

	dbCache := sq.NewStmtCache(ldb)
	db := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(dbCache)

	// Ensure required permissions exist in the database
	var (
		requiredPermissions = map[string]string{
			"role_admins": "Role Admins",
			"sig_admins":  "SIG Admins",
		}
		id int
	)

	for k, v := range requiredPermissions {
		err = db.Select("id").
			From("permissions").
			Where(sq.Eq{"name": k}).
			Where(sq.Eq{"namespace": config.Namespace}).
			QueryRow().Scan(&id)

		switch err {
		case nil:
			sugar.Infof("%s (%d) found", k, id)
		case sql.ErrNoRows:
			sugar.Infof("%s NOT found, creating", k)
			err = db.Insert("permissions").
				Columns("namespace", "name", "description").
				Values(config.Namespace, k, v).
				Suffix("RETURNING \"id\"").
				QueryRow().Scan(&id)
			if err != nil {
				sugar.Error(err)
				return nil, err
			}
		default:
			sugar.Error(err)
			return nil, err
		}
	}

	rh := &rolesHandler{
		db:        db,
		Logger:    log,
		clients:   clients,
		namespace: config.Namespace,
	}

	// Start sync thread
	syncControl = make(chan syncData, 1)
	go rh.syncThread()

	return rh, nil
}

func (h *rolesHandler) GetRoleKeys(_ context.Context, _ *rolesrv.NilMessage, response *rolesrv.StringList) error {
	response.Value = roleKeys
	return nil
}

func (h *rolesHandler) GetRoleTypes(_ context.Context, _ *rolesrv.NilMessage, response *rolesrv.StringList) error {
	response.Value = roleTypes
	return nil
}

func (h *rolesHandler) AddRole(ctx context.Context, request *rolesrv.Role, _ *rolesrv.NilMessage) error {
	var (
		sugar = h.Sugar()
		count int
	)

	// Type, Name and ShortName are required so let's check for those
	if len(request.Type) == 0 {
		return errors.New("type is required")
	}

	if len(request.ShortName) == 0 {
		return errors.New("short name is required")
	}

	if len(request.Name) == 0 {
		return errors.New("name is required")
	}

	if !validListItem(request.Type, roleTypes) {
		return fmt.Errorf("`%s` isn't a valid Role Type", request.Type)
	}

	err := h.db.Select("COUNT(*)").
		From("roles").
		Where(sq.Eq{"role_nick": request.ShortName}).
		Where(sq.Eq{"namespace": h.namespace}).
		QueryRow().Scan(&count)
	if err != nil {
		return fmt.Errorf("error: %s", err)
	}

	if count > 0 {
		return fmt.Errorf("role `%s` (%s) already exists", request.Name, request.ShortName)
	}

	_, err = h.db.Insert("roles").
		Columns("namespace", "color", "hoist", "joinable", "managed", "mentionable", "name", "permissions",
			"position", "role_nick", "sig", "sync", "chat_type").
		Values(h.namespace, request.Color, request.Hoist, request.Joinable, request.Managed, request.Mentionable,
			request.Name, request.Permissions, request.Position, request.ShortName, request.Sig, request.Sync,
			request.Type).
		Query()
	if err != nil {
		return fmt.Errorf("error adding role: %s", err)
	}

	err = h.addDiscordRole(ctx, request.Name)
	if err != nil {
		sugar.Error(err)
		return err
	}

	return nil
}

func (h *rolesHandler) UpdateRole(ctx context.Context, request *rolesrv.UpdateInfo, _ *rolesrv.NilMessage) error {
	var sugar = h.Sugar()

	if len(request.Name) == 0 {
		return errors.New("name is required")
	}

	if len(request.Key) == 0 {
		return errors.New("key is required")
	}

	if len(request.Value) == 0 {
		return errors.New("value is required")
	}

	// check if key is in validListItem
	if !validListItem(request.Key, roleKeys) {
		return fmt.Errorf("`%s` isn't a valid role key", request.Key)
	}

	_, err := h.db.Update("roles").
		Set(request.Key, request.Value).
		Where(sq.Eq{"name": request.Name}).
		Where(sq.Eq{"namespace": h.namespace}).
		Query()
	if err != nil {
		newErr := fmt.Errorf("error updating role: %s", err)
		sugar.Error(newErr)
		return newErr
	}

	err = h.updateDiscordRole(ctx, request.Key, request.Value)
	if err != nil {
		sugar.Error(err)
		return err
	}

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

func (h *rolesHandler) RemoveRole(ctx context.Context, request *rolesrv.Role, _ *rolesrv.NilMessage) error {
	var sugar = h.Sugar()

	if len(request.ShortName) == 0 {
		return errors.New("short name is required")
	}

	_, err := h.db.Delete("roles").
		Where(sq.Eq{"role_nick": request.ShortName}).
		Where(sq.Eq{"namespace": h.namespace}).
		Query()
	if err != nil {
		newErr := fmt.Errorf("error deleting role: %s", err)
		sugar.Error(newErr)
		return newErr
	}

	err = h.removeDiscordRole(ctx, request.ShortName)
	if err != nil {
		sugar.Error(err)
		return err
	}

	return nil
}

type Role struct {
	Color       int32  `db:"color"`
	Hoist       bool   `db:"hoist"`
	Joinable    bool   `db:"joinable"`
	Managed     bool   `db:"managed"`
	Mentionable bool   `db:"mentionable"`
	Name        string `db:"name"`
	Permissions int32  `db:"permissions"`
	Position    int32  `db:"position"`
	ShortName   string `db:"role_nick"`
	Sig         bool   `db:"sig"`
	Sync        bool   `db:"sync"`
	Type        string `db:"chat_type"`
}

func (h *rolesHandler) getRoles() (*rolesrv.GetRolesResponse, error) {
	var (
		rs        rolesrv.GetRolesResponse
		sugar     = h.Sugar()
		charTotal int
	)

	rows, err := h.db.Select("color", "hoist", "joinable", "managed", "mentionable", "name", "permissions",
		"position", "role_nick", "sig", "sync").
		From("roles").
		Where(sq.Eq{"namespace": h.namespace}).
		Query()
	if err != nil {
		newErr := fmt.Errorf("error getting roles: %s", err)
		sugar.Error(newErr)
		return nil, newErr
	}
	defer func() {
		if err = rows.Close(); err != nil {
			sugar.Error(err)
		}
	}()

	var role rolesrv.Role
	for rows.Next() {
		err = rows.Scan(
			&role.Color,
			&role.Hoist,
			&role.Joinable,
			&role.Managed,
			&role.Mentionable,
			&role.Name,
			&role.Permissions,
			&role.Position,
			&role.ShortName,
			&role.Sig,
			&role.Sync,
		)
		if err != nil {
			newErr := fmt.Errorf("error scanning role row: %s", err)
			sugar.Error(newErr)
			return nil, newErr
		}
		charTotal += len(role.ShortName) + len(role.Name) + 15 // Guessing on bool excess
		rs.Roles = append(rs.Roles, &role)
		sugar.Infof("Added role: %+v", role)
	}

	if charTotal >= 2000 {
		return nil, errors.New("too many roles (exceeds Discord 2k character limit)")
	}

	return &rs, nil
}

func (h *rolesHandler) GetRoles(_ context.Context, _ *rolesrv.NilMessage, response *rolesrv.GetRolesResponse) error {
	roles, err := h.getRoles()
	if err != nil {
		return fmt.Errorf("error getting roles: %s", err)
	}

	response = roles
	return nil
}

func (h *rolesHandler) getRole(shortName string) (rolesrv.Role, error) {
	var (
		chatType, name                                   string
		color, position, permissions                     int32
		hoist, managed, mentionable, sig, joinable, sync bool
	)

	err := h.db.Select("chat_type", "name", "color", "hoist", "position", "permissions", "managed",
		"mentionable", "sig", "joinable", "sync").
		From("roles").
		Where(sq.Eq{"role_nick": shortName}).
		Where(sq.Eq{"namespace": h.namespace}).
		QueryRow().Scan(&chatType, &name, &color, &hoist, &position, &permissions, &managed,
		&mentionable, &sig, &joinable, &sync)
	if err != nil {
		return rolesrv.Role{}, fmt.Errorf("error getting roles: %s", err)
	}

	return rolesrv.Role{
		ShortName:   shortName,
		Type:        chatType,
		Name:        name,
		Color:       color,
		Hoist:       hoist,
		Position:    position,
		Permissions: permissions,
		Managed:     managed,
		Mentionable: mentionable,
		Sig:         sig,
		Joinable:    joinable,
		Sync:        sync,
	}, nil
}

func (h *rolesHandler) GetRole(_ context.Context, request *rolesrv.Role, response *rolesrv.Role) error {
	role, err := h.getRole(request.ShortName)
	if err != nil {
		return fmt.Errorf("error getting roles: %s", err)
	}

	response = &role
	return nil
}

func (h *rolesHandler) sendMessage(ctx context.Context, channelId, message string, sendMessage bool) {
	sugar := h.Sugar()

	if sendMessage {
		_, err := h.clients.discord.SendMessage(ctx, &discord.SendMessageRequest{ChannelId: channelId, Message: message})
		if err != nil {
			msg := fmt.Sprintf("sendMessage: %s", err.Error())
			sugar.Error(msg)
		}
	}
}

// This is part of sync
//func ignoreRole(roleName string) bool {
//	for i := range ignoredRoles {
//		log.Infof("Checking %s == %s", roleName, ignoredRoles[i])
//		if roleName == ignoredRoles[i] {
//			log.Infof("Ignoring: %s", ignoredRoles[i])
//			return true
//		}
//	}
//
//	return false
//}

func (h *rolesHandler) GetRoleMembership(_ context.Context, request *rolesrv.RoleMembershipRequest, response *rolesrv.RoleMembershipResponse) error {
	var (
		err    error
		id     int
		userID int
		sugar  = h.Sugar()
	)

	id, err = h.getRoleID(request.Name)
	if err != nil {
		return fmt.Errorf("error getting filter ID (%s): %s", request.Name, err)
	}

	rows, err := h.db.Select("user_id").
		From("filter_membership").
		InnerJoin("role_filters USING (filter)").
		Where(sq.Eq{"role_filters.role": id}).
		Where(sq.Eq{"namespace": h.namespace}).
		Query()
	if err != nil {
		sugar.Error(err)
		return err
	}

	for rows.Next() {
		err = rows.Scan(&userID)
		if err != nil {
			newErr := fmt.Errorf("error scanning user_id (%s): %s", request.Name, err)
			sugar.Error(newErr)
			return newErr
		}

		response.Members = append(response.Members, fmt.Sprintf("%d", userID))
	}

	return nil
}

func (h *rolesHandler) isRoleMember(userID, roleID string) (bool, error) {
	var (
		err   error
		count int
		sugar = h.Sugar()
	)

	err = h.db.Select("COUNT(*)").
		From("filter_membership").
		InnerJoin("role_filters USING (filter)").
		Where(sq.Eq{"role_filters.role": roleID}).
		Where(sq.Eq{"namespace": h.namespace}).
		Where(sq.Eq{"user_id": userID}).
		QueryRow().Scan(&count)
	if err != nil {
		sugar.Error(err)
		return false, err
	}

	if count == 1 {
		return true, nil
	}

	return false, nil
}

//
// Filter related stuff
//

func (h *rolesHandler) GetFilters(_ context.Context, _ *rolesrv.NilMessage, response *rolesrv.FilterList) error {
	var (
		sugar             = h.Sugar()
		name, description string
		charTotal         int
	)

	rows, err := h.db.Select("name", "description").
		From("filters").
		Where(sq.Eq{"namespace": h.namespace}).
		Query()
	if err != nil {
		newErr := fmt.Errorf("error getting filters: %s", err)
		sugar.Error(newErr)
		return newErr
	}
	defer func() {
		if err = rows.Close(); err != nil {
			sugar.Error(err)
		}
	}()

	for rows.Next() {
		err = rows.Scan(&name, &description)
		if err != nil {
			newErr := fmt.Errorf("error scanning filter row: %s", err)
			sugar.Error(newErr)
			return newErr
		}
		charTotal += len(name) + len(description)
		response.FilterList = append(response.FilterList, &rolesrv.Filter{
			Name:        name,
			Description: description,
		})
	}

	if len(response.FilterList) == 0 {
		return errors.New("no filters")
	}

	if charTotal >= 2000 {
		return errors.New("too many filters (exceeds Discord 2k character limit)")
	}

	return nil
}

func (h *rolesHandler) AddFilter(_ context.Context, request *rolesrv.Filter, _ *rolesrv.NilMessage) error {
	var sugar = h.Sugar()

	if len(request.Name) == 0 {
		return errors.New("name is required")
	}

	if len(request.Description) == 0 {
		return errors.New("description is required")
	}

	_, err := h.db.Insert("filters").
		Columns("namespace", "name", "description").
		Values(h.namespace, request.Name, request.Description).
		Query()
	if err != nil {
		// TODO: Catch `pq: duplicate key value violates unique constraint` and return a friendly error
		newErr := fmt.Errorf("error adding filter (%s): %s", request.Name, err)
		sugar.Error(newErr)
		return newErr
	}

	return nil
}

func (h rolesHandler) getFilterID(name string) (int, error) {
	var (
		err   error
		id    int
		sugar = h.Sugar()
	)

	if name == "" {
		return 0, errors.New("getFilterID: name is required")
	}

	sugar.Infof("Looking up filter: %s", name)
	err = h.db.Select("id").
		From("filters").
		Where(sq.Eq{"name": name}).
		Where(sq.Eq{"namespace": h.namespace}).
		QueryRow().Scan(&id)

	return id, err
}

func (h rolesHandler) getRoleID(name string) (int, error) {
	var (
		err error
		id  int
	)

	err = h.db.Select("id").
		From("roles").
		Where(sq.Eq{"name": name}).
		Where(sq.Eq{"namespace": h.namespace}).
		QueryRow().Scan(&id)

	return id, err
}

func (h *rolesHandler) RemoveFilter(_ context.Context, request *rolesrv.Filter, _ *rolesrv.NilMessage) error {
	var (
		err   error
		id    int
		sugar = h.Sugar()
	)

	id, err = h.getFilterID(request.Name)
	if err != nil {
		newErr := fmt.Errorf("error getting filter ID (%s): %s", request.Name, err)
		sugar.Error(newErr)
		return newErr
	}

	// Delete all the filter members
	_, err = h.db.Delete("filter_membership").
		Where(sq.Eq{"filter": id}).
		Where(sq.Eq{"namespace": h.namespace}).
		Query()
	if err != nil {
		newErr := fmt.Errorf("error deleting filter members (%s): %s", request.Name, err)
		sugar.Error(newErr)
		return newErr
	}

	// Delete the filter
	_, err = h.db.Delete("filters").
		Where(sq.Eq{"name": request.Name}).
		Where(sq.Eq{"namespace": h.namespace}).
		Query()
	if err != nil {
		newErr := fmt.Errorf("error deleting filter (%s): %s", request.Name, err)
		sugar.Error(newErr)
		return newErr
	}

	return nil
}

// TODO: Rename this GetFilterMembers
func (h *rolesHandler) GetMembers(_ context.Context, request *rolesrv.Filter, response *rolesrv.MemberList) error {
	var (
		err    error
		id     int
		userId int
		sugar  = h.Sugar()
	)

	if request.Name == "" {
		return errors.New("name is required")
	}

	id, err = h.getFilterID(request.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			sugar.Info("Filter '%s' doesn't exist in namespace '%s'", request.Name, h.namespace)
			return fmt.Errorf("filter doesn't exist: %s", request.Name)
		}
		newErr := fmt.Errorf("error getting filter ID (%s): %s", request.Name, err)
		sugar.Error(newErr)
		return newErr
	}

	rows, err := h.db.Select("user_id").
		From("filter_membership").
		Where(sq.Eq{"id": id}).
		Where(sq.Eq{"namespace": h.namespace}).
		Query()
	if err != nil {
		newErr := fmt.Errorf("error getting filter members (%s): %s", request.Name, err)
		sugar.Error(newErr)
		return newErr
	}
	defer func() {
		if err = rows.Close(); err != nil {
			sugar.Error(err)
		}
	}()

	for rows.Next() {
		err = rows.Scan(&userId)
		if err != nil {
			newErr := fmt.Errorf("error scanning filter member id (%s): %s", request.Name, err)
			sugar.Error(newErr)
			return newErr
		}

		// TODO: Maybe this should be a int in the protobuf
		response.Members = append(response.Members, fmt.Sprintf("%d", userId))
	}

	if len(response.Members) == 0 {
		return errors.New("no filter members")
	}

	return nil
}

// TODO: Rename this AddFilterMembers
func (h *rolesHandler) AddMembers(ctx context.Context, request *rolesrv.Members, _ *rolesrv.NilMessage) error {
	var (
		err     error
		id      int
		roles   []string
		addList []string
		sugar   = h.Sugar()
		cancel  func()
	)

	ctx, cancel = context.WithTimeout(ctx, time.Second*20)
	defer cancel()

	id, err = h.getFilterID(request.Filter)
	if err != nil {
		newErr := fmt.Errorf("error getting filter ID (%s): %s", request.Name, err)
		sugar.Error(newErr)
		return newErr
	}

	for _, member := range request.Name {
		_, err = h.db.Insert("filter_membership").
			Columns("namespace", "filter", "user_id").
			Values(h.namespace, id, member).Query()
		if err != nil {
			// TODO: I need to catch and ignore already exists
			newErr := fmt.Errorf("error adding filter members (%s): %s", request.Name, err)
			sugar.Error(newErr)
			return newErr
		}
		// TODO: do sync
		// 1) Get all roles that use this filter
		roles, err = h.getFilterRoles(id)
		if err != nil {
			newErr := fmt.Errorf("error getting role list for filter (%d): %s", id, err)
			sugar.Error(newErr)
			return newErr
		}

		// 2) check membership of roles
		for _, r := range roles {
			isMember, err := h.isRoleMember(member, r)
			if err != nil {
				newErr := fmt.Errorf("error checking role membership (%s:%d): %s", member, r, err)
				sugar.Error(newErr)
				return newErr
			}
			if isMember {
				addList = append(addList, r)
			}
		}

		//	3) update roles (remove)
		_, err = h.clients.discord.UpdateMember(ctx, &discord.UpdateMemberRequest{
			Operation: discord.MemberUpdateOperation_ADD_OR_UPDATE_ROLES,
			UserId:    member,
			RoleIds:   addList,
		})
	}

	return nil
}

// TODO: Rename this RemoveFilterMembers
func (h *rolesHandler) RemoveMembers(ctx context.Context, request *rolesrv.Members, _ *rolesrv.NilMessage) error {
	var (
		err        error
		id         int
		roles      []string
		removeList []string
		sugar      = h.Sugar()
		cancel     func()
	)

	ctx, cancel = context.WithTimeout(ctx, time.Second*20)
	defer cancel()

	id, err = h.getFilterID(request.Filter)
	if err != nil {
		newErr := fmt.Errorf("error getting filter ID (%s): %s", request.Name, err)
		sugar.Error(newErr)
		return newErr
	}

	for _, member := range request.Name {
		_, err = h.db.Delete("filter_membership").
			Where(sq.Eq{"user_id": member}).
			Where(sq.Eq{"filter": id}).
			Where(sq.Eq{"namespace": h.namespace}).
			Query()
		if err != nil {
			newErr := fmt.Errorf("error removing filter members (%s): %s", request.Name, err)
			sugar.Error(newErr)
			return newErr
		}

		// TODO: do sync
		// 1) Get all roles that use this filter
		roles, err = h.getFilterRoles(id)
		if err != nil {
			newErr := fmt.Errorf("error getting role list for filter (%d): %s", id, err)
			sugar.Error(newErr)
			return newErr
		}

		// 2) check membership of roles
		for _, r := range roles {
			isMember, err := h.isRoleMember(member, r)
			if err != nil {
				newErr := fmt.Errorf("error checking role membership (%s:%d): %s", member, r, err)
				sugar.Error(newErr)
				return newErr
			}
			if isMember {
				removeList = append(removeList, r)
			}
		}

		//	3) update roles (remove)
		_, err = h.clients.discord.UpdateMember(ctx, &discord.UpdateMemberRequest{
			Operation: discord.MemberUpdateOperation_REMOVE_ROLES,
			UserId:    member,
			RoleIds:   removeList,
		})
	}

	return nil
}

func (h *rolesHandler) getFilterRoles(filterID int) ([]string, error) {
	var (
		sugar  = h.Sugar()
		roleID int
		roles  []string
	)

	rows, err := h.db.Select("role").
		From("role_filters").
		Where(sq.Eq{"filter": filterID}).
		Where(sq.Eq{"namespace": h.namespace}).
		Query()
	if err != nil {
		newErr := fmt.Errorf("error getting roles that use filterID (%d): %s", filterID, err)
		sugar.Error(newErr)
		return nil, newErr
	}
	defer func() {
		if err = rows.Close(); err != nil {
			sugar.Error(err)
		}
	}()

	for rows.Next() {
		err = rows.Scan(&roleID)
		if err != nil {
			newErr := fmt.Errorf("error scanning filter role id (%d): %s", filterID, err)
			sugar.Error(newErr)
			return nil, newErr
		}

		roles = append(roles, fmt.Sprint("%d", roleID))
	}

	return roles, nil
}

func (h *rolesHandler) GetDiscordUser(ctx context.Context, request *rolesrv.GetDiscordUserRequest, response *rolesrv.GetDiscordUserResponse) error {
	members, err := h.clients.discord.GetAllMembersAsSlice(ctx, &discord.GetAllMembersRequest{})
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

	return errors.New("user not found")
}

func (h *rolesHandler) GetDiscordUserList(ctx context.Context, _ *rolesrv.NilMessage, response *rolesrv.GetDiscordUserListResponse) error {
	members, err := h.clients.discord.GetAllMembersAsSlice(ctx, &discord.GetAllMembersRequest{})
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

func (h *rolesHandler) SyncToChatService(_ context.Context, _ *rolesrv.SyncRequest, _ *rolesrv.NilMessage) error {
	//syncControl <- syncData{ChannelId: request.ChannelId, UserId: request.UserId, SendMessage: request.SendMessage}
	// This is a no-op now
	return nil
}

func (h *rolesHandler) sendDualMessage(msg, channelId string, sendMessage bool) {
	ctx := context.Background()
	sugar := h.Sugar()

	sugar.Info(msg)
	h.sendMessage(ctx, channelId, common.SendSuccess(msg), sendMessage)
}

func (h *rolesHandler) syncThread() {
	for {
		request := <-syncControl
		h.sendDualMessage("This command does nothing for now", request.ChannelId, request.SendMessage)

		//t1 := time.Now()
		//
		//h.sendDualMessage("Starting Role Sync", request.ChannelId, request.SendMessage)
		//
		//err := h.syncRoles(request.ChannelId, request.UserId, request.SendMessage)
		//if err != nil {
		//	h.Logger.Error(fmt.Sprintf("syncRoles error: %v", err))
		//}
		//
		//msg := fmt.Sprintf("Completed Role Sync [%s]", time.Since(t1))
		//h.sendDualMessage(msg, request.ChannelId, request.SendMessage)
		//
		//t2 := time.Now()
		//h.sendDualMessage("Starting Member Sync", request.ChannelId, request.SendMessage)
		//
		//err = h.syncMembers(request.ChannelId, request.UserId, request.SendMessage)
		//if err != nil {
		//	h.Logger.Error(fmt.Sprintf("syncMembers error: %v", err))
		//}
		//
		//msg = fmt.Sprintf("Completed Member Sync [%s]", time.Since(t2))
		//h.sendDualMessage(msg, request.ChannelId, request.SendMessage)
		//
		//msg = fmt.Sprintf("Completed All Syncing [%s]", time.Since(t1))
		//h.sendDualMessage(msg, request.ChannelId, request.SendMessage)

		// Sleep 15 minutes after each sync
		//time.Sleep(15 * time.Minute)
	}
}

func (h *rolesHandler) ListUserRoles(_ context.Context, request *rolesrv.ListUserRolesRequest, response *rolesrv.ListUserRolesResponse) error {
	var (
		shortName, chatType, name                        string
		color, position, permissions                     int32
		hoist, managed, mentionable, sig, joinable, sync bool

		sugar = h.Sugar()
	)

	rows, err := h.db.Select("roles.role_nick", "roles.chat_type", "roles.name", "roles.color",
		"roles.hoist", "roles.position", "roles.permissions", "roles.managed", "roles.mentionable",
		"roles.sig", "roles.joinable", "roles.sync").
		From("filters").
		Join("filter_membership ON filters.id = filter_membership.filter").
		Join("role_filters ON filters.id = role_filters.filter").
		Join("roles ON role_filters.role = roles.id").
		Where(sq.Eq{"filter_membership.user_id": request.UserId}).
		Where(sq.Eq{"filters.namespace": h.namespace}).
		Query()
	if err != nil {
		newErr := fmt.Errorf("error getting user roles (%s): %s", request.UserId, err)
		sugar.Error(newErr)
		return newErr
	}
	defer func() {
		if err = rows.Close(); err != nil {
			sugar.Error(err)
		}
	}()

	for rows.Next() {
		err = rows.Scan(&shortName, &chatType, &name, &color, &hoist, &position, &permissions, &managed,
			&mentionable, &sig, &joinable, &sync)
		if err != nil {
			return fmt.Errorf("error scanning role for userID (%s): %s", request.UserId, err)
		}

		response.Roles = append(response.Roles, &rolesrv.Role{
			ShortName:   shortName,
			Type:        chatType,
			Name:        name,
			Color:       color,
			Hoist:       hoist,
			Position:    position,
			Permissions: permissions,
			Managed:     managed,
			Mentionable: mentionable,
			Sig:         sig,
			Joinable:    joinable,
			Sync:        sync,
		})
	}

	return nil
}
