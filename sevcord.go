package sevcord

import "github.com/bwmarrin/discordgo"

type Client struct {
	dg *discordgo.Session

	commands map[string]SlashCommandObject
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
		dg:       dg,
		commands: make(map[string]SlashCommandObject),
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
	_, err := c.dg.ApplicationCommandBulkOverwrite(c.dg.State.User.ID, "", cmds)
	return err
}

func (c *Client) Close() error {
	return c.dg.Close()
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
	Respond(*Response)
	Guild() string
}
