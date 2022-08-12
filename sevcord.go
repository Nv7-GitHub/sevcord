package sevcord

// TODO: Locale support (note that it needs to be in User info too)

import (
	"log"
	"sync"

	"github.com/bwmarrin/discordgo"
)

var Logger = log.Default()

type componentHandler struct {
	handlers map[string]any // map[componentid]handler
	followup *string
}

type modalHandler struct {
	handler  ModalHandler
	followup *string
}

type Client struct {
	dg *discordgo.Session

	commands     map[string]SlashCommandObject
	contextMenus map[string]*ContextMenuCommand

	componentHandlers map[string]componentHandler // map[interactionid]handler
	modalHandlers     map[string]modalHandler     // map[interactionid]handler
	lock              *sync.RWMutex
}

func NewClient(token string) (*Client, error) {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}
	err = dg.Open()
	if err != nil {
		return nil, err
	}
	return &Client{
		dg:                dg,
		commands:          make(map[string]SlashCommandObject),
		contextMenus:      make(map[string]*ContextMenuCommand),
		componentHandlers: make(map[string]componentHandler),
		modalHandlers:     make(map[string]modalHandler),
		lock:              &sync.RWMutex{},
	}, nil
}

type SlashCommandObject interface {
	name() string
	build() *discordgo.ApplicationCommandOption
	isGroup() bool
}

func (c *Client) HandleSlashCommand(cmd SlashCommandObject) {
	c.commands[cmd.name()] = cmd
}

func (c *Client) HandleContextMenuCommand(cmd *ContextMenuCommand) {
	c.contextMenus[cmd.Name] = cmd
}

func (c *Client) Start() error {
	c.dg.Identify.Intents = discordgo.IntentsAllWithoutPrivileged // TODO: Configureable
	c.dg.AddHandler(c.interactionHandler)

	// Build commands
	cmds := make([]*discordgo.ApplicationCommand, 0, len(c.commands))
	for _, cmd := range c.commands {
		v := cmd.build()
		cmds = append(cmds, &discordgo.ApplicationCommand{
			Name:        v.Name,
			Description: v.Description,
			Options:     v.Options,
			Type:        discordgo.ChatApplicationCommand,
		})
	}
	for _, cmd := range c.contextMenus {
		cmds = append(cmds, &discordgo.ApplicationCommand{
			Name: cmd.Name,
			Type: discordgo.ApplicationCommandType(cmd.Kind),
		})
	}
	_, err := c.dg.ApplicationCommandBulkOverwrite(c.dg.State.User.ID, "", cmds)
	return err
}

func (c *Client) Close() error {
	return c.dg.Close()
}
