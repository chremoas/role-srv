package handler

func (h *rolesHandler) syncMembers(channelId, userId string, sendMessage bool) error {
	//
	// Not doing this right now
	//

	//sugar := h.Sugar()
	//var roleNameMap = make(map[string]string)
	//var idToNameMap = make(map[string]string)
	//var discordMemberships = make(map[string]*sets.StringSet)
	//var chremoasMemberships = make(map[string]*sets.StringSet)
	//var updateMembers = make(map[string]*sets.StringSet)
	//
	//// Discord limit is 1000, should probably make this a config option. -brian
	//var numberPerPage int32 = 1000
	//var memberCount = 1
	//var memberId = ""
	//
	//t := time.Now()
	//
	//// Need to pre-populate the membership sets with all the users so we can pick up users with no roles.
	//for memberCount > 0 {
	//	//longCtx, _ := context.WithTimeout(context.Background(), time.Second * 20)
	//
	//	members, err := h.clients.discord.GetAllMembers(context.Background(), &discord.GetAllMembersRequest{NumberPerPage: numberPerPage, After: memberId})
	//	if err != nil {
	//		msg := fmt.Sprintf("syncMembers: GetAllMembers: %s", err.Error())
	//		h.sendMessage(context.Background(), channelId, common.SendFatal(msg), true)
	//		sugar.Error(msg)
	//		return err
	//	}
	//
	//	for m := range members.Members {
	//		userId := members.Members[m].User.Id
	//		if _, ok := discordMemberships[userId]; !ok {
	//			discordMemberships[userId] = sets.NewStringSet()
	//		}
	//
	//		idToNameMap[userId] = members.Members[m].User.Username
	//
	//		for r := range members.Members[m].Roles {
	//			discordMemberships[userId].Add(members.Members[m].Roles[r].Name)
	//		}
	//
	//		if _, ok := chremoasMemberships[userId]; !ok {
	//			chremoasMemberships[userId] = sets.NewStringSet()
	//		}
	//
	//		oldNum, _ := strconv.Atoi(members.Members[m].User.Id)
	//		newNum, _ := strconv.Atoi(memberId)
	//
	//		if oldNum > newNum {
	//			memberId = members.Members[m].User.Id
	//		}
	//	}
	//
	//	memberCount = len(members.Members)
	//}
	//
	//h.sendDualMessage(
	//	fmt.Sprintf("Got all Discord members [%s]", time.Since(t)),
	//	channelId,
	//	sendMessage,
	//)
	//
	//t = time.Now()
	//
	//// Get all the Roles from discord and create a map of their name to their Id
	//discordRoles, err := h.clients.discord.GetAllRoles(context.Background(), &discord.GuildObjectRequest{})
	//if err != nil {
	//	msg := fmt.Sprintf("syncMembers: GetAllRoles: %s", err.Error())
	//	h.sendMessage(context.Background(), channelId, common.SendFatal(msg), true)
	//	sugar.Error(msg)
	//	return err
	//}
	//
	//for d := range discordRoles.Roles {
	//	roleNameMap[discordRoles.Roles[d].Name] = discordRoles.Roles[d].Id
	//}
	//
	//h.sendDualMessage(
	//	fmt.Sprintf("Got all Discord roles [%s]", time.Since(t)),
	//	channelId,
	//	sendMessage,
	//)
	//
	//t = time.Now()
	//
	//// Get all the Chremoas roles and build membership Sets
	//chremoasRoles, err := h.getRoles()
	//if err != nil {
	//	msg := fmt.Sprintf("syncMembers: getRoles: %s", err.Error())
	//	h.sendMessage(context.Background(), channelId, common.SendFatal(msg), true)
	//	sugar.Error(msg)
	//	return err
	//}
	//
	//h.sendDualMessage(
	//	fmt.Sprintf("Got all Chremoas roles [%s]", time.Since(t)),
	//	channelId,
	//	sendMessage,
	//)
	//
	//t = time.Now()
	//
	//for r := range chremoasRoles {
	//	sugar.Debugf("Checking role: %s", chremoasRoles[r])
	//	role, err := h.getRole(chremoasRoles[r])
	//	if err != nil {
	//		msg := fmt.Sprintf("syncMembers: getRole: %s: %s", chremoasRoles[r], err.Error())
	//		h.sendMessage(context.Background(), channelId, common.SendFatal(msg), true)
	//		sugar.Error(msg)
	//		return err
	//	}
	//
	//	if role["Sync"] == "0" || role["Sync"] == "false" {
	//		continue
	//	}
	//
	//	membership, err := h.getRoleMembership(chremoasRoles[r])
	//	if err != nil {
	//		msg := fmt.Sprintf("syncMembers: getRoleMembership: %s", err.Error())
	//		h.sendMessage(context.Background(), channelId, common.SendFatal(msg), true)
	//		sugar.Error(msg)
	//		return err
	//	}
	//
	//	roleName, err := h.getRole(chremoasRoles[r])
	//	if err != nil {
	//		msg := fmt.Sprintf("syncMembers: getRole: %s", err.Error())
	//		h.sendMessage(context.Background(), channelId, common.SendFatal(msg), true)
	//		sugar.Error(msg)
	//		return err
	//	}
	//
	//	//roleId := roleNameMap[roleName["Name"]]
	//
	//	for m := range membership.Set {
	//		sugar.Debugf("Key is: %s", m)
	//		if len(m) != 0 {
	//			sugar.Debugf("Set is %v", chremoasMemberships[m])
	//			if chremoasMemberships[m] == nil {
	//				chremoasMemberships[m] = sets.NewStringSet()
	//			}
	//			chremoasMemberships[m].Add(roleName["Name"])
	//		}
	//	}
	//}
	//
	//h.sendDualMessage(
	//	fmt.Sprintf("Got all role Memberships [%s]", time.Since(t)),
	//	channelId,
	//	sendMessage,
	//)
	//
	//t = time.Now()
	//
	//for m := range chremoasMemberships {
	//	if discordMemberships[m] == nil {
	//		sugar.Debugf("not in discord: %v", m)
	//		continue
	//	}
	//
	//	// Get the list of memberships that are in chremoas but not discord (need to be added to discord)
	//	diff := chremoasMemberships[m].Difference(discordMemberships[m])
	//	diff2 := discordMemberships[m].Difference(chremoasMemberships[m])
	//
	//	if diff.Len() != 0 || diff2.Len() != 0 {
	//		if !ignoreRole(idToNameMap[m]) {
	//			for r := range chremoasMemberships[m].Set {
	//				if _, ok := updateMembers[m]; !ok {
	//					updateMembers[m] = sets.NewStringSet()
	//				}
	//				updateMembers[m].Add(roleNameMap[r])
	//			}
	//		}
	//	}
	//}
	//
	//// Apply the membership sets to discord overwriting anything that's there.
	//h.sendDualMessage(
	//	fmt.Sprintf("Updating %d discord users", len(updateMembers)),
	//	channelId,
	//	sendMessage,
	//)
	//
	//noSyncList := h.Redis.KeyName("members:no_sync")
	//sugar.Infof("noSyncList: %v", noSyncList)
	//for m := range updateMembers {
	//	// Don't sync people who we don't want to mess with. Always put the Discord Server Owner here
	//	// because we literally can't sync them no matter what.
	//	noSync, _ := h.Redis.Client.SIsMember(noSyncList, m).Result()
	//	if noSync {
	//		sugar.Infof("Skipping noSync user: %s", m)
	//		continue
	//	}
	//
	//	ctx, _ := context.WithTimeout(context.Background(), time.Second*20)
	//	_, err = h.clients.discord.UpdateMember(ctx, &discord.UpdateMemberRequest{
	//		Operation: discord.MemberUpdateOperation_ADD_OR_UPDATE_ROLES,
	//		UserId:    m,
	//		RoleIds:   updateMembers[m].ToSlice(),
	//	})
	//	if err != nil {
	//		msg := fmt.Sprintf("syncMembers: UpdateMember: %s", err.Error())
	//		h.sendMessage(context.Background(), channelId, common.SendFatal(msg), true)
	//		sugar.Error(msg)
	//	}
	//	sugar.Infof("Updating Discord User: %s", m)
	//}
	//
	//h.sendDualMessage(
	//	fmt.Sprintf("Updated Discord Roles [%s]", time.Since(t)),
	//	channelId,
	//	sendMessage,
	//)

	return nil
}
