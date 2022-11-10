package sevcord

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var Logger = log.Default()

type Sevcord struct {
	dg       *discordgo.Session // Note: only use this to create cmds, give one provided with handlers for user
	commands map[string]SlashCommandObject
}

func (s *Sevcord) RegisterSlashCommand(cmd SlashCommandObject) {
	s.commands[cmd.name()] = cmd
}

func New(token string) (*Sevcord, error) {
	dg, err := discordgo.New("Bot " + strings.TrimSpace(token))
	if err != nil {
		return nil, err
	}
	return &Sevcord{
		dg:       dg,
		commands: make(map[string]SlashCommandObject),
	}, nil
}

func (s *Sevcord) Listen() {
	// Build commands
	cmds := make([]*discordgo.ApplicationCommand, 0, len(s.commands))
	for _, cmd := range s.commands {
		v := cmd.dg()
		cmds = append(cmds, &discordgo.ApplicationCommand{
			Name:                     v.Name,
			Description:              v.Description,
			Options:                  v.Options,
			Type:                     discordgo.ChatApplicationCommand,
			DefaultMemberPermissions: cmd.permissions(),
		})
	}

	// Handlers
	s.dg.AddHandler(s.interactionHandler)
	s.dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		_, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, "", cmds)
		if err != nil {
			Logger.Println("Error updating commands", err)
		}
		Logger.Println("Bot Ready")
	})

	// Open
	s.dg.Identify.Intents = discordgo.IntentsNone
	s.dg.Open()

	// Wait
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	fmt.Println("Gracefully shutting down...")

	// Close
	s.dg.Close()
}
