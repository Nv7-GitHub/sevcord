package sevcord

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
)

var Logger = log.Default()

// MiddlewareFunc accepts context and returns whether or not to continue
type MiddlewareFunc func(ctx Ctx) (ok bool)

type Sevcord struct {
	lock *sync.RWMutex

	dg             *discordgo.Session // Note: only use this to create cmds, give one provided with handlers for user
	middleware     []MiddlewareFunc
	commands       map[string]SlashCommandObject
	buttonHandlers map[string]ButtonHandler
	selectHandlers map[string]SelectHandler
	modalHandlers  map[string]ModalHandler
}

func (s *Sevcord) RegisterSlashCommand(cmd SlashCommandObject) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.commands[cmd.name()] = cmd
}

func New(token string) (*Sevcord, error) {
	dg, err := discordgo.New("Bot " + strings.TrimSpace(token))
	if err != nil {
		return nil, err
	}
	return &Sevcord{
		lock:           &sync.RWMutex{},
		dg:             dg,
		middleware:     make([]MiddlewareFunc, 0),
		commands:       make(map[string]SlashCommandObject),
		buttonHandlers: make(map[string]ButtonHandler),
		selectHandlers: make(map[string]SelectHandler),
		modalHandlers:  make(map[string]ModalHandler),
	}, nil
}

// AddMiddleware adds middleware, a function that is run before every command handler is called. Middleware is run in the order it is added
func (s *Sevcord) AddMiddleware(m MiddlewareFunc) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.middleware = append(s.middleware, m)
}

func (s *Sevcord) AddButtonHandler(id string, handler ButtonHandler) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.buttonHandlers[id] = handler
}

func (s *Sevcord) AddSelectHandler(id string, handler SelectHandler) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.selectHandlers[id] = handler
}

// Dg gets the global discordgo session. NOTE: Only use this to add handlers/intents, use the one provided with Ctx for anything else
func (s *Sevcord) Dg() *discordgo.Session {
	return s.dg
}

func (s *Sevcord) Listen() {
	s.lock.RLock()
	// Build commands
	cmds := make([]*discordgo.ApplicationCommand, 0, len(s.commands))
	for _, cmd := range s.commands {
		v := cmd.dg()
		dmPermission := false
		cmds = append(cmds, &discordgo.ApplicationCommand{
			Name:        v.Name,
			Description: v.Description,
			Options:     v.Options,
			Type:        discordgo.ChatApplicationCommand,

			DefaultMemberPermissions: cmd.permissions(),
			DMPermission:             &dmPermission,
		})
	}
	s.lock.RUnlock()

	// Handlers
	s.dg.AddHandler(s.interactionHandler)
	s.dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		_, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, "", cmds)
		if err != nil {
			Logger.Println("Error updating commands", err)
		}
		Logger.Println("Bot Ready")
	})

	// Open
	s.dg.Identify.Intents = discordgo.IntentsNone
	s.dg.Open()

	// Wait
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	fmt.Println("Gracefully shutting down...")

	// Close
	s.dg.Close()
}
