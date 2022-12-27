package main

import (
	_ "embed"
	"fmt"
	"math/rand"
	"time"

	"github.com/Nv7-Github/sevcord/v2"
	"github.com/bwmarrin/discordgo"
)

//go:embed token.txt
var token string

func main() {
	bot, err := sevcord.New(token)
	if err != nil {
		panic(err)
	}
	// 1 in 10 chance of not being able to use bot
	rand.Seed(time.Now().UnixNano())
	bot.AddMiddleware(func(ctx sevcord.Ctx, cmd string) bool {
		v := rand.Intn(9)
		if v == 0 {
			ctx.Respond(sevcord.NewMessage("Unlucky"))
			return false
		}
		return true
	})
	// Ping + button example
	bot.RegisterSlashCommand(sevcord.NewSlashCommand("ping", "Is the bot ok? + Button demo", func(ctx sevcord.Ctx, params []any) {
		ctx.Acknowledge()
		msg := ""
		if params[0] != nil {
			msg = " " + params[0].(string)
		}
		ctx.Respond(sevcord.NewMessage("Pong!" + msg).
			AddComponentRow(sevcord.NewButton("Click me!", sevcord.ButtonStylePrimary, "click", ctx.Author().User.ID)))
	}, sevcord.NewOption("echo", "Echoed in the response", sevcord.OptionKindString, false).AutoComplete(func(ctx sevcord.Ctx, params any) []sevcord.Choice {
		return []sevcord.Choice{sevcord.NewChoice("Hello", "Hello"), sevcord.NewChoice("World", "World")}
	})))
	// Select menu example
	bot.RegisterSlashCommand(sevcord.NewSlashCommand("select", "Select menu demo", func(ctx sevcord.Ctx, params []any) {
		ctx.Acknowledge()
		ctx.Respond(sevcord.NewMessage("Check out these select menus").
			AddComponentRow(
				sevcord.NewSelect("Select menu", "select", "String").
					Option(sevcord.NewSelectOption("Option 1", "First option", "1")).
					Option(sevcord.NewSelectOption("Option 2", "Second option", "2")).
					Option(sevcord.NewSelectOption("Default", "This is already selected", "default").
							SetDefault(true).
							WithEmoji(sevcord.ComponentEmojiDefault('😎'))).
					SetRange(0, 3), // Allows users to select unlimited instead of default of 1
			),
		)
	}))
	bot.RegisterSlashCommand(sevcord.NewSlashCommand("auto-select", "Auto-populated select menu demo", func(ctx sevcord.Ctx, a []any) {
		ctx.Acknowledge()
		ctx.Respond(sevcord.NewMessage("Check out these auto-populated select menus").
			AddComponentRow(
				sevcord.NewSelect("User menu", "user_select", "User").SetKind(sevcord.SelectKindUser).SetRange(0, 25),
			).
			AddComponentRow(
				sevcord.NewSelect("Role menu", "role_select", "Role").
					SetKind(sevcord.SelectKindRole).
					SetRange(0, 25),
			).
			AddComponentRow(
				sevcord.NewSelect("Mentionable menu", "userrole_select", "Mentionable").
					SetKind(sevcord.SelectKindMentionable).
					SetRange(0, 25),
			).
			AddComponentRow(
				sevcord.NewSelect("Channel menu", "channel_select", "Channel").
					SetKind(sevcord.SelectKindChannel).
					ChannelMenuFilter(discordgo.ChannelTypeGuildText).
					SetRange(0, 25),
			).
			AddComponentRow(
				sevcord.NewSelect("Voice channel menu", "voice_select", "Voice").
					SetKind(sevcord.SelectKindChannel).
					ChannelMenuFilter(discordgo.ChannelTypeGuildVoice).
					SetRange(0, 25),
			))
	}))
	// Modal example
	bot.RegisterSlashCommand(sevcord.NewSlashCommand("modal", "Modal demo", func(ctxV sevcord.Ctx, params []any) {
		ctx := ctxV.(*sevcord.InteractionCtx)
		ctx.Modal(sevcord.NewModal("Modal", func(ctx sevcord.Ctx, values []string) {
			ctx.Acknowledge()
			ctx.Respond(sevcord.NewMessage(fmt.Sprintf("You entered: `%v`", values[0])))
		}).
			Input(sevcord.NewModalInput("text", "Text input", sevcord.ModalInputStyleSentence, 240)).
			Input(sevcord.NewModalInput("paragraph", "Paragraph input", sevcord.ModalInputStyleParagraph, 2400)),
		)
	}))
	// Button example handler
	bot.AddButtonHandler("click", func(ctx sevcord.Ctx, params string) {
		// Uses params to see whether author is pressing
		if ctx.Author().User.ID == params {
			ctx.Respond(sevcord.NewMessage("The author of this message clicked me!").
				AddComponentRow(sevcord.NewButton("Click me!", sevcord.ButtonStylePrimary, "click", params)))
		} else {
			ctx.Respond(sevcord.NewMessage(fmt.Sprintf("<@%s> clicked me!", ctx.Author().User.ID)).
				AddComponentRow(sevcord.NewButton("Click me!", sevcord.ButtonStylePrimary, "click", params)))
		}
	})
	// Select menu example handler
	selectHandler := func(ctx sevcord.Ctx, params string, options []string) {
		ctx.Acknowledge() // That way it makes a new ephemeral message instead of updating the original
		ctx.Respond(sevcord.NewMessage(fmt.Sprintf("You Selected: `%v` (Select Type: **%s**)", options, params)))
	}
	bot.AddSelectHandler("select", selectHandler)
	bot.AddSelectHandler("user_select", selectHandler)
	bot.AddSelectHandler("role_select", selectHandler)
	bot.AddSelectHandler("userrole_select", selectHandler)
	bot.AddSelectHandler("channel_select", selectHandler)
	bot.AddSelectHandler("voice_select", selectHandler)
	// Message handler
	bot.SetMessageHandler(func(ctx sevcord.Ctx, content string) {
		if content == "ping" {
			ctx.Acknowledge()
			ctx.Respond(sevcord.NewMessage("Pong!"))
		}
	})
	bot.Listen()
}
