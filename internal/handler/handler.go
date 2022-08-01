package handler

import (
	"github.com/bwmarrin/discordgo"
	"github.com/seventv/common/utils"
	"github.com/seventv/compactdisc/internal/global"
	"go.uber.org/zap"
)

func Register(gctx global.Context, session *discordgo.Session) {
if gctx.Config().Discord.DefaultRoleId != "" {
    session.AddHandler(messageCreate(gctx))
    session.AddHandler(guildMemberAdd(gctx))
} else {
    zap.S().Warnw("default role id not set")
}
	session.AddHandler(guildMemberAdd(gctx))
}

// gives the user the default role if they dont already have it
// this is for users already in the guild who do not yet have it
func messageCreate(gctx global.Context) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if gctx.Config().Discord.DefaultRoleId == "" {
			zap.S().Fatalw("default role id not set")
		}

		if !utils.Contains(m.Member.Roles, gctx.Config().Discord.DefaultRoleId) {
			finalRoles := append(m.Member.Roles, gctx.Config().Discord.DefaultRoleId)

			if err := s.GuildMemberEdit(m.GuildID, m.Author.ID, finalRoles); err != nil {
				zap.S().Errorw("failed to add default role to user", "error", err)
			}
		}
	}
}

// gives the user the default role on join
func guildMemberAdd(gctx global.Context) func(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	return func(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
		if gctx.Config().Discord.DefaultRoleId == "" {
			zap.S().Fatalw("default role id not set")
		}

		finalRoles := append(m.Roles, gctx.Config().Discord.DefaultRoleId)

		if err := s.GuildMemberEdit(m.GuildID, m.User.ID, finalRoles); err != nil {
			zap.S().Errorw("failed to add default role to user", "error", err)
		}
	}
}
