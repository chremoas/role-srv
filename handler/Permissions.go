package handler

import (
	"github.com/chremoas/role-srv/proto"
	"github.com/micro/go-micro/client"
	"golang.org/x/net/context"
)

type permissionsHandler struct {
	Client    client.Client
}

func NewPermissionsHandler() chremoas_permissions.PermissionsHandler {
	return &permissionsHandler{}
}

func (ph *permissionsHandler) Perform(ctx context.Context, request *chremoas_permissions.PermissionsRequest, response *chremoas_permissions.PerformResponse) error {
	return nil
}

func (ph *permissionsHandler) AddPermisssions(ctx context.Context, request *chremoas_permissions.PermissionsRequest, response *chremoas_permissions.PermissionsResponse) error {
	return nil
}

func (ph *permissionsHandler) RemovePermissions(ctx context.Context, request *chremoas_permissions.PermissionsRequest, response *chremoas_permissions.PermissionsResponse) error {
	return nil
}

func (ph *permissionsHandler) GetPermissions(ctx context.Context, request *chremoas_permissions.PermissionsRequest, response *chremoas_permissions.PermissionsResponse) error {
	var permissions []chremoas_permissions.AdminRoles

	permissions = append(permissions, chremoas_permissions.AdminRoles_SERVER_ADMIN)

	response.PermissionsList = permissions

	return nil
}