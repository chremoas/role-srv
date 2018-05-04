// Code generated by protoc-gen-micro. DO NOT EDIT.
// source: roles.proto

/*
Package chremoas_roles is a generated protocol buffer package.

It is generated from these files:
	roles.proto

It has these top-level messages:
	NilMessage
	SyncRequest
	StringList
	Role
	UpdateInfo
	GetRolesResponse
	RoleSyncResponse
	FilterList
	Filter
	Members
	MemberList
*/
package chremoas_roles

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	client "github.com/micro/go-micro/client"
	server "github.com/micro/go-micro/server"
	context "context"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ client.Option
var _ server.Option

// Client API for Roles service

type RolesService interface {
	AddRole(ctx context.Context, in *Role, opts ...client.CallOption) (*NilMessage, error)
	UpdateRole(ctx context.Context, in *UpdateInfo, opts ...client.CallOption) (*NilMessage, error)
	RemoveRole(ctx context.Context, in *Role, opts ...client.CallOption) (*NilMessage, error)
	GetRoles(ctx context.Context, in *NilMessage, opts ...client.CallOption) (*GetRolesResponse, error)
	GetRole(ctx context.Context, in *Role, opts ...client.CallOption) (*Role, error)
	SyncRoles(ctx context.Context, in *SyncRequest, opts ...client.CallOption) (*RoleSyncResponse, error)
	GetRoleKeys(ctx context.Context, in *NilMessage, opts ...client.CallOption) (*StringList, error)
	GetRoleTypes(ctx context.Context, in *NilMessage, opts ...client.CallOption) (*StringList, error)
	SyncMembers(ctx context.Context, in *SyncRequest, opts ...client.CallOption) (*NilMessage, error)
	GetFilters(ctx context.Context, in *NilMessage, opts ...client.CallOption) (*FilterList, error)
	AddFilter(ctx context.Context, in *Filter, opts ...client.CallOption) (*NilMessage, error)
	RemoveFilter(ctx context.Context, in *Filter, opts ...client.CallOption) (*NilMessage, error)
	GetMembers(ctx context.Context, in *Filter, opts ...client.CallOption) (*MemberList, error)
	AddMembers(ctx context.Context, in *Members, opts ...client.CallOption) (*NilMessage, error)
	RemoveMembers(ctx context.Context, in *Members, opts ...client.CallOption) (*NilMessage, error)
}

type rolesService struct {
	c    client.Client
	name string
}

func NewRolesService(name string, c client.Client) RolesService {
	if c == nil {
		c = client.NewClient()
	}
	if len(name) == 0 {
		name = "chremoas.roles"
	}
	return &rolesService{
		c:    c,
		name: name,
	}
}

