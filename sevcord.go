package sevcord

import "github.com/bwmarrin/discordgo"

type Client struct {
	dg    *discordgo.Session
	appID string

	commands map[string]slashCommandHandleable
}

func NewClient(appID string, token string) (*Client, error) {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}
	return &Client{
		dg:       dg,
		appID:    appID,
		commands: make(map[string]slashCommandHandleable),
	}, nil
}

type slashCommandHandleable interface {
	name() string
	build() *discordgo.ApplicationCommandOption
	isGroup() bool
}

func (c *Client) HandleSlashCommand(cmd slashCommandHandleable) {
	c.commands[cmd.name()] = cmd
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
	_, err := c.dg.ApplicationCommandBulkOverwrite(c.appID, "", cmds)
	if err != nil {
		return err
	}
	return c.dg.Open()
}

// TODO: Embed builder, component builder, message commands

type Response struct {
	content    string
	embed      *discordgo.MessageEmbed
	components []discordgo.MessageComponent
}

func MessageResponse(message string) *Response {
	return &Response{content: message}
}

type Ctx interface {
	Acknowledge()
	Respond(Response)
	Guild() string
}
