package client

import (
	rolesrv "github.com/chremoas/role-srv/proto"
	"strings"
)

func (r Roles) GetSyncRequest(sender string, sendMessage bool) *rolesrv.SyncRequest {
	s := strings.Split(sender, ":")
	return &rolesrv.SyncRequest{ChannelId: s[0], UserId: s[1], SendMessage: sendMessage}
}