func (c *rolesService) AddRole(ctx context.Context, in *Role, opts ...client.CallOption) (*NilMessage, error) {
	req := c.c.NewRequest(c.name, "Roles.AddRole", in)
	out := new(NilMessage)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rolesService) UpdateRole(ctx context.Context, in *UpdateInfo, opts ...client.CallOption) (*NilMessage, error) {
	req := c.c.NewRequest(c.name, "Roles.UpdateRole", in)
	out := new(NilMessage)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rolesService) RemoveRole(ctx context.Context, in *Role, opts ...client.CallOption) (*NilMessage, error) {
	req := c.c.NewRequest(c.name, "Roles.RemoveRole", in)
	out := new(NilMessage)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rolesService) GetRoles(ctx context.Context, in *NilMessage, opts ...client.CallOption) (*GetRolesResponse, error) {
	req := c.c.NewRequest(c.name, "Roles.GetRoles", in)
	out := new(GetRolesResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rolesService) GetRole(ctx context.Context, in *Role, opts ...client.CallOption) (*Role, error) {
	req := c.c.NewRequest(c.name, "Roles.GetRole", in)
	out := new(Role)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rolesService) SyncRoles(ctx context.Context, in *SyncRequest, opts ...client.CallOption) (*RoleSyncResponse, error) {
	req := c.c.NewRequest(c.name, "Roles.SyncRoles", in)
	out := new(RoleSyncResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rolesService) GetRoleKeys(ctx context.Context, in *NilMessage, opts ...client.CallOption) (*StringList, error) {
	req := c.c.NewRequest(c.name, "Roles.GetRoleKeys", in)
	out := new(StringList)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rolesService) GetRoleTypes(ctx context.Context, in *NilMessage, opts ...client.CallOption) (*StringList, error) {
	req := c.c.NewRequest(c.name, "Roles.GetRoleTypes", in)
	out := new(StringList)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rolesService) SyncMembers(ctx context.Context, in *SyncRequest, opts ...client.CallOption) (*NilMessage, error) {
	req := c.c.NewRequest(c.name, "Roles.SyncMembers", in)
	out := new(NilMessage)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rolesService) GetFilters(ctx context.Context, in *NilMessage, opts ...client.CallOption) (*FilterList, error) {
	req := c.c.NewRequest(c.name, "Roles.GetFilters", in)
	out := new(FilterList)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rolesService) AddFilter(ctx context.Context, in *Filter, opts ...client.CallOption) (*NilMessage, error) {
	req := c.c.NewRequest(c.name, "Roles.AddFilter", in)
	out := new(NilMessage)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rolesService) RemoveFilter(ctx context.Context, in *Filter, opts ...client.CallOption) (*NilMessage, error) {
	req := c.c.NewRequest(c.name, "Roles.RemoveFilter", in)
	out := new(NilMessage)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rolesService) GetMembers(ctx context.Context, in *Filter, opts ...client.CallOption) (*MemberList, error) {
	req := c.c.NewRequest(c.name, "Roles.GetMembers", in)
	out := new(MemberList)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rolesService) AddMembers(ctx context.Context, in *Members, opts ...client.CallOption) (*NilMessage, error) {
	req := c.c.NewRequest(c.name, "Roles.AddMembers", in)
	out := new(NilMessage)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rolesService) RemoveMembers(ctx context.Context, in *Members, opts ...client.CallOption) (*NilMessage, error) {
	req := c.c.NewRequest(c.name, "Roles.RemoveMembers", in)
	out := new(NilMessage)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Roles service

type RolesHandler interface {
	AddRole(context.Context, *Role, *NilMessage) error
	UpdateRole(context.Context, *UpdateInfo, *NilMessage) error
	RemoveRole(context.Context, *Role, *NilMessage) error
	GetRoles(context.Context, *NilMessage, *GetRolesResponse) error
	GetRole(context.Context, *Role, *Role) error
	SyncRoles(context.Context, *SyncRequest, *RoleSyncResponse) error
	GetRoleKeys(context.Context, *NilMessage, *StringList) error
	GetRoleTypes(context.Context, *NilMessage, *StringList) error
	SyncMembers(context.Context, *SyncRequest, *NilMessage) error
	GetFilters(context.Context, *NilMessage, *FilterList) error
	AddFilter(context.Context, *Filter, *NilMessage) error
	RemoveFilter(context.Context, *Filter, *NilMessage) error
	GetMembers(context.Context, *Filter, *MemberList) error
	AddMembers(context.Context, *Members, *NilMessage) error
	RemoveMembers(context.Context, *Members, *NilMessage) error
}

func RegisterRolesHandler(s server.Server, hdlr RolesHandler, opts ...server.HandlerOption) {
	type roles interface {
		AddRole(ctx context.Context, in *Role, out *NilMessage) error
		UpdateRole(ctx context.Context, in *UpdateInfo, out *NilMessage) error
		RemoveRole(ctx context.Context, in *Role, out *NilMessage) error
		GetRoles(ctx context.Context, in *NilMessage, out *GetRolesResponse) error
		GetRole(ctx context.Context, in *Role, out *Role) error
		SyncRoles(ctx context.Context, in *SyncRequest, out *RoleSyncResponse) error
		GetRoleKeys(ctx context.Context, in *NilMessage, out *StringList) error
		GetRoleTypes(ctx context.Context, in *NilMessage, out *StringList) error
		SyncMembers(ctx context.Context, in *SyncRequest, out *NilMessage) error
		GetFilters(ctx context.Context, in *NilMessage, out *FilterList) error
		AddFilter(ctx context.Context, in *Filter, out *NilMessage) error
		RemoveFilter(ctx context.Context, in *Filter, out *NilMessage) error
		GetMembers(ctx context.Context, in *Filter, out *MemberList) error
		AddMembers(ctx context.Context, in *Members, out *NilMessage) error
		RemoveMembers(ctx context.Context, in *Members, out *NilMessage) error
	}
	type Roles struct {
		roles
	}
	h := &rolesHandler{hdlr}
	s.Handle(s.NewHandler(&Roles{h}, opts...))
}

type rolesHandler struct {
	RolesHandler
}

func (h *rolesHandler) AddRole(ctx context.Context, in *Role, out *NilMessage) error {
	return h.RolesHandler.AddRole(ctx, in, out)
}

func (h *rolesHandler) UpdateRole(ctx context.Context, in *UpdateInfo, out *NilMessage) error {
	return h.RolesHandler.UpdateRole(ctx, in, out)
}

func (h *rolesHandler) RemoveRole(ctx context.Context, in *Role, out *NilMessage) error {
	return h.RolesHandler.RemoveRole(ctx, in, out)
}

func (h *rolesHandler) GetRoles(ctx context.Context, in *NilMessage, out *GetRolesResponse) error {
	return h.RolesHandler.GetRoles(ctx, in, out)
}

func (h *rolesHandler) GetRole(ctx context.Context, in *Role, out *Role) error {
	return h.RolesHandler.GetRole(ctx, in, out)
}

func (h *rolesHandler) SyncRoles(ctx context.Context, in *SyncRequest, out *RoleSyncResponse) error {
	return h.RolesHandler.SyncRoles(ctx, in, out)
}

func (h *rolesHandler) GetRoleKeys(ctx context.Context, in *NilMessage, out *StringList) error {
	return h.RolesHandler.GetRoleKeys(ctx, in, out)
}

func (h *rolesHandler) GetRoleTypes(ctx context.Context, in *NilMessage, out *StringList) error {
	return h.RolesHandler.GetRoleTypes(ctx, in, out)
}

func (h *rolesHandler) SyncMembers(ctx context.Context, in *SyncRequest, out *NilMessage) error {
	return h.RolesHandler.SyncMembers(ctx, in, out)
}

func (h *rolesHandler) GetFilters(ctx context.Context, in *NilMessage, out *FilterList) error {
	return h.RolesHandler.GetFilters(ctx, in, out)
}

func (h *rolesHandler) AddFilter(ctx context.Context, in *Filter, out *NilMessage) error {
	return h.RolesHandler.AddFilter(ctx, in, out)
}

func (h *rolesHandler) RemoveFilter(ctx context.Context, in *Filter, out *NilMessage) error {
	return h.RolesHandler.RemoveFilter(ctx, in, out)
}

func (h *rolesHandler) GetMembers(ctx context.Context, in *Filter, out *MemberList) error {
	return h.RolesHandler.GetMembers(ctx, in, out)
}

func (h *rolesHandler) AddMembers(ctx context.Context, in *Members, out *NilMessage) error {
	return h.RolesHandler.AddMembers(ctx, in, out)
}

func (h *rolesHandler) RemoveMembers(ctx context.Context, in *Members, out *NilMessage) error {
	return h.RolesHandler.RemoveMembers(ctx, in, out)
}
