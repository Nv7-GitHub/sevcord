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
	Permissions *int // Discordgo permissions bit mask, leave nil to allow everyone to use the command, only works for top-level commands
}

type SlashCommand struct {
	Name        string
	Description string
	Options     []Option
	Permissions *int // Discordgo permissions bit mask, leave nil to allow everyone to use the command, only works for top-level commands
	Handler     SlashCommandHandler
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

type Choice struct {
	Name  string
	Value string
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
