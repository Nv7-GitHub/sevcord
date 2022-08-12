package sevcord

import (
	"errors"
	"image"
	"io"
	"sort"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Color represents a discord color, and is the decimal representation of a hex color. For example: blue would be `Color(0x0000FF)`
type Color int

// Permissions
type Permissions int64

const (
	// Text channel permissions
	PermissionViewChannel           = Permissions(discordgo.PermissionViewChannel)
	PermissionSendMessages          = Permissions(discordgo.PermissionSendMessages)
	PermissionSendTTSMessages       = Permissions(discordgo.PermissionSendTTSMessages)
	PermissionManageMessages        = Permissions(discordgo.PermissionManageMessages)
	PermissionEmbedLinks            = Permissions(discordgo.PermissionEmbedLinks)
	PermissionAttachFiles           = Permissions(discordgo.PermissionAttachFiles)
	PermissionReadMessageHistory    = Permissions(discordgo.PermissionReadMessageHistory)
	PermissionMentionEveryone       = Permissions(discordgo.PermissionMentionEveryone)
	PermissionUseExternalEmojis     = Permissions(discordgo.PermissionUseExternalEmojis)
	PermissionUseSlashCommands      = Permissions(discordgo.PermissionUseSlashCommands)
	PermissionManageThreads         = Permissions(discordgo.PermissionManageThreads)
	PermissionCreatePublicThreads   = Permissions(discordgo.PermissionCreatePublicThreads)
	PermissionCreatePrivateThreads  = Permissions(discordgo.PermissionCreatePrivateThreads)
	PermissionUseExternalStickers   = Permissions(discordgo.PermissionUseExternalStickers)
	PermissionSendMessagesInThreads = Permissions(discordgo.PermissionSendMessagesInThreads)
	PermissionAddReactions          = Permissions(discordgo.PermissionAddReactions)

	// Voice channel permissions
	PermissionVoiceConnect         = Permissions(discordgo.PermissionVoiceConnect)
	PermissionVoiceSpeak           = Permissions(discordgo.PermissionVoiceSpeak)
	PermissionVoiceMuteMembers     = Permissions(discordgo.PermissionVoiceMuteMembers)
	PermissionVoiceDeafenMembers   = Permissions(discordgo.PermissionVoiceDeafenMembers)
	PermissionVoiceMoveMembers     = Permissions(discordgo.PermissionVoiceMoveMembers)
	PermissionVoiceUseVAD          = Permissions(discordgo.PermissionVoiceUseVAD)
	PermissionVoicePrioritySpeaker = Permissions(discordgo.PermissionVoicePrioritySpeaker)
	PermissionVoiceStreamVideo     = Permissions(discordgo.PermissionVoiceStreamVideo)
	PermissionUseActivities        = Permissions(discordgo.PermissionUseActivities)

	// Management permissions
	PermissionChangeNickname      = Permissions(discordgo.PermissionChangeNickname)
	PermissionManageNicknames     = Permissions(discordgo.PermissionManageNicknames)
	PermissionManageRoles         = Permissions(discordgo.PermissionManageRoles)
	PermissionManageWebhooks      = Permissions(discordgo.PermissionManageWebhooks)
	PermissionManageEmojis        = Permissions(discordgo.PermissionManageEmojis)
	PermissionManageEvents        = Permissions(discordgo.PermissionManageEvents)
	PermissionCreateInstantInvite = Permissions(discordgo.PermissionCreateInstantInvite)
	PermissionKickMembers         = Permissions(discordgo.PermissionKickMembers)
	PermissionBanMembers          = Permissions(discordgo.PermissionBanMembers)
	PermissionAdministrator       = Permissions(discordgo.PermissionAdministrator)
	PermissionManageChannels      = Permissions(discordgo.PermissionManageChannels)
	PermissionManageServer        = Permissions(discordgo.PermissionManageServer)
	PermissionViewAuditLogs       = Permissions(discordgo.PermissionViewAuditLogs)
	PermissionViewGuildInsights   = Permissions(discordgo.PermissionViewGuildInsights)
	PermissionModerateMembers     = Permissions(discordgo.PermissionModerateMembers)
)

func CombinePermissions(perms ...Permissions) Permissions {
	var p Permissions
	for _, perm := range perms {
		p |= perm
	}
	return p
}

func (p Permissions) Add(perms ...Permissions) Permissions {
	out := p
	for _, perm := range perms {
		out |= perm
	}
	return out
}

func (p Permissions) Has(perm Permissions) bool {
	return p&perm == perm
}

// Responses
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

// Users
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
	Guild          *string // Whether the following fields are filled out or not (if nil, then no, otherwise it is the Guild ID)
	JoinedAt       time.Time
	Nickname       string
	Deaf           bool     // Deafened on guild level
	Mute           bool     // Muted on guild level
	GuildAvatarURL string   // Blank if no custom avatar
	Roles          []string // IDs of roles
	Permissions    Permissions

	userflags UserFlag
}

