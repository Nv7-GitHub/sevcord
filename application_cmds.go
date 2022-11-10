package sevcord

import "github.com/bwmarrin/discordgo"

type SlashCommandObject interface {
	name() string
	dg() *discordgo.ApplicationCommandOption
	isGroup() bool
	permissions() *int64
}

// NOTE: Can only have 2 levels of subcommands
type SlashCommandGroup struct {
	Name        string
	Description string
	Children    []SlashCommandObject
	Permissions *int
}

func NewSlashCommandGroup(name string, description string, children ...SlashCommandObject) *SlashCommandGroup {
	return &SlashCommandGroup{Name: name, Description: description, Children: children}
}

// RequirePermissions accepts a discordgo permissions bit mask
func (s *SlashCommandGroup) RequirePermissions(p int) {
	s.Permissions = &p
}

type SlashCommand struct {
	Name        string
	Description string
	Options     []Option
	Permissions *int
	Handler     SlashCommandHandler
}

func NewSlashCommand(name, description string, handler SlashCommandHandler, options ...Option) *SlashCommand {
	return &SlashCommand{Name: name, Description: description, Options: options, Handler: handler}
}

// RequirePermissions accepts a discordgo permissions bit mask
func (s *SlashCommand) RequirePermissions(p int) {
	s.Permissions = &p
}

