package operations

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/seventv/compactdisc"
	"github.com/seventv/compactdisc/internal/global"
	"go.uber.org/zap"
)

func SendMessage(gctx global.Context, ctx context.Context, req compactdisc.Request[compactdisc.RequestPayloadSendMessage]) error {
	channelID, ok := gctx.Config().Discord.Channels[req.Data.Channel]
	if !ok {
		return nil
	}

	var webhook *discordgo.Webhook

	if req.Data.Webhook {
		hooks, _ := gctx.Inst().Discord.Session().ChannelWebhooks(channelID)
		if len(hooks) > 0 {
			webhook = hooks[0]
		}
	}

	z := zap.S().Named("api/SendMessage").With(
		"channel_id", channelID,
	)

	var (
		msg *discordgo.Message
		err error
	)

	if webhook != nil {
		msg, err = gctx.Inst().Discord.Session().WebhookExecute(webhook.ID, webhook.Token, true, &discordgo.WebhookParams{
			Content:         req.Data.Message.Content,
			Username:        gctx.Inst().Discord.Identity().Username,
			AvatarURL:       gctx.Inst().Discord.Identity().AvatarURL("128"),
			TTS:             req.Data.Message.TTS,
			Files:           req.Data.Message.Files,
			Components:      req.Data.Message.Components,
			Embeds:          req.Data.Message.Embeds,
			AllowedMentions: req.Data.Message.AllowedMentions,
		})
	} else {
		msg, err = gctx.Inst().Discord.Session().ChannelMessageSendComplex(channelID, &req.Data.Message)
	}

	if err != nil {
		z.Errorw("failed to send message", "error", err)
		return err
	}

	z.With("message_id", msg.ID).Info("message sent")

	return nil
}
