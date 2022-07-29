package sevcord

import (
	"sync"

	"github.com/bwmarrin/discordgo"
)

type Client struct {
	dg *discordgo.Session

	commands map[string]SlashCommandObject

	componentHandlers map[string]map[string]interface{} // map[interactionid]map[componentid]handler
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
		componentHandlers: make(map[string]map[string]interface{}),
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

type Response struct {
	content    string
	embed      *discordgo.MessageEmbed
	components [][]Component
}

func MessageResponse(message string) *Response {
	return &Response{content: message}
}

type EmbedBuilder struct {
	e *discordgo.MessageEmbed
}

func NewEmbedBuilder(title string) *EmbedBuilder {
	e := &discordgo.MessageEmbed{
		Title: title,
	}
	return &EmbedBuilder{e: e}
}

func (e *EmbedBuilder) Description(description string) *EmbedBuilder {
	e.e.Description = description
	return e
}

func (e *EmbedBuilder) Author(name, url, iconurl string) *EmbedBuilder {
	e.e.Author = &discordgo.MessageEmbedAuthor{
		Name:    name,
		URL:     url,
		IconURL: iconurl,
	}
	return e
}

func (e *EmbedBuilder) Color(color int) *EmbedBuilder {
	e.e.Color = color
	return e
}

func (e *EmbedBuilder) Footer(text, iconurl string) *EmbedBuilder {
	e.e.Footer = &discordgo.MessageEmbedFooter{
		Text:    text,
		IconURL: iconurl,
	}
	return e
}

func (e *EmbedBuilder) Image(url string) *EmbedBuilder {
	e.e.Image = &discordgo.MessageEmbedImage{
		URL: url,
	}
	return e
}

func (e *EmbedBuilder) Thumbnail(url string) *EmbedBuilder {
	e.e.Thumbnail = &discordgo.MessageEmbedThumbnail{
		URL: url,
	}
	return e
}

func (e *EmbedBuilder) Timestamp(timestamp string) *EmbedBuilder {
	e.e.Timestamp = timestamp
	return e
}

func (e *EmbedBuilder) Field(title string, description string, inline bool) *EmbedBuilder {
	e.e.Fields = append(e.e.Fields, &discordgo.MessageEmbedField{
		Name:   title,
		Value:  description,
		Inline: inline,
	})
	return e
}

func EmbedResponse(e *EmbedBuilder) *Response {
	return &Response{embed: e.e}
}

func (r *Response) ComponentRow(components ...Component) *Response {
	r.components = append(r.components, components)
	return r
}

type Ctx interface {
	Acknowledge()
	Respond(*Response)
	Guild() string
}
