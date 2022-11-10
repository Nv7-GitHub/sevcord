package sevcord

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

type Sevcord struct {
	dg       *discordgo.Session // Note: only use this to create cmds, give one provided with handlers for user
	commands map[string]SlashCommandObject
}

func (s *Sevcord) RegisterSlashCommand(cmd SlashCommandObject) {
	s.commands[cmd.name()] = cmd
}

func New(token string) (*Sevcord, error) {
	dg, err := discordgo.New(strings.TrimSpace(token))
	if err != nil {
		return nil, err
	}
	return &Sevcord{
		dg:       dg,
		commands: make(map[string]SlashCommandObject),
	}, nil
}

func (s *Sevcord) Start() {
	s.dg.AddHandler(s.interactionHandler)
	s.dg.Identify.Intents = discordgo.IntentsNone
	s.dg.Open()
}
