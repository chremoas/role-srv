package handler

import (
	"github.com/chremoas/role-srv/proto"
	"github.com/micro/go-micro/client"
	"golang.org/x/net/context"
)

type permissionsHandler struct {
	Client client.Client
}

func NewPermissionsHandler() chremoas_role.PermissionsHandler {
	return &permissionsHandler{}
}

func (h *permissionsHandler) Perform(ctx context.Context, request *chremoas_role.PermissionsRequest, response *chremoas_role.PerformResponse) error {
	return nil
}

func (h *permissionsHandler) AddPermisssions(ctx context.Context, request *chremoas_role.PermissionsRequest, response *chremoas_role.PermissionsResponse) error {
	return nil
}

func (h *permissionsHandler) RemovePermissions(ctx context.Context, request *chremoas_role.PermissionsRequest, response *chremoas_role.PermissionsResponse) error {
	return nil
}

func (h *permissionsHandler) GetPermissions(ctx context.Context, request *chremoas_role.PermissionsRequest, response *chremoas_role.PermissionsResponse) error {
	return nil
}
