package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/seventv/compactdisc/internal/global"
	"go.uber.org/zap"
)

type Command struct {
	Data    *discordgo.ApplicationCommand
	Handler CommandHandler
}

func DefineCommand(data *discordgo.ApplicationCommand, handler CommandHandler) *Command {
	return &Command{
		Data:    data,
		Handler: handler,
	}
}

type CommandHandler func(session *discordgo.Session, interaction *discordgo.InteractionCreate) error

func Setup(gctx global.Context) error {
	disc := gctx.Inst().Discord.Session()
	appID := gctx.Inst().Discord.Identity().ID
	guildID := gctx.Config().Discord.GuildID

	registeredCommands, _ := disc.ApplicationCommands(appID, guildID)
	for _, cmd := range registeredCommands {
		disc.ApplicationCommandDelete(appID, guildID, cmd.ID)
	}

	commands := []*Command{
		UserInfo(gctx, appID, guildID),
	}

	for _, cmd := range commands {
		_, err := disc.ApplicationCommandCreate(appID, gctx.Config().Discord.GuildID, cmd.Data)
		if err != nil {
			zap.S().Errorw("failed to setup commands", "error", err)
		}

		disc.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if err := cmd.Handler(s, i); err != nil {
				zap.S().Errorw("failed to handle command", "command", cmd.Data.Name, "error", err)

				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content:         err.Error(),
						AllowedMentions: &discordgo.MessageAllowedMentions{},
						Flags:           discordgo.MessageFlagsEphemeral,
					},
				})
			}
		})
	}

	return nil
}
