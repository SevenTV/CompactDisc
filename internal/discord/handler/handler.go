package handler

import (
	"github.com/bwmarrin/discordgo"
	"github.com/seventv/common/utils"
	"go.uber.org/zap"
)

var (
	DefaultRoleId = ""
)

func Register(session *discordgo.Session) {
	if DefaultRoleId == "" {
		zap.S().Fatalw("DefaultRoleId is not set")

		return
	}

	session.AddHandler(messageCreate)
	session.AddHandler(guildMemberAdd)
}

// gives the user the default role if they dont already have it
// this is for users already in the guild who do not yet have it
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if !utils.Contains(m.Member.Roles, DefaultRoleId) {
		finalRoles := append(m.Member.Roles, DefaultRoleId)

		if err := s.GuildMemberEdit(m.GuildID, m.Author.ID, finalRoles); err != nil {
			zap.S().Errorw("failed to add default role to user", "error", err)
		}
	}
}

// gives the user the default role on join
func guildMemberAdd(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	finalRoles := append(m.Roles, DefaultRoleId)

	if err := s.GuildMemberEdit(m.GuildID, m.User.ID, finalRoles); err != nil {
		zap.S().Errorw("failed to add default role to user", "error", err)
	}
}
