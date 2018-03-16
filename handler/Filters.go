package handler

import (
	"errors"
	"fmt"
	//discord "github.com/chremoas/discord-gateway/proto"
	rolesrv "github.com/chremoas/role-srv/proto"
	"github.com/chremoas/services-common/config"
	redis "github.com/chremoas/services-common/redis"
	//"github.com/micro/go-micro"
	//"github.com/micro/go-micro/client"
	"golang.org/x/net/context"
	//"regexp"
	//"github.com/fatih/structs"
	//"strconv"
	"strings"
)

type filtersHandler struct {
	//Client client.Client
	Redis  *redis.Client
}

func NewFiltersHandler(config *config.Configuration) rolesrv.FiltersHandler {
	addr := fmt.Sprintf("%s:%d", config.Redis.Host, config.Redis.Port)
	redisClient := redis.Init(addr, config.Redis.Password, config.Redis.Database, config.LookupService("srv", "perms"))

	_, err := redisClient.Client.Ping().Result()
	if err != nil {
		panic(err)
	}

	return &filtersHandler{Redis: redisClient}
}

func (h *filtersHandler) GetFilters(ctx context.Context, request *rolesrv.NilMessage, response *rolesrv.FilterList) error {
	filters, err := h.Redis.Client.Keys(h.Redis.KeyName("filter_description:*")).Result()

	if err != nil {
		return err
	}

	for filter := range filters {
		filterDescription, err := h.Redis.Client.Get(filters[filter]).Result()

		if err != nil {
			return err
		}

		filterName := strings.Split(filters[filter], ":")

		response.FilterList = append(response.FilterList,
			&rolesrv.Filter{Name: filterName[len(filterName)-1], Description: filterDescription})
	}

	return nil
}

func (h *filtersHandler) AddFilter(ctx context.Context, request *rolesrv.Filter, response *rolesrv.NilMessage) error {
	filterName := h.Redis.KeyName(fmt.Sprintf("filter_description:%s", request.Name))

	// Type and Name are required so let's check for those
	if len(request.Name) == 0 {
		return errors.New("Name is required.")
	}

	if len(request.Description) == 0 {
		return errors.New("Description is required.")
	}

	exists, err := h.Redis.Client.Exists(filterName).Result()

	if err != nil {
		return err
	}

	if exists == 1 {
		return fmt.Errorf("Filter `%s` already exists.", request.Name)
	}

	_, err = h.Redis.Client.Set(filterName, request.Description, 0).Result()

	if err != nil {
		return err
	}

	response = &rolesrv.NilMessage{}

	return nil
}

func (h *filtersHandler) RemoveFilter(ctx context.Context, request *rolesrv.Filter, response *rolesrv.NilMessage) error {
	filterName := h.Redis.KeyName(fmt.Sprintf("filter_description:%s", request.Name))
	filterMembers := h.Redis.KeyName(fmt.Sprintf("filter_members:%s", request.Name))

	exists, err := h.Redis.Client.Exists(filterName).Result()

	if err != nil {
		return err
	}

	if exists == 0 {
		return fmt.Errorf("Filter `%s` doesn't exists.", request.Name)
	}

	members, err := h.Redis.Client.SMembers(filterMembers).Result()

	if len(members) > 0 {
		return fmt.Errorf("Filter `%s` not empty.", request.Name)
	}

	_, err = h.Redis.Client.Del(filterName).Result()

	if err != nil {
		return err
	}

	response = &rolesrv.NilMessage{}
	return nil
}

func (h *filtersHandler) GetMembers(ctx context.Context, request *rolesrv.Filter, response *rolesrv.MemberList) error {
	var memberlist []string
	filterName := h.Redis.KeyName(fmt.Sprintf("filter_members:%s", request.Name))

	filters, err := h.Redis.Client.SMembers(filterName).Result()

	if err != nil {
		return err
	}

	for filter := range filters {
		memberlist = append(memberlist, filters[filter])
	}

	response.Members = memberlist
	return nil
}

func (h *filtersHandler) AddMembers(ctx context.Context, request *rolesrv.Members, response *rolesrv.NilMessage) error {
	filterName := h.Redis.KeyName(fmt.Sprintf("filter_members:%s", request.Filter))
	filterDesc := h.Redis.KeyName(fmt.Sprintf("filter_description:%s", request.Filter))

	exists, err := h.Redis.Client.Exists(filterDesc).Result()

	if err != nil {
		return err
	}

	if exists == 0 {
		return fmt.Errorf("Filter `%s` doesn't exists.", request.Filter)
	}

	for member := range request.Name {
		_, err = h.Redis.Client.SAdd(filterName, request.Name[member]).Result()
	}

	if err != nil {
		return err
	}

	response = &rolesrv.NilMessage{}
	return nil
}

func (h *filtersHandler) RemoveMembers(ctx context.Context, request *rolesrv.Members, response *rolesrv.NilMessage) error {
	filterName := h.Redis.KeyName(fmt.Sprintf("filter_members:%s", request.Filter))
	filterDesc := h.Redis.KeyName(fmt.Sprintf("filter_description:%s", request.Filter))

	exists, err := h.Redis.Client.Exists(filterDesc).Result()

	if err != nil {
		return err
	}

	if exists == 0 {
		return fmt.Errorf("Filter `%s` doesn't exists.", request.Filter)
	}

	//isMember, err := h.Redis.Client.SIsMember(filterName, request.Name).Result()
	//if !isMember {
	//	return fmt.Errorf("`%s` not a member of filter '%s'", request.Name, request.Filter)
	//}

	for member := range request.Name {
		_, err = h.Redis.Client.SRem(filterName, request.Name[member]).Result()
	}

	if err != nil {
		return err
	}

	response = &rolesrv.NilMessage{}
	return nil
}