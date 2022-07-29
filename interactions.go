package sevcord

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var Logger = log.Default()

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
	OptionKindUser                         // *User
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
	c          *Client
	i          *discordgo.Interaction
	s          *discordgo.Session
	followup   bool // ID of FollowupMessageCreate
	followupid string
	component  bool
}

func (i *interactionCtx) Acknowledge() {
	if i.component {
		return
	}

	err := i.s.InteractionRespond(i.i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		Logger.Println(err)
	}
	i.followup = true
}

func (i *interactionCtx) Respond(r *Response) {
	i.send(r, false)
}

func (i *interactionCtx) Edit(r *Response) {
	i.send(r, true)
}

func (i *interactionCtx) send(r *Response, edit bool) {
	var embs []*discordgo.MessageEmbed
	if r.embed != nil {
		embs = []*discordgo.MessageEmbed{r.embed}
	}

	var comps []discordgo.MessageComponent
	if r.components != nil {
		handlers := make(map[string]interface{})
		comps = make([]discordgo.MessageComponent, 0, len(r.components))
		for ind, r := range r.components {
			row := make([]discordgo.MessageComponent, 0, len(r))
			for j, c := range r {
				id := fmt.Sprintf("%d_%d", j, ind)
				handlers[id] = c.handler()
				v := c.build(i.i.ID + ":" + id)
				row = append(row, v)
			}
			comps = append(comps, discordgo.ActionsRow{
				Components: row,
			})
		}

		i.c.lock.Lock()
		i.c.componentHandlers[i.i.ID] = handlers
		i.c.lock.Unlock()
	}

	if i.followup {
		if edit && !i.component {
			_, err := i.s.FollowupMessageEdit(i.i, i.followupid, &discordgo.WebhookEdit{
				Content:    r.content,
				Embeds:     embs,
				Components: comps,
			})
			if err != nil {
				Logger.Println(err)
				return
			}
			return
		}

		msg, err := i.s.FollowupMessageCreate(i.i, true, &discordgo.WebhookParams{
			Content:    r.content,
			Embeds:     embs,
			Components: comps,
		})
		if err != nil {
			Logger.Println(err)
			return
		}
		i.followupid = msg.ID
		return
	}

	typ := discordgo.InteractionResponseChannelMessageWithSource
	if i.component && edit {
		typ = discordgo.InteractionResponseUpdateMessage
	}
	err := i.s.InteractionRespond(i.i, &discordgo.InteractionResponse{
		Type: typ,
		Data: &discordgo.InteractionResponseData{
			Content:    r.content,
			Components: comps,
			Embeds:     embs,
			Flags:      1 << 6, // All non-acknowledged responses are ephemeral
		},
	})
	if err != nil {
		Logger.Println(err)
		return
	}
}

func (i *interactionCtx) Guild() string {
	return i.i.GuildID
}

func (i *interactionCtx) session() *discordgo.Session {
	return i.s
}

func (i *interactionCtx) User() *User {
	if i.i.Member == nil {
		return userFromUser(i.i.User)
	}
	return userFromMember(i.i.Member)
}

func (i *interactionCtx) Channel() string {
	return i.i.ChannelID
}

func (i *interactionCtx) Modal(m *Modal) {
	comps := make([]discordgo.MessageComponent, len(m.Inputs))
	for ind, inp := range m.Inputs {
		comps[ind] = &discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.TextInput{
					CustomID:    strconv.Itoa(ind),
					Label:       inp.Label,
					Style:       discordgo.TextInputStyle(inp.Style),
					Placeholder: inp.Placeholder,
					Required:    inp.Required,
					MinLength:   inp.MinLength,
					MaxLength:   inp.MaxLength,
				},
			},
		}
	}

	i.c.lock.Lock()
	i.c.modalHandlers[i.i.ID] = m.Handler
	i.c.lock.Unlock()

	err := i.s.InteractionRespond(i.i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			Title:      m.Title,
			Components: comps,
			CustomID:   i.i.ID,
		},
	})
	if err != nil {
		Logger.Println(err)
	}
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
							res := vopt.Autocomplete(optToAny(opt, dat, s))
							choices := make([]*discordgo.ApplicationCommandOptionChoice, len(res))
							for i, choice := range res {
								choices[i] = &discordgo.ApplicationCommandOptionChoice{
									Name:  choice.Name,
									Value: choice.Value,
								}
							}

							err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
								Type: discordgo.InteractionApplicationCommandAutocompleteResult,
								Data: &discordgo.InteractionResponseData{
									Choices: choices,
								},
							})
							if err != nil {
								Logger.Println(err)
							}
							return
						}
					}
				}
			}
		}

		opts := make(map[string]any, len(dat.Options))
		for _, opt := range cmdOpts {
			opts[opt.Name] = optToAny(opt, dat, s)
		}

		pars := make([]any, len(v.(*SlashCommand).Options))
		for i, opt := range v.(*SlashCommand).Options {
			pars[i] = opts[opt.Name]
		}

		v.(*SlashCommand).Handler(pars, ctx)

	case discordgo.InteractionMessageComponent:
		dat := i.MessageComponentData()
		ctx.component = true
		parts := strings.Split(dat.CustomID, ":")
		c.lock.RLock()
		handlers, exists := c.componentHandlers[parts[0]]
		c.lock.RUnlock()
		if !exists {
			return
		}
		h, exists := handlers[parts[1]]
		if !exists {
			return
		}
		switch h := h.(type) {
		case ButtonHandler:
			h(ctx)

		case SelectHandler:
			h(ctx, dat.Values)
		}

	case discordgo.InteractionModalSubmit:
		dat := i.ModalSubmitData()
		ctx.component = true
		c.lock.RLock()
		handler, exists := c.modalHandlers[dat.CustomID]
		c.lock.RUnlock()
		if !exists {
			return
		}
		vals := make([]string, len(dat.Components))
		for i, comp := range dat.Components {
			vals[i] = comp.(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
		}
		handler(ctx, vals)
	}
}

func optToAny(opt *discordgo.ApplicationCommandInteractionDataOption, i discordgo.ApplicationCommandInteractionData, s *discordgo.Session) any {
	switch opt.Type {
	case discordgo.ApplicationCommandOptionString:
		return opt.StringValue()

	case discordgo.ApplicationCommandOptionInteger:
		return opt.IntValue()

	case discordgo.ApplicationCommandOptionBoolean:
		return opt.BoolValue()

	case discordgo.ApplicationCommandOptionUser:
		return userFromUser(opt.UserValue(s))

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

type ModalHandler func(Ctx, []string)

type Modal struct {
	Title   string
	Inputs  []ModalInput
	Handler ModalHandler
}

type ModalInputStyle int

const (
	ModalInputStyleSentence  ModalInputStyle = 1
	ModalInputStyleParagraph ModalInputStyle = 2
)

type ModalInput struct {
	Label       string
	Placeholder string
	Style       ModalInputStyle
	Required    bool
	MinLength   int
	MaxLength   int
}
