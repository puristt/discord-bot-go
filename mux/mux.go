// Package mux provides a simple Discord message route multiplexer that
// parses messages and then executes a matching registered handler, if found.
// mux can be used with both Disgord and the DiscordGo library.
package mux

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/puristt/discord-bot-go/otobot"
)

// Route holds information about a specific message route handler
type Route struct {
	Pattern     string      // match pattern that should trigger this route handler
	Description string      // short description of this route
	Help        string      // detailed help string for this route
	Handler     HandlerFunc // route handler function to call
}

// HandlerFunc is the function signature required for a message route handler.
type HandlerFunc func(*discordgo.Session, *discordgo.Message)

// Mux is the main struct for all mux methods.
type Mux struct {
	Routes  []*Route
	Default *Route
	Prefix  string
}

// New returns a new Discord message route mux
func New() *Mux {
	m := &Mux{}
	m.Prefix = "-dg "
	return m
}

// Route allows you to register a route
func (m *Mux) Route(pattern, desc string, cb HandlerFunc) (*Route, error) {

	r := Route{}
	r.Pattern = pattern
	r.Description = desc
	r.Handler = cb
	m.Routes = append(m.Routes, &r) //to store all added routes.

	return &r, nil
}

// OnMessageCreate is a DiscordGo Event Handler function. This must be
// registered using the DiscordGo.Session.AddHandler function.  This function
// will receive all Discord messages and parse them for matches to registered
// routes.
func (m *Mux) OnMessageCreate(ds *discordgo.Session, dm *discordgo.MessageCreate) {
	// This isn't required in this specific example but it's a good practice.
	if dm.Author.ID == ds.State.User.ID {
		return
	}

	if strings.Contains("blitz", dm.Content) {
		ds.ChannelMessageSend(dm.ChannelID, "Gaza geldim, hizmete hazirim!")
	}

	if strings.HasPrefix(dm.Content, "-play") {
		query := strings.Trim(dm.Content, "-play ")
		if strings.Compare(query, "") == 0 {
			return
		}
		otobot.PlaySong(query, dm)
	}

	//stop command stops playing song
	if strings.Compare(dm.Content, "-stop") == 0 {
		otobot.StopSong(dm)
	}

	//skip command skips playing song
	if strings.Compare(dm.Content, "-skip") == 0 {
		otobot.SkipSong(dm)
	}

	if strings.Compare(dm.Content, "-showq") == 0 {
		otobot.ShowPlayQueue(dm)
	}
}
