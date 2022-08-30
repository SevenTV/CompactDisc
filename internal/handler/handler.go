package handler

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/seventv/common/utils"
	"github.com/seventv/compactdisc/internal/global"
	"go.uber.org/zap"
)

func Register(gctx global.Context, session *discordgo.Session) {
	session.AddHandler(messageCreate(gctx))
	session.AddHandler(messageDelete(gctx))
	session.AddHandler(guildMemberAdd(gctx))
}

// messageCreate is a handler for messages
func messageCreate(gctx global.Context) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Member == nil {
			return
		}

		// Store the message in cache for 24h
		key := gctx.Inst().Redis.ComposeKey("compactdisc", "cache", "message", m.ID)

		j, _ := json.Marshal(m.Message)
		if err := gctx.Inst().Redis.SetEX(gctx, key, utils.B2S(j), time.Hour*72); err != nil {
			zap.S().Errorw("failed to store message in cache", "error", err)
		}

		// Assign default role if user does not have it
		if gctx.Config().Discord.DefaultRoleId != "" && !utils.Contains(m.Member.Roles, gctx.Config().Discord.DefaultRoleId) {
			finalRoles := append(m.Member.Roles, gctx.Config().Discord.DefaultRoleId)

			if _, err := s.GuildMemberEdit(m.GuildID, m.Author.ID, &discordgo.GuildMemberParams{
				Roles: &finalRoles,
			}); err != nil {
				zap.S().Errorw("failed to add default role to user", "error", err)
			}
		}
	}
}

func messageDelete(gctx global.Context) func(s *discordgo.Session, m *discordgo.MessageDelete) {
	return func(s *discordgo.Session, m *discordgo.MessageDelete) {
		key := gctx.Inst().Redis.ComposeKey("compactdisc", "cache", "message", m.ID)

		data, err := gctx.Inst().Redis.Get(gctx, key)
		if err != nil {
			zap.S().Errorw("failed to get message from cache", "error", err)
			return
		}

		var msg *discordgo.Message
		if err := json.Unmarshal([]byte(data), &msg); err != nil {
			zap.S().Errorw("failed to unmarshal message from cache", "error", err)

			return
		}

		// Create an embed of the deleted message
		fields := []*discordgo.MessageEmbedField{}
		if msg.Content != "" {
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:  "Content",
				Value: msg.Content,
			})
		}

		if len(msg.Attachments) > 0 {
			fields = append(fields, &discordgo.MessageEmbedField{
				Name: "Attachments",
				Value: func() string {
					a := make([]string, len(msg.Attachments))
					for i, v := range msg.Attachments {
						a[i] = v.URL
					}

					return strings.Join(a, "\n")
				}(),
			})
		}

		embed := &discordgo.MessageEmbed{
			Author: &discordgo.MessageEmbedAuthor{
				Name:    msg.Author.Username,
				IconURL: msg.Author.AvatarURL("128"),
			},
			Description: fmt.Sprintf("❌ **Message by %s deleted in <#%s>**\n", msg.Author.Mention(), msg.ChannelID),
			Color:       0xFF0000,
			Timestamp:   msg.Timestamp.Format(time.RFC3339),
			Fields:      fields,
		}

		if _, err := s.ChannelMessageSendEmbed(gctx.Config().Discord.Channels["mod_logs"], embed); err != nil {
			zap.S().Errorw("failed to send embed", "error", err)
		}
	}
}

// guildMemberAdd is a handler for new joins
func guildMemberAdd(gctx global.Context) func(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	return func(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
		if gctx.Config().Discord.DefaultRoleId != "" {
			finalRoles := append(m.Roles, gctx.Config().Discord.DefaultRoleId)

			if _, err := s.GuildMemberEdit(m.GuildID, m.User.ID, &discordgo.GuildMemberParams{
				Roles: &finalRoles,
			}); err != nil {
				zap.S().Errorw("failed to add default role to user", "error", err)
			}
		}
	}
}
