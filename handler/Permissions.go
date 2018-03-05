package handler

import (
	"errors"
	"github.com/chremoas/role-srv/proto"
	"github.com/chremoas/services-common/config"
	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro"
	"golang.org/x/net/context"
)

type permissionsHandler struct {
	Client client.Client
}

func NewPermissionsHandler(conf *config.Configuration, service micro.Service) chremoas_role.PermissionsHandler {
	return &permissionsHandler{}
}

func (h *permissionsHandler) Perform(ctx context.Context, request *chremoas_role.PermissionsRequest, response *chremoas_role.PerformResponse) error {
	return errors.New("Not Implemented")
}

func (h *permissionsHandler) AddPermisssions(ctx context.Context, request *chremoas_role.PermissionsRequest, response *chremoas_role.PermissionsResponse) error {
	return errors.New("Not Implemented")
}

func (h *permissionsHandler) RemovePermissions(ctx context.Context, request *chremoas_role.PermissionsRequest, response *chremoas_role.PermissionsResponse) error {
	return errors.New("Not Implemented")
}

func (h *permissionsHandler) GetPermissions(ctx context.Context, request *chremoas_role.PermissionsRequest, response *chremoas_role.PermissionsResponse) error {
	return errors.New("Not Implemented")
}
