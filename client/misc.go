package client

import (
	rolesrv "github.com/chremoas/role-srv/proto"
	"strings"
	"fmt"
	"context"
	"bytes"
)

var unknownUserError = "User not found"

func (r Roles) GetSyncRequest(sender string, sendMessage bool) *rolesrv.SyncRequest {
	s := strings.Split(sender, ":")
	return &rolesrv.SyncRequest{ChannelId: s[0], UserId: s[1], SendMessage: sendMessage}
}

func (r Roles) MapName(ctx context.Context, members []string) (buffer bytes.Buffer, err error) {
	users, err := r.RoleClient.GetDiscordUserList(ctx, &rolesrv.NilMessage{})

	var found = false
	for m := range members {
		if len(members[m]) > 0 {
			for u := range users.Users {
				if members[m] == users.Users[u].Id {
					buffer.WriteString(fmt.Sprintf("\t%s\n", users.Users[u].Nick))
					found = true
				}
			}

			if !found {
				buffer.WriteString(fmt.Sprintf("\t%s\n", members[m]))
				found = false
			}
		}
	}

	return buffer, err
}