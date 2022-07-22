package handler

import (
	"context"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/seventv/common/utils"
	"github.com/seventv/compactdisc/client"
	"github.com/seventv/compactdisc/internal/global"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

func SyncUser(gctx global.Context, ctx context.Context, req client.Request[client.RequestPayloadSyncUser]) error {
	userID := req.Data.UserID

	appRoles, err := gctx.Inst().Query.Roles(ctx, bson.M{})
	if err != nil {
		return err
	}

	user, err := gctx.Inst().Query.Users(ctx, bson.M{"_id": userID}).First()
	if err != nil {
		return err
	}

	con, ind, _ := user.Connections.Discord()
	if ind == -1 {
		return nil // ignore, because the user does not have a discord connection
	}

	dis := gctx.Inst().Discord.Session()
	guildID := gctx.Config().Discord.GuildID

	z := zap.S().Named("api/SyncUser").With(
		"user_id", userID.Hex(),
		"guild_id", guildID,
		"discord_id", con.ID,
	)

	member, err := dis.State.Member(guildID, con.ID)
	if err != nil { // member is not in state, so we must fetch them
		member, err = dis.GuildMember(guildID, con.ID)
		_ = dis.State.MemberAdd(member)
	}

	if err != nil {
		z.Infow("user is not in the guild", "error", err)
		return nil // ignore, because the user is not a member of the guild
	}

	botRank := 0
	botMember, err := dis.State.Member(guildID, dis.State.User.ID)

	if err != nil {
		z.Errorw("bot is not in the guild", "error", err)
		return err
	}

	roles, _ := dis.GuildRoles(guildID)
	if roles == nil {
		roles = []*discordgo.Role{}
	}

	guildRoles := make(map[string]*discordgo.Role)
	for _, rol := range roles {
		guildRoles[rol.ID] = rol

		if utils.Contains(botMember.Roles, rol.ID) && rol.Position > botRank {
			botRank = rol.Position
		}
	}

	// Go through the user's roles and sync their discord roles with it
	finalRoles := make([]string, len(member.Roles))
	copy(finalRoles, member.Roles)

	add := []string{}
	del := []string{}

	for _, rol := range appRoles {
		if rol.DiscordID == 0 {
			continue // ignore, because the role is not linked to discord
		}

		roleID := strconv.Itoa(int(rol.DiscordID))
		grole, ok := guildRoles[roleID]

		if !ok {
			continue // ignore, because the role is not in the guild
		}

		if !utils.Contains(user.RoleIDs, rol.ID) { // user does not have this role
			if grole.Position >= botRank || grole.Managed {
				continue // ignore, because the bot cannot edit this role
			}

			// remove the role from the result
			if pos := utils.SliceIndexOf(finalRoles, roleID); pos != -1 {
				finalRoles = utils.SliceRemove(finalRoles, pos)

				del = append(del, grole.Name)
			}
		} else { // user has this role
			if utils.Contains(finalRoles, roleID) {
				continue // role is already attributed in discord
			}

			// add the role to the result
			finalRoles = append(finalRoles, roleID)
			add = append(add, grole.Name)
		}
	}

	if len(add) == 0 && len(del) == 0 {
		z.Info("user's roles are in sync")
		return nil
	}

	if err := dis.GuildMemberEdit(gctx.Config().Discord.GuildID, member.User.ID, finalRoles); err != nil {
		z.Errorw("failed to update discord roles")
		return err
	}

	z.Infow("roles updated", "added", add, "removed", del)

	return nil
}
