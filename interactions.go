package sevcord

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

func (s *Sevcord) checkMiddleware(ctx Ctx) bool {
	s.lock.RLock()
	for _, mid := range s.middleware {
		s.lock.RUnlock()
		if !mid(ctx) {
			return false
		}
		s.lock.RLock()
	}
	s.lock.RUnlock()
	return true
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
							res := vopt.Autocomplete(ctx, opt.Value)
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
		if s.checkMiddleware(ctx) {
			v.(*SlashCommand).Handler(ctx, pars)
		}

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

		case discordgo.SelectMenuComponent, discordgo.ChannelSelectMenuComponent, discordgo.RoleSelectMenuComponent, discordgo.UserSelectMenuComponent, discordgo.MentionableSelectMenuComponent:
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
