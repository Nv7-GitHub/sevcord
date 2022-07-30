package sevcord

// TODO: Locale support (note that it needs to be in User info too)

import (
	"image"
	"io"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

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

	commands map[string]SlashCommandObject

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

type File struct {
	Name        string
	ContentType string
	Reader      io.Reader
}

type Response struct {
	content    string
	embed      *discordgo.MessageEmbed
	components [][]Component
	files      []*discordgo.File
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

func (r *Response) File(f File) *Response {
	if r.files == nil {
		r.files = make([]*discordgo.File, 0)
	}
	r.files = append(r.files, &discordgo.File{
		Name:        f.Name,
		Reader:      f.Reader,
		ContentType: f.ContentType,
	})
	return r
}

type UserFlag int

const (
	UserFlagDiscordEmployee           UserFlag = 1 << 0
	UserFlagDiscordPartner            UserFlag = 1 << 1
	UserFlagHypeSquadEvents           UserFlag = 1 << 2
	UserFlagBugHunterLevel1           UserFlag = 1 << 3
	UserFlagHouseBravery              UserFlag = 1 << 6
	UserFlagHouseBrilliance           UserFlag = 1 << 7
	UserFlagHouseBalance              UserFlag = 1 << 8
	UserFlagEarlySupporter            UserFlag = 1 << 9
	UserFlagTeamUser                  UserFlag = 1 << 10
	UserFlagSystem                    UserFlag = 1 << 12
	UserFlagBugHunterLevel2           UserFlag = 1 << 14
	UserFlagVerifiedBot               UserFlag = 1 << 16
	UserFlagVerifiedBotDeveloper      UserFlag = 1 << 17
	UserFlagDiscordCertifiedModerator UserFlag = 1 << 18
)

type User struct {
	ID            string
	Username      string
	Discriminator string // 4 numbers after name
	BannerColor   int
	Bot           bool
	System        bool
	AvatarURL     string
	BannerURL     string

	// Guild user options
	Guild          bool // Whether the following fields are filled out or not
	JoinedAt       time.Time
	Nickname       string
	Deaf           bool   // Deafened on guild level
	Mute           bool   // Muted on guild level
	GuildAvatarURL string // Blank if no custom avatar

	userflags UserFlag
}

// Avatar gets the user's avatar without making a network request
func (u *User) Avatar(ctx Ctx) (image.Image, error) {
	return ctx.session().UserAvatar(u.ID)
}
func (u *User) HasFlag(flag UserFlag) bool {
	return u.userflags&flag != 0
}

func userFromUser(d *discordgo.User) *User {
	return &User{
		ID:            d.ID,
		Username:      d.Username,
		Discriminator: d.Discriminator,
		BannerColor:   d.AccentColor,
		Bot:           d.Bot,
		System:        d.System,
		AvatarURL:     d.AvatarURL(""),
		BannerURL:     d.BannerURL(""),
		userflags:     UserFlag(d.Flags),
	}
}

func userFromMember(d *discordgo.Member) *User {
	u := userFromUser(d.User)
	u.Guild = true
	u.JoinedAt = d.JoinedAt
	u.Nickname = d.Nick
	u.Deaf = d.Deaf
	u.Mute = d.Mute
	if d.Avatar != "" {
		u.GuildAvatarURL = d.AvatarURL("")
	}
	return u
}

type ChannelType int

const (
	ChannelTypeGuildText          ChannelType = 0
	ChannelTypeDM                 ChannelType = 1
	ChannelTypeGuildVoice         ChannelType = 2
	ChannelTypeGroupDM            ChannelType = 3
	ChannelTypeGuildCategory      ChannelType = 4
	ChannelTypeGuildNews          ChannelType = 5
	ChannelTypeGuildStore         ChannelType = 6
	ChannelTypeGuildNewsThread    ChannelType = 10
	ChannelTypeGuildPublicThread  ChannelType = 11
	ChannelTypeGuildPrivateThread ChannelType = 12
	ChannelTypeGuildStageVoice    ChannelType = 13
)

type Channel struct {
	Guild  bool // Whether its a guild channel
	ID     string
	Name   string
	Topic  string
	Type   ChannelType
	NSFW   bool
	Parent string // ID of category, ID of channel a thread is in if its a thread

	// Group DMs
	Icon       string
	Recipients []*User

	// Voice channels
	Bitrate   int
	UserLimit int
}

func channelFromChannel(d *discordgo.Channel) *Channel {
	v := &Channel{
		Guild:     d.GuildID != "",
		ID:        d.ID,
		Name:      d.Name,
		Topic:     d.Topic,
		Type:      ChannelType(d.Type),
		NSFW:      d.NSFW,
		Parent:    d.ParentID,
		Icon:      d.Icon,
		Bitrate:   d.Bitrate,
		UserLimit: d.UserLimit,
	}
	if d.Recipients != nil {
		v.Recipients = make([]*User, len(d.Recipients))
		for i, r := range d.Recipients {
			v.Recipients[i] = userFromUser(r)
		}
	}
	return v
}

func GetChannel(c Ctx, id string) (*Channel, error) {
	d, err := c.session().Channel(id)
	if err != nil {
		return nil, err
	}
	return channelFromChannel(d), nil
}

func GetUser(c Ctx, id string) (*User, error) {
	d, err := c.session().User(id)
	if err != nil {
		return nil, err
	}
	return userFromUser(d), nil
}

func GetGuildUser(c Ctx, id string, guild string) (*User, error) {
	d, err := c.session().GuildMember(guild, id)
	if err != nil {
		return nil, err
	}
	return userFromMember(d), nil
}

type Ctx interface {
	Acknowledge()
	Respond(*Response)
	Edit(*Response)
	Modal(*Modal)

	Guild() string
	Channel() string
	User() *User

	session() *discordgo.Session
}
