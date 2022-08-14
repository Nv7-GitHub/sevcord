package sevcord

import (
	"github.com/bwmarrin/discordgo"
)

type MessageHandler func(Ctx, *Message)

func (c *Client) messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	msg := messageFromDg(m.Message)
	ctx := &MessageCtx{
		c: c,
		s: s,
		m: m,
	}
	(*c.mHandler)(ctx, msg)
}