func (s *SlashCommandGroup) name() string { return s.Name }
func (s *SlashCommandGroup) dg() *discordgo.ApplicationCommandOption {
	children := make([]*discordgo.ApplicationCommandOption, len(s.Children))
	for i, child := range s.Children {
		children[i] = child.dg()
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
func (s *SlashCommandGroup) permissions() *int64 {
	if s.Permissions != nil {
		v := int64(*s.Permissions)
		return &v
	}
	return nil
}
func (s *SlashCommand) name() string { return s.Name }
func (s *SlashCommand) dg() *discordgo.ApplicationCommandOption {
	opts := make([]*discordgo.ApplicationCommandOption, len(s.Options))
	for i, opt := range s.Options {
		opts[i] = &discordgo.ApplicationCommandOption{
			Name:         opt.Name,
			Description:  opt.Description,
			Type:         opt.Kind.dg(),
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
func (s *SlashCommand) permissions() *int64 {
	if s.Permissions != nil {
		v := int64(*s.Permissions)
		return &v
	}
	return nil
}

type SlashCommandAttachment struct {
	Filename    string
	URL         string
	ProxyURL    string // Use this to download
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

func (o OptionKind) dg() discordgo.ApplicationCommandOptionType {
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

func NewOption(name, description string, kind OptionKind, required bool) Option {
	return Option{Name: name, Description: description, Kind: kind, Required: required}
}

func (o Option) AddChoices(c Choice) Option {
	if o.Choices == nil {
		o.Choices = make([]Choice, 0)
	}
	o.Choices = append(o.Choices, c)
	return o
}

func (o Option) AutoComplete(a AutocompleteHandler) Option {
	o.Autocomplete = a
	return o
}

type Choice struct {
	Name  string
	Value string
}

func NewChoice(name, value string) Choice {
	return Choice{Name: name, Value: value}
}

type AutocompleteHandler func(Ctx, any) []Choice
type SlashCommandHandler func(Ctx, []any)
type ContextMenuHandler func(Ctx, string)

type ContextMenuKind int

const (
	ContextMenuKindMessage = ContextMenuKind(discordgo.MessageApplicationCommand) // The string passed is message ID
	ContextMenuKindUser    = ContextMenuKind(discordgo.UserApplicationCommand)    // The string passed is User ID
)

type ContextMenuCommand struct {
	Kind    ContextMenuKind
	Name    string
	Handler ContextMenuHandler
}

// Components

type ButtonHandler func(ctx Ctx, params string)

type Button struct {
	Label string
	Style ButtonStyle

	// Handler
	Handler string // ID of handler
	Params  string // Params to pass to handler

	// Optional
	Emoji    *ComponentEmoji
	URL      string // Only link button can have this
	Disabled bool
}

func NewButton(label string, style ButtonStyle, handler string, params string) *Button {
	return &Button{
		Label:    label,
		Style:    style,
		Handler:  handler,
		Params:   params,
		Disabled: false,
	}
}

func (b *Button) WithEmoji(emoji ComponentEmoji) *Button {
	b.Emoji = &emoji
	return b
}

func (b *Button) SetURL(url string) *Button {
	b.URL = url
	return b
}

func (b *Button) SetDisabled(disabled bool) *Button {
	b.Disabled = disabled
	return b
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

func (c ComponentEmoji) dg() discordgo.ComponentEmoji {
	return discordgo.ComponentEmoji{
		Name:     c.name,
		ID:       c.id,
		Animated: c.animated,
	}
}

const componentSeperator = "|"

type ButtonStyle int

const (
	ButtonStylePrimary   ButtonStyle = 1
	ButtonStyleSecondary ButtonStyle = 2
	ButtonStyleSuccess   ButtonStyle = 3
	ButtonStyleDanger    ButtonStyle = 4
	ButtonStyleLink      ButtonStyle = 5
)

func (b *Button) dg() discordgo.MessageComponent {
	v := discordgo.Button{
		Label:    b.Label,
		Style:    discordgo.ButtonStyle(b.Style),
		Disabled: b.Disabled,
		URL:      b.URL,
		CustomID: b.Handler + componentSeperator + b.Params,
	}
	if b.Emoji != nil {
		v.Emoji = b.Emoji.dg()
	}
	if b.Style == ButtonStyleLink {
		v.CustomID = ""
	}
	return v
}

type SelectHandler func(ctx Ctx, params string, selected []string)

type Select struct {
	Placeholder string
	Options     []SelectOption

	// Handler
	Handler string // ID of handler
	Params  string // Params to pass to handler

	// Optional
	MinValues int
	MaxValues int
	Disabled  bool
}

func NewSelect(placeholder string, handler string, params string) *Select {
	return &Select{
		Placeholder: placeholder,
		Handler:     handler,
		Params:      params,
		Disabled:    false,
		MinValues:   1,
		MaxValues:   1,
	}
}

func (s *Select) Option(option SelectOption) *Select {
	s.Options = append(s.Options, option)
	return s
}

func (s *Select) SetRange(min, max int) *Select {
	s.MinValues = min
	s.MaxValues = max
	return s
}

func (s *Select) SetDisabled(disabled bool) *Select {
	s.Disabled = disabled
	return s
}

func (s *Select) dg() discordgo.MessageComponent {
	v := discordgo.SelectMenu{
		Placeholder: s.Placeholder,
		Options:     make([]discordgo.SelectMenuOption, len(s.Options)),
		MinValues:   &s.MinValues,
		MaxValues:   s.MaxValues,
		Disabled:    s.Disabled,
		CustomID:    s.Handler + componentSeperator + s.Params,
	}
	for i, opt := range s.Options {
		v.Options[i] = discordgo.SelectMenuOption{
			Label:       opt.Label,
			Value:       opt.ID,
			Description: opt.Description,
			Default:     opt.Default,
		}
		if opt.Emoji != nil {
			v.Options[i].Emoji = opt.Emoji.dg()
		}
	}
	return v
}

type SelectOption struct {
	Label       string
	Description string
	ID          string // Must be unique

	// Optional
	Emoji   *ComponentEmoji
	Default bool // Whether it is automatically ticked
}

func NewSelectOption(label, description, id string) SelectOption {
	return SelectOption{
		Label:       label,
		Description: description,
		ID:          id,
	}
}

func (s SelectOption) WithEmoji(emoji ComponentEmoji) SelectOption {
	s.Emoji = &emoji
	return s
}

func (s SelectOption) SetDefault(defaulted bool) SelectOption {
	s.Default = defaulted
	return s
}

// Modals
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

func NewModalInput(label, placeholder string, style ModalInputStyle, maxLength int) ModalInput {
	return ModalInput{
		Label:       label,
		Placeholder: placeholder,
		Style:       style,
		Required:    true,
		MinLength:   1,
		MaxLength:   maxLength,
	}
}

func (m ModalInput) SetRequired(required bool) ModalInput {
	m.Required = required
	return m
}

func (m ModalInput) SetLength(min, max int) ModalInput {
	m.MinLength = min
	m.MaxLength = max
	return m
}

func NewModal(title string, handler ModalHandler) Modal {
	return Modal{
		Title:   title,
		Handler: handler,
		Inputs:  make([]ModalInput, 0),
	}
}

func (m Modal) Input(inp ModalInput) Modal {
	m.Inputs = append(m.Inputs, inp)
	return m
}
