package sevcord

import (
	"github.com/bwmarrin/discordgo"
)

// NOTE: Can only have 2 of these and then a command
type SlashCommandGroup struct {
	Name        string
	Description string
	Children    []SlashCommandObject
}

// TODO: Application command perms
type SlashCommand struct {
	Name        string
	Description string
	Options     []Option
	Handler     SlashCommandHandler
}

func (s *SlashCommandGroup) name() string { return s.Name }
func (s *SlashCommandGroup) build() *discordgo.ApplicationCommandOption {
	children := make([]*discordgo.ApplicationCommandOption, len(s.Children))
	for i, child := range s.Children {
		children[i] = child.build()
		if child.isGroup() {
			children[i].Type = discordgo.ApplicationCommandOptionSubCommandGroup
		} else {
			children[i].Type = discordgo.ApplicationCommandOptionSubCommand
		}
	}
	return &discordgo.ApplicationCommandOption{
		Name:        s.Name,
		Description: s.Description,
		Options:     children,
	}
}
func (s *SlashCommandGroup) isGroup() bool { return true }
func (s *SlashCommand) name() string       { return s.Name }
func (s *SlashCommand) build() *discordgo.ApplicationCommandOption {
	opts := make([]*discordgo.ApplicationCommandOption, len(s.Options))
	for i, opt := range s.Options {
		opts[i] = &discordgo.ApplicationCommandOption{
			Name:         opt.Name,
			Description:  opt.Description,
			Type:         opt.Kind.build(),
			Required:     opt.Required,
			Autocomplete: opt.Autocomplete != nil,
		}
		if opt.Choices != nil {
			opts[i].Choices = make([]*discordgo.ApplicationCommandOptionChoice, len(opt.Choices))
			for j, choice := range opt.Choices {
				opts[i].Choices[j] = &discordgo.ApplicationCommandOptionChoice{
					Name:  choice.Name,
					Value: choice.Value,
				}
			}
		}
	}
	return &discordgo.ApplicationCommandOption{
		Name:        s.Name,
		Description: s.Description,
		Options:     opts,
	}
}
func (s *SlashCommand) isGroup() bool { return false }

type SlashCommandAttachment struct {
	Filename    string
	URL         string
	ContentType string
}

type OptionKind int

const (
	OptionKindString     OptionKind = iota // string
	OptionKindInt                          // int
	OptionKindBool                         // bool
	OptionKindUser                         // user id (string)
	OptionKindChannel                      // channel id (string)
	OptionKindRole                         // role id (string)
	OptionKindFloat                        // float64
	OptionKindAttachment                   // *SlashCommandAttachment
)

func (o OptionKind) build() discordgo.ApplicationCommandOptionType {
	return [...]discordgo.ApplicationCommandOptionType{discordgo.ApplicationCommandOptionString, discordgo.ApplicationCommandOptionInteger, discordgo.ApplicationCommandOptionBoolean, discordgo.ApplicationCommandOptionUser, discordgo.ApplicationCommandOptionChannel, discordgo.ApplicationCommandOptionRole, discordgo.ApplicationCommandOptionNumber, discordgo.ApplicationCommandOptionAttachment}[o]
}

type Option struct {
	Name        string
	Description string
	Kind        OptionKind
	Required    bool

	Choices      []Choice            // Optional
	Autocomplete AutocompleteHandler // Optional
}

type Choice struct {
	Name  string
	Value string
}

type AutocompleteHandler func(any) []Choice
type SlashCommandHandler func([]any, Ctx)

type interactionCtx struct {
	c         *Client
	i         *discordgo.Interaction
	s         *discordgo.Session
	followup  bool // ID of followup
	component bool
}

