package discord

import (
	"context"

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
	ses, err := discordgo.New(token)
	if err != nil {
		return nil, err
	}

	// Open connection to discord gateway
	if err := ses.Open(); err != nil {
		return nil, err
	}

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
