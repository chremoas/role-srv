package handler

import (
	"fmt"
	rolesrv "github.com/chremoas/role-srv/proto"
	"github.com/chremoas/services-common/config"
	redis "github.com/chremoas/services-common/redis"
	"github.com/fatih/structs"
	"golang.org/x/net/context"
	"strings"
)

type rulesHandler struct {
	//Client client.Client
	Redis *redis.Client
}

func NewRulesHandler(config *config.Configuration) rolesrv.RulesHandler {
	addr := fmt.Sprintf("%s:%d", config.Redis.Host, config.Redis.Port)
	redisClient := redis.Init(addr, config.Redis.Password, config.Redis.Database, config.LookupService("srv", "perms"))

	_, err := redisClient.Client.Ping().Result()
	if err != nil {
		panic(err)
	}

	return &rulesHandler{Redis: redisClient}
}

func (h *rulesHandler) AddRule(ctx context.Context, request *rolesrv.Rule, response *rolesrv.NilMessage) error {
	ruleName := h.Redis.KeyName(fmt.Sprintf("rule:%s", request.Name))
	roleName := h.Redis.KeyName(fmt.Sprintf("role:%s", request.Rule.Role))
	filterA := h.Redis.KeyName(fmt.Sprintf("filter:%s", request.Rule.FilterA))
	filterB := h.Redis.KeyName(fmt.Sprintf("filter:%s", request.Rule.FilterB))

	// There has got to be a better way to do this.
	// Check if rule exists
	exists, err := h.Redis.Client.Exists(ruleName).Result()

	if err != nil {
		return err
	}

	if exists == 1 {
		return fmt.Errorf("Rule `%s` already exists.", request.Name)
	}

	// Check if role exists
	exists, err = h.Redis.Client.Exists(roleName).Result()

	if err != nil {
		return err
	}

	if exists == 0 {
		return fmt.Errorf("Role `%s` doesn't exists.", request.Rule.Role)
	}

	// Check if filter A exists
	exists, err = h.Redis.Client.Exists(filterA).Result()

	if err != nil {
		return err
	}

	if exists == 0 {
		return fmt.Errorf("FilterA `%s` doesn't exists.", request.Rule.FilterA)
	}

	// Check if filter B exists
	exists, err = h.Redis.Client.Exists(filterB).Result()

	if err != nil {
		return err
	}

	if exists == 0 {
		return fmt.Errorf("FilterB `%s` doesn't exists.", request.Rule.FilterB)
	}

	_, err = h.Redis.Client.HMSet(ruleName, structs.Map(request.Rule)).Result()

	if err != nil {
		return err
	}

	response = &rolesrv.NilMessage{}
	return nil
}

func (h *rulesHandler) UpdateRule(ctx context.Context, request *rolesrv.Rule, response *rolesrv.NilMessage) error {
	return nil
}

func (h *rulesHandler) RemoveRule(ctx context.Context, request *rolesrv.Rule, response *rolesrv.NilMessage) error {
	ruleName := h.Redis.KeyName(fmt.Sprintf("rule:%s", request.Name))

	exists, err := h.Redis.Client.Exists(ruleName).Result()

	if err != nil {
		return err
	}

	if exists == 0 {
		return fmt.Errorf("Rule `%s` doesn't exists.", request.Name)
	}

	_, err = h.Redis.Client.Del(ruleName).Result()

	if err != nil {
		return err
	}

	response = &rolesrv.NilMessage{}
	return nil
}

func (h *rulesHandler) GetRules(ctx context.Context, request *rolesrv.NilMessage, response *rolesrv.RulesList) error {
	rules, err := h.Redis.Client.Keys(h.Redis.KeyName("rule:*")).Result()

	if err != nil {
		return err
	}

	for rule := range rules {
		if err != nil {
			return err
		}

		ruleName := strings.Split(rules[rule], ":")

		response.Rules = append(response.Rules, &rolesrv.Rule{Name: ruleName[len(ruleName)-1]})
	}

	return nil
}

func (h *rulesHandler) GetRule(ctx context.Context, request *rolesrv.Rule, response *rolesrv.Rule) error {
	ruleName := h.Redis.KeyName(fmt.Sprintf("rule:%s", request.Name))

	rule, err := h.Redis.Client.HGetAll(ruleName).Result()

	if err != nil {
		return err
	}

	response.Name = request.Name
	response.Rule = &rolesrv.RuleInfo{
		Role:    rule["Role"],
		FilterA: rule["FilterA"],
		FilterB: rule["FilterB"],
	}

	return nil
}
