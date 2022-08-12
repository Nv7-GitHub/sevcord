package sevcord

import "github.com/bwmarrin/discordgo"

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
