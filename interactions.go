package sevcord

import (
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type InteractionCtx struct {
	dg           *discordgo.Session
	i            *discordgo.Interaction
	s            *Sevcord
	acknowledged bool
	component    bool // If component, then update
}

func (i *InteractionCtx) Dg() *discordgo.Session {
	return i.dg
}

func (i *InteractionCtx) Acknowledge() error {
	i.acknowledged = true
	if i.component { // if component, then make it so that response will be ephemeral instead of update
		return nil
	}
	return i.dg.InteractionRespond(i.i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
}

func (i *InteractionCtx) Respond(msg MessageSend) error {
	b := msg.dg()
	if i.acknowledged && !i.component {
		_, err := i.dg.FollowupMessageCreate(i.i, true, &discordgo.WebhookParams{
			Content:    b.Content,
			Files:      b.Files,
			Embeds:     b.Embeds,
			Components: b.Components,
		})
		return err
	}
	if i.component && !i.acknowledged { // if not acknowledged, then update instead of ephemeral (no non-ephemeral response allowed on components since no one needs that)
		return i.dg.InteractionRespond(i.i, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content:    b.Content,
				Files:      b.Files,
				Embeds:     b.Embeds,
				Components: b.Components,
			},
		})
	}
	return i.dg.InteractionRespond(i.i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:    b.Content,
			Files:      b.Files,
			Embeds:     b.Embeds,
			Components: b.Components,
			Flags:      1 << 6, // Ephemeral
		},
	})
}

func (i *InteractionCtx) Author() *discordgo.Member {
	return i.i.Member
}

func (i *InteractionCtx) Modal(m Modal) error {
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

	i.s.lock.Lock()
	i.s.modalHandlers[i.i.ID] = m.Handler
	i.s.lock.Unlock()

	return i.dg.InteractionRespond(i.i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			Title:      m.Title,
			Components: comps,
			CustomID:   i.i.ID,
		},
	})
}

func (i *InteractionCtx) Channel() string {
	return i.i.ChannelID
}

func (i *InteractionCtx) Guild() string {
	return i.i.GuildID
}

func (s *Sevcord) interactionHandler(dg *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx := &InteractionCtx{
		dg: dg,
		i:  i.Interaction,
		s:  s,
	}

	switch i.Type {
	case discordgo.InteractionApplicationCommand, discordgo.InteractionApplicationCommandAutocomplete:
		dat := i.ApplicationCommandData()
		s.lock.RLock()
		v, exists := s.commands[dat.Name]
		s.lock.RUnlock()
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
							res := vopt.Autocomplete(ctx, optToAny(opt, dat, dg))
							choices := make([]*discordgo.ApplicationCommandOptionChoice, len(res))
							for i, choice := range res {
								choices[i] = &discordgo.ApplicationCommandOptionChoice{
									Name:  choice.Name,
									Value: choice.Value,
								}
							}

							err := dg.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
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
			opts[opt.Name] = optToAny(opt, dat, dg)
		}

		pars := make([]any, len(v.(*SlashCommand).Options))
		for i, opt := range v.(*SlashCommand).Options {
			pars[i] = opts[opt.Name]
		}

		// Check midleware
		s.lock.RLock()
		for _, mid := range s.middleware {
			s.lock.RUnlock()
			if !mid(ctx) {
				return
			}
			s.lock.RLock()
		}
		s.lock.RUnlock()

		v.(*SlashCommand).Handler(ctx, pars)

	case discordgo.InteractionMessageComponent:
		ctx.component = true
		dat := i.MessageComponentData()
		parts := strings.SplitN(dat.CustomID, "|", 2)
		switch dat.ComponentType {
		case discordgo.ButtonComponent:
			s.lock.RLock()
			v, exists := s.buttonHandlers[parts[0]]
			s.lock.RUnlock()
			if !exists {
				return
			}
			v(ctx, parts[1])

		case discordgo.SelectMenuComponent:
			s.lock.RLock()
			v, exists := s.selectHandlers[parts[0]]
			s.lock.RUnlock()
			if !exists {
				return
			}

			v(ctx, parts[1], dat.Values)
		}

	case discordgo.InteractionModalSubmit:
		dat := i.ModalSubmitData()
		ctx.component = true
		s.lock.RLock()
		handler, exists := s.modalHandlers[dat.CustomID]
		s.lock.RUnlock()
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
		return opt.UserValue(s)

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
			ProxyURL:    att.ProxyURL,
			ContentType: att.ContentType,
		}

	default:
		return nil
	}
}