func (i *interactionCtx) Acknowledge() {
	if i.component {
		return
	}

	i.s.InteractionRespond(i.i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	i.followup = true
}

func (i *interactionCtx) Respond(r *Response) {
	var embs []*discordgo.MessageEmbed
	if r.embed != nil {
		embs = []*discordgo.MessageEmbed{r.embed}
	}

	if i.followup {
		msg, err := i.s.FollowupMessageCreate(i.i, true, &discordgo.WebhookParams{
			Content:    r.content,
			Embeds:     embs,
			Components: r.components,
		})
		if err != nil {
			return
		}
		if r.componentHandlers != nil {
			i.c.componentHandlers[msg.ID] = r.componentHandlers
		}
		return
	}

	typ := discordgo.InteractionResponseChannelMessageWithSource
	if i.component {
		typ = discordgo.InteractionResponseUpdateMessage
	}
	err := i.s.InteractionRespond(i.i, &discordgo.InteractionResponse{
		Type: typ,
		Data: &discordgo.InteractionResponseData{
			Content:    r.content,
			Components: r.components,
			Embeds:     embs,
			Flags:      1 << 6, // All non-acknowledged responses are ephemeral
		},
	})
	if err != nil {
		return
	}
	if r.componentHandlers != nil {
		i.c.componentHandlers[i.i.ID] = r.componentHandlers // TODO: Make this the ID of the ephemeral message instead of interaction id so ephemeral components work
	}
}

func (i *interactionCtx) Guild() string {
	return i.i.GuildID
}

func (c *Client) interactionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx := &interactionCtx{
		i: i.Interaction,
		s: s,
		c: c,
	}

	switch i.Type {
	case discordgo.InteractionApplicationCommand, discordgo.InteractionApplicationCommandAutocomplete:
		dat := i.ApplicationCommandData()
		v, exists := c.commands[dat.Name]
		if !exists {
			return
		}
		cmdOpts := dat.Options
		if v.isGroup() {
			var opt *discordgo.ApplicationCommandInteractionDataOption
			for v.isGroup() {
				if opt == nil {
					opt = dat.Options[0]
				} else {
					cmdOpts = opt.Options
					opt = opt.Options[0]
				}

				for _, val := range v.(*SlashCommandGroup).Children {
					if val.name() == opt.Name {
						v = val
						cmdOpts = opt.Options
						break
					}
				}
			}
		}

		// If autocomplete
		if i.Type == discordgo.InteractionApplicationCommandAutocomplete {
			for _, opt := range cmdOpts {
				if opt.Focused {
					for _, vopt := range v.(*SlashCommand).Options {
						if opt.Name == vopt.Name {
							res := vopt.Autocomplete(optToAny(opt, dat))
							choices := make([]*discordgo.ApplicationCommandOptionChoice, len(res))
							for i, choice := range res {
								choices[i] = &discordgo.ApplicationCommandOptionChoice{
									Name:  choice.Name,
									Value: choice.Value,
								}
							}

							s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
								Type: discordgo.InteractionApplicationCommandAutocompleteResult,
								Data: &discordgo.InteractionResponseData{
									Choices: choices,
								},
							})
							return
						}
					}
				}
			}
		}

		opts := make(map[string]any, len(dat.Options))
		for _, opt := range cmdOpts {
			opts[opt.Name] = optToAny(opt, dat)
		}

		pars := make([]any, len(v.(*SlashCommand).Options))
		for i, opt := range v.(*SlashCommand).Options {
			pars[i] = opts[opt.Name]
		}

		v.(*SlashCommand).Handler(pars, ctx)

	case discordgo.InteractionMessageComponent:
		dat := i.MessageComponentData()
		ctx.component = true
		handlers, exists := c.componentHandlers[i.Message.ID]
		if !exists {
			return
		}
		h, exists := handlers[dat.CustomID]
		if !exists {
			return
		}
		switch h := h.(type) {
		case ButtonHandler:
			h(ctx)

		case SelectHandler:
			h(ctx, dat.Values)
		}
	}
}

