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
	"github.com/prometheus/common/log"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/net/context"

	sq "github.com/Masterminds/squirrel"

	rolesrv "github.com/chremoas/role-srv/proto"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type rolesHandler struct {
	db *sqlx.DB
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
var ignoredRoles []string

// Role keys are database columns we're allowed up update
var roleKeys = []string{"Name", "Color", "Hoist", "Position", "Permissions", "Joinable", "Managed", "Mentionable", "Sync"}
var roleTypes = []string{"internal", "discord"}

func NewRolesHandler(config *config.Configuration, service micro.Service, log *zap.Logger) rolesrv.RolesHandler {
	var (
		sugar = log.Sugar()
		err   error
		c     = service.Client()
	)

	clients := clientList{
		discord: discord.NewDiscordGatewayService(config.LookupService("gateway", "discord"), c),
	}

	ignoredRoles = viper.GetStringSlice("bot.ignoredRoles")

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s %s",
		viper.GetString("database.host"),
		viper.GetInt("database.port"),
		viper.GetString("database.username"),
		viper.GetString("database.password"),
		viper.GetString("database.database"),
		viper.GetString("database.options"),
	)

	db, err := sqlx.Connect(viper.GetString("database.driver"), dsn)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	// Ensure required permissions exist in the database
	var (
		requiredPermissions = map[string]string{
			"role_admins": "Role Admins",
			"sig_admins":  "SIG Admins",
		}
		id int
	)

	tx := db.MustBegin()

	for k, v := range requiredPermissions {
		err = tx.Get(&id, "SELECT id FROM permissions WHERE namespace = $1 AND name = $2", config.Namespace, k)

		switch err {
		case nil:
			sugar.Infof("%s found: %d", k, id)
		case sql.ErrNoRows:
			sugar.Infof("%s NOT found, creating", k)
			tx.MustExec("INSERT INTO permissions (namespace, name, description) VALUES ($1, $2, $3)", config.Namespace, k, v)
		default:
			sugar.Infof("error: %s\n", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		panic(err)
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

	return rh
}

func (h *rolesHandler) GetRoleKeys(_ context.Context, _ *rolesrv.NilMessage, response *rolesrv.StringList) error {
	response.Value = roleKeys
	return nil
}

func (h *rolesHandler) GetRoleTypes(_ context.Context, _ *rolesrv.NilMessage, response *rolesrv.StringList) error {
	response.Value = roleTypes
	return nil
}

func (h *rolesHandler) AddRole(_ context.Context, request *rolesrv.Role, _ *rolesrv.NilMessage) error {
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

	rows, err := sq.Select("id").
		From("roles").
		Where(sq.Eq{"role_nick": request.ShortName}).
		Where(sq.Eq{"namespace": h.namespace}).
		RunWith(h.db).Query()
	if rows != nil {
		defer rows.Close()
	}
	//err := h.db.Get(&id, "SELECT id FROM roles WHERE role_nick = $1", request.ShortName)
	switch err {
	case nil:
		return fmt.Errorf("role `%s` (%s) already exists", request.Name, request.ShortName)
	case sql.ErrNoRows:
		_, err = sq.Insert("roles").
			Columns("namespace", "name", "role_nick", "chat_type").
			Values(h.namespace, request.Name, request.ShortName, request.Type).
			RunWith(h.db).Query()
		if err != nil {
			return fmt.Errorf("error adding role: %s", err)
		}
	default:
		return fmt.Errorf("error: %s", err)
	}

	return nil
}

func (h *rolesHandler) UpdateRole(_ context.Context, request *rolesrv.UpdateInfo, _ *rolesrv.NilMessage) error {
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

	_, err := sq.Update("roles").
		Set(request.Key, request.Value).
		Where(sq.Eq{"name": request.Name}).
		Where(sq.Eq{"namespace": h.namespace}).
		RunWith(h.db).Query()
	if err != nil {
		return fmt.Errorf("error updating role: %s", err)
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

func (h *rolesHandler) RemoveRole(_ context.Context, request *rolesrv.Role, _ *rolesrv.NilMessage) error {
	if len(request.ShortName) == 0 {
		return errors.New("short name is required")
	}

	_, err := sq.Delete("roles").
		Where(sq.Eq{"role_nick": request.ShortName}).
		Where(sq.Eq{"namespace": h.namespace}).
		RunWith(h.db).Query()
	if err != nil {
		return fmt.Errorf("error deleting role: %s", err)
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

func (h *rolesHandler) GetRoles(_ context.Context, _ *rolesrv.NilMessage, response *rolesrv.GetRolesResponse) error {
	rows, err := sq.Select("*").
		From("roles").
		RunWith(h.db).Query()
	if err != nil {
		return fmt.Errorf("error getting roles: %s", err)
	}
	defer rows.Close()

	var role Role
	for rows.Next() {
		err = rows.Scan(&role)
		if err != nil {
			return fmt.Errorf("error scanning role row: %s", err)
		}
		response.Roles = append(response.Roles, &rolesrv.Role{
			ShortName: role.ShortName,
			Name:      role.Name,
			Sig:       role.Sig,
			Joinable:  role.Joinable,
			Sync:      role.Sync,
		})
	}

	return nil
}

func (h *rolesHandler) GetRole(_ context.Context, request *rolesrv.Role, response *rolesrv.Role) error {
	var role Role

	err := sq.Select("*").
		From("roles").
		Where(sq.Eq{"role_nick": request.ShortName}).
		Where(sq.Eq{"namespace": h.namespace}).
		RunWith(h.db).QueryRow().Scan(&role)
	if err != nil {
		return fmt.Errorf("error getting roles: %s", err)
	}

	response = &rolesrv.Role{
		ShortName:   role.ShortName,
		Type:        role.Type,
		Name:        role.Name,
		Color:       role.Color,
		Hoist:       role.Hoist,
		Position:    role.Position,
		Permissions: role.Permissions,
		Managed:     role.Managed,
		Mentionable: role.Mentionable,
		Sig:         role.Sig,
		Joinable:    role.Joinable,
		Sync:        role.Sync,
	}
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

func ignoreRole(roleName string) bool {
	for i := range ignoredRoles {
		log.Infof("Checking %s == %s", roleName, ignoredRoles[i])
		if roleName == ignoredRoles[i] {
			log.Infof("Ignoring: %s", ignoredRoles[i])
			return true
		}
	}

	return false
}

func (h *rolesHandler) GetRoleMembership(_ context.Context, request *rolesrv.RoleMembershipRequest, response *rolesrv.RoleMembershipResponse) error {
	var (
		err error
		id int
		user_id int
	)

	id, err = h.getRoleID(request.Name)
	if err != nil {
		return fmt.Errorf("error getting filter ID (%s): %s", request.Name, err)
	}

	rows, err := sq.Select("user_id").
		From("filter_membership").
		Join("role_filters USING (filter)").
		Where(sq.Eq{"role_filters.role": id}).
		Where(sq.Eq{"namespace": h.namespace}).
		RunWith(h.db).Query()

	for rows.Next() {
		err := rows.Scan(&user_id)
		if err != nil {
			return fmt.Errorf("error scanning user_id (%s): %s", request.Name, err)
		}

		response.Members = append(response.Members, fmt.Sprintf("%d", user_id))
	}

	return nil
}

//
// Filter related stuff
//

type Filter struct {
	Name        string `db:"name"`
	Description string `db:"description"`
}

func (h *rolesHandler) GetFilters(_ context.Context, _ *rolesrv.NilMessage, response *rolesrv.FilterList) error {
	rows, err := sq.Select("*").
		From("filters").
		RunWith(h.db).Query()
	if err != nil {
		return fmt.Errorf("error getting filters: %s", err)
	}
	defer rows.Close()

	var filter Filter
	for rows.Next() {
		err = rows.Scan(&filter)
		if err != nil {
			return fmt.Errorf("error scanning filter row: %s", err)
		}
		response.FilterList = append(response.FilterList, &rolesrv.Filter{
			Name:        filter.Name,
			Description: filter.Description,
		})
	}

	return nil
}

func (h *rolesHandler) AddFilter(_ context.Context, request *rolesrv.Filter, _ *rolesrv.NilMessage) error {
	if len(request.Name) == 0 {
		return errors.New("name is required")
	}

	if len(request.Description) == 0 {
		return errors.New("description is required")
	}

	_, err := sq.Insert("filters").
		Columns("name", "description").
		Values(request.Name, request.Description).
		RunWith(h.db).Query()
	if err != nil {
		return fmt.Errorf("error adding filter (%s): %s", request.Name, err)
	}

	return nil
}

func (h rolesHandler) getFilterID(name string) (int, error) {
	var (
		err error
		id int
	)

	err = sq.Select("id").
		From("filters").
		Where(sq.Eq{"name": name}).
		Where(sq.Eq{"namespace": h.namespace}).
		RunWith(h.db).QueryRow().Scan(&id)

	return id, err
}

func (h rolesHandler) getRoleID(name string) (int, error) {
	var (
		err error
		id int
	)

	err = sq.Select("id").
		From("roles").
		Where(sq.Eq{"name": name}).
		Where(sq.Eq{"namespace": h.namespace}).
		RunWith(h.db).QueryRow().Scan(&id)

	return id, err
}

func (h *rolesHandler) RemoveFilter(_ context.Context, request *rolesrv.Filter, _ *rolesrv.NilMessage) error {
	var (
		err error
		id int
	)

	id, err = h.getFilterID(request.Name)
	if err != nil {
		return fmt.Errorf("error getting filter ID (%s): %s", request.Name, err)
	}

	// Delete all the filter members
	_, err = sq.Delete("filter_members").
		Where(sq.Eq{"filter": id}).
		Where(sq.Eq{"namespace": h.namespace}).
		RunWith(h.db).Query()
	if err != nil {
		return fmt.Errorf("error deleting filter members (%s): %s", request.Name, err)
	}

	// Delete the filter
	_, err = sq.Delete("filters").
		Where(sq.Eq{"name": request.Name}).
		Where(sq.Eq{"namespace": h.namespace}).
		RunWith(h.db).Query()
	if err != nil {
		return fmt.Errorf("error deleting filter (%s): %s", request.Name, err)
	}

	return nil
}

// TODO: Rename this GetFilterMembers
func (h *rolesHandler) GetMembers(_ context.Context, request *rolesrv.Filter, response *rolesrv.MemberList) error {
	var (
		err error
		id int
		userId int
	)

	id, err = h.getFilterID(request.Name)
	if err != nil {
		return fmt.Errorf("error getting filter ID (%s): %s", request.Name, err)
	}

	rows, err := sq.Select("user_id").
		From("filter_membership").
		Where(sq.Eq{"id": id}).
		Where(sq.Eq{"namespace": h.namespace}).
		RunWith(h.db).Query()
	if err != nil {
		return fmt.Errorf("error getting filter members (%s): %s", request.Name, err)
	}

	for rows.Next() {
		err = rows.Scan(userId)
		if err != nil {
			return fmt.Errorf("error scanning filter member id (%s): %s", request.Name, err)
		}

		// TODO: Maybe this should be a int in the protobuf
		response.Members = append(response.Members, fmt.Sprintf("%d", userId))
	}

	return nil
}

// TODO: Rename this AddFilterMembers
func (h *rolesHandler) AddMembers(_ context.Context, request *rolesrv.Members, _ *rolesrv.NilMessage) error {
	var (
		err error
		id int
	)

	id, err = h.getFilterID(request.Filter)
	if err != nil {
		return fmt.Errorf("error getting filter ID (%s): %s", request.Name, err)
	}

	for _, member := range request.Name {
		_, err = sq.Insert("filter_membership").
			Columns("namespace", "filter", "user_id").
			Values(h.namespace, id, member).RunWith(h.db).Query()
		if err != nil {
			// TODO: I need to catch and ignore already exists
			return fmt.Errorf("error adding filter members (%s): %s", request.Name, err)
		}
	}

	return nil
}

// TODO: Rename this RemoveFilterMembers
func (h *rolesHandler) RemoveMembers(_ context.Context, request *rolesrv.Members, _ *rolesrv.NilMessage) error {
	var (
		err error
		id int
	)

	id, err = h.getFilterID(request.Filter)
	if err != nil {
		return fmt.Errorf("error getting filter ID (%s): %s", request.Name, err)
	}

	for _, member := range request.Name {
		_, err = sq.Delete("filter_membership").
			Where(sq.Eq{"user_id": member}).
			Where(sq.Eq{"filter": id}).
			Where(sq.Eq{"namespace": h.namespace}).
			RunWith(h.db).Query()
		if err != nil {
			return fmt.Errorf("error removing filter members (%s): %s", request.Name, err)
		}
	}
	return nil
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

	return errors.New("User not found")
}

func (h *rolesHandler) GetDiscordUserList(ctx context.Context, request *rolesrv.NilMessage, response *rolesrv.GetDiscordUserListResponse) error {
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

func (h *rolesHandler) SyncToChatService(ctx context.Context, request *rolesrv.SyncRequest, response *rolesrv.NilMessage) error {
	syncControl <- syncData{ChannelId: request.ChannelId, UserId: request.UserId, SendMessage: request.SendMessage}
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

		t1 := time.Now()

		h.sendDualMessage("Starting Role Sync", request.ChannelId, request.SendMessage)

		err := h.syncRoles(request.ChannelId, request.UserId, request.SendMessage)
		if err != nil {
			h.Logger.Error(fmt.Sprintf("syncRoles error: %v", err))
		}

		msg := fmt.Sprintf("Completed Role Sync [%s]", time.Since(t1))
		h.sendDualMessage(msg, request.ChannelId, request.SendMessage)

		t2 := time.Now()
		h.sendDualMessage("Starting Member Sync", request.ChannelId, request.SendMessage)

		err = h.syncMembers(request.ChannelId, request.UserId, request.SendMessage)
		if err != nil {
			h.Logger.Error(fmt.Sprintf("syncMembers error: %v", err))
		}

		msg = fmt.Sprintf("Completed Member Sync [%s]", time.Since(t2))
		h.sendDualMessage(msg, request.ChannelId, request.SendMessage)

		msg = fmt.Sprintf("Completed All Syncing [%s]", time.Since(t1))
		h.sendDualMessage(msg, request.ChannelId, request.SendMessage)

		// Sleep 15 minutes after each sync
		//time.Sleep(15 * time.Minute)
	}
}

func (h *rolesHandler) ListUserRoles(ctx context.Context, request *rolesrv.ListUserRolesRequest, response *rolesrv.ListUserRolesResponse) error {
	rows, err := sq.Select("roles.*").
		From("filters").
		Join("filter_membership ON filters.id = filter_membership.filter").
		Join("roles_filters ON filters.id = role_filters.filter").
		Join("roles ON roles_filters.role = roles.id").
		Where(sq.Eq{"filter_membership.user_id": request.UserId}).
		Where(sq.Eq{"namespace": h.namespace}).
		RunWith(h.db).Query()
	if err != nil {
		return fmt.Errorf("error getting user roles (%s): %s", request.UserId, err)
	}

	var role Role
	for rows.Next() {
		err = rows.Scan(&role)
		if err != nil {
			return fmt.Errorf("error scanning role for userID (%s): %s", request.UserId, err)
		}

		response.Roles = append(response.Roles, &rolesrv.Role{
			ShortName:   role.ShortName,
			Type:        role.Type,
			Name:        role.Name,
			Color:       role.Color,
			Hoist:       role.Hoist,
			Position:    role.Position,
			Permissions: role.Permissions,
			Managed:     role.Managed,
			Mentionable: role.Mentionable,
			Sig:         role.Sig,
			Joinable:    role.Joinable,
			Sync:        role.Sync,
		})
	}

	return nil
}
