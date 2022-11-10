package sevcord

import (
	"github.com/bwmarrin/discordgo"
)

type InteractionCtx struct {
	dg           *discordgo.Session
	i            *discordgo.Interaction
	acknowledged bool
}

func (i *InteractionCtx) Dg() *discordgo.Session {
	return i.dg
}

func (i *InteractionCtx) Acknowledge() error {
	i.acknowledged = true
	return i.dg.InteractionRespond(i.i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
}

func (i *InteractionCtx) Respond(msg MessageSend) error {
	b := msg.dg()
	if i.acknowledged {
		_, err := i.dg.FollowupMessageCreate(i.i, true, &discordgo.WebhookParams{
			Content:    b.Content,
			Files:      b.Files,
			Embeds:     b.Embeds,
			Components: b.Components,
		})
		return err
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

func (s *Sevcord) interactionHandler(dg *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx := &InteractionCtx{
		dg: dg,
		i:  i.Interaction,
	}

	switch i.Type {
	case discordgo.InteractionApplicationCommand, discordgo.InteractionApplicationCommandAutocomplete:
		dat := i.ApplicationCommandData()
		v, exists := s.commands[dat.Name]
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

		v.(*SlashCommand).Handler(ctx, pars)
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