func optToAny(opt *discordgo.ApplicationCommandInteractionDataOption, i discordgo.ApplicationCommandInteractionData) any {
	switch opt.Type {
	case discordgo.ApplicationCommandOptionString:
		return opt.StringValue()

	case discordgo.ApplicationCommandOptionInteger:
		return opt.IntValue()

	case discordgo.ApplicationCommandOptionBoolean:
		return opt.BoolValue()

	case discordgo.ApplicationCommandOptionUser:
		return opt.Value.(string)

	case discordgo.ApplicationCommandOptionChannel:
		return opt.Value.(string)

	case discordgo.ApplicationCommandOptionRole:
		return opt.Value.(string)

	case discordgo.ApplicationCommandOptionNumber:
		return opt.FloatValue()

	case discordgo.ApplicationCommandOptionAttachment:
		id := opt.Value.(string)
		att := i.Resolved.Attachments[id]
		return &SlashCommandAttachment{
			Filename:    att.Filename,
			URL:         att.URL,
			ContentType: att.ContentType,
		}

	default:
		return nil
	}
}

type ButtonHandler func(ctx Ctx)

type Button struct {
	Label   string
	Style   ButtonStyle
	Handler ButtonHandler

	// Optional
	Emoji    *ComponentEmoji
	URL      string // Only link button can have this
	Disabled bool
}

type ComponentEmoji struct {
	name     string // Make this the raw emoji for a builtin emoji, otherwise the custom emoji name and use ID
	id       string
	animated bool
}

func ComponentEmojiDefault(emoji rune) *ComponentEmoji {
	return &ComponentEmoji{
		name: string(emoji),
	}
}

func ComponentEmojiCustom(name, id string, animated bool) *ComponentEmoji {
	return &ComponentEmoji{
		name:     name,
		id:       id,
		animated: animated,
	}
}

type ButtonStyle int

const (
	ButtonStylePrimary   ButtonStyle = 1
	ButtonStyleSecondary ButtonStyle = 2
	ButtonStyleSuccess   ButtonStyle = 3
	ButtonStyleDanger    ButtonStyle = 4
	ButtonStyleLink      ButtonStyle = 5
)

func (b *Button) build(customid string) discordgo.MessageComponent {
	v := discordgo.Button{
		Label:    b.Label,
		Style:    discordgo.ButtonStyle(b.Style),
		Disabled: b.Disabled,
		URL:      b.URL,
		CustomID: customid,
	}
	if b.Emoji != nil {
		v.Emoji = discordgo.ComponentEmoji{
			Name:     b.Emoji.name,
			ID:       b.Emoji.id,
			Animated: b.Emoji.animated,
		}
	}
	if b.Style == ButtonStyleLink {
		v.CustomID = ""
	}
	return v
}

func (b *Button) handler() any {
	return b.Handler
}

type SelectHandler func(ctx Ctx, opts []string) // opts is the IDs of the options that are selected

type Select struct {
	Placeholder string
	Options     []SelectOption
	Handler     SelectHandler

	// Optional
	MinValues int
	MaxValues int
	Disabled  bool
}

func (s *Select) build(customid string) discordgo.MessageComponent {
	v := discordgo.SelectMenu{
		Placeholder: s.Placeholder,
		Options:     make([]discordgo.SelectMenuOption, len(s.Options)),
		MinValues:   &s.MinValues,
		MaxValues:   s.MaxValues,
		Disabled:    s.Disabled,
		CustomID:    customid,
	}
	for i, opt := range s.Options {
		v.Options[i] = discordgo.SelectMenuOption{
			Label:       opt.Label,
			Value:       opt.ID,
			Description: opt.Description,
			Default:     opt.Default,
		}
		if opt.Emoji != nil {
			v.Options[i].Emoji = discordgo.ComponentEmoji{
				Name:     opt.Emoji.name,
				ID:       opt.Emoji.id,
				Animated: opt.Emoji.animated,
			}
		}
	}
	return v
}

func (s *Select) handler() any {
	return s.Handler
}

type SelectOption struct {
	Label       string
	Description string
	ID          string // Must be unique

	// Optional
	Emoji   *ComponentEmoji
	Default bool // Whether it is automatically ticked
}

type Component interface {
	build(customid string) discordgo.MessageComponent
	handler() any
}
