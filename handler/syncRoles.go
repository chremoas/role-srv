package handler

func (h *rolesHandler) syncRoles(channelId, userId string, sendMessage bool) error {
	//
	// Not doing this right now
	//

	//ctx := context.Background()
	//var matchDiscordError = regexp.MustCompile(`^The role '.*' already exists$`)
	//chremoasRoleSet := sets.NewStringSet()
	//discordRoleSet := sets.NewStringSet()
	//sugar := h.Sugar()
	//var chremoasRoleData = make(map[string]map[string]string)
	//
	//chremoasRoles, err := h.getRoles()
	//if err != nil {
	//	msg := fmt.Sprintf("syncRoles: h.getRoles(): %s", err.Error())
	//	h.sendMessage(ctx, channelId, common.SendFatal(msg), true)
	//	sugar.Error(msg)
	//	return err
	//}
	//
	//for role := range chremoasRoles {
	//	roleName := h.Redis.KeyName(fmt.Sprintf("role:%s", chremoasRoles[role]))
	//	c, err := h.Redis.Client.HGetAll(roleName).Result()
	//
	//	if err != nil {
	//		msg := fmt.Sprintf("syncRoles: HGetAll(): %s", err.Error())
	//		h.sendMessage(ctx, channelId, common.SendFatal(msg), true)
	//		sugar.Error(msg)
	//		return err
	//	}
	//
	//	sugar.Debugf("Checking %s: %s", c["Name"], c["Sync"])
	//	if c["Sync"] == "1" || c["Sync"] == "true" {
	//		chremoasRoleSet.Add(c["Name"])
	//
	//		if _, ok := chremoasRoleData[c["Name"]]; !ok {
	//			chremoasRoleData[c["Name"]] = make(map[string]string)
	//		}
	//		chremoasRoleData[c["Name"]] = c
	//	}
	//}
	//
	//discordRoles, err := h.clients.discord.GetAllRoles(ctx, &discord.GuildObjectRequest{})
	//if err != nil {
	//	msg := fmt.Sprintf("syncRoles: GetAllRoles: %s", err.Error())
	//	h.sendMessage(ctx, channelId, common.SendFatal(msg), true)
	//	sugar.Error(msg)
	//	return err
	//}
	//
	//ignoreSet := sets.NewStringSet()
	//ignoreSet.Add(viper.GetString("bot.botRole"))
	//ignoreSet.Add("@everyone")
	//for i := range ignoredRoles {
	//	ignoreSet.Add(ignoredRoles[i])
	//}
	//
	//for role := range discordRoles.Roles {
	//	if !ignoreSet.Contains(discordRoles.Roles[role].Name) {
	//		discordRoleSet.Add(discordRoles.Roles[role].Name)
	//	}
	//}
	//
	//toAdd := chremoasRoleSet.Difference(discordRoleSet)
	//toDelete := discordRoleSet.Difference(chremoasRoleSet)
	//toUpdate := discordRoleSet.Intersection(chremoasRoleSet)
	//
	//sugar.Debugf("toAdd: %v", toAdd)
	//sugar.Debugf("toDelete: %v", toDelete)
	//sugar.Debugf("toUpdate: %v", toUpdate)
	//
	//for r := range toAdd.Set {
	//	_, err := h.clients.discord.CreateRole(ctx, &discord.CreateRoleRequest{Name: r})
	//
	//	if err != nil {
	//		if matchDiscordError.MatchString(err.Error()) {
	//			// The role list was cached most likely so we'll pretend we didn't try
	//			// to create it just now. -brian
	//			sugar.Debugf("syncRoles added: %s", r)
	//			continue
	//		} else {
	//			msg := fmt.Sprintf("syncRoles: CreateRole() attempting to create '%s': %s", r, err.Error())
	//			h.sendMessage(ctx, channelId, common.SendFatal(msg), true)
	//			sugar.Error(msg)
	//			return err
	//		}
	//	}
	//
	//	sugar.Debugf("syncRoles added: %s", r)
	//}
	//
	//for r := range toDelete.Set {
	//	_, err := h.clients.discord.DeleteRole(ctx, &discord.DeleteRoleRequest{Name: r})
	//
	//	if err != nil {
	//		msg := fmt.Sprintf("syncRoles: DeleteRole() Error Deleting '%s': %s", r, err.Error())
	//		h.sendMessage(ctx, channelId, common.SendFatal(msg), true)
	//		sugar.Error(msg)
	//		return err
	//	}
	//
	//	sugar.Debugf("syncRoles removed: %s", r)
	//}
	//
	//for r := range toUpdate.Set {
	//	color, _ := strconv.ParseInt(chremoasRoleData[r]["Color"], 10, 64)
	//	perm, _ := strconv.ParseInt(chremoasRoleData[r]["Permissions"], 10, 64)
	//	position, _ := strconv.ParseInt(chremoasRoleData[r]["Position"], 10, 64)
	//	hoist, _ := strconv.ParseBool(chremoasRoleData[r]["Hoist"])
	//	mention, _ := strconv.ParseBool(chremoasRoleData[r]["Mentionable"])
	//	managed, _ := strconv.ParseBool(chremoasRoleData[r]["Managed"])
	//
	//	editRequest := &discord.EditRoleRequest{
	//		Name:     chremoasRoleData[r]["Name"],
	//		Color:    color,
	//		Perm:     perm,
	//		Position: position,
	//		Hoist:    hoist,
	//		Mention:  mention,
	//		Managed:  managed,
	//	}
	//
	//	longCtx, _ := context.WithTimeout(ctx, time.Minute*5)
	//	_, err := h.clients.discord.EditRole(longCtx, editRequest)
	//	if err != nil {
	//		msg := fmt.Sprintf("syncRoles: EditRole(): %s", err.Error())
	//		h.sendMessage(ctx, channelId, common.SendFatal(msg), true)
	//		sugar.Error(msg)
	//		return err
	//	}
	//
	//	sugar.Debugf("syncRoles updated: %s", r)
	//}

	return nil
}
