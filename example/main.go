package main

import (
	_ "embed"

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
		ctx.Respond(sevcord.NewMessage("Pong!" + msg))
	}, sevcord.NewOption("echo", "Echoed in the response", sevcord.OptionKindString, false).AutoComplete(func(ctx sevcord.Ctx, params any) []sevcord.Choice {
		return []sevcord.Choice{sevcord.NewChoice("Hello", "Hello"), sevcord.NewChoice("World", "World")}
	})))
	bot.Listen()
}
