package sevcord

import (
	"strconv"

	"github.com/bwmarrin/discordgo"
)

type Ctx interface {
	Acknowledge()
	Respond(*Response)
	Edit(*Response)

	Guild() string
	Channel() string
	User() *User

	session() *discordgo.Session
}

type InteractionCtx struct {
	c          *Client
	i          *discordgo.Interaction
	s          *discordgo.Session
	followup   bool // ID of FollowupMessageCreate
	followupid string
	component  bool
}

func (i *InteractionCtx) Acknowledge() {
	err := i.s.InteractionRespond(i.i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		Logger.Println(err)
	}
	i.followup = true
	i.component = false
}

func (i *InteractionCtx) Respond(r *Response) {
	i.send(r, false)
}

func (i *InteractionCtx) Edit(r *Response) {
	i.send(r, true)
}

func (i *InteractionCtx) send(r *Response, edit bool) {
	var embs []*discordgo.MessageEmbed
	if r.embed != nil {
		embs = []*discordgo.MessageEmbed{r.embed}
	}

	var followup *string = nil
	if i.followup {
		followup = &i.followupid
	}
	comps := r.registerComponents(i.c, i.i.ID, followup)

	if i.followup && !(edit && i.component) {
		if edit && !i.component {
			_, err := i.s.FollowupMessageEdit(i.i, i.followupid, &discordgo.WebhookEdit{
				Content:    r.content,
				Embeds:     embs,
				Components: comps,
				Files:      r.files,
			})
			if err != nil {
				Logger.Println(err)
				return
			}
			return
		}

		msg, err := i.s.FollowupMessageCreate(i.i, true, &discordgo.WebhookParams{
			Content:    r.content,
			Embeds:     embs,
			Components: comps,
			Files:      r.files,
		})
		if err != nil {
			Logger.Println(err)
			return
		}
		i.followupid = msg.ID
		return
	}

	typ := discordgo.InteractionResponseChannelMessageWithSource
	if edit {
		typ = discordgo.InteractionResponseUpdateMessage
		i.component = false
	}
	err := i.s.InteractionRespond(i.i, &discordgo.InteractionResponse{
		Type: typ,
		Data: &discordgo.InteractionResponseData{
			Content:    r.content,
			Components: comps,
			Embeds:     embs,
			Files:      r.files,
			Flags:      1 << 6, // All non-acknowledged responses are ephemeral
		},
	})
	if err != nil {
		Logger.Println(err)
		return
	}
}

func (i *InteractionCtx) Guild() string {
	return i.i.GuildID
}

func (i *InteractionCtx) session() *discordgo.Session {
	return i.s
}

func (i *InteractionCtx) User() *User {
	if i.i.Member == nil {
		return userFromUser(i.i.User)
	}
	return userFromMember(i.i.Member)
}

func (i *InteractionCtx) Channel() string {
	return i.i.ChannelID
}

func (i *InteractionCtx) Modal(m *Modal) {
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

	v := modalHandler{handler: m.Handler}
	if i.followup {
		v.followup = &i.followupid
	}

	i.c.lock.Lock()
	i.c.modalHandlers[i.i.ID] = v
	i.c.lock.Unlock()

	err := i.s.InteractionRespond(i.i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			Title:      m.Title,
			Components: comps,
			CustomID:   i.i.ID,
		},
	})
	if err != nil {
		Logger.Println(err)
	}
}

type MessageCtx struct {
	c      *Client
	s      *discordgo.Session
	m      *discordgo.MessageCreate
	typing bool
	respid string
}

func (m *MessageCtx) Acknowledge() {
	m.s.ChannelTyping(m.m.ChannelID)
	m.typing = true
}

func (m *MessageCtx) Respond(r *Response) {
	if m.typing {
		m.s.ChannelTyping(m.m.ChannelID)
		m.typing = false
	}

	var embs []*discordgo.MessageEmbed
	if r.embed != nil {
		embs = []*discordgo.MessageEmbed{r.embed}
	}
	res, err := m.s.ChannelMessageSendComplex(m.m.ChannelID, &discordgo.MessageSend{
		Content:    r.content,
		Embeds:     embs,
		Components: r.registerComponents(m.c, m.m.ID, nil),
		Files:      r.files,
	})
	if err != nil {
		Logger.Println(err)
		return
	}

	m.respid = res.ID
}

func (m *MessageCtx) Edit(r *Response) {
	if m.typing {
		m.s.ChannelTyping(m.m.ChannelID)
		m.typing = false
	}

	var embs []*discordgo.MessageEmbed
	if r.embed != nil {
		embs = []*discordgo.MessageEmbed{r.embed}
	}
	var cont *string = nil
	if r.content != "" {
		cont = &r.content
	}
	_, err := m.s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Content:    cont,
		Embeds:     embs,
		Components: r.registerComponents(m.c, m.m.ID, nil),
	})
	if err != nil {
		Logger.Println(err)
		return
	}
}

func (m *MessageCtx) Guild() string   { return m.m.GuildID }
func (m *MessageCtx) Channel() string { return m.m.ChannelID }
func (m *MessageCtx) User() *User {
	if m.m.Member != nil {
		return userFromMember(m.m.Member)
	}
	return userFromUser(m.m.Author)
}
