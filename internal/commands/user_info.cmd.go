package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/seventv/common/structures/v3"
	"github.com/seventv/common/utils"
	"github.com/seventv/compactdisc/internal/global"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserInfo retrieves information about a 7TV user via their Discord connection
func UserInfo(gctx global.Context, appID string, guildID string) *Command {
	return DefineCommand(
		&discordgo.ApplicationCommand{
			ID:                       appID,
			ApplicationID:            appID,
			GuildID:                  guildID,
			Type:                     discordgo.UserApplicationCommand,
			Name:                     "User Info",
			DefaultMemberPermissions: utils.PointerOf(int64(discordgo.PermissionManageRoles)),
			DescriptionLocalizations: &map[discordgo.Locale]string{},
		},
		func(session *discordgo.Session, interaction *discordgo.InteractionCreate) error {
			data := interaction.ApplicationCommandData()
			userID := data.TargetID

			// Fetch the user
			user, err := gctx.Inst().Query.Users(gctx, bson.M{
				"connections": bson.M{"$elemMatch": bson.M{
					"platform": structures.UserConnectionPlatformDiscord,
					"id":       userID,
				}},
			}).First()
			if err != nil {
				return err
			}

			avatarURL := ""
			if user.AvatarID != "" {
				avatarURL = fmt.Sprintf("https://%s/pp/%s/%s", gctx.Config().CdnURL, user.ID.Hex(), user.AvatarID)
			} else {
				for _, con := range user.Connections {
					if con.Platform == structures.UserConnectionPlatformTwitch {
						if con, err := structures.ConvertUserConnection[structures.UserConnectionDataTwitch](con); err == nil {
							avatarURL = con.Data.ProfileImageURL
						}
					} else if con.Platform == structures.UserConnectionPlatformDiscord {
						if con, err := structures.ConvertUserConnection[structures.UserConnectionDataDiscord](con); err == nil {
							avatarURL = fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.png", con.Data.ID, con.Data.Avatar)
						}
					}
				}
			}

			// Format an embed
			embed := &discordgo.MessageEmbed{
				Type:        discordgo.EmbedTypeRich,
				Title:       fmt.Sprintf("%s (%s)", user.DisplayName, user.Username),
				URL:         user.WebURL(gctx.Config().WebsiteURL),
				Description: user.Biography,
				Author: &discordgo.MessageEmbedAuthor{
					URL:     user.WebURL(gctx.Config().WebsiteURL),
					Name:    user.DisplayName,
					IconURL: avatarURL,
				},
				Timestamp: user.ID.Timestamp().Format(time.RFC3339),
				Color:     int(user.GetHighestRole().Color),
				Fields: func() []*discordgo.MessageEmbedField {
					fields := make([]*discordgo.MessageEmbedField, 0)

					// Add the user's connections
					for _, con := range user.Connections {
						fields = append(fields, &discordgo.MessageEmbedField{
							Name:   string(con.Platform),
							Value:  con.ID,
							Inline: true,
						})
					}

					// Add the user's roles
					fields = append(fields, &discordgo.MessageEmbedField{
						Name: "Roles",
						Value: func() string {
							roleList := make([]string, len(user.Roles))
							for i, role := range user.Roles {
								roleList[i] = role.Name
							}

							return strings.Join(roleList, "\n")
						}(),
						Inline: false,
					})

					editorIDs := make([]primitive.ObjectID, len(user.Editors))
					for i, editor := range user.Editors {
						editorIDs[i] = editor.ID
					}

					editors, err := gctx.Inst().Query.Users(gctx, bson.M{"_id": bson.M{"$in": editorIDs}}).Items()

					if err == nil {
						fields = append(fields, &discordgo.MessageEmbedField{
							Name: "Editors",
							Value: func() string {
								editorList := make([]string, len(user.Editors))
								for i, u := range editors {
									editorList[i] = fmt.Sprintf("[%s (%s)](%s)", u.DisplayName, u.Username, u.WebURL(gctx.Config().WebsiteURL))
								}

								if len(editorList) == 0 {
									return "None"
								}

								return strings.Join(editorList, "\n")
							}(),
							Inline: false,
						})
					}

					return fields
				}(),
			}

			return session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("**[user]** [%s (%s)](%s)", user.DisplayName, user.Username, user.WebURL(gctx.Config().WebsiteURL)),
					Embeds:  []*discordgo.MessageEmbed{embed},
				},
			})
		},
	)
}
