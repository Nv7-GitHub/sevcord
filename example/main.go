package main

import (
	_ "embed"
	"fmt"

	"github.com/Nv7-Github/sevcord"
)

//go:embed token.txt
var token string

func main() {
	bot, err := sevcord.New(token)
	if err != nil {
		panic(err)
	}
	bot.RegisterSlashCommand(sevcord.NewSlashCommand("ping", "Is the bot ok?", func(ctx sevcord.Ctx, params []any) {
		ctx.Acknowledge()
		msg := ""
		if params[0] != nil {
			msg = " " + params[0].(string)
		}
		ctx.Respond(sevcord.NewMessage("Pong!" + msg).
			AddComponentRow(sevcord.NewButton("Click me!", sevcord.ButtonStylePrimary, "click", ctx.Author().ID)))
	}, sevcord.NewOption("echo", "Echoed in the response", sevcord.OptionKindString, false).AutoComplete(func(ctx sevcord.Ctx, params any) []sevcord.Choice {
		return []sevcord.Choice{sevcord.NewChoice("Hello", "Hello"), sevcord.NewChoice("World", "World")}
	})))
	bot.AddButtonHandler("click", func(ctx sevcord.Ctx, params string) {
		// Uses params to see whether author is pressing
		if ctx.Author().ID == params {
			ctx.Respond(sevcord.NewMessage("The author of this message clicked me!").
				AddComponentRow(sevcord.NewButton("Click me!", sevcord.ButtonStylePrimary, "click", params)))
		} else {
			ctx.Respond(sevcord.NewMessage(fmt.Sprintf("<@%s> clicked me!", ctx.Author().ID)).
				AddComponentRow(sevcord.NewButton("Click me!", sevcord.ButtonStylePrimary, "click", params)))
		}
	})
	bot.Listen()
}
