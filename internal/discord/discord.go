package discord

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type Instance interface {
	Session() *discordgo.Session
	Identity() *discordgo.User
}

type discordInst struct {
	ses *discordgo.Session
}

func New(ctx context.Context, token string) (Instance, error) {
	ses, err := discordgo.New(fmt.Sprintf("Bot %s", token))
	if err != nil {
		return nil, err
	}

	// Open connection to discord gateway
	ses.Identify.Intents = discordgo.MakeIntent(ses.Identify.Intents | discordgo.IntentsGuildMembers | discordgo.IntentDirectMessages)
	if err := ses.Open(); err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		ses.Close()
	}()

	return &discordInst{
		ses: ses,
	}, nil
}

func (di *discordInst) Session() *discordgo.Session {
	return di.ses
}

func (di *discordInst) Identity() *discordgo.User {
	return di.ses.State.User
}
