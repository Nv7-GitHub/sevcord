package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/Nv7-Github/sevcord"
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
						Handler: func(args []any, ctx sevcord.Ctx) {
							ctx.Acknowledge()
							ctx.Edit(sevcord.MessageResponse("Hello! You said " + args[0].(string)))
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
						Handler: func(args []any, ctx sevcord.Ctx) {
							ctx.Edit(sevcord.MessageResponse("Hello! You said " + args[0].(string)))
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
						Autocomplete: func(val any) []sevcord.Choice {
							return []sevcord.Choice{{Name: "Hey", Value: "Hey"}, {Name: "Hello", Value: "Hello"}, {Name: "Bye", Value: "Bye"}} // Up to 25
						},
					},
				},
				Handler: func(args []any, ctx sevcord.Ctx) {
					ctx.Edit(sevcord.MessageResponse("Hello! You said " + args[0].(string)))
				},
			},
			&sevcord.SlashCommand{
				Name:        "components",
				Description: "Test components",
				Options:     []sevcord.Option{},
				Handler: func(args []any, ctx sevcord.Ctx) {
					ctx.Acknowledge()

					r := sevcord.MessageResponse("Components").ComponentRow(
						&sevcord.Button{Label: "Button", Style: sevcord.ButtonStylePrimary, Handler: func(ctx sevcord.Ctx) { ctx.Edit(sevcord.MessageResponse("Button pressed")) }},
						&sevcord.Button{Label: "Secondary Button", Style: sevcord.ButtonStyleSecondary, Handler: func(ctx sevcord.Ctx) { ctx.Edit(sevcord.MessageResponse("Secondary button pressed")) }},
						&sevcord.Button{Label: "Danger Button", Style: sevcord.ButtonStyleDanger, Handler: func(ctx sevcord.Ctx) { ctx.Edit(sevcord.MessageResponse("Dange button pressed")) }},
						&sevcord.Button{Label: "Success Button", Style: sevcord.ButtonStyleSuccess, Handler: func(ctx sevcord.Ctx) { ctx.Edit(sevcord.MessageResponse("Success button pressed")) }},
						&sevcord.Button{Label: "Disabled Button", Style: sevcord.ButtonStylePrimary, Disabled: true},
					).ComponentRow(
						&sevcord.Button{Label: "Emoji Button", Style: sevcord.ButtonStyleSuccess, Emoji: sevcord.ComponentEmojiDefault('😳'), Handler: func(ctx sevcord.Ctx) { ctx.Edit(sevcord.MessageResponse("Emoji button pressed")) }},
						&sevcord.Button{Label: "Link Button", Style: sevcord.ButtonStyleLink, URL: "https://github.com/Nv7-Github/sevcord"},
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
