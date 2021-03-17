package client

import (
	"bytes"
	"context"
	"fmt"
	rolesrv "github.com/chremoas/role-srv/proto"
	"strings"
)

func (r Roles) GetSyncRequest(sender string, sendMessage bool) *rolesrv.SyncRequest {
	s := strings.Split(sender, ":")
	return &rolesrv.SyncRequest{ChannelId: s[0], UserId: s[1], SendMessage: sendMessage}
}

func (r Roles) MapName(ctx context.Context, members []string) (buffer bytes.Buffer, names []string, err error) {
	users, err := r.RoleClient.GetDiscordUserList(ctx, &rolesrv.NilMessage{})
	if err != nil {
		fmt.Printf("GetDiscordUserList error: %v", err)
		return buffer, nil, err
	}
	var found = false
	var name string

	for m := range members {
		fmt.Printf("Checking member: %v", m)
		if len(members[m]) > 0 {
			for u := range users.Users {
				fmt.Printf("Checking users: %v", u)
				if members[m] == users.Users[u].Id {
					if len(users.Users[u].Nick) != 0 {
						name = users.Users[u].Nick
					} else {
						name = users.Users[u].Username
					}
					buffer.WriteString(fmt.Sprintf("\t%s\n", name))
					names = append(names, name)
					found = true
				}
			}

			if !found {
				buffer.WriteString(fmt.Sprintf("\t%s\n", members[m]))
				found = false
			}
		}
	}

	return buffer, names, err
}
