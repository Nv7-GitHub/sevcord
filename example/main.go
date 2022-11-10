package main

import (
	_ "embed"
	"fmt"

	"github.com/Nv7-Github/sevcord/v2"
)

//go:embed token.txt
var token string

func main() {
	bot, err := sevcord.New(token)
	if err != nil {
		panic(err)
	}
	bot.RegisterSlashCommand(sevcord.NewSlashCommand("ping", "Is the bot ok? + Button demo", func(ctx sevcord.Ctx, params []any) {
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
	bot.RegisterSlashCommand(sevcord.NewSlashCommand("select", "Select menu demo", func(ctx sevcord.Ctx, params []any) {
		ctx.Acknowledge()
		ctx.Respond(sevcord.NewMessage("Check out these select menus").
			AddComponentRow(
				sevcord.NewSelect("Select menu", "select", "").
					Option(sevcord.NewSelectOption("Option 1", "First option", "1")).
					Option(sevcord.NewSelectOption("Option 2", "Second option", "2")).
					Option(sevcord.NewSelectOption("Default", "This is already selected", "default").
							SetDefault(true).
							WithEmoji(*sevcord.ComponentEmojiDefault('😎'))).
					SetRange(0, 3), // Allows users to select unlimited instead of default of 1
			),
		)
	}))
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
	bot.AddSelectHandler("select", func(ctx sevcord.Ctx, params string, options []string) {
		ctx.Acknowledge() // That way it makes a new ephemeral message instead of updating the original
		err := ctx.Respond(sevcord.NewMessage(fmt.Sprintf("You Selected: `%v`", options)))
		fmt.Println(err)
	})
	bot.Listen()
}
