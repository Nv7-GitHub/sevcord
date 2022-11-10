package sevcord

import "github.com/bwmarrin/discordgo"

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
}