// Avatar gets the user's avatar without making a network request
func (u *User) Avatar(ctx Ctx) (image.Image, error) {
	return ctx.session().UserAvatar(u.ID)
}
func (u *User) HasFlag(flag UserFlag) bool {
	return u.userflags&flag != 0
}
func (u *User) GiveRole(c Ctx, roleID string) error {
	return c.session().GuildMemberRoleAdd(*u.Guild, u.ID, roleID)
}
func (u *User) RemoveRole(c Ctx, roleID string) error {
	return c.session().GuildMemberRoleRemove(*u.Guild, u.ID, roleID)
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
	u.Guild = &d.GuildID
	u.JoinedAt = d.JoinedAt
	u.Nickname = d.Nick
	u.Deaf = d.Deaf
	u.Mute = d.Mute
	u.Roles = d.Roles
	u.Permissions = Permissions(d.Permissions)
	if d.Avatar != "" {
		u.GuildAvatarURL = d.AvatarURL("")
	}
	return u
}

// Channels
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

// Roles

type Role struct {
	// Not used for creating role
	ID       string
	Position int  // Position in role hierarchy
	Managed  bool // If its managed by an integration (if true cannot give/take role)

	// Required to create role
	Name        string
	Mentionable bool
	Hoist       bool  // Shows up seperately in member list
	Color       Color // Hex color of role
	Permissions Permissions
}

func roleFromDg(r *discordgo.Role) *Role {
	return &Role{
		ID:          r.ID,
		Name:        r.Name,
		Mentionable: r.Mentionable,
		Hoist:       r.Hoist,
		Color:       Color(r.Color),
		Position:    r.Position,
		Managed:     r.Managed,
		Permissions: Permissions(r.Permissions),
	}
}

// GetGuildRole gets a role from a guild (NOTE: This will try to get the role from the cache if possible)
func GetGuildRole(c Ctx, id string, guild string) (*Role, error) {
	v, err := c.session().State.Role(guild, id)
	if err == nil {
		return roleFromDg(v), nil
	}
	r, err := c.session().GuildRoles(guild)
	if err != nil {
		return nil, err
	}
	for _, v := range r {
		if v.ID == id {
			return roleFromDg(v), nil
		}
	}

	return nil, errors.New("sevcord: role not found")
}

// GetGuildRoles gets a sorted list of all roles in a guild
func GetGuildRoles(c Ctx, guild string) ([]*Role, error) {
	r, err := c.session().GuildRoles(guild)
	if err != nil {
		return nil, err
	}
	roles := make([]*Role, len(r))
	for i, v := range r {
		roles[i] = roleFromDg(v)
	}
	sort.Slice(roles, func(i, j int) bool {
		return roles[i].Position < roles[j].Position
	})
	return roles, nil
}

// CreateRole creates the role and returns the role created
func CreateRole(c Ctx, r *Role, guild string) (*Role, error) {
	dgr, err := c.session().GuildRoleCreate(guild)
	if err != nil {
		return nil, err
	}
	dgr, err = c.session().GuildRoleEdit(guild, dgr.ID, r.Name, int(r.Color), r.Hoist, int64(r.Permissions), r.Mentionable)
	if err != nil {
		return nil, err
	}
	return roleFromDg(dgr), nil
}

// DeleteRole deletes a role
func DeleteRole(c Ctx, id string, guild string) error {
	return c.session().GuildRoleDelete(guild, id)
}
