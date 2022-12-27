package sevcord

import (
	"strconv"

	"github.com/bwmarrin/discordgo"
)

// MessageCtx represents a message context
type MessageCtx struct {
	m            *discordgo.Message
	d            *discordgo.Session
	acknowledged bool
}

func (m *MessageCtx) Dg() *discordgo.Session {
	return m.d
}

func (m *MessageCtx) Acknowledge() error {
	if m.acknowledged {
		return nil
	}
	m.acknowledged = true
	return m.d.ChannelTyping(m.m.ChannelID)
}

func (m *MessageCtx) Respond(msg MessageSend) error {
	v := msg.Dg()
	v.Reference = &discordgo.MessageReference{
		MessageID: m.m.ID,
		ChannelID: m.m.ChannelID,
		GuildID:   m.m.GuildID,
	}
	_, err := m.d.ChannelMessageSendComplex(m.m.ChannelID, v)
	return err
}

func (m *MessageCtx) Author() *discordgo.Member {
	v := m.m.Member
	v.User = m.m.Author
	return v
}

func (m *MessageCtx) Channel() string {
	return m.m.ChannelID
}

func (m *MessageCtx) Guild() string {
	return m.m.GuildID
}

func (m *MessageCtx) Message() *discordgo.Message {
	return m.m
}

// InteractionCtx represents context for an interaction
type InteractionCtx struct {
	dg           *discordgo.Session
	i            *discordgo.Interaction
	s            *Sevcord
	acknowledged bool
	component    bool // If component, then update
	modal        bool
}

func (i *InteractionCtx) Dg() *discordgo.Session {
	return i.dg
}

func (i *InteractionCtx) Acknowledge() error {
	if i.acknowledged {
		return nil
	}

	i.acknowledged = true
	if i.component { // if component, then make it so that response will be ephemeral instead of update
		return nil
	}
	return i.dg.InteractionRespond(i.i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
}

func (i *InteractionCtx) Respond(msg MessageSend) error {
	b := msg.Dg()
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
		typ := discordgo.InteractionResponseUpdateMessage
		if i.modal {
			typ = discordgo.InteractionResponseChannelMessageWithSource
		}
		return i.dg.InteractionRespond(i.i, &discordgo.InteractionResponse{
			Type: typ,
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
