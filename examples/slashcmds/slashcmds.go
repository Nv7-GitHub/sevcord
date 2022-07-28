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
						Handler: func(args []interface{}, ctx sevcord.Ctx) {
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
						Handler: func(args []interface{}, ctx sevcord.Ctx) {
							ctx.Respond(sevcord.MessageResponse("Hello! You said " + args[0].(string)))
						},
					},
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

	err = c.Close()
	if err != nil {
		panic(err)
	}
}
