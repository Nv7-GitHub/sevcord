package sevcord

import "github.com/bwmarrin/discordgo"

type Sevcord struct {
	dg *discordgo.Session // Note: only use this to create cmds, give one provided with handlers for user
}
