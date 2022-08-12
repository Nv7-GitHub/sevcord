package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/Nv7-Github/sevcord"
	"github.com/Nv7-Github/sevcord/sevutil"
)

var token = flag.String("token", "", "Discord token")

func init() { flag.Parse() }

func main() {
	c, err := sevcord.NewClient(*token)
	if err != nil {
		panic(err)
	}
	c.HandleSlashCommand(&sevcord.SlashCommandGroup{
		Name:        "test",
		Description: "Test command top level",
		Children: []sevcord.SlashCommandObject{
			&sevcord.SlashCommandGroup{
				Name:        "subtest",
				Description: "Test command sub command group",
				Children: []sevcord.SlashCommandObject{
					&sevcord.SlashCommand{
						Name:        "subsubtest",
						Description: "Test command sub command",
						Options: []sevcord.Option{
							{
								Name:        "option",
								Description: "Test command option",
								Kind:        sevcord.OptionKindString,
								Required:    true,
							},
						},
						Handler: func(ctx sevcord.Ctx, args []any) {
							ctx.Acknowledge()
							ctx.Respond(sevcord.MessageResponse("Hello! You said " + args[0].(string)))
						},
					},
					&sevcord.SlashCommand{
						Name:        "ephemeral",
						Description: "Test command sub command with ephemeral response",
						Options: []sevcord.Option{
							{
								Name:        "option",
								Description: "Test command option",
								Kind:        sevcord.OptionKindString,
								Required:    true,
							},
						},
						Handler: func(ctx sevcord.Ctx, args []any) {
							ctx.Respond(sevcord.MessageResponse("Hello! You said " + args[0].(string)))
						},
					},
				},
			},
			&sevcord.SlashCommand{
				Name:        "autocomplete",
				Description: "Test autocomplete",
				Options: []sevcord.Option{
					{
						Name:        "option",
						Description: "Test command option",
						Kind:        sevcord.OptionKindString,
						Required:    true,
						Autocomplete: func(c sevcord.Ctx, val any) []sevcord.Choice {
							return []sevcord.Choice{{Name: "Hey", Value: "Hey"}, {Name: "Hello", Value: "Hello"}, {Name: "Bye", Value: "Bye"}} // Up to 25
						},
					},
				},
				Handler: func(ctx sevcord.Ctx, args []any) {
					ctx.Respond(sevcord.MessageResponse("Hello! You said " + args[0].(string)))
				},
			},
			&sevcord.SlashCommand{
				Name:        "components",
				Description: "Test components",
				Options:     []sevcord.Option{},
				Handler: func(ctx sevcord.Ctx, args []any) {
					ctx.Acknowledge()

					r := sevcord.MessageResponse("Components").ComponentRow(
						&sevcord.Button{Label: "Button", Style: sevcord.ButtonStylePrimary, Handler: func(ctx sevcord.Ctx) { ctx.Edit(sevcord.MessageResponse("Button pressed")) }},
						&sevcord.Button{Label: "Secondary Button", Style: sevcord.ButtonStyleSecondary, Handler: func(ctx sevcord.Ctx) { ctx.Edit(sevcord.MessageResponse("Secondary button pressed")) }},
						&sevcord.Button{Label: "Danger Button", Style: sevcord.ButtonStyleDanger, Handler: func(ctx sevcord.Ctx) { ctx.Edit(sevcord.MessageResponse("Danger button pressed")) }},
						&sevcord.Button{Label: "Success Button", Style: sevcord.ButtonStyleSuccess, Handler: func(ctx sevcord.Ctx) { ctx.Edit(sevcord.MessageResponse("Success button pressed")) }},
						&sevcord.Button{Label: "Disabled Button", Style: sevcord.ButtonStylePrimary, Disabled: true},
					).ComponentRow(
						&sevcord.Button{Label: "Emoji Button", Style: sevcord.ButtonStylePrimary, Emoji: sevcord.ComponentEmojiDefault('😳'), Handler: func(ctx sevcord.Ctx) { ctx.Edit(sevcord.MessageResponse("Emoji button pressed")) }},
						&sevcord.Button{Label: "Link Button", Style: sevcord.ButtonStyleLink, URL: "https://github.com/Nv7-Github/sevcord"},
						&sevcord.Button{Label: "Modal Button", Style: sevcord.ButtonStylePrimary, Handler: func(ctx sevcord.Ctx) {
							ctx.(*sevcord.InteractionCtx).Modal(&sevcord.Modal{
								Title: "Modal",
								Inputs: []sevcord.ModalInput{
									{
										Label:       "Short Input",
										Placeholder: "This is a short input",
										Style:       sevcord.ModalInputStyleSentence,
									},
									{
										Label:       "Long Input",
										Placeholder: "This is a long input",
										Style:       sevcord.ModalInputStyleParagraph,
									},
								},
								Handler: func(ctx sevcord.Ctx, vals []string) {
									ctx.Edit(sevcord.MessageResponse(fmt.Sprintf("You entered `%v`", vals)))
								},
							})
						}},
					).ComponentRow(
						&sevcord.Select{
							Placeholder: "What is your favorite color?",
							Options: []sevcord.SelectOption{
								{Label: "Red", Description: "Red", ID: "red", Emoji: sevcord.ComponentEmojiDefault('🔴')},
								{Label: "Green", Description: "Green", ID: "green", Emoji: sevcord.ComponentEmojiDefault('🟢')},
								{Label: "Blue", Description: "Blue", ID: "blue", Emoji: sevcord.ComponentEmojiDefault('🔵')},
								{Label: "Yellow", Description: "Yellow", ID: "yellow", Emoji: sevcord.ComponentEmojiDefault('🟡')},
								{Label: "Orange", Description: "Orange", ID: "orange", Emoji: sevcord.ComponentEmojiDefault('🟠')},
								{Label: "Purple", Description: "Purple", ID: "purple", Emoji: sevcord.ComponentEmojiDefault('🟣')},
							},
							Handler: func(ctx sevcord.Ctx, vals []string) {
								ctx.Edit(sevcord.MessageResponse(fmt.Sprintf("You selected `%v`", vals))) // NOTE: Use ctx.Respond to make the values selected not change
							},
							MaxValues: 6,
						},
					)
					ctx.Respond(r)
				},
			},
			&sevcord.SlashCommand{
				Name:        "pageswitcher",
				Description: "Test page switcher",
				Options:     []sevcord.Option{},
				Handler: func(ctx sevcord.Ctx, args []any) {
					ctx.Acknowledge()
					items := []any{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, "a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}
					sevutil.NewPageSwitcher(ctx, &sevutil.PageSwitcher{
						Title:   "Test Page Switcher",
						Content: sevutil.PSGetterFromItems(items, 10),
					})
				},
			},
		},
	})
	c.HandleContextMenuCommand(&sevcord.ContextMenuCommand{
		Kind: sevcord.ContextMenuKindMessage,
		Name: "Message Command",
		Handler: func(ctx sevcord.Ctx, id string) {
			ctx.Respond(sevcord.MessageResponse(fmt.Sprintf("This message's ID is `%s`!", id)))
		},
	})
	c.HandleContextMenuCommand(&sevcord.ContextMenuCommand{
		Kind: sevcord.ContextMenuKindUser,
		Name: "User Command",
		Handler: func(ctx sevcord.Ctx, id string) {
			ctx.Respond(sevcord.MessageResponse(fmt.Sprintf("That user is <@%s>!", id)))
		},
	})

	fmt.Println("Starting...")
	err = c.Start()
	if err != nil {
		panic(err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	fmt.Println("Press Ctrl+C to exit")
	<-stop

	fmt.Println("Gracefully shutting down...")
	err = c.Close()
	if err != nil {
		panic(err)
	}
}
