package sevcord

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

func (c *Client) interactionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx := &InteractionCtx{
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
							res := vopt.Autocomplete(ctx, optToAny(opt, dat, s))
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

		v.(*SlashCommand).Handler(ctx, pars)

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
		h, exists := handlers.handlers[parts[1]]
		if !exists {
			return
		}
		if handlers.followup != nil {
			ctx.followup = true
			ctx.followupid = *handlers.followup
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
		if handler.followup != nil {
			ctx.followup = true
			ctx.followupid = *handler.followup
		}
		vals := make([]string, len(dat.Components))
		for i, comp := range dat.Components {
			vals[i] = comp.(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
		}
		handler.handler(ctx, vals)
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
