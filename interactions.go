package sevcord

import "github.com/bwmarrin/discordgo"

// NOTE: Can only have 2 of these and then a command
type SlashCommandGroup struct {
	Name        string
	Description string
	Children    []slashCommandHandleable
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
			Autocomplete: opt.Autocomplete != nil, // TODO: Autocomplete handler
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
	i        *discordgo.Interaction
	s        *discordgo.Session
	followup bool
	button   bool
}

func (i *interactionCtx) Acknowledge() {
	i.s.InteractionRespond(i.i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	i.followup = true
}

func (i *interactionCtx) Respond(r Response) {
	if i.followup {
		i.s.InteractionRespond(i.i, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})
	}

	typ := discordgo.InteractionResponseChannelMessageWithSource
	if i.button {
		typ = discordgo.InteractionResponseUpdateMessage
	}
	i.s.InteractionRespond(i.i, &discordgo.InteractionResponse{
		Type: typ,
		Data: &discordgo.InteractionResponseData{
			Content:    r.content,
			Components: r.components,
			Embeds:     []*discordgo.MessageEmbed{r.embed},
			Flags:      1 << 6, // All non-acknowledged responses are ephemeral
		},
	})
}

func (i *interactionCtx) Guild() string {
	return i.i.GuildID
}

func (c *Client) interactionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx := &interactionCtx{
		i: i.Interaction,
		s: s,
	}

	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		i := i.ApplicationCommandData()
		v, exists := c.commands[i.Name]
		if !exists {
			return
		}
		if v.isGroup() {
			var opt *discordgo.ApplicationCommandInteractionDataOption
			for v.isGroup() {
				if opt == nil {
					opt = i.Options[0]
				} else {
					opt = opt.Options[0]
				}

				for _, val := range v.(*SlashCommandGroup).Children {
					if val.name() == opt.Name {
						v = val
						break
					}
				}
			}
		}

		opts := make(map[string]any, len(i.Options))
		for _, opt := range i.Options {
			opts[opt.Name] = optToAny(opt, i)
		}

		pars := make([]any, len(v.(*SlashCommand).Options))
		for i, opt := range v.(*SlashCommand).Options {
			pars[i] = opts[opt.Name]
		}

		v.(*SlashCommand).Handler(pars, ctx)
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
