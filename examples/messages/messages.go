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
	c, err := sevcord.NewClient(*token, &sevcord.ClientParams{Messages: true})
	if err != nil {
		panic(err)
	}

	// Handlers

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
