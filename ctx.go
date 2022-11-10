package sevcord

import (
	"io"

	"github.com/bwmarrin/discordgo"
)

type MessageSend struct {
	content string
	embeds  []EmbedBuilder
	files   []struct {
		name        string
		contentType string
		reader      io.Reader
	}
	reply *struct {
		messageID string
		channelID string
		guildID   string
	}
}

type EmbedBuilder struct {
	url         string
	title       string
	description string
	timestamp   string
	color       int
	footerText  string
	footerIcon  string
	image       string
	thumbnail   string
	authorURL   string
	authorName  string
	authorIcon  string
	fields      []struct {
		name   string
		value  string
		inline bool
	}
}

type Ctx interface {
	Dg() *discordgo.Session        // Allows access to underlying discordgo session
	Acknowledge() error            // Indicates progress
	Respond(msg MessageSend) error // Displays message to user (note: in interactions, if not acknowledged this will be ephemeral)
}

// Builder methods
func NewEmbed() EmbedBuilder {
	return EmbedBuilder{fields: make([]struct {
		name   string
		value  string
		inline bool
	}, 0)}
}

func (e EmbedBuilder) URL(url string) EmbedBuilder {
	e.url = url
	return e
}

func (e EmbedBuilder) Title(title string) EmbedBuilder {
	e.title = title
	return e
}

func (e EmbedBuilder) Description(description string) EmbedBuilder {
	e.description = description
	return e
}

func (e EmbedBuilder) Timestamp(timestamp string) EmbedBuilder {
	e.timestamp = timestamp
	return e
}

func (e EmbedBuilder) Color(color int) EmbedBuilder {
	e.color = color
	return e
}

func (e EmbedBuilder) Footer(text, iconURL string) EmbedBuilder {
	e.footerText = text
	e.footerIcon = iconURL
	return e
}

func (e EmbedBuilder) Image(url string) EmbedBuilder {
	e.image = url
	return e
}

func (e EmbedBuilder) Thumbnail(url string) EmbedBuilder {
	e.thumbnail = url
	return e
}

func (e EmbedBuilder) Author(url, name, iconURL string) EmbedBuilder {
	e.authorURL = url
	e.authorName = name
	e.authorIcon = iconURL
	return e
}

func (e EmbedBuilder) AddField(name, value string, inline bool) EmbedBuilder {
	e.fields = append(e.fields, struct {
		name   string
		value  string
		inline bool
	}{name, value, inline})
	return e
}

func NewMessage(content string) MessageSend {
	return MessageSend{content: content,
		files: make([]struct {
			name        string
			contentType string
			reader      io.Reader
		}, 0),
		embeds: make([]EmbedBuilder, 0),
	}
}

func (m MessageSend) Content(content string) MessageSend {
	m.content = content
	return m
}

func (m MessageSend) AddEmbed(embed EmbedBuilder) MessageSend {
	m.embeds = append(m.embeds, embed)
	return m
}

func (m MessageSend) AddFile(name, contentType string, reader io.Reader) MessageSend {
	m.files = append(m.files, struct {
		name        string
		contentType string
		reader      io.Reader
	}{name, contentType, reader})
	return m
}

func (m MessageSend) Reply(messageID, channelID, guildID string) MessageSend {
	m.reply = &struct {
		messageID string
		channelID string
		guildID   string
	}{messageID, channelID, guildID}
	return m
}

func (e EmbedBuilder) dg() *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		URL:         e.url,
		Title:       e.title,
		Description: e.description,
		Timestamp:   e.timestamp,
		Color:       e.color,
	}
	if e.image != "" {
		embed.Image = &discordgo.MessageEmbedImage{
			URL: e.image,
		}
	}
	if e.footerText != "" || e.footerIcon != "" {
		embed.Footer = &discordgo.MessageEmbedFooter{
			Text:    e.footerText,
			IconURL: e.footerIcon,
		}
	}
	if e.thumbnail != "" {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: e.thumbnail,
		}
	}
	if e.authorURL != "" || e.authorName != "" || e.authorIcon != "" {
		embed.Author = &discordgo.MessageEmbedAuthor{
			URL:     e.authorURL,
			Name:    e.authorName,
			IconURL: e.authorIcon,
		}
	}
	fields := make([]*discordgo.MessageEmbedField, len(e.fields))
	for i, field := range e.fields {
		fields[i] = &discordgo.MessageEmbedField{
			Name:   field.name,
			Value:  field.value,
			Inline: field.inline,
		}
	}
	embed.Fields = fields
	return embed
}

func (m MessageSend) dg() *discordgo.MessageSend {
	msg := &discordgo.MessageSend{
		Content: m.content,
		Embeds:  make([]*discordgo.MessageEmbed, len(m.embeds)),
		Files:   make([]*discordgo.File, len(m.files)),
	}
	for i, embed := range m.embeds {
		msg.Embeds[i] = embed.dg()
	}
	for i, file := range m.files {
		msg.Files[i] = &discordgo.File{
			Name:        file.name,
			ContentType: file.contentType,
			Reader:      file.reader,
		}
	}
	if m.reply != nil {
		msg.Reference = &discordgo.MessageReference{
			MessageID: m.reply.messageID,
			ChannelID: m.reply.channelID,
			GuildID:   m.reply.guildID,
		}
	}
	return msg
}
